# Unit B `budget` — Frontend Components

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: B (`budget` / ダメ予算)
**Stage**: Functional Design (Construction Phase)
**Depth**: Standard
**Frontend Stack**: Next.js 15 App Router (TypeScript) + TanStack Query + Jotai + AWS Amplify Hosting

---

## 0. このドキュメントの目的

Unit B `budget` のフロントエンド側コンポーネント階層・props/state・ユーザインタラクション・バリデーション・API 連携を定義する。実装は Code Generation で行うが、本ドキュメントは UI レベルの仕様確定に責任を持つ。

**凍結された Unit 間 Interface 契約**: REST API パス（`GET /api/wallet`、`POST /api/wallet/budget`）と `useWallet` フックの公開シグネチャは [unit-interfaces.md](../../interfaces/unit-interfaces.md) §3.3 / §9 で凍結済み。本ドキュメントの §3.1 `useWallet` は凍結契約と完全一致させ、エラー時の UI 制御はフック内部で `query.isError` を直接参照する方式（公開 API には含めない）に統一する。差異が必要な場合は先に unit-interfaces.md を更新する。

---

## 1. コンポーネント階層

```
app/budget/page.tsx (BudgetSetupScreen) ── Unit B 主担当
└── components/
    ├── BudgetForm.tsx          ── Unit B (初回設定 / 総額入力モデル)
    │   ├── QuickBudgetButtons.tsx  ── Unit B
    │   └── BudgetNumberInput.tsx   ── Unit B
    ├── AddBudgetForm.tsx       ── Unit B (既存ユーザ向け / 追加金額入力モデル、2026-05-27 追加)
    └── BudgetSubmitButton.tsx  ── Unit B

app/page.tsx (MainScreen) ── Unit C 主、Unit B は BalanceDisplay を提供
└── components/
    └── BalanceDisplay.tsx      ── Unit B（Unit C の MainScreen 上部に埋め込まれる）
        ├── BalanceAmount.tsx
        ├── BalanceLabel.tsx
        └── ResetCountdown.tsx
```

**hooks/**:
- `useWallet.ts` ── Unit B（残高取得）
- `useSetBudget.ts` ── Unit B（予算設定 mutation）

---

## 2. ページ・コンポーネント詳細

### 2.1 `BudgetSetupScreen` (`app/budget/page.tsx`)

**目的**: 初回登録後の予算設定 / 既存ユーザの予算変更画面

**ユースケース**: UC-B-01（初回設定）、UC-B-02（変更）

**フォーム出し分け** (背景: `aidlc-docs/audit.md` 2026-05-27T11:50:00Z エントリ):
- 初回 (Wallet 未作成): `BudgetForm` (総額入力モデル)
- 既存ユーザ: `AddBudgetForm` (追加金額モデル、`mutate(currentBudget + addition)` で総額送信)

両者とも backend には予算総額が渡るため API 契約は不変。

**レイアウト (初回 / BudgetForm 経路)**:

```
┌──────────────────────────────────────┐
│      ゴロゴロPay                      │
│                                       │
│   ダメ予算を設定する                   │
│                                       │
│   月間ダメ予算を 1,000〜100,000 円    │
│   (1,000 円刻み) で設定してください    │
│                                       │
│   よく使われる金額:                    │ ← QuickBudgetButtons
│   [10,000][30,000★][50,000]          │
│   [80,000][100,000]                   │
│                                       │
│   ┌─────────────────────────────┐    │
│   │ ¥ [   30,000   ]            │    │ ← BudgetNumberInput
│   └─────────────────────────────┘    │
│                                       │
│   ┌─────────────────────────────┐    │
│   │ ダメ予算を設定する             │    │
│   └─────────────────────────────┘    │
└──────────────────────────────────────┘
```

**AddBudgetForm 経路 (既存ユーザ)**:
- クイック追加チップ `[+10,000][+20,000][+30,000][+50,000]`
- カスタム追加金額入力 (1,000 円刻み)
- 合計プレビュー `合計: ¥{total} (現在 ¥{current})`
- `currentBudget + addition > 100,000` のチップは disabled
- 送信ボタン「ダメ予算を追加する」 → `mutate(currentBudget + addition)`

**Props**: なし（ページコンポーネント）

**State** (Client Component):
```typescript
interface BudgetSetupState {
  monthlyBudget: number | null;     // 入力値
  validationError: string | null;   // インライン表示用
  isSubmitting: boolean;            // 送信中フラグ
}
```

**初期値**:
- BudgetForm: `monthlyBudget` 初期 `30_000`
- AddBudgetForm: `addition` 初期 `10_000` (最初のクイックチップ)

**バリデーション**: VR-B-01 (1,000-100,000 円) / VR-B-02 (1,000 円刻み) は両 form 共通。AddBudgetForm では「合計が 100,000 円を超えない」も検証。送信時に `useSetBudget().mutate(...)` で `POST /api/wallet/budget`。

**API 連携**:
- `POST /api/wallet/budget` を `useSetBudget` フック経由で呼び出し（凍結 IF §3.3）
- 成功時: `queryClient.invalidateQueries(['wallet'])` で残高キャッシュ無効化、`router.push('/')` で MainScreen へ遷移
- 失敗時: `validationError` にエラーメッセージ表示

**ストーリー対応**: US-0-03

---

### 2.2 `BudgetNumberInput` (`components/BudgetNumberInput.tsx`)

**目的**: 数値入力フィールド + リアルタイムバリデーション

**Props**:
```typescript
interface BudgetNumberInputProps {
  value: number | null;
  onChange: (value: number | null) => void;
  onValidationError: (message: string | null) => void;
}
```

**HTML 仕様** (Q-B11 補足 = γ: 1,000 円刻み):
```tsx
<input
  type="number"
  min={1000}
  max={100000}
  step={1000}
  value={value ?? ''}
  onChange={handleChange}
  inputMode="numeric"
  aria-label="月間ダメ予算"
/>
```

**バリデーション** (VR-B-01, VR-B-02):
- 範囲外（`< 1000` または `> 100000`）: `"1,000 〜 100,000 円の範囲で入力してください"`
- 1,000 円刻み外（`% 1000 != 0`）: `"1,000 円単位で入力してください"`
- 空欄: バリデーションエラーは出さない（送信ボタン側で抑制）

**プレフィックス表示**: `¥` を input の左側に CSS で固定表示

**注**: `<input type="number">` は値を `string` で扱うので、内部で `parseInt` 変換 + NaN チェック実装。

---

### 2.3 `QuickBudgetButtons` (`components/QuickBudgetButtons.tsx`)

**目的**: よく使う金額のクイック選択ボタン群

**Props**:
```typescript
interface QuickBudgetButtonsProps {
  selected: number | null;       // 現在選択中の値（ハイライト表示用）
  onSelect: (value: number) => void;
}
```

**ボタン定義** (Q-B11 = A):
```typescript
const QUICK_BUDGETS = [10_000, 30_000, 50_000, 80_000, 100_000] as const;
```

**表示**:
- 5 個のボタンを横並び（モバイル時は 2 行 = 3+2 で折り返し）
- 各ボタンに金額ラベル（例: `¥10,000`）
- `selected` と一致するボタンはハイライト表示（背景色変更）
- ペルソナ初期推奨値 30,000 円のみアクセント色 + `★` 装飾

**インタラクション**:
- ボタンタップで `onSelect(value)` 発火
- ハイライト切替は親コンポーネント（`BudgetSetupScreen`）の State に従う

---

### 2.4 `BudgetSubmitButton` (`components/BudgetSubmitButton.tsx`)

**目的**: 予算設定 CTA ボタン

**Props**:
```typescript
interface BudgetSubmitButtonProps {
  monthlyBudget: number | null;
  validationError: string | null;
  isSubmitting: boolean;
  onSubmit: (monthlyBudget: number) => void;
}
```

**表示**:
- ラベル: `"ダメ予算を設定する"`（送信中は `"設定中..."`）
- 大きな CTA ボタン（画面下部固定 or BudgetForm 直下）
- ダメ化UX のトーンに合わせ、自虐的なコピー

**Disabled 条件**:
- `monthlyBudget == null`
- `validationError != null`
- `isSubmitting == true`

**インタラクション**:
- クリック時に `onSubmit(monthlyBudget)` 発火（親が mutation 実行）

---

### 2.5 `BalanceDisplay` (`components/BalanceDisplay.tsx`)

**目的**: 残高を MainScreen 上部に常時表示する帯

**ユースケース**: UC-B-03

**レイアウト** (Q-B12 = A):

```
┌──────────────────────────────────────┐
│                                       │
│   残りダメ予算                         │ ← BalanceLabel（小フォント）
│                                       │
│   ¥29,150                             │ ← BalanceAmount（特大フォント、80% 超で赤）
│                                       │
│   月末リセットまで あと 12 日          │ ← ResetCountdown（小フォント）
│                                       │
└──────────────────────────────────────┘
```

**Props**:
```typescript
interface BalanceDisplayProps {
  // 通常は props なしで内部で useWallet() を使う
  // ただしテスタビリティのため inject できる設計
  balance?: number;
  monthlyBudget?: number;
}
```

**State**:
- 内部で `useWallet()` を呼び `balance`, `monthlyBudget` 取得
- 内部で `now` を計算して翌月 1 日までの日数を算出

**警告色判定**:
```typescript
const consumptionRate = 1 - balance / monthlyBudget;
const isWarning = consumptionRate > 0.8;
const amountColor = isWarning ? 'red' : 'black';
```

**Loading / エラー状態**:
- `useWallet().isLoading == true` の間はスケルトン表示
- エラー時の表示（`"残高取得失敗"` + リトライボタン）は ErrorBoundary または `BalanceDisplay` 内で直接 `useQuery({ queryKey: ['wallet'] })` を購読して `query.isError` を判定する方式とする（凍結 IF の `useWallet` には `isError` を含めないため、§3.1 注を参照）

**ストーリー対応**: US-0-04（残高初期表示）、US-1-02（残高常時可視化）

---

### 2.6 `BalanceAmount` (`components/BalanceAmount.tsx`)

**目的**: 残高金額の特大表示

**Props**:
```typescript
interface BalanceAmountProps {
  balance: number;
  isWarning: boolean;
}
```

**表示**:
- フォーマット: `Intl.NumberFormat('ja-JP').format(balance)` で 3 桁区切り → `¥29,150`
- フォントサイズ: 画面幅の 12〜18% 相当（モバイル）、CSS で `clamp()` 使用
- `isWarning == true` 時は文字色を赤（CSS class 切替）

---

### 2.7 `BalanceLabel` (`components/BalanceLabel.tsx`)

**目的**: 「残りダメ予算」ラベル

**Props**: なし（固定文言）

**表示**: 小フォントで `"残りダメ予算"`

---

### 2.8 `ResetCountdown` (`components/ResetCountdown.tsx`)

**目的**: 月末リセットまでの日数カウントダウン（退化ループ動線、ダメ化UX 強化）

**Props**: なし（内部で `now` を計算）

**ロジック**:
```typescript
function computeRemainingDays(): number {
  const now = new Date();
  const nextMonth = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  const diffMs = nextMonth.getTime() - now.getTime();
  return Math.ceil(diffMs / (1000 * 60 * 60 * 24));
}
```

**表示**: `"月末リセットまで あと N 日"`（N == 1 のとき `"明日"` も可）

**再計算**: コンポーネントマウント時のみ計算、日付変わると再描画は ResetCountdown 親側 `BalanceDisplay` のリレンダーに依存（毎時間 invalidate しなくても精度十分）

---

## 3. カスタムフック

### 3.1 `useWallet`

**目的**: 残高 + 月間予算の取得・キャッシュ

**シグネチャ**（[unit-interfaces.md §9](../../interfaces/unit-interfaces.md) で凍結。本ドキュメントは凍結契約と完全一致させる）:
```typescript
function useWallet(): {
  balance: number;
  monthlyBudget: number;
  isLoading: boolean;
  refetch(): void;
};
```

**実装方針** (Q-B6 = A):
```typescript
function useWallet() {
  const query = useQuery({
    queryKey: ['wallet'],
    queryFn: () => api.getWallet(),  // GET /api/wallet（凍結 IF §3.3）
    staleTime: 30_000,         // 30 秒間 fresh
  });
  return {
    balance: query.data?.balance ?? 0,
    monthlyBudget: query.data?.monthlyBudget ?? 0,
    isLoading: query.isLoading,
    refetch: query.refetch,
  };
}
```

**注 (エラー時の UI 制御)**:
- 凍結 IF は `isError` を公開シグネチャに含めない方針。エラー UI が必要なコンポーネント（例: `BalanceDisplay`）は内部で `useQuery({ queryKey: ['wallet'], ... })` を直接呼ぶか、ErrorBoundary でラップして対応する
- 公開シグネチャを変更したい場合は先に unit-interfaces.md §9 を v1.1 に改定し、影響 Unit に共有してから本書を更新する

---

### 3.2 `useSetBudget`

**目的**: 予算設定 mutation（新規追加、Application Design では未定義だった）

**シグネチャ**:
```typescript
function useSetBudget(): {
  mutate(monthlyBudget: number): void;
  isPending: boolean;
  error: Error | null;
};
```

**実装方針**:
```typescript
function useSetBudget() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const mutation = useMutation({
    mutationFn: (monthlyBudget: number) => api.postWalletBudget(monthlyBudget),  // POST /api/wallet/budget（凍結 IF §3.3）
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallet'] });  // PR-B-05 自動 invalidate
      router.push('/');                                          // MainScreen へ遷移
    },
  });
  return {
    mutate: mutation.mutate,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
```

**注**: API クライアント `api.postWalletBudget` は Unit B の hooks 内で定義し、Unit 横串の `lib/api.ts` に集約する。

---

## 4. ルーティング

| Path | コンポーネント | 認証要否 | 利用ストーリー |
|---|---|---|---|
| `/budget` | `BudgetSetupScreen` | 必須 | US-0-03 |
| `/` | `MainScreen` (Unit C) — `BalanceDisplay` 埋め込み | 必須 | US-0-04, US-1-02 |
| `/budget-empty` | `BudgetEmptyScreen` (Unit E) — `BalanceDisplay` 埋め込み可 | 必須 | US-1-04（遷移先） |

**遷移ルール**:
- `/budget` 送信成功 → `/` へ遷移（`router.push('/')`）
- `/` 描画時に `useWallet()` で予算未設定（404 / NotFound）と判定された場合 → `/budget` にリダイレクト（middleware または useEffect）
- `/` 描画時に `balance == 0` を検出 → `/budget-empty` へリダイレクト（Unit E 側ロジック）

---

## 5. バリデーション仕様まとめ

| 入力 | ルール | エラーメッセージ |
|---|---|---|
| `monthlyBudget` 範囲 | `1000 ≤ x ≤ 100000` | `"1,000 〜 100,000 円の範囲で入力してください"` |
| `monthlyBudget` 刻み | `% 1000 == 0` | `"1,000 円単位で入力してください"` |
| 空欄 | `null` 不可 | （送信ボタン Disabled で抑制） |

**フロント側**: リアルタイム検証で送信前に弾く（NFR-DEG-01 摩擦排除）
**バックエンド側**: 重複検証（VR-B-01, VR-B-02、防御的）

---

## 6. アクセシビリティ

- `BudgetNumberInput`: `aria-label="月間ダメ予算"`、`inputMode="numeric"` でモバイルテンキー
- `QuickBudgetButtons`: 各ボタンに `aria-label` 付与（例: `"30,000 円を選択"`）
- `BalanceDisplay`: `role="status"` + `aria-live="polite"` で残高変化を読み上げ
- 警告色（赤）には文字記号（例: `⚠`）も併記して色覚配慮

---

## 7. 状態管理（Jotai atoms）

Unit B 単独で必要な atom:

```typescript
// state/atoms.ts

import { atomWithQuery } from 'jotai-tanstack-query';

// walletQuery: TanStack Query を Jotai 連携
export const walletQueryAtom = atomWithQuery(() => ({
  queryKey: ['wallet'],
  queryFn: () => api.getWallet(),
  staleTime: 30_000,
}));

// 派生 atom: 残高のみ
export const walletBalanceAtom = atom((get) => {
  const result = get(walletQueryAtom);
  return result.data?.balance ?? 0;
});

// 派生 atom: 月間予算のみ
export const monthlyBudgetAtom = atom((get) => {
  const result = get(walletQueryAtom);
  return result.data?.monthlyBudget ?? 0;
});
```

**注**: TanStack Query を Jotai に統合するか、`useQuery` 直接利用かは Application Design で `walletBalanceAtom (derived from walletQuery)` と記述されているので Jotai 統合を採用。

---

## 8. ダメ化UX 観点まとめ

| 体験 | 該当コンポーネント / フック |
|---|---|
| NFR-DEG-01 低摩擦オンボーディング | QuickBudgetButtons（1 タップ完結）+ BudgetNumberInput（リアルタイム検証） |
| NFR-DEG-03 残高常時可視化 | BalanceDisplay（MainScreen 上部固定） + ResetCountdown（消化圧力） |
| NFR-DEG-04 退化ループ | ResetCountdown が「あと N 日で失効」を演出、SetBudget 増額誘導との連携（Unit E） |
| 即時甘やかし（PR-B-02） | useSetBudget の onSuccess で `invalidateQueries` → 即座に新残高が表示 |

---

## 9. 既存設計（Application Design / 凍結 IF）との差分

| 項目 | 既存定義 | 本ドキュメント | 理由 |
|---|---|---|---|
| `BudgetSetupScreen` | 「予算額の入力・バリデーション」 | クイックボタン 5 個 + 数値入力（1,000 円刻み） | Q-B11 = A, Q-B11 補足 = γ |
| `BalanceDisplay` | （components.md §2.4 MainScreen 内に「大きな残高」と概略記載のみ） | 残高 + ラベル + リセット日カウントダウンの 3 要素 | Q-B12 = A |
| `useWallet` | 凍結 IF §9: `{ balance, monthlyBudget, isLoading, refetch }` | 同（凍結契約と完全一致） | 凍結 IF 準拠。エラー UI は ErrorBoundary または直接 useQuery で対応（§3.1 注） |
| `useSetBudget` | （Application Design では未定義、凍結 IF にも記載なし） | 新規追加（Unit B 内部の utility hook） | Unit 境界をまたがないため凍結 IF への追加不要 |

凍結 IF の公開シグネチャは Wave 1 並列化のため変更不可。`useSetBudget` は Unit B 内部の実装詳細であり、凍結 IF §9 への追記は行わない。

---

## 10. 審査観点へのトレーサビリティ

| 審査観点 | 対応 |
|---|---|
| ビジネス意図の明確さ | ResetCountdown のダメ化UX 動線、BudgetSubmitButton の自虐コピーで「ダメ化」テーマを UI に貫通 |
| 創造性とテーマ適合性 | 30,000 円の `★` 装飾でペルソナ初期値を視覚的に示唆、コピーのトーンを統一 |
| Unit 分解の適切さ | BalanceDisplay は Unit B が提供して Unit C の MainScreen に埋め込まれる役割分担を明示 |
| ドキュメント品質 | ASCII モック + props/state 型定義 + バリデーション + アクセシビリティを横断的に記述 |
