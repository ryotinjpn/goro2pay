# Unit E (`metrics`) — Infrastructure Design

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Infrastructure Design
**Unit**: E — `metrics`（ダメ化メトリクス）
**Depth**: Standard
**Related**:
- Plan: [metrics-infrastructure-design-plan.md](../../plans/metrics-infrastructure-design-plan.md)
- NFR Design: [nfr-design-patterns.md](../nfr-design/nfr-design-patterns.md) / [logical-components.md](../nfr-design/logical-components.md)
- 凍結契約: [unit-interfaces.md](../../interfaces/unit-interfaces.md) §6
- 継承元: [order/infrastructure-design/](../../order/infrastructure-design/) / [suggest/infrastructure-design/](../../suggest/infrastructure-design/)

**方針**: Q-I1〜Q-I3 全 A。Unit E は DynamoDB テーブル新規なし・IAM/env 追加なし。`modules/api_gateway/routes.tf` への 3 route 追記 + 既存 tftest への assert 追記のみ。

---

## 1. 論理コンポーネント → インフラ マッピング

| LC | インフラリソース | module | 変更種別 |
|---|---|---|---|
| LC-ME-01 MetricsService | 共有 API Lambda 上の Go コード | `modules/lambda_api/`（既存） | **コード追加のみ** |
| LC-ME-02 MetricsHandler | `GET /api/metrics` route | `modules/api_gateway/routes.tf`（既存） | **追記** |
| LC-ME-03 BudgetRaiseService | 共有 API Lambda 上の Go コード | `modules/lambda_api/`（既存） | **コード追加のみ** |
| LC-ME-04 BudgetRaiseHandler | `GET /api/budget/raise/recommendation` + `POST /api/budget/raise` route | `modules/api_gateway/routes.tf`（既存） | **追記** |
| LC-ME-05〜06 DTO / 純関数 | Go コード | Lambda 内部 | — |
| LC-ME-07〜11 Frontend | Amplify Hosting（既存） | `modules/amplify/`（変更なし） | **変更なし** |

**新規 Terraform module**: なし（Q-I1=A）

---

## 2. modules/api_gateway/routes.tf への追記（Q-I2=A）

3 エンドポイントすべて `authorization_type = "JWT"` で追加する。

```hcl
# ----------------------------------------------------------------------------
# Unit E: GET /api/metrics / GET /api/budget/raise/recommendation /
#         POST /api/budget/raise (凍結契約 §6.2)
# integration は既存 api_lambda を再利用 (single Lambda)。
# ----------------------------------------------------------------------------

resource "aws_apigatewayv2_route" "get_metrics" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/metrics"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "get_budget_raise_recommendation" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/budget/raise/recommendation"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "post_budget_raise" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/budget/raise"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}
```

---

## 3. 環境変数・IAM（既存カバー済み、追記不要）

### 3.1 環境変数（lambda_api module）

Unit E が必要とする環境変数はすべて既存 `modules/lambda_api/api_lambda.tf` に注入済み。

| 環境変数名 | 参照テーブル | 注入元 |
|---|---|---|
| `DDB_TABLE_WALLET` | GoroPay_Wallet | `var.wallet_table_name`（Unit B 実装済み） |
| `DDB_TABLE_BUDGET_SETTINGS` | GoroPay_BudgetSettings | `var.budget_settings_table_name`（Unit B 実装済み） |
| `ORDER_HISTORY_TABLE_NAME` | GoroPay_OrderHistory | `var.order_history_table_name`（Unit C 実装済み） |

**追記不要**。

### 3.2 IAM ポリシー

Unit E が必要とするすべての DynamoDB アクセス権限は既存 policy でカバー済み。

| アクセス | Action | 付与元 module |
|---|---|---|
| Wallet.GetItem | `dynamodb:GetItem` | `modules/budget/iam.tf`（Unit B） |
| BudgetSettings.GetItem | `dynamodb:GetItem` | `modules/budget/iam.tf`（Unit B） |
| BudgetSettings.PutItem / UpdateItem（BudgetSettingsWriter） | `dynamodb:PutItem`, `dynamodb:UpdateItem` | `modules/budget/iam.tf`（Unit B） |
| OrderHistory.Query（CountThisMonth） | `dynamodb:Query` | `modules/order_history/`（Unit C） |

**追記不要**。Unit B の `BudgetSettingsReadWrite` Sid に PutItem/UpdateItem が含まれるため、`BudgetRaiseService.Accept` の書き込みも権限内。

---

## 4. Terraform テスト（Q-I3=A）

`modules/api_gateway/tests/api_gateway_basic.tftest.hcl` の `resources_present` run に Unit E の 3 route assert を追記する。

```hcl
# 既存 resources_present run 末尾に追記
assert {
  condition     = aws_apigatewayv2_route.get_metrics.route_key == "GET /api/metrics"
  error_message = "GET /api/metrics route が存在すること"
}
assert {
  condition     = aws_apigatewayv2_route.get_metrics.authorization_type == "JWT"
  error_message = "GET /api/metrics は JWT 認証必須"
}
assert {
  condition     = aws_apigatewayv2_route.get_budget_raise_recommendation.route_key == "GET /api/budget/raise/recommendation"
  error_message = "GET /api/budget/raise/recommendation route が存在すること"
}
assert {
  condition     = aws_apigatewayv2_route.post_budget_raise.route_key == "POST /api/budget/raise"
  error_message = "POST /api/budget/raise route が存在すること"
}
```

---

## 5. 変更ファイル一覧

| ファイル | 変更種別 | 内容 |
|---|---|---|
| `infra/modules/api_gateway/routes.tf` | 追記 | 3 route リソース追加 |
| `infra/modules/api_gateway/tests/api_gateway_basic.tftest.hcl` | 追記 | Unit E route の assert 4 件追加 |

**変更なし**:
- `infra/modules/lambda_api/` — env / IAM 追加なし
- `infra/modules/budget/` — 変更なし
- `infra/modules/order_history/` — 変更なし
- `infra/modules/amplify/` — 変更なし（Frontend は既存 Amplify Hosting で配信）
