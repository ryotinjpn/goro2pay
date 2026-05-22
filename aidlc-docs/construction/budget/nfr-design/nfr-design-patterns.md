# Budget Unit — NFR Design Patterns

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: NFR Design / Construction
**Predecessors**: NFR Requirements 完了 (PR #74)

本ドキュメントは Unit B の **設計パターン集** を定義する。NFR Requirements で確定した数値・しきい値を実現するための論理レベルのパターンを記述し、実装コードは Code Generation で書く。

参照: [budget-nfr-design-plan.md](../../plans/budget-nfr-design-plan.md), [nfr-requirements.md](../nfr-requirements/nfr-requirements.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md), [Functional Design](../functional-design/)

---

## 1. Reliability Patterns（信頼性）

### P-REL-01: `DeductConditional` — ConditionalCheckFailedException の即マップ (Q-D1=A)

**目的**: `UpdateItem` の `ConditionExpression: balance >= :amount` 失敗（残高不足）を即座に `ErrInsufficientBalance` にマップし、クライアントに 402 を返す。

**パターン**: 即マップ（リトライなし）

**設計根拠**: 残高不足は再試行しても変わらない。競合書き込みによる偶発的な失敗はハッカソン規模では考慮不要（`ConditionalCheckFailedException` はスロットリングではなく条件不成立）。デモ規模の同時ユーザ数では競合が連続する確率は無視できる。

**設計責務**:
- `WalletRepository.Deduct()` は Go SDK から `ConditionalCheckFailedException` を受け取ったら、**リトライなしで即座に** `wallet.ErrInsufficientBalance` を返す
- Handler 層は `ErrInsufficientBalance` を HTTP 402 にマップ（[unit-interfaces.md §8](../../interfaces/unit-interfaces.md)）
- `ConditionalCheckFailedException` 以外のエラーは 500 として伝播

**論理シグネチャ**:
```go
// WalletRepository 内
if isConditionalCheckFailed(err) {
    return wallet.ErrInsufficientBalance
}
```

---

### P-REL-02: `SetBudget` 部分失敗 — エラーをそのまま返す (Q-D2=A)

**目的**: `SetBudget` の 2 操作（BudgetSettings Upsert + Wallet 作成/差分調整）の一方が失敗した場合、エラーをそのまま返し、クライアント側の再試行に委ねる。

**パターン**: エラーそのまま返す（補償トランザクションなし）

**設計根拠**: `SetBudget` は Functional Design §2.2 で PBT P-2「べき等性」が保証されている。再試行で一貫した状態に収束するため、補償トランザクションの実装コストに見合わない。DLQ・非同期リカバリは MVP/ハッカソン範囲外。

**設計責務**:
- `WalletService.SetBudget()` は 1 操作目成功・2 操作目失敗の場合、2 操作目のエラーをそのまま返す
- クライアント（Frontend）は失敗時に再試行可能（`SetBudget` のべき等性を信頼）
- ロールバック・DLQ・非同期リカバリは実装しない

---

### P-REL-03: `BudgetResetLog` 冪等性 — PutItem with ConditionExpression (NFR-REL-03)

**目的**: 月次リセット処理の多重実行を防止する。

**パターン**: `PutItem` with `ConditionExpression: attribute_not_exists(resetDate) AND attribute_not_exists(userId)`

**設計責務**:
- `BudgetResetLogRepository.Save()` は当該 ConditionExpression を付けて PutItem を実行
- 既存レコードが存在する場合（多重実行検知）、`ConditionalCheckFailedException` をそのまま伝播
- `WalletService.ResetAll()` は BudgetResetLog 書き込み失敗を ERROR ログ出力後に継続する

---

## 2. Performance Patterns（性能）

### P-PERF-01: Deduct p95 ≤ 500ms 達成のための制約 (NFR-PERF-01 / Issue #25)

**目的**: Deduct API の p95 応答時間を 500ms 以内に収める。

**設計上の制約**:
- 冪等性チェック（`IdempotencyRepository.TryAcquire`）と残高更新（`WalletRepository.Deduct`）は **逐次実行**（1 往復ずつ）
- DynamoDB は **ConsistentRead: true** を使用（Wallet テーブル、NFR-REL-01 整合）
- Lambda コールドスタート対策は本 MVP では実装しない（Provisioned Concurrency は使わない）
- DynamoDB Provisioned 1 WCU で対応（ハッカソン規模、バースト容量内）
- P-REL-01 の通り Deduct ではリトライなし → 追加レイテンシが生じない

---

## 3. Observability Patterns（観測性）

### P-OBS-01: 構造化ログ — `log/slog` + `ContextAwareSlogHandler` 再利用 (Q-D5=A)

**目的**: Unit B の 9 項目構造化ログ（Unit A の 8 項目 + `amount` + `newBalance`）を、ハンドラ呼び出し側で項目を意識せずに自動付与する。

**パターン**: Unit A の `ContextAwareSlogHandler`（LC-AUTH-05）を **そのまま再利用**。`amount` と `newBalance` は handler 内で明示的な attrs として渡す。

**ライブラリ**: `log/slog`（Go 1.21 標準）— **Unit A と統一**（外部依存ゼロ、JSON ハンドラ組み込み）

**設計根拠**: `zerolog` や `zap` はサードパーティ依存が追加される。ハッカソン規模では `slog` のパフォーマンスで十分。Unit A が先行採用済みのため統一することでライブラリ管理コストを削減。

**設計責務**:
- `ContextAwareSlogHandler`（LC-AUTH-05、`apps/api/internal/logging/handler.go`）を Unit B の Lambda でも **同一インスタンス** として初期化する（アプリ起動時 `slog.SetDefault(...)` で登録）
- `amount` / `newBalance` は **handler 内で明示的 attrs** として渡す（Unit A の 8 項目は context から自動抽出される）:
  ```go
  slog.InfoContext(ctx, "deduct success",
      "action", "deduct",
      "amount", req.Amount,
      "newBalance", result.NewBalance,
  )
  ```
- 9 項目: `level`, `timestamp`, `userId`, `action`, `traceId`, `requestId`, `userAgent`, `amount`（該当 API のみ）, `newBalance`（該当 API のみ）
- 平文 userId/email を絶対に出力しない（Unit A NFR-OBS-01 の制約を Unit B も継承）

### P-OBS-02: `ResetAll` エラーログパターン (Q-D8=A)

**目的**: 月初リセットで一部ユーザが失敗した場合、CloudWatch Logs でユーザ単位のエラーを追跡しつつ集計サマリも記録する。

**パターン**: ユーザごとに ERROR ログ + 全ユーザ処理後に集計サマリ（INFO）

**設計責務**:
- `WalletService.ResetAll()` は全ユーザを処理しながら失敗を `errors []error` に収集（中断なし）
- **各失敗ユーザ処理直後**に ERROR ログを出力:
  ```go
  slog.ErrorContext(ctx, "reset failed for user",
      "action", "monthly_reset",
      "targetUserId", userID,
      "error", err.Error(),
  )
  ```
- **全ユーザ処理後**に INFO ログで集計サマリを出力:
  ```go
  slog.InfoContext(ctx, "monthly reset complete",
      "action", "monthly_reset",
      "processedUsers", total,
      "failedUsers", len(errors),
  )
  ```
- 成功ユーザの個別 INFO ログは出力しない（C 案は不採用、ログ量抑制）

---

## 4. Testing Patterns（テスト）

### P-TEST-01: PBT — `gopter` による残高不変条件テスト (Q-D6=B)

**目的**: 残高不変条件 / `SetBudget` べき等性 / バリデーション境界値を Property-Based Testing で検証する。

**パターン**: 通常の `go test` の一部として実行（Build tag・nightly 分離なし）

**ライブラリ**: `github.com/leanovate/gopter` — **Unit A と統一**（Unit A の P-TEST-01 採用に合わせる、Shrink サポートありで残高境界値テストに適している）

**設計根拠**: Unit A が `gopter` を選定済み。アプリケーション全体でライブラリを統一することでバージョン管理コストを削減。`rapid` への変更は Unit A との不整合を生むため採用しない。

**設計責務**:
- テストファイルを `apps/api/internal/wallet/` 以下に配置
- Generation 数は Unit A 同様デフォルト（100）を使用
- CI では `go test ./...` で自動実行

**3 プロパティ** (NFR-TEST-01 / nfr-requirements.md §8.1):
1. **残高不変条件**: 任意の `amount > 0`, `balance >= amount` で `Deduct` が成功し `newBalance == balance - amount` となること
2. **`SetBudget` べき等性**: 同じ `monthlyBudget` を 2 回 `SetBudget` しても結果が同じになること
3. **バリデーション境界値**: `monthlyBudget` が `[1, 100_000]` の範囲外の値は必ず `ErrBudgetOutOfRange` を返すこと

---

## 5. Degradation Patterns（デグレード演出）

### P-DEG-01: `BalanceDisplay` ローディング状態 — 初回スケルトン + 2 回目以降 placeholderData (Q-D3=B+初回スケルトン)

**目的**: NFR-DEG-03「残高常時可視化」を達成する。初回はスケルトン、2 回目以降は前回値を薄く表示。

**パターン**: 初回ロード → スケルトン UI / 2 回目以降のフェッチ中 → `placeholderData` で前回値を薄く表示

**設計責務**:
- `useWallet()` フックは TanStack Query の `useQuery` を内部で利用
  - `staleTime: 30 * 1000`（30 秒）
  - `placeholderData: keepPreviousData`（2 回目以降のフェッチ中に前回データを保持）
- `BalanceDisplay` コンポーネントの表示分岐:
  - `isLoading === true`（データが一度もない初回）→ **スケルトン UI**（グレーのプレースホルダ）
  - `isFetching === true && data !== undefined`（2 回目以降のフェッチ中）→ **薄い表示**（opacity 0.5 等）
  - `data !== undefined && !isFetching` → **通常表示**
- `refetch()` を `useWallet` フックから公開し、外部（Unit C）からの即時 invalidate に対応

**TanStack Query 設定**:
```ts
const query = useQuery({
  queryKey: ['balance'],
  queryFn: fetchBalance,
  staleTime: 30_000,
  placeholderData: keepPreviousData,
})
```

---

### P-DEG-02: 残高枯渇演出 — `InsufficientBalanceModal` (Q-D4=B)

**目的**: NFR-DEG-04「残高枯渇演出」を達成する。402 受信時にモーダルでダメ化文言 + 増額誘導を表示。

**パターン**: モーダル演出

**設計根拠**: モーダルは「残りダメ予算が足りません」等のダメ化文言を大きく演出できる。NFR-DEG-05 のダメ化トーンを強調する UI として適切。タップ数はインライン（A 案）より増えるが、コンテキスト切替（画面遷移、C 案）より少ない。

**設計責務**:
- Unit C の `useOrder()` が 402 を受信 → Jotai atom `insufficientBalanceAtom` に `true` を書き込み
- `InsufficientBalanceModal` が atom を購読し open
- 表示内容: 「残りダメ予算が足りません…」（ダメ化文言）+ 増額誘導ボタン（`/budget` 画面へ遷移）+ 閉じるボタン
- Modal が閉じると atom を `false` にリセット
- 文言・アニメーション詳細は Code Generation で確定

**論理フロー**:
```
Unit C useOrder() → 402 受信
  → insufficientBalanceAtom.set(true)
  → InsufficientBalanceModal が open
  → 「増額する」ボタン → /budget へ遷移
  → 閉じる → insufficientBalanceAtom.set(false)
```

---

### P-DEG-03: `Deduct` 成功後の TanStack Query invalidate — Unit C コールバック方式 (Q-D7=A)

**目的**: NFR-DEG-03「残高常時可視化」の即時更新を実現する。`Deduct` 成功後（Unit C 注文完了後）に `GetBalance` のキャッシュを即座に無効化する。

**パターン**: Unit C の注文完了 callback で `invalidateQueries(['balance'])` を呼ぶ

**設計根拠**: 注文フローが残高更新の責任を持つことで処理の流れが明確。Jotai atom 購読（B 案）と比較して実装がシンプル。`refetchInterval`（C 案）は 30s 遅延があり即時性に劣る。

**設計責務**:
- Unit C の `useOrder()` フック内で注文成功時:
  ```ts
  const queryClient = useQueryClient()
  // ... 注文処理 ...
  queryClient.invalidateQueries({ queryKey: ['balance'] })
  ```
- Unit B の `useWallet()` フックは `queryKey: ['balance']` を使用（Unit C との **クロスユニット契約**）
- **クロスユニット依存の明記**: `['balance']` query key は Unit B と Unit C の暗黙的な契約であるため [unit-interfaces.md §9.1](../../interfaces/unit-interfaces.md) に凍結する
- TanStack Query の 30s stale + 即 invalidate の組み合わせにより、通常時は 30s、注文直後は即時更新

---

## 6. パターン → NFR トレーサビリティ

| パターン | 対応 NFR / Q-D 質問 |
|---|---|
| P-REL-01 (ConditionalCheck 即マップ) | NFR-REL-01 / Q-D1 |
| P-REL-02 (SetBudget 部分失敗) | FD §2.2 べき等性 / Q-D2 |
| P-REL-03 (BudgetResetLog 冪等性) | NFR-REL-03 / Q-N4 |
| P-PERF-01 (Deduct 500ms 制約) | NFR-PERF-01 / Issue #25 |
| P-OBS-01 (slog + ContextAwareSlogHandler) | NFR-OBS-01 / Q-D5 |
| P-OBS-02 (ResetAll ログ) | NFR-OBS-01 / Q-D8 |
| P-TEST-01 (gopter PBT) | NFR-TEST-01 / Q-D6 |
| P-DEG-01 (BalanceDisplay Loading) | NFR-DEG-03 / Q-D3 |
| P-DEG-02 (InsufficientBalance Modal) | NFR-DEG-04 / Q-D4 |
| P-DEG-03 (invalidate from Unit C) | NFR-DEG-03 / Q-D7 |

---

## 7. Logical Components 連携

各パターンが利用する論理コンポーネントの一覧と相互関係は [logical-components.md](./logical-components.md) を参照。

---

## 8. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | DynamoDB テーブル物理設計（PK/SK/GSI/TTL 詳細）、Lambda IAM Role、EventBridge Scheduler の Terraform 実装 |
| **Code Generation** | `WalletRepository.Deduct()` の ConditionalCheck 判定実装、`SetBudget` 2 操作シーケンス、`ContextAwareSlogHandler` 再利用コード（LC-AUTH-05）、`gopter` PBT テスト実装、`BalanceDisplay` の skeleton/placeholderData 実装、`InsufficientBalanceModal` 文言・アニメーション、`useWallet` TanStack Query 設定 |
