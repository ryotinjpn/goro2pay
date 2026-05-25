# Frontend Summary — Unit B (`budget`)

## ディレクトリ構造

```
web/
├── lib/api/wallet.ts                  # ★ getWallet / postWalletBudget (apiClient.request 経由)
├── hooks/
│   ├── useWallet.ts                    # ★ TanStack Query (queryKey: ['balance'], stale 30s)
│   └── useSetBudget.ts                 # ★ mutation + invalidate ['balance'] + router.push('/')
├── state/budget.ts                    # ★ insufficientBalanceAtom (jotai)
├── components/budget/                  # ★ 新規ディレクトリ
│   ├── BalanceDisplay.tsx              # 残高表示 (skeleton / placeholderData / 警告色)
│   ├── BudgetForm.tsx                  # クイックボタン 5 + 数値入力 + バリデーション + 送信
│   └── InsufficientBalanceModal.tsx    # 残高枯渇モーダル
├── app/
│   ├── (authenticated)/budget/page.tsx # ★ BudgetSetupScreen
│   └── page.tsx                        # ★ MainScreen に BalanceDisplay + Modal を最小追記
├── hooks/useOrder.ts                  # ★ 402 受信時に insufficientBalanceAtom.set(true) 追記
└── tests/
    ├── budget-state.test.ts            # ★ atom の挙動
    ├── wallet-api.test.ts              # ★ getWallet / postWalletBudget
    ├── BalanceDisplay.test.tsx         # ★ 4 シナリオ
    ├── BudgetForm.test.tsx             # ★ 6 シナリオ
    └── InsufficientBalanceModal.test.tsx  # ★ 4 シナリオ
```

## API ラッパ (`lib/api/wallet.ts`)

凍結 IF §3.3 の `Response` を `apiClient.request` 経由で取得し、JSON parse + `ApiError` (orders.ts と統一) を throw する。

```typescript
export async function getWallet(): Promise<WalletResponse>
export async function postWalletBudget(monthlyBudget: number): Promise<SetBudgetResponse>
```

## Hooks

### `useWallet` (LC-BUDGET-08)

凍結 IF §9 の 4 フィールドのみ公開:
```typescript
{ balance: number, monthlyBudget: number, isLoading: boolean, refetch(): void }
```

内部実装:
- `useQuery({ queryKey: ['balance'], staleTime: 30_000, placeholderData: keepPreviousData, retry: 0 })`
- queryKey は **クロスユニット契約** (unit-interfaces.md §9.1) — Unit C `useOrder` が invalidate

### `useSetBudget`

mutation + 自動 invalidate + redirect:
```typescript
useMutation({
  mutationFn: postWalletBudget,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['balance'] });
    router.push('/');
  },
})
```

## State (`state/budget.ts`)

```typescript
export const insufficientBalanceAtom = atom<boolean>(false);
```

クロスユニット契約:
- Unit C `useOrder` の onError で `err instanceof ApiError && err.status === 402` のとき `setInsufficientBalance(true)`
- Unit B `InsufficientBalanceModal` が `useAtom(insufficientBalanceAtom)` で購読 → open

## Components

### `BalanceDisplay` (LC-BUDGET-09 / P-DEG-01)

`useWallet` ではなく内部で `useQuery` を直接 subscribe (FD §3.1 注 — `isFetching` / `isError` を扱うため凍結 IF を回避)。

表示分岐:
- `query.isLoading` → スケルトン (`data-testid="budget-balance-skeleton"`、`animation: pulse`)
- `query.isError || !query.data` → エラー + 再試行ボタン
- 通常 → 残高 + 月間予算 + リセット日カウントダウン
- `query.isFetching=true && data` → opacity 0.5 (placeholderData 表示)

警告色: 消化率 > 0.8 で `⚠ ¥XX,XXX` + 赤色文字。

### `BudgetForm`

- クイックボタン 5 個 `[10000, 30000★, 50000, 80000, 100000]`
- 数値入力 (`type="number" min={1000} max={100000} step={1000}`)
- リアルタイムバリデーション (VR-B-01 範囲 / VR-B-02 1000 円刻み)
- 送信ボタン: validation + isPending で disabled
- 送信時: `mutate(monthlyBudget)`

### `InsufficientBalanceModal` (LC-BUDGET-10 / P-DEG-02)

- `useAtom(insufficientBalanceAtom)` で open 制御
- ESC キー / 背景クリックで close
- 「もっとダメになる」ボタン → `router.push('/budget')` + atom false
- 「閉じる」ボタン → atom false (遷移なし)
- a11y: `role="dialog"`, `aria-modal="true"`, `aria-labelledby`

## Page (`app/(authenticated)/budget/page.tsx`)

`BudgetSetupScreen`:
- `useWallet()` で既存予算を取得
- 既存ユーザは prefill (タイトル: 「ダメ予算を変更する」)
- 新規ユーザは BudgetForm 内で 30,000 円デフォルト (タイトル: 「ダメ予算を設定する」)

## MainScreen 統合 (`app/page.tsx`)

最小追記:
```tsx
<section style={{ marginBottom: 24 }}>
  <BalanceDisplay />
</section>
// (既存 GoroButton + OrderHistoryList をそのまま保持)
<InsufficientBalanceModal />
```

## Unit C 連携 (`hooks/useOrder.ts`) 最小追記

`onError` に 1 ブロック追加:
```tsx
if (err instanceof ApiError && err.status === 402) {
  setInsufficientBalance(true);
}
```
既存の `mapOrderError` 経由 `/budget-empty` ナビゲートはそのまま保持。modal と navigate が併存する。

## テスト集計

| ファイル | テスト数 | カバー |
|---|---|---|
| `tests/budget-state.test.ts` | 3 | atom 初期値 / set / reset |
| `tests/wallet-api.test.ts` | 4 | getWallet 200/404、postWalletBudget 200/400 |
| `tests/BalanceDisplay.test.tsx` | 4 | スケルトン / 通常 / 警告色 / エラー |
| `tests/BudgetForm.test.tsx` | 6 | クイックボタン / バリデーション 2 / disabled / submit / isPending |
| `tests/InsufficientBalanceModal.test.tsx` | 4 | atom false→render なし / atom true→open / 増額遷移 / 閉じる |

合計 **21 テスト追加**、既存 54 と合わせて **vitest 75 PASS**。
