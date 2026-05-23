// vitest 上で `@testing-library/jest-dom` の matcher (toBeInTheDocument 等) を
// 型レベルで有効化するための ambient declaration 取り込みファイル。
// tsconfig.json で `types: [...]` を使うと `@types/*` の自動取り込みが
// 無効化されるため、include 経由で本ファイルだけを取り込む方式に揃える。
import "@testing-library/jest-dom/vitest";
