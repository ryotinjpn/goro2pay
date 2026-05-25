# Unit E (`metrics`) — NFR Design Patterns

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Design
**Unit**: E — `metrics`（ダメ化メトリクス）
**Depth**: Standard
**Related**:
- Plan: [metrics-nfr-design-plan.md](../../plans/metrics-nfr-design-plan.md)
- NFR Requirements: [nfr-requirements.md](../nfr-requirements/nfr-requirements.md)（NFRE-E01〜E10）
- Functional Design: [business-logic-model.md](../functional-design/business-logic-model.md)
- 凍結契約: [unit-interfaces.md](../../interfaces/unit-interfaces.md) §6

**方針**: Q-DD1〜Q-DD5 全 A/B 確定。Bedrock なし・新規テーブルなし の軽量 Unit として、Unit C/D パターンを最大限再利用しつつ Unit E 固有の errgroup 並列パターンと DEG UX パターンを追加定義する。

---

## 1. パターン識別子規則

- 形式: `P-ME-{名前}-{番号}`（Pattern / Metrics / E）
- Unit C パターン再利用時は「**再利用: P-xxx**」と明記し、差分のみ記述
- 新規定義: `P-ME-PARALLEL-01`（errgroup 並列読取）/ `P-ME-FE-DEG-01`（DEG UX）

---

## 2. Unit C パターン再利用マップ（Q-DD1=A）

| Unit C パターン | Unit E での再利用 | 差分 |
|---|---|---|
| **P-INIT-01** Lambda Cold Start | package-level DynamoDB client を metrics / budget_raise / repo 各 package の `init()` で初期化 | なし（同一パターン） |
| **P-DI-01** Manual DI | `main.go` で MetricsService / BudgetRaiseService を手動配線 | metrics の依存グラフ（§3） |
| **P-MOCK-01** Function-Field Mock | BudgetSettingsReader / WalletReader / OrderHistoryReader をフィールド差し替えでモック | なし |
| **P-OBS-01** Latency Measurement | `GetMetrics` の総所要時間を `bedrockLatencyMs` ではなく `metricsLatencyMs` として計測 | フィールド名のみ差替 |
| **P-OBS-03** Layered Logging | 正常は `slog.Info`、ErrNoBudgetSet は `slog.Warn`、DynamoDB 障害は `slog.Error` | なし |
| **P-FE-LOAD-01** Loading State | `useMetrics` / `useBudgetRaise` の `isLoading` 中は skeleton/spinner 表示 | なし |

**採用しないパターン**:
- P-RETRY-01（Bedrock なし）
- P-PLAN-01（Bedrock なし）
- P-OBS-02（LogSummary）→ Q-DD4=B で採用見送り、P-OBS-01 の latency 計測のみで対応

---

## 3. P-ME-PARALLEL-01: MetricsService 並列読取（Q-DD2=A）

**由来**: NFRE-E01（P95 500ms）、nfr-requirements.md §2.2

**目的**: `GetMetrics` が DynamoDB を 3 回読む（BudgetSettings / Wallet / CountThisMonth）うち、独立した BudgetSettings + Wallet を `errgroup` で並列化してレイテンシを最小化する。

**設計**:

```go
// apps/api/internal/metrics/service.go
func (s *MetricsService) GetMetrics(ctx context.Context, userID string) (*Metrics, error) {
    // Phase 1: 並列読取（BudgetSettings + Wallet は互いに独立）
    g, gctx := errgroup.WithContext(ctx)
    var budgetSettings *budget.BudgetSettings
    var wallet *budget.Wallet

    g.Go(func() error {
        bs, err := s.BudgetSettingsReader.Get(gctx, userID)
        budgetSettings = bs
        return err
    })
    g.Go(func() error {
        w, err := s.WalletReader.Get(gctx, userID)
        wallet = w
        return err
    })
    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("metrics: parallel read: %w", err)
    }

    // Phase 2: NotFound / ErrNoBudgetSet チェック（並列完了後）
    if budgetSettings == nil || budgetSettings.MonthlyBudget == 0 {
        return nil, ErrNoBudgetSet
    }
    if wallet == nil {
        return nil, ErrNoBudgetSet // Wallet と BudgetSettings は同時作成前提
    }

    // Phase 3: CountThisMonth を直列呼出（BudgetSettings 有無確認後）
    start := time.Now()
    count, err := s.OrderHistoryReader.CountThisMonth(gctx, userID)
    if err != nil {
        return nil, fmt.Errorf("metrics: count this month: %w", err)
    }
    metricsLatencyMs := time.Since(start).Milliseconds()
    slog.Info("metrics.GetMetrics", "userId", userID, "latencyMs", metricsLatencyMs, "count", count)

    // Phase 4: 計算（純関数）
    return computeMetrics(budgetSettings, wallet, count), nil
}
```

**重要な設計決定**:

| 決定 | 理由 |
|---|---|
| CountThisMonth は Phase 3（直列）| BudgetSettings 存在確認（ErrNoBudgetSet）後に呼ぶ必要がある。不要な DynamoDB アクセスを避ける |
| `errgroup.WithContext` を使用 | いずれかの goroutine がエラーを返したとき、もう一方の DynamoDB 呼出を context cancel で打ち切れる |
| エラー時は即 return（部分返却なし）| NFRE-E04。Metrics の一部だけ返すと UI に矛盾が生じる |

---

## 4. P-ME-FE-DEG-01: DEG UX（ダメ化演出）パターン（Q-DD3=A）

**由来**: NFRE-E08（80% 超警告色）/ NFRE-E09（RaiseModal 強制表示）/ NFR-DEG-03〜05

**目的**: ユーザーの退化ループを最大化する Frontend 演出を 1 パターンに集約し、Code Generation で一貫した実装を保証する。

### 4.1 ThresholdExceeded CSS 切替（NFRE-E08）

**トリガー**: `Metrics.ThresholdExceeded === true`（consumptionRate > 0.8 は Backend 判定済み）

```tsx
// MetricsPanel: プログレスバーと消化率テキストの条件スタイル
const barClass = metrics.thresholdExceeded
  ? "bg-red-500 animate-pulse"
  : "bg-blue-500"

const rateClass = metrics.thresholdExceeded
  ? "text-red-500 animate-pulse font-bold"
  : "text-gray-700"
```

**実装ルール**:
- `animate-pulse` は Tailwind のデフォルト（opacity 0.5 ↔ 1.0 の 2 秒サイクル）をそのまま使用
- `ThresholdExceeded` の判定は Backend 専権（Frontend での rate > 0.8 判定の重複禁止）

### 4.2 RaiseModal 強制自動表示（NFRE-E09）

**トリガー**: `BudgetEmptyScreen` のマウント

```tsx
// BudgetEmptyScreen: useState の初期値を true にしてマウント時に自動オープン
const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true)
// sessionStorage / localStorage によるフラグ管理なし（逃げられない）
```

**実装ルール**:
- `useState(true)` で初期化し、`useEffect` で open する二段階は不要
- 「今月はがんばる」ボタンで `setIsRaiseModalOpen(false)` にしても、再度 `/budget-empty` に来ると再びマウントされ自動表示
- `isRaiseModalOpen` を localStorage 等で永続化しない（退化ループを閉じない）

### 4.3 ダメコピー一覧（NFR-DEG-05）

| 要素 | コピー |
|---|---|
| BudgetEmptyScreen タイトル | 「今月はもうダメになれません」 |
| BudgetEmptyScreen サブテキスト | 「翌月 1 日に予算がリセットされます」 |
| RaiseModal タイトル | 「翌月予算を増額しますか？」 |
| RaiseModal 推奨額ラベル | 「推奨: ¥{X,XXX}（あなたには必要です）」 |
| RaiseModal 拒否ボタン | 「今月はがんばる」 |
| RaiseModal 承認ボタン | 「増額する」 |

---

## 5. P-E-PBT-01/02: Property-Based Testing（NFRE-E07）

Unit E の PBT は Unit C の `P-PBT-01`（gopter フレームワーク）を継承し、Unit E 固有の 2 プロパティを定義する。

### P-E-PBT-01: ConsumptionRate 不変条件

```go
// apps/api/internal/metrics/service_pbt_test.go
properties.Property("ConsumptionRate は常に [0.0, 1.0] に収まる", prop.ForAll(
    func(budget, remaining uint32) bool {
        b := int(budget%100_000) + 1          // [1, 100_000]
        r := int(remaining) % (b + 1)         // [0, b]
        rate := computeConsumptionRate(b, r)
        return rate >= 0.0 && rate <= 1.0
    },
    gen.UInt32(), gen.UInt32(),
))
```

### P-E-PBT-02: ComputeRecommendedBudget 不変条件

```go
properties.Property("推奨予算は現予算以上かつ 100_000 以下", prop.ForAll(
    func(current uint32) bool {
        c := int(current%100_000) + 1
        rec := computeRecommendedBudget(c)
        return rec >= c && rec <= 100_000
    },
    gen.UInt32(),
))
```

**採用ライブラリ**: `gopter`（P-PBT-01 継承、go.mod 登録済み）

---

## 6. NFR トレーサビリティ（NFRE-Exx → パターン）

| NFRE-E | パターン |
|---|---|
| NFRE-E01（P95 500ms） | P-ME-PARALLEL-01 + P-OBS-01（latency 計測） |
| NFRE-E02（P95 200ms） | P-INIT-01（cold start 短縮）|
| NFRE-E03（P95 300ms） | P-INIT-01 |
| NFRE-E04（DynamoDB 障害 → 500） | P-ME-PARALLEL-01（errgroup エラー集約）|
| NFRE-E05（ErrNoBudgetSet → 400） | P-ME-PARALLEL-01（Phase 2 判定）|
| NFRE-E06（新規テーブルなし） | — |
| NFRE-E07（PBT） | P-E-PBT-01 / P-E-PBT-02 |
| NFRE-E08（animate-pulse） | P-ME-FE-DEG-01 §4.1 |
| NFRE-E09（RaiseModal 強制） | P-ME-FE-DEG-01 §4.2 |
| NFRE-E10（AttachUserID） | P-INIT-01 / P-DI-01（middleware 配線） |

---

## 7. 文書管理

- **凍結契約への影響**: なし（P-ME-* は Unit E 内部パターン）
- **次ステージ**: Infrastructure Design（Standard）。logical-components.md と合わせて引き継ぐ
