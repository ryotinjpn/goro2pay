# `modules/observability`

横串 Observability: SNS Topic + CloudWatch Logs Metric Filter ×3 + Alarms ×3 + AWS Budgets。Unit C / D / E で共有再利用可能。

## リソース

- `aws_sns_topic.alarms` (`gp-{env}-alarms`) + email subscription
- Metric Filter ×3:
  - `place_order_latency` → `PlaceOrderLatencyMs` (NFRC-C13-1)
  - `bedrock_retry`        → `BedrockRetryCount` (NFRC-C13-2)
  - `fallback_triggered`   → `FallbackTriggeredCount` (NFRC-C13-3)
- Alarms ×3 (上記 metric に対応)
- AWS Budgets `gp-{env}-bedrock-budget` ($5/月、80% + 100% 通知)

## Outputs

| Output | 用途 |
|---|---|
| `sns_topic_arn` | Unit D/E 追加時に同 Topic を再利用 |

## 仕様根拠

- NFRC-C13 (3 アラーム閾値: p95 3000ms / リトライ 5/5min / フォールバック 3/5min)
- NFRC-C14 (CloudWatch Logs メトリクスフィルタのみ、PutMetricData は不実装)
- NFRC-C20 / Q-I7 = A (Bedrock 月 $5、80% + 100% 通知)
- Q-I4 = C (しきい値を `variables.tf` で外部化、環境別 tfvars で上書き可能)
- Q-I5 = C (横串 SNS Topic、Unit D/E 再利用)
- Q-I6 = A (`var.alarm_email` + email 購読、確認メール承認は手動)
