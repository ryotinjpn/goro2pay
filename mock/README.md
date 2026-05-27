# goro2pay Mock (React Sandbox)

PR #94 の design-spec をもとに作った、スタンドアロンの React モック。

## 起動

```bash
cd mock
npm install
npm run dev
# → http://localhost:5173
```

## 含まれている画面

| Path | 画面 | design-spec 参照 |
|---|---|---|
| `/` | ランディング (デモボタン1回でスロット → CTA 表示) | §5.1 |
| `/main` | メイン (idle / suggested / slot / dead を内部 state で遷移) | §3 |
| `/main/complete` | 注文完了 (`いい判断だ。`、5 秒後に `/main` へ自動遷移) | §5.4 |

## デバッグセレクタ

`/main` 画面の右上に小さなボタン群があり、`auto / idle / suggested / dead` の状態を強制できます (プレゼン中の状態切り替え用)。

## モックの範囲

- API は `src/lib/fakeApi.ts` の `setTimeout` で擬似応答 (1.4 秒)
- Auth (サインアップ / ログイン / モーダル) は対象外
- a11y / レスポンシブ / テストは最低限のみ
- 本番アプリ (`web/`) には影響しない独立サンドボックス
