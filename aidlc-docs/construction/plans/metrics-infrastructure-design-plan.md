# Unit E `metrics` — Infrastructure Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard
**Stage**: Infrastructure Design (Construction Phase, per-unit)

---

## 0. このドキュメントの目的

Unit E の Infrastructure Design 成果物（infrastructure-design.md / deployment-architecture.md）生成前に、インフラ設計上の意思決定を収集する。

方針: 既存コード調査により Unit E は DynamoDB テーブル新規なし・IAM/env 既存カバー済みのため質問数は最小限（Q-I1〜Q-I3）。

---

## 1. Plan 実施チェックリスト

- [x] Step 1: 既存インフラ module / routes.tf / iam.tf を調査済み
- [x] Step 2: Unit E 固有インフラ要件を特定（API Gateway 3 route 追加のみ）
- [x] Step 3: ヒアリング質問を抽出（Q-I1〜Q-I3）
- [x] Step 4: チャットで 1 問ずつ提示してヒアリング（Q-I1〜Q-I3 全問完了）
- [x] Step 5: 回答を `[Answer]:` タグへ反映
- [x] Step 6: infrastructure-design.md / deployment-architecture.md 生成

---

## 2. 調査結果サマリ

| 項目 | 調査結果 | Unit E での対応 |
|---|---|---|
| 新規 DynamoDB テーブル | なし（NFRE-E06） | **対応不要** |
| 新規 Terraform module | 不要 | 既存 module 追記のみ |
| `DDB_TABLE_WALLET` env | `lambda_api/api_lambda.tf` に既存変数 `var.wallet_table_name` で注入済み | **追記不要** |
| `DDB_TABLE_BUDGET_SETTINGS` env | 同上、`var.budget_settings_table_name` で注入済み | **追記不要** |
| `ORDER_HISTORY_TABLE_NAME` env | 同上、`var.order_history_table_name` で注入済み | **追記不要** |
| BudgetSettings R/W IAM | `modules/budget/iam.tf` の `api_lambda` policy に `GetItem/PutItem/UpdateItem` 含む | **追記不要** |
| Wallet R IAM | 同上 | **追記不要** |
| OrderHistory R IAM | `modules/order_history/` にて付与済み（Unit C 実装） | **追記不要** |
| API Gateway routes | `GET /api/metrics` / `GET /api/budget/raise/recommendation` / `POST /api/budget/raise` が未追加 | **追記必要** |

---

## 3. 質問

### Q-I1: 新規 Terraform module の要否

既存調査の通り、Unit E は DynamoDB テーブル新規なし・IAM/env 追加なし。`modules/api_gateway/routes.tf` への 3 route 追記のみで足りるか確認。

A) 新規 module は不要。`modules/api_gateway/routes.tf` に 3 route を追記するだけで Unit E のインフラは完結する（調査通り）
B) Unit E 専用 IAM policy を新規 `modules/metrics/iam.tf` として切り出す（BudgetSettings Write を Unit B の policy と分離したい場合）
X) Other

[Answer]: A — 新規 module 不要。routes.tf への 3 route 追記のみ

---

### Q-I2: 3 エンドポイントの認証方式

凍結契約 §6.2 で全エンドポイントが「認証: 必須」と定義されている。Terraform route の `authorization_type` について。

A) 3 エンドポイントすべて `authorization_type = "JWT"` + `authorizer_id = aws_apigatewayv2_authorizer.cognito.id`（Unit B/C と同一パターン、認証必須）
B) 一部（例: GET /api/metrics）を `NONE` にしてバックエンド側で JWT 検証する（Unit A middleware を信頼するパターン）
X) Other

[Answer]: A — 全エンドポイント JWT 認証必須

---

### Q-I3: Terraform test の追加方針

`modules/api_gateway/tests/api_gateway_basic.tftest.hcl` は既存の route（logout / health / orders / wallet）の assert を持つ。Unit E の 3 route を追加するとき。

A) 既存の `api_gateway_basic.tftest.hcl` の `resources_present` run に Unit E route の assert を追記する（1 ファイルで管理、シンプル）
B) Unit E 専用の `api_gateway_metrics_routes.tftest.hcl` を新規作成する（route 単位でファイルを分ける）
X) Other

[Answer]: A — 既存 api_gateway_basic.tftest.hcl に追記

---

## 4. 矛盾チェックメモ（回答収集後に記入）

| 観点 | チェック内容 | 結果 |
|---|---|---|
| Q-I1 × NFRE-E10 | AttachUserID middleware と IAM policy の整合 | ✅ Unit A middleware は API Gateway JWT 認証後に userID を context 注入。IAM は Unit B/C policy でカバー済み。矛盾なし |
| Q-I2 × 凍結契約 §6.2 | 認証必須定義と Terraform route 設定の整合 | ✅ 全 3 route が JWT 認証必須。凍結契約 §6.2「認証: 必須」と一致 |
| Q-I3 × 既存 tftest 数 | test 追加後の総 tftest 数が計画通りか | ✅ 既存 api_gateway_basic.tftest.hcl に 3 assert 追記。ファイル数は変わらず管理シンプル |
