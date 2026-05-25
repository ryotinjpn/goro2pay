# Frontend Components — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. コンポーネント一覧

| コンポーネント | 種別 | 配置 | 担当ストーリー |
|---|---|---|---|
| `useMetrics` | Hook | `web/hooks/useMetrics.ts` | US-3-01, US-3-02, US-3-03 |
| `useBudgetRaise` | Hook | `web/hooks/useBudgetRaise.ts` | US-3-04, US-X-03 |
| `MetricsPanel` | Component | `web/components/MetricsPanel.tsx` | US-3-01, US-3-02 |
| `BudgetEmptyScreen` | Page | `web/app/budget-empty/page.tsx` | US-3-03, US-3-04 |
| `RaiseModal` | Component | `web/components/RaiseModal.tsx` | US-3-04, US-X-03 |

---

## 2. Hooks

### 2.1 useMetrics

```typescript
// web/hooks/useMetrics.ts
interface Metrics {
  damageCount: number
  consumptionRate: number   // 0.0 .. 1.0
  monthlyBudget: number
  remainingBalance: number
  thresholdExceeded: boolean
  summaryText: string
}

function useMetrics(): {
  metrics: Metrics | undefined
  isLoading: boolean
  isError: boolean
  errorCode: string | undefined  // "ERR_NO_BUDGET_SET" など
}
```

**実装方針**:
- TanStack Query `useQuery({ queryKey: ['metrics'], queryFn: fetchMetrics })`
- `fetchMetrics` → GET /api/metrics
- エラーレスポンスから `errorCode` を抽出して返す

---

### 2.2 useBudgetRaise

```typescript
// web/hooks/useBudgetRaise.ts
interface Recommendation {
  currentMonthlyBudget: number
  recommendedMonthlyBudget: number
}

interface BudgetRaiseResult {
  newMonthlyBudget: number
  appliedFrom: string  // ISO 8601
}

function useBudgetRaise(): {
  recommendation: Recommendation | undefined
  isLoadingRecommendation: boolean
  accept: (newMonthlyBudget: number) => Promise<void>
  isAccepting: boolean
  acceptResult: BudgetRaiseResult | undefined
}
```

**実装方針**:
- `recommendation`: `useQuery({ queryKey: ['budget-raise-recommendation'], queryFn: fetchRecommendation })`
- `accept`: `useMutation` → POST /api/budget/raise
- `accept` 成功後に `queryClient.invalidateQueries(['metrics'])` + `queryClient.invalidateQueries(['wallet'])` で関連クエリを更新

---

## 3. Components

### 3.1 MetricsPanel

**配置**: `web/components/MetricsPanel.tsx`（`app/page.tsx` の MainScreen に組み込み）

**Props**:
```typescript
interface MetricsPanelProps {
  // なし（useMetrics hook を内部で呼ぶ）
}
```

**表示内容**:
- ダメ化回数: `{damageCount} 回`
- 消化率プログレスバー: `consumptionRate * 100%`
  - `thresholdExceeded = true` → バー・数値を赤色（警告色、NFR-DEG-03）
  - `thresholdExceeded = false` → 通常色
- `summaryText` テキスト表示

**状態管理**:
- Loading 中: スケルトン表示
- `errorCode === "ERR_NO_BUDGET_SET"`: コンポーネント内で `/budget` へ `router.push`
- `remainingBalance === 0`: コンポーネント内で `/budget-empty` へ `router.push`（BR-FE01）

**MetricsPanel の MainScreen への組み込みイメージ**:
```
app/page.tsx (MainScreen)
  +-- BalanceDisplay         (Unit B: 残高大表示)
  +-- MetricsPanel           (Unit E: 回数・消化率)
  +-- SuggestionCard         (Unit D: 先回りサジェスト)
  +-- GoroButton             (Unit C: ご飯めんどくさいボタン)
```

---

### 3.2 RaiseModal

**配置**: `web/components/RaiseModal.tsx`

**Props**:
```typescript
interface RaiseModalProps {
  isOpen: boolean
  onClose: () => void
  onAccept: (newBudget: number) => Promise<void>
  recommendation: {
    currentMonthlyBudget: number
    recommendedMonthlyBudget: number
  } | undefined
  isAccepting: boolean
  acceptResult: { newMonthlyBudget: number; appliedFrom: string } | undefined
}
```

**表示内容**:
- タイトル: 「翌月予算を増額しますか？」
- 推奨額表示: `推奨: ¥{recommendedMonthlyBudget:,}`（NFR-DEG-05: 自虐コピー）
- 「増額する」ボタン → `onAccept(recommendedMonthlyBudget)` を呼び出し
- 「今月はがんばる」ボタン → `onClose()`
- Accept 成功後: 「翌月 {appliedFrom 月} から ¥{newMonthlyBudget:,} が適用されます」を表示してモーダルを閉じる

**ローディング状態**:
- `isAccepting = true` → 「増額する」ボタンを disabled にしてスピナー表示

---

## 4. Page

### 4.1 BudgetEmptyScreen

**配置**: `web/app/budget-empty/page.tsx`

**表示内容**:
- タイトル: 「今月はもうダメになれません」（NFR-DEG-05: 自虐コピー）
- サブテキスト: 「翌月 1 日に予算がリセットされます」
- RaiseModal をマウント時に自動オープン（BR-FE02）

**状態管理**:
```typescript
const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true) // 自動オープン
const { recommendation, accept, isAccepting, acceptResult } = useBudgetRaise()
```

**ユーザーフロー**:
```
BudgetEmptyScreen マウント
  └── RaiseModal 自動オープン
        |-- 「増額する」押下 → accept(recommendedBudget) → 成功メッセージ → モーダル閉
        |-- 「今月はがんばる」押下 → モーダル閉（BudgetEmptyScreen は表示継続）
```

---

## 5. API クライアント関数

```typescript
// web/lib/api/metrics.ts
export async function fetchMetrics(): Promise<Metrics> {
  const res = await apiClient.get('/api/metrics')
  if (!res.ok) {
    const err = await res.json()
    throw { code: err.code, status: res.status }
  }
  return res.json()
}

export async function fetchRecommendation(): Promise<Recommendation> {
  const res = await apiClient.get('/api/budget/raise/recommendation')
  if (!res.ok) throw new Error('fetch recommendation failed')
  return res.json()
}

export async function postBudgetRaise(newMonthlyBudget: number): Promise<BudgetRaiseResult> {
  const res = await apiClient.post('/api/budget/raise', { newMonthlyBudget })
  if (!res.ok) throw new Error('budget raise failed')
  return res.json()
}
```

---

## 6. ユーザーインタラクションフロー

```
[MainScreen]
  MetricsPanel 表示
    |
    |-- consumptionRate > 0.8 → 警告色で表示（不安の演出）
    |-- remainingBalance === 0 → /budget-empty へ遷移
    |
    注文完了（GoroButton）
      └── useOrder mutation 成功
            └── invalidateQueries(['metrics']) → MetricsPanel 自動更新

[BudgetEmptyScreen]
  RaiseModal 自動表示
    |
    |-- 「増額する」→ POST /api/budget/raise
    |      └── 成功 → 「翌月から適用」メッセージ表示
    |
    └── 「今月はがんばる」→ モーダル閉（画面は維持）
```
