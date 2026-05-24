# Unit C — Frontend Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order`)

## 生成ファイル

### lib + state (Step 8)

| ファイル | 用途 |
|---|---|
| `web/lib/ulid.ts` | LC-34、Frontend ULID 発行 (idempotencyKey 用) |
| `web/lib/toasts.ts` | LC-27、自虐トースト 3 種 + getRandomToast 純関数 |
| `web/lib/errorMappers.ts` | LC-26、`mapOrderError(err) → OrderErrorAction` 純関数 |
| `web/lib/api/orders.ts` | LC-33、`placeOrder` / `fetchOrderHistory` API ラッパ + `ApiError` class |
| `web/state/toastAtoms.ts` | LC-28、Jotai `toastsAtom` |

### Hooks (Step 9)

| ファイル | 用途 |
|---|---|
| `web/hooks/useDisableLock.ts` | LC-25、連打抑制 1 秒振る舞い hook |
| `web/hooks/useToast.ts` | LC-29、トースト表示 API |
| `web/hooks/useOrderHistory.ts` | LC-23、TanStack Query + `keepPreviousData` |
| `web/hooks/useOrder.ts` | LC-22、PlaceOrder mutation + 連打抑制 + invalidate |

### Components (Step 10)

| ファイル | 用途 |
|---|---|
| `web/components/order/Toast.tsx` | LC-31、個別トースト |
| `web/components/order/ToastHost.tsx` | LC-30、トースト一覧 (最大 3 件キュー) |
| `web/components/order/GoroButton.tsx` | LC-21、「ご飯めんどくさい」ボタン |
| `web/components/order/OrderHistoryList.tsx` | LC-24、履歴リスト + opacity 0.5 fetching 表示 |
| `web/components/order/OrderHistorySkeleton.tsx` | LC-24、初回ロード スケルトン |

### Pages (Step 11)

| ファイル | 用途 |
|---|---|
| `web/app/layout.tsx` (編集) | `<ToastHost />` を `<AppProviders>` 配下に配置 |
| `web/app/page.tsx` (編集) | MainScreen に `<GoroButton />` + `<OrderHistoryList />` 統合 |
| `web/app/order/[id]/complete/page.tsx` | LC-32、5 秒自動遷移 |

### テスト (Step 12)

| ファイル | テスト数 |
|---|---|
| `web/tests/toasts.test.ts` | 4 |
| `web/tests/errorMappers.test.ts` | 6 |
| `web/tests/useDisableLock.test.tsx` | 4 (fakeTimers で時間境界検証) |
| `web/tests/useToast.test.tsx` | 3 |
| `web/tests/orders-api.test.ts` | 5 (vi.mock で apiClient 差替) |
| `web/tests/ToastHost.test.tsx` | 4 (4 件追加で 3 件表示) |

## NFR Design ↔ コード対応

| パターン | 実装ファイル |
|---|---|
| P-FE-ERR-01 Order Error Mapping | `errorMappers.ts` |
| P-FE-TOAST-01 Random Toast Variant | `toasts.ts` |
| P-FE-TOAST-02 Toast Host & Queue | `ToastHost.tsx` + `toastAtoms.ts` + `useToast.ts` |
| P-FE-LOAD-01 Order History Loading State | `useOrderHistory.ts` (`placeholderData: keepPreviousData`) + `OrderHistoryList.tsx` |
| P-FE-LOCK-01 Disable Lock Hook | `useDisableLock.ts` |

## NFR 達成根拠

| NFR | 実装 |
|---|---|
| NFRC-C19 (TanStack Query 60s + invalidate) | `useOrderHistory.ts` の staleTime/gcTime/refetchOnWindowFocus 設定、`useOrder.ts` の onSuccess invalidate |
| NFRC-C22 (エラー UX、自虐トースト 3 種、連打抑制 1 秒) | mapOrderError + getRandomToast + useDisableLock の組合せ |
| NFRC-C24 (Bedrock 本文非表示) | Frontend は表示しない (Bedrock 結果は store / menu のみ) |
| NFR-DEG-03 (履歴常時可視化) | placeholderData: keepPreviousData |
| NFR-DEG-05 (自虐コピー) | 3 種トースト文言、空状態「履歴がまだありません」 |
| NFR-DEG-01 (1 タップ動線) | GoroButton 1 タップ → 完了画面 5 秒 → MainScreen 自動遷移 |

## テスト結果サマリ

```
Test Files  12 passed (12)
     Tests  44 passed (44)
```

Unit C 新規 26 テスト + Unit A 既存 18 テスト、すべて通過。

## 後続ステージへの引き継ぎ

- Unit B `BalanceDisplay` は Unit C の MainScreen に統合する想定 (Unit B Code Generation で実装、Unit C 側は variant を残しておく)
- Unit D サジェストカードは MainScreen に追加 (履歴 5 件以上時、Unit D Code Generation で実装)
- E2E テスト (Playwright) は Build & Test ステージで追加検討
