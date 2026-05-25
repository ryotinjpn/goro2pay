# Unit D (`suggest`) — Infrastructure Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Infrastructure Design
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Prerequisite**: NFR Design 承認済み（PR #102 マージ済み）、Unit C Infrastructure Design（PR #80 マージ済み）= module 構成の継承元
**Related**: [suggest/nfr-design/](../suggest/nfr-design/)（LC-SUGGEST-*）、継承元 [order/infrastructure-design/](../order/infrastructure-design/)、凍結契約 [unit-interfaces.md](../interfaces/unit-interfaces.md) §5・§7・§10

---

## 1. Plan の目的

Unit D `suggest` の論理コンポーネント（LC-SUGGEST-*）を AWS リソースにマッピングする。Unit D は Bedrock・観測性の Terraform module を **Unit C から再利用**し、suggest 固有の新規は **`modules/suggestion/`（DynamoDB）** + `GET /api/suggest` ルート + `lambda_api` への policy/env 追記に限られる。terraform-module-design / terraform-coding-rule / terraform-test 規約準拠。

## 2. Unit C インフラからの継承・差分

| 項目 | Unit C（既存） | Unit D |
|---|---|---|
| `modules/bedrock/` | Unit C/D **共有候補**として作成済み。bedrock_policy_arn は lambda_api に付与済み | **再利用**（新規 Bedrock IAM 不要） |
| `modules/observability/` | 横串、Unit D/E 再利用可 | **再利用**（suggest アラーム追加は要否判断 Q-DI5） |
| `modules/order_history/` | Unit C 専用 DynamoDB | — |
| `modules/suggestion/` | — | **新規**（GoroPay_Suggestion、本 Plan の主対象） |
| `modules/api_gateway/routes.tf` | Unit C 2 ルート追記済み | `GET /api/suggest` 1 本追記 |
| `modules/lambda_api/` | order_history + bedrock policy 付与済み | suggestion policy + DDB_TABLE_SUGGESTION env 追記 |

## 3. 生成する Artifacts（Plan 承認 + 回答後）

1. `aidlc-docs/construction/suggest/infrastructure-design/infrastructure-design.md`
2. `aidlc-docs/construction/suggest/infrastructure-design/deployment-architecture.md`

## 4. 作業手順（Checkboxes）

- [x] §5 の確認質問（Q-DI1〜Q-DI6）にユーザが回答（全 A）
- [x] 曖昧さ・矛盾を点検（Unit C module 規約整合、矛盾なし）
- [x] ユーザによる Plan 承認（2026-05-25「ok」）
- [x] `infrastructure-design.md` 生成
- [x] `deployment-architecture.md` 生成
- [ ] 完了メッセージ（2-option ゲート）
- [ ] aidlc-state.md / audit.md 更新、push + PR

---

## 5. 確認質問（Q-DI1 〜 Q-DI6）

> 1 問ずつ対話形式で提示。**(推奨)** は調査に基づく既定案。「全部推奨で」で一括採用可。

#### Q-DI1 — 新規 `modules/suggestion/`（DynamoDB）
`GoroPay_Suggestion`（PK `suggestionId` / TTL `expiresAt` 30分 / 1RCU・1WCU / GSI なし、NFRD-D08）を Unit C `modules/order_history/` と同型の module として新規作成するか。

- A) **新規 `modules/suggestion/`**（terraform-module-design 準拠: main.tf / locals.tf / variables.tf / outputs.tf + tests/）。table + `dynamodb_policy_arn` を output **(推奨)**
- B) 既存 module に相乗り（別案、指定）
- C) Other

[Answer]: **A**（新規 `modules/suggestion/`、terraform-module-design 準拠。GoroPay_Suggestion: PK suggestionId / TTL expiresAt 30分 / 1RCU・1WCU / GSI なし。table + dynamodb_policy_arn を output）

#### Q-DI2 — Bedrock IAM の扱い
Unit D も Bedrock を呼ぶが、`modules/bedrock/` は Unit C/D 共有候補として作成済みで、bedrock_policy_arn は lambda_api に付与済み。

- A) **既存 `modules/bedrock/` を再利用、新規 Bedrock IAM は作らない**（共有 API Lambda に既に付与済み。InferSuggestion も同 policy で動く） **(推奨)**
- B) suggest 用に Bedrock IAM を別途定義
- C) Other

[Answer]: **A**（既存 `modules/bedrock/` を再利用、新規 Bedrock IAM なし。共有 API Lambda に bedrock_policy_arn 付与済みのため InferSuggestion も同 policy で動作）

#### Q-DI3 — API Gateway ルート
`GET /api/suggest`（Cognito Authorizer 必須）を `modules/api_gateway/routes.tf` に追記するか。

- A) **`modules/api_gateway/routes.tf` に `GET /api/suggest` を 1 本追記**（Unit C ルートと同型、Cognito Authorizer + lambda 統合） **(推奨)**
- B) 別構成（指定）
- C) Other

[Answer]: **A**（`modules/api_gateway/routes.tf` に `GET /api/suggest` を追記、Cognito Authorizer + Lambda 統合、Unit C ルートと同型）

#### Q-DI4 — lambda_api への追記
共有 API Lambda に suggestion テーブルへのアクセス権と env を追加するか。

- A) **`module.lambda_api` の `additional_policy_arns` に `module.suggestion.dynamodb_policy_arn` を追加、env に `DDB_TABLE_SUGGESTION` を追加**（凍結契約 §10 の変数名）。Lambda 設定（256MB/arm64/10s）は Unit C のまま変更なし **(推奨)**
- B) 別構成（指定）
- C) Other

[Answer]: **A**（lambda_api の additional_policy_arns に suggestion.dynamodb_policy_arn を追加、env に DDB_TABLE_SUGGESTION 追加。Lambda 設定は変更なし）

#### Q-DI5 — 観測性アラーム（フォールバック率）
`modules/observability/` を再利用し、suggest のフォールバック率アラーム（`fallbackUsed=true` のメトリクスフィルタ + Alarm）を今追加するか。

- A) **今は追加せず、observability module の再利用余地だけ残す**（suggest はメイン機能でなく、過剰なアラームを避ける。必要時に追加）。構造化ログ（fallbackUsed）は出すので後から filter 追加可 **(推奨)**
- B) suggest フォールバック率アラームを今追加（Unit C NFRC-C13-3 と同方式）
- C) Other

[Answer]: **A**（今はアラーム追加せず observability module 再利用余地のみ残す。fallbackUsed はログに出し後付け可能）

#### Q-DI6 — Terraform テスト
`modules/suggestion/` の tftest（スキーマ assertion: PK suggestionId / TTL expiresAt / 1RCU1WCU）を Unit C と同様に追加するか。

- A) **`modules/suggestion/tests/dynamodb_schema.tftest.hcl` を追加**（mock_provider、PK/TTL/capacity を assert。Unit C order_history と同型） **(推奨)**
- B) tftest なし
- C) Other

[Answer]: **A**（`modules/suggestion/tests/dynamodb_schema.tftest.hcl` を追加、mock_provider で PK/TTL/capacity を assert。Unit C order_history と同型）

---

## 6. 矛盾チェック観点

1. `GoroPay_Suggestion` のキー/TTL が凍結契約 §5.3 / NFRD-D08 と一致するか
2. Bedrock IAM 再利用が NFRD-D16 / Unit C `modules/bedrock/` と矛盾しないか
3. env 変数名が凍結契約 §10（`DDB_TABLE_SUGGESTION`）と一致するか
4. module 分割が terraform-module-design 規約・Unit C 構造と整合するか

## 7. 完了条件

- Q-DI1〜Q-DI6 回答 + 矛盾なし確認 + Plan 承認
- `infrastructure-design.md` / `deployment-architecture.md` 生成
- 2-option 完了ゲート確認

## 8. 次ステージ（Code Generation）への引き継ぎ

`modules/suggestion/` の Terraform 実体、`GET /api/suggest` ルート、lambda_api 追記、`SuggestService` 系 Go 実装、`useSuggestion`/`SuggestBubble` 実装を Code Generation に引き継ぐ。
