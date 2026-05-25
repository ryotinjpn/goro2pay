# NFR Requirements — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. NFR 一覧

| ID | カテゴリ | 要件 | 根拠 |
|---|---|---|---|
| NFRE-E01 | Performance | GET /api/metrics は p95 500ms 以内 | 3 DynamoDB 読み取り、BudgetSettings+Wallet 並列化で削減 |
| NFRE-E02 | Performance | GET /api/budget/raise/recommendation は p95 200ms 以内 | 1 DynamoDB 読み取りのみ |
| NFRE-E03 | Performance | POST /api/budget/raise は p95 300ms 以内 | 1 読み取り + 1 書き込み |
| NFRE-E04 | Reliability | いずれかの DynamoDB 読み取りが失敗したら HTTP 500 を返す（部分返却なし） | Q-N2=A。メトリクス信頼性優先 |
| NFRE-E05 | Reliability | `ErrNoBudgetSet` は HTTP 400 で返す（5xx にしない） | BR-M01。クライアントが /budget へ誘導できるようにする |
| NFRE-E06 | Scalability | Unit E は新規 DynamoDB テーブルなし。既存テーブルの読み取り負荷のみ | unit-of-work.md 確定済み |
| NFRE-E07 | Maintainability | ConsumptionRate と ComputeRecommendedBudget に gopter で PBT を適用 | Q-N3=A。PBT Partial 方針（Extension 設定）と整合 |
| NFRE-E08 | Usability | ThresholdExceeded=true 時に MetricsPanel を赤色 + animate-pulse で表示 | Q-N4=B。NFR-DEG-03: 不安の演出 |
| NFRE-E09 | Usability | BudgetEmptyScreen マウントのたびに RaiseModal を自動表示（再表示制御なし） | Q-N5=A。NFR-DEG-04: 退化ループを最大化 |
| NFRE-E10 | Security | 全エンドポイントに Unit A の AttachUserID middleware を適用 | unit-interfaces.md §2。Unit E は認証機構を持たない |

---

## 2. パフォーマンス詳細

### 2.1 GET /api/metrics のレイテンシ構造

```
GetMetrics レイテンシ内訳 (目標 p95 500ms):

[並列] BudgetSettingsReader.Get  〜 5ms  (DynamoDB GetItem)
[並列] WalletReader.Get          〜 5ms  (DynamoDB GetItem)
       ↓ errgroup で待機
[直列] CountThisMonth            〜10ms  (DynamoDB Query + FilterExpression)
       ↓
       ConsumptionRate 計算      〜 0ms  (純関数)
       SummaryText 生成          〜 0ms  (文字列フォーマット)
       ↓
合計                              〜15ms  (DynamoDB レイテンシのみ)
```

Lambda コールドスタートや API Gateway オーバーヘッドを含めても p95 500ms は十分達成可能。

### 2.2 DynamoDB 読み取り並列化パターン (Q-N1=C)

```go
// errgroup で BudgetSettings + Wallet を並列取得
g, ctx := errgroup.WithContext(ctx)
var budgetSettings *BudgetSettings
var wallet *Wallet

g.Go(func() error {
    bs, err := s.BudgetSettings.Get(ctx, userID)
    budgetSettings = bs
    return err
})
g.Go(func() error {
    w, err := s.Wallet.Get(ctx, userID)
    wallet = w
    return err
})
if err := g.Wait(); err != nil {
    return nil, err
}

// NotFound チェック後に CountThisMonth を直列呼び出し
if budgetSettings == nil || budgetSettings.MonthlyBudget == 0 {
    return nil, ErrNoBudgetSet
}
count, err := s.OrderHistory.CountThisMonth(ctx, userID)
```

---

## 3. 信頼性

### 3.1 エラーハンドリング方針 (Q-N2=A)

| エラー種別 | HTTP | 挙動 |
|---|---|---|
| BudgetSettings/Wallet DynamoDB 障害 | 500 | errgroup がエラーを捕捉して即返却 |
| OrderHistoryReader DynamoDB 障害 | 500 | 部分返却なし |
| BudgetSettings 未作成 (nil) | 400 / ERR_NO_BUDGET_SET | クライアントが /budget へ誘導 |
| 認証失敗 | 401 | Unit A middleware が先処理 |

---

## 4. テスト戦略

### 4.1 PBT 適用対象 (Q-N3=A)

**P-E-PBT-01: ConsumptionRate 不変条件**

```
プロパティ: 任意の (monthlyBudget ∈ [1, 100_000], remainingBalance ∈ [0, monthlyBudget]) に対して
  0.0 <= computeConsumptionRate(monthlyBudget, remainingBalance) <= 1.0
```

**P-E-PBT-02: ComputeRecommendedBudget 不変条件**

```
プロパティ: 任意の currentBudget ∈ [1, 100_000] に対して
  computeRecommendedBudget(currentBudget) >= currentBudget  (減少しない)
  computeRecommendedBudget(currentBudget) <= 100_000        (上限を超えない)
```

### 4.2 テーブルテスト

| 対象 | テストケース |
|---|---|
| GetMetrics | 正常 / BudgetSettings 未作成 / DynamoDB 障害 |
| ComputeRecommendedBudget | 通常値 / 上限クランプ (67,000 → 100,000) / 上限ちょうど (100,000 → 100,000) |
| Accept | 正常 / 範囲外 (0 / 100,001) / BudgetSettings 未作成 |

---

## 5. NFR-DEG（ダメ化 UX）要件

### 5.1 NFRE-E08: 警告色演出 (Q-N4=B)

`ThresholdExceeded = true`（ConsumptionRate > 0.8）のとき:
- プログレスバー: `bg-red-500 animate-pulse`
- 消化率テキスト: `text-red-500 animate-pulse font-bold`
- 通常時: `bg-blue-500` / `text-gray-700`

### 5.2 NFRE-E09: RaiseModal 強制再表示 (Q-N5=A)

- BudgetEmptyScreen マウントのたびに `isRaiseModalOpen = true` で初期化
- `sessionStorage` や `localStorage` によるフラグ管理なし
- 「今月はがんばる」で閉じても次回アクセス時に再表示（逃げられない退化ループ）

### 5.3 NFR-DEG-05: 自虐コピー

| 画面 | コピー |
|---|---|
| RaiseModal タイトル | 「翌月予算を増額しますか？」 |
| RaiseModal 推奨額 | 「推奨: ¥{X,XXX}（あなたには必要です）」 |
| RaiseModal 拒否ボタン | 「今月はがんばる」 |
| BudgetEmptyScreen タイトル | 「今月はもうダメになれません」 |
| BudgetEmptyScreen サブテキスト | 「翌月 1 日に予算がリセットされます」 |
