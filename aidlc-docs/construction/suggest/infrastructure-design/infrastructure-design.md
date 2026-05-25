# Unit D (`suggest`) — Infrastructure Design

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Infrastructure Design
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Related**: Plan [suggest-infrastructure-design-plan.md](../../plans/suggest-infrastructure-design-plan.md)、[suggest/nfr-design/logical-components.md](../nfr-design/logical-components.md)（LC-SUGGEST-*）、継承元 [order/infrastructure-design/](../../order/infrastructure-design/)、凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md) §5・§7・§10
**方針**: Q-DI1〜Q-DI6（全 A）。新規は `modules/suggestion/` のみ。Bedrock/observability は Unit C module を再利用。terraform-module-design / terraform-coding-rule / terraform-test 準拠。

---

## 1. 論理コンポーネント → インフラ マッピング

| LC | インフラリソース | module |
|---|---|---|
| LC-SUGGEST-04 SuggestionStore | DynamoDB `GoroPay_Suggestion` | **`modules/suggestion/`（新規）** |
| LC-SUGGEST-01/02/05/06 Backend | 共有 API Lambda 上の Go コード | `modules/lambda_api/`（既存、env/policy 追記） |
| LC-SUGGEST-05 SuggestionBuilder（Bedrock 呼出） | Bedrock 権限 | `modules/bedrock/`（既存・**再利用**） |
| LC-SUGGEST-02 SuggestHandler（GET /api/suggest） | API Gateway route | `modules/api_gateway/routes.tf`（既存、1 本追記） |
| LC-SUGGEST-10/11 Frontend | Amplify Hosting（既存） | `modules/amplify/`（変更なし、共有ビルド） |
| 観測性（SuggestLogSummary） | CloudWatch Logs（既存ロググループ） | `modules/observability/`（再利用、新規アラームなし Q-DI5） |

---

## 2. 新規 `modules/suggestion/`（Q-DI1=A）

terraform-module-design 準拠のファイル分割（`main.tf` / `locals.tf` / `variables.tf` / `outputs.tf` / `tests/`）。

### 2.1 DynamoDB テーブル定義（概念）

```hcl
# main.tf（概念。Code Generation で実体化）
resource "aws_dynamodb_table" "suggestion" {
  name         = local.table_name                # "GoroPay_Suggestion"（env prefix は locals）
  billing_mode = "PROVISIONED"                    # Q-DI1 / NFRD-D08
  read_capacity  = 1
  write_capacity = 1
  hash_key     = "suggestionId"                  # PK（凍結契約 §5.3）

  attribute {
    name = "suggestionId"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"                 # TTL 30分（BR-D08 / NFRD-D05）
    enabled        = true
  }

  point_in_time_recovery { enabled = false }      # MVP（Unit C order_history と同方針）
  server_side_encryption { enabled = true }       # マネージドデフォルト（NFR-SEC-02）

  tags = local.tags                               # Unit=suggest 個別タグ
}
```

- **GSI**: なし（suggestionId 単独 Get で完結、NFRD-D08）
- **属性**: PK `suggestionId` + TTL `expiresAt` のみキー定義。`userId` / `plan`(JSON) / `createdAt` は非キー属性（スキーマレス）

### 2.2 IAM ポリシー（DynamoDB アクセス）

```hcl
# suggestion テーブルへの最小権限ポリシー（lambda_api にアタッチ）
# 許可アクション: GetItem / PutItem（Save/Get のみ。Query/Scan 不要）
# Resource: aws_dynamodb_table.suggestion.arn
```

- 出力 `dynamodb_policy_arn` を lambda_api の `additional_policy_arns` に渡す（§4）

### 2.3 outputs

| output | 用途 |
|---|---|
| `dynamodb_table_name` | lambda_api の `DDB_TABLE_SUGGESTION` env |
| `dynamodb_policy_arn` | lambda_api の `additional_policy_arns` |

---

## 3. `modules/api_gateway/routes.tf` 追記（Q-DI3=A）

```hcl
# Unit D: GET /api/suggest (GetSuggestion)
resource "aws_apigatewayv2_route" "get_suggest" {
  api_id             = var.api_id
  route_key          = "GET /api/suggest"
  target             = "integrations/${aws_apigatewayv2_integration.lambda.id}"  # 既存 lambda 統合を共有
  authorization_type = "JWT"                                                     # Cognito Authorizer（Unit A）
  authorizer_id      = var.cognito_authorizer_id
}
```

- Unit C の `POST/GET /api/orders` と同型。既存 Lambda 統合・Authorizer を共有

---

## 4. `modules/lambda_api/` 追記（Q-DI4=A）

`envs/dev/main.tf` の `module "lambda_api"` 呼出を更新：

```hcl
module "lambda_api" {
  source = "../../modules/lambda_api"
  # ... 既存 ...
  additional_policy_arns = [
    module.order_history.dynamodb_policy_arn,
    module.bedrock.bedrock_policy_arn,
    module.suggestion.dynamodb_policy_arn,     # ★ Unit D 追加
  ]
  order_history_table_name = module.order_history.dynamodb_table_name
  suggestion_table_name    = module.suggestion.dynamodb_table_name  # ★ Unit D 追加 → env DDB_TABLE_SUGGESTION
}
```

- Lambda 設定（256MB / arm64 / 10s）は Unit C のまま **変更なし**（NFRD-D15）
- env `DDB_TABLE_SUGGESTION`（凍結契約 §10）を Lambda に注入

---

## 5. 再利用 module（変更なし）

| module | 再利用内容 |
|---|---|
| `modules/bedrock/`（Q-DI2=A） | bedrock_policy_arn は lambda_api に付与済み。InferSuggestion も同 policy で動作。**新規 Bedrock IAM なし** |
| `modules/observability/`（Q-DI5=A） | 既存ロググループ・X-Ray なし方針を踏襲。suggest フォールバック率アラームは **今は追加しない**（fallbackUsed ログは出力、後付け可） |
| `modules/amplify/` | Frontend は共有ビルド（useSuggestion/SuggestBubble は web/ に含まれ既存パイプラインでデプロイ） |
| `modules/cognito/` | 認証（Unit A、変更なし） |

---

## 6. `envs/dev/main.tf` 配線（追記）

```hcl
module "suggestion" {
  source = "../../modules/suggestion"
  # name_prefix / tags / env など共通変数
}
```

`outputs.tf` への追加は任意（テーブル名は lambda_api 経由で参照されるため必須ではない）。

---

## 7. Terraform テスト（Q-DI6=A）

`modules/suggestion/tests/dynamodb_schema.tftest.hcl`（mock_provider 利用）:

| assert | 内容 |
|---|---|
| PK | `hash_key == "suggestionId"` |
| TTL | `ttl.attribute_name == "expiresAt"` かつ `enabled == true` |
| capacity | `billing_mode == "PROVISIONED"` / `read_capacity == 1` / `write_capacity == 1` |
| 暗号化 | `server_side_encryption.enabled == true` |

Unit C `order_history` の tftest と同型。

---

## 8. NFR / 凍結契約 トレーサビリティ

| 項目 | 反映 |
|---|---|
| NFRD-D08（1RCU/1WCU） | §2.1 PROVISIONED 1/1 |
| NFRD-D05（TTL30分） | §2.1 ttl expiresAt |
| NFRD-D15（Lambda 256MB/arm64/10s） | §4 変更なし |
| NFRD-D16（Bedrock 再利用） | §5 modules/bedrock 再利用 |
| NFRD-D10（カスタムメトリクスなし） | §5 observability 新規アラームなし |
| 凍結契約 §5.3（GoroPay_Suggestion キー） | §2.1 |
| 凍結契約 §10（DDB_TABLE_SUGGESTION） | §4 env |

---

## 9. 文書管理

- **凍結契約への影響**: なし
- **次ステージ**: Code Generation。`modules/suggestion/` 実体 + lambda_api/routes 追記 + Go/Frontend 実装
