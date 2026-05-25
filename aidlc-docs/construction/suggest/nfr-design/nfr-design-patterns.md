# Unit D (`suggest`) — NFR Design Patterns

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Design
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Related**: Plan [suggest-nfr-design-plan.md](../../plans/suggest-nfr-design-plan.md)、NFR-R [nfr-requirements.md](../nfr-requirements/nfr-requirements.md)（NFRD-Dxx）、継承元 [order/nfr-design/nfr-design-patterns.md](../../order/nfr-design/nfr-design-patterns.md)（P-*）
**方針**: Q-DD1〜Q-DD6（全 A）に基づき、Bedrock/観測性/DI/テストの基盤は Unit C パターンを再利用、suggest 固有 orchestration を新規定義。

---

## 1. パターン識別子規則

- 形式: `P-SG-{名前}-{番号}`（Pattern / suggest）
- Unit C パターン再利用時は「**再利用: P-xxx**」と明記し、差分のみ記述

---

## 2. Unit C パターン再利用マップ（Q-DD1=A）

| Unit C パターン | suggest での再利用 | 差分 |
|---|---|---|
| **P-RETRY-01** Bedrock Retry Classification | `InferSuggestion` のリトライ判定にそのまま使用（共有 `RetryClassifier` = LC-ORDER-07） | なし |
| **P-OBS-01** Latency Measurement | `bedrockLatencyMs` 計測に `Measure(ctx, "bedrock", ...)` を流用 | なし |
| **P-OBS-02** LogSummary | → `P-SG-OBS-01`（SuggestLogSummary）として項目を差し替え | ログ項目（§4） |
| **P-OBS-03** Layered Logging | そのまま流用（INFO/WARN の層別） | なし |
| **P-INIT-01** Lambda Cold Start | package-level の Bedrock/DynamoDB client を共有（LC-ORDER-14/15 再利用） | なし |
| **P-DI-01** Manual Dependency Injection | `main.go` で SuggestService 系を配線 | suggest の依存グラフ（§3） |
| **P-MOCK-01** Function-Field Mock | SuggestService/Builder テストで Bedrock/Fallback を closure 差し替え | なし |
| **P-PBT-01** Property-Based Testing | → `P-SG-PBT-01`（軽量版、§5） | 範囲縮小 |
| **P-FE-LOAD-01** Loading State | `useSuggestion` の `isLoading` に流用 | なし |
| P-FE-ERR-01 / P-FE-TOAST-01〜02 / P-FE-LOCK-01 | **再利用しない**（suggest は表示のみ、エラーはカード非表示に丸め、NFRD-D17） | — |

---

## 3. P-SG-BUILD-01: Suggestion Construction Strategy（Q-DD2=A）

P-PLAN-01（Unit C の Plan Construction）の suggest 版。GetSuggestion の中核 orchestration を `SuggestionBuilder` に集約する。

### 依存（DI、P-DI-01 流用）
- `OrderHistoryReader`（Unit C 所有、読取）
- `BedrockAdapter`（共有 LC-ORDER-05）
- `RetryClassifier`（共有 LC-ORDER-07、P-RETRY-01）
- `FallbackSuggestProvider`（共有 LC-ORDER-09）
- `SuggestionStore`（Unit D 所有 LC-SUGGEST-04）

### orchestration 擬似コード
```
SuggestionBuilder.Build(ctx, userID):
  history = OrderHistoryReader.ListRecent(userID, 30)

  // 履歴十分判定（NFRD: 5件、BR-D01）
  if count(history, last30d) < 5: return NoSuggestion

  // 抑制判定（BR-D03）— SuggestService 側で実施し Builder には十分性のみ渡す設計も可
  // （本パターンでは Builder は推論+フォールバックに集中、抑制は SuggestService）

  // Bedrock 推論 + リトライ（P-RETRY-01 / NFRD-D03）
  for attempt in 1..2:
    ctx2 = WithTimeout(ctx, 1500ms)
    out, err = bedrock.InferSuggestion(ctx2, brief(history))
    if err == nil && out.HasSuggestion: 
      return Generated(out.Plan, out.Title, fallbackUsed=false)
    if !RetryClassifier.ShouldRetry(err): break   // 永続エラーは即抜け

  // フォールバック（BR-D05/D06 / NFRD-D04）
  plan = FallbackSuggestProvider.BuildFromHistory(history)
  if plan == nil: return NoSuggestion
  return Generated(plan, title=nil, fallbackUsed=true)
```

### SuggestService の責務（Builder の外側）
- 抑制判定（直近 3h 同カテゴリ注文、BR-D03）
- Builder 呼び出し → Generated なら suggestionId 採番 + SuggestionStore.Save（TTL 30分）
- SuggestLogSummary への記録（P-SG-OBS-01）

> Builder と抑制判定を分離することで、Builder は「推論+フォールバック」に純化しテスト容易性を保つ。

---

## 4. P-SG-OBS-01: SuggestLogSummary（Q-DD4=A、P-OBS-02 流用）

リクエスト単位で構造化ログを 1 行に集約する LogSummary パターン（P-OBS-02 と同型）。

| 項目群 | 項目 |
|---|---|
| 共通 8 項目（Unit A LC-AUTH-05） | level / timestamp / userId / action / traceId / requestId / email_hash / userAgent |
| suggest 固有 7 項目（NFRD-D11） | suggestionId / hasSuggestion / fallbackUsed / historyCount / suppressedRecentOrder / bedrockLatencyMs / bedrockAttempt |

- **action**: `"get_suggestion"` / `"resolve_suggestion"`
- **PII 排除**: Bedrock プロンプト/レスポンス本文は記録しない（NFRD-D16、構造的に LogSummary に本文フィールドを持たせない）
- **カスタムメトリクスなし**（NFRD-D10）。フォールバック率等は `fallbackUsed` のログをメトリクスフィルタで集計
- **層別（P-OBS-03）**: 正常は INFO 1 行、Bedrock フォールバック発動は WARN

---

## 5. P-SG-PBT-01: suggest Property-Based Testing（Q-DD6=A、P-PBT-01 軽量版）

| プロパティ | 対象 | フレームワーク |
|---|---|---|
| 最頻判定の決定性 | `FallbackSuggestProvider.BuildFromHistory`: 同じ履歴入力 → 同じ最頻 Plan、同点は最新 OrderedAt 優先 | gopter（backend） |
| レスポンス→props ラウンドトリップ | `GET /api/suggest` JSON → `useSuggestion` props → 再シリアライズで不変 | fast-check（frontend） |

- Bedrock は P-MOCK-01（Function-Field Mock）で決定論化
- 主要分岐（履歴十分 / 抑制 / フォールバック）は通常ユニットテスト + 境界値（履歴 4/5/6 件）で担保

---

## 6. P-SG-FE-01: Frontend Suggestion Display（Q-DD5=A）

| 要素 | パターン |
|---|---|
| `useSuggestion` | TanStack Query（queryKey `['suggestion']`、staleTime 実質無限、refetchOnWindowFocus:false、retry:0）。`isLoading` は P-FE-LOAD-01 流用。取得失敗は `hasSuggestion=false` 相当に丸めカード非表示（NFRD-D17） |
| `SuggestBubble` | GoroButton 内蔵の表示専用コンポーネント。`screenState==='suggested'` 時に「そろそろだろ。」表示（design spec §3.3、BR-D13/D16）。非インタラクティブ |
| エラー/連打 | **専用ハンドリングなし**（Unit C の toast/lock は不採用）。「押す。」押下時の注文送信は Unit C `useOrder` に委譲 |

---

## 7. NFR トレーサビリティ（NFRD-Dxx → パターン）

| NFRD-D | パターン |
|---|---|
| NFRD-D03（Bedrock 1.5s×1） | P-RETRY-01 再利用 + P-SG-BUILD-01 |
| NFRD-D04（フォールバック） | P-SG-BUILD-01 |
| NFRD-D05（TTL30分/失効nil） | P-SG-BUILD-01（Save）+ LC-SUGGEST-04 |
| NFRD-D10（カスタムメトリクスなし） | P-SG-OBS-01 |
| NFRD-D11（構造化ログ） | P-SG-OBS-01（P-OBS-02 流用）+ P-OBS-03 |
| NFRD-D12（軽量 PBT） | P-SG-PBT-01 |
| NFRD-D13（mock） | P-MOCK-01 再利用 |
| NFRD-D15（256MB/arm64/cold） | P-INIT-01 再利用 |
| NFRD-D16（モデル/PII） | P-INIT-01 + P-SG-BUILD-01（履歴のみ送信） |
| NFRD-D17（FE マウント1回/失敗丸め） | P-SG-FE-01 |

---

## 8. 文書管理

- **凍結契約への影響**: なし（P-SG-* は Unit D 内部、共有 Adapter の IF は凍結契約 §7 のまま）
- **次ステージ**: Infrastructure Design（Standard）。logical-components.md と合わせて引き継ぐ
