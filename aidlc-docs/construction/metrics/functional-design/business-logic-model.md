# Business Logic Model — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. ユースケース概要

Unit E は「退化ループの完成点」として機能する。ユーザーが今月何回・いくらダメになったかを可視化し、残高 0 になったユーザーを翌月予算増額へと誘導することで、ダメ化 UX のフェーズ 3 体験（US-X-03）を実現する。

```
[MainScreen]
  |-- MetricsPanel (DamageCount / ConsumptionRate / 警告色)
  |     |-- useMetrics → GET /api/metrics
  |     |-- useOrder mutation 成功後に invalidateQueries で自動更新
  |     |-- remainingBalance === 0 → router.push('/budget-empty')
  |
[BudgetEmptyScreen]
  |-- マウント時に RaiseModal を自動表示
  |-- useBudgetRaise → GET /api/budget/raise/recommendation
  |                  → POST /api/budget/raise (Accept)
```

---

## 2. ビジネスロジック詳細

### 2.1 MetricsService.GetMetrics

**入力**: `userID string`

**処理フロー**:

```
1. BudgetSettingsReader.Get(userID)
   - NotFound / monthlyBudget = 0 → return ErrNoBudgetSet

2. WalletReader.Get(userID)
   - NotFound → return ErrNoBudgetSet (BudgetSettings と Wallet は同時作成前提)

3. OrderHistoryReader.CountThisMonth(ctx, userID)
   - unit-interfaces.md 凍結済みメソッド（追加実装不要）
   - 実装側が当月 JST で件数を返す

4. ConsumptionRate 計算
   - spent = monthlyBudget - remainingBalance
   - rate = float64(spent) / float64(monthlyBudget)
   - clamp: rate = min(max(rate, 0.0), 1.0)
     ※ monthlyBudget > 0 は step 1 で保証済み
     ※ remainingBalance < 0 は Unit B NFR-REL-01 で発生しないが防御クランプを設ける

5. ThresholdExceeded = rate > 0.8

6. SummaryText 生成 (Backend 固定フォーマット)
   - 消化額 = monthlyBudget - remainingBalance
   - "今月のダメ化回数: {DamageCount} 回、消化額 ¥{spent:,}"

7. return Metrics{
     DamageCount:     count,
     ConsumptionRate: rate,
     MonthlyBudget:   monthlyBudget,
     RemainingBalance: remainingBalance,
     ThresholdExceeded: rate > 0.8,
     SummaryText:     summaryText,
   }
```

**エラー一覧**:

| エラー | HTTP | 条件 |
|---|---|---|
| `ErrNoBudgetSet` | 400 / ERR_NO_BUDGET_SET | BudgetSettings 未作成 or monthlyBudget=0 |
| `ErrUnauthorized` | 401 | JWT 検証失敗（Unit A middleware が先処理） |
| 内部エラー | 500 | DynamoDB 障害等 |

---

### 2.2 BudgetRaiseService.ComputeRecommendedBudget

**入力**: `userID string`

**処理フロー**:

```
1. BudgetSettingsReader.Get(userID) → currentBudget
   - NotFound → return ErrNoBudgetSet

2. recommended = min(int(float64(currentBudget) * 1.5), 100_000)
   - 上限 100,000 円（unit-of-work.md 確定済み）
   - 小数点以下切り捨て

3. return recommended
```

**例**:

| currentBudget | recommended |
|---|---|
| 30,000 | 45,000 |
| 70,000 | 100,000 (上限クランプ) |
| 100,000 | 100,000 (変化なし) |

---

### 2.3 BudgetRaiseService.Accept

**入力**: `userID string`, `newMonthlyBudget int`

**処理フロー**:

```
1. バリデーション
   - 1 <= newMonthlyBudget <= 100,000 → OK
   - それ以外 → return ErrInvalidBudget

2. effectiveFrom 計算 (サーバサイド、JST)
   - now = time.Now().In(Asia/Tokyo)
   - effectiveFrom = 翌月 1 日 00:00:00 JST
     例: 2026-05-25 → 2026-06-01T00:00:00+09:00

3. BudgetSettingsWriter.Set(ctx, userID, newMonthlyBudget, effectiveFrom)
   - Unit B が公開する Write Interface
   - 当月 Wallet.Balance は据え置き（翌月リセット時に新 monthlyBudget が適用）

4. return BudgetRaiseResult{
     NewMonthlyBudget: newMonthlyBudget,
     AppliedFrom:      effectiveFrom,
   }
```

**エラー一覧**:

| エラー | HTTP | 条件 |
|---|---|---|
| `ErrInvalidBudget` | 400 | newMonthlyBudget < 1 or > 100,000 |
| `ErrNoBudgetSet` | 400 | BudgetSettings 未作成 |
| `ErrUnauthorized` | 401 | JWT 検証失敗 |

---

## 3. フロントエンド側ビジネスロジック

### 3.1 BudgetEmpty 遷移判定

```typescript
// useMetrics hook 内
useEffect(() => {
  if (metrics?.remainingBalance === 0) {
    router.push('/budget-empty')
  }
}, [metrics?.remainingBalance])
```

- `useMetrics` の `isError` かつエラーコード `ERR_NO_BUDGET_SET` → `/budget` へリダイレクト
- `remainingBalance === 0` → `/budget-empty` へリダイレクト

### 3.2 MetricsPanel 更新トリガー

```typescript
// useOrder hook 内 (Unit C 実装済み、Unit E で invalidateQueries を追加)
onSuccess: () => {
  queryClient.invalidateQueries({ queryKey: ['metrics'] })
  queryClient.invalidateQueries({ queryKey: ['wallet'] })
}
```

Unit C の `useOrder` mutation 成功時に `['metrics']` クエリを無効化し、MetricsPanel が自動で最新データを取得する。

### 3.3 RaiseModal 自動表示

```typescript
// BudgetEmptyScreen
const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true) // マウント時に自動オープン
```

BudgetEmptyScreen のマウントと同時に `isRaiseModalOpen = true` で RaiseModal を表示する。ユーザーはモーダルを閉じることができるが、画面は BudgetEmptyScreen のまま残る。

---

## 4. API ハンドラ処理フロー

### 4.1 GET /api/metrics

```
1. UserID = auth.UserIDFromContext(c)
2. metrics, err = metricsService.GetMetrics(ctx, userID)
3. エラーマッピング:
   - ErrNoBudgetSet → 400 {"code": "ERR_NO_BUDGET_SET", "message": "..."}
   - その他 → 500
4. 200 OK: MetricsResponse{...}
```

### 4.2 GET /api/budget/raise/recommendation

```
1. UserID = auth.UserIDFromContext(c)
2. recommended, err = budgetRaiseService.ComputeRecommendedBudget(ctx, userID)
3. 200 OK: {"currentMonthlyBudget": X, "recommendedMonthlyBudget": Y}
```

### 4.3 POST /api/budget/raise

```
1. UserID = auth.UserIDFromContext(c)
2. req.Bind → BudgetRaiseRequest{NewMonthlyBudget int}
3. result, err = budgetRaiseService.Accept(ctx, userID, req.NewMonthlyBudget)
4. 200 OK: BudgetRaiseResponse{NewMonthlyBudget, AppliedFrom}
```

---

## 5. 担当ストーリーとの対応

| ストーリー | 対応ロジック |
|---|---|
| US-3-01 ダメ化回数の表示 | GetMetrics.DamageCount（当月 OrderHistory 件数） |
| US-3-02 消化率と警告色 | GetMetrics.ConsumptionRate + ThresholdExceeded（>0.8 で赤） |
| US-3-03 残高 0 時の BudgetEmptyScreen | useMetrics → remainingBalance===0 → router.push |
| US-3-04 増額誘導モーダル | BudgetEmptyScreen マウント → RaiseModal 自動表示 → Accept |
| US-X-03 退化ループの完成 | Accept 後に翌月から増額適用、ループが継続 |
