# Unit C — Code Generation Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order` / 代行手配コア / MVP の心臓)
**Construction Depth**: Comprehensive

## 全成果物一覧

### Backend (apps/api/)

| カテゴリ | ファイル数 |
|---|---|
| Adapter (bedrock / delivery / fallback) | 14 (実装 + テスト + mock) |
| Repository (order_history) | 3 (実装 + テスト + inmemory) |
| Order 中核 (service / plan_builder / logger / event_logger / types / wallet) | 12 (実装 + 単体 + 統合 9 シナリオ + PBT) |
| 横串 (middleware / observability) | 4 (LatencyMiddleware + Measure ヘルパ + テスト) |
| Handler (order) | 2 (実装 + テスト) |
| main.go 配線 + wallet_stub.go | 2 (編集 + 新規) |
| ドキュメントサマリ | 4 (business-logic / api-layer / repository-layer / adapters) |

**合計**: 約 41 ファイル

### Frontend (web/)

| カテゴリ | ファイル数 |
|---|---|
| lib (ulid / toasts / errorMappers / api/orders) | 4 |
| state (toastAtoms) | 1 |
| hooks (useDisableLock / useToast / useOrder / useOrderHistory) | 4 |
| components (Toast / ToastHost / GoroButton / OrderHistoryList / Skeleton) | 5 |
| pages (layout 編集 / page 編集 / order/[id]/complete) | 3 |
| Vitest テスト | 6 |
| ドキュメントサマリ | 1 |

**合計**: 約 24 ファイル

### Infrastructure (infra/)

| カテゴリ | ファイル数 |
|---|---|
| `modules/order_history/` | 6 (5 .tf + 1 README + 1 tftest) |
| `modules/bedrock/` | 7 (6 .tf + 1 README + 1 tftest) |
| `modules/observability/` | 7 (6 .tf + 1 README + 1 tftest) |
| 既存 module 編集 (api_gateway / lambda_api) | 5 (routes.tf / variables.tf / iam.tf / api_lambda.tf / outputs.tf) |
| envs/dev 編集 (locals / main / outputs) | 3 |
| ドキュメントサマリ | 1 |

**合計**: 約 28 ファイル

### ドキュメント

| ファイル |
|---|
| `business-logic-summary.md` |
| `api-layer-summary.md` |
| `repository-layer-summary.md` |
| `adapters-summary.md` |
| `frontend-summary.md` |
| `infrastructure-summary.md` |
| `deployment-runbook.md` |
| `code-generation-summary.md` (本ファイル) |

**全合計**: 約 **101 ファイル** (Backend 41 + Frontend 24 + Infra 28 + Docs 8)

## ステップ別チェック結果

| Step | 内容 | 状態 |
|---|---|---|
| Step 1 | 環境準備 (go.mod / package.json / ディレクトリ) | ✅ |
| Step 2 | Backend Adapter 層 (Bedrock / Delivery / Fallback) | ✅ |
| Step 3 | Backend Repository 層 (OrderHistory) | ✅ |
| Step 4 | Backend Order 中核 (Service / PlanBuilder / LogSummary / EventLogger / PBT) | ✅ |
| Step 5 | Backend 観測性 (LatencyMiddleware / Measure ヘルパ) | ✅ |
| Step 6 | Backend Handler + main.go 配線 | ✅ |
| Step 7 | Backend サマリ 4 種 | ✅ |
| Step 8 | Frontend lib + state | ✅ |
| Step 9 | Frontend hooks | ✅ |
| Step 10 | Frontend components | ✅ |
| Step 11 | Frontend pages 配置 | ✅ |
| Step 12 | Frontend テスト | ✅ |
| Step 13 | Frontend サマリ | ✅ |
| Step 14 | Infrastructure 新規 module 3 種 | ✅ |
| Step 15 | Infrastructure 既存 module 追記 | ✅ |
| Step 16 | Infrastructure envs/dev 追記 | ✅ |
| Step 17 | Infrastructure サマリ | ✅ |
| Step 18 | Documentation (deployment-runbook + 各 summary) | ✅ |
| Step 19 | 完了確認 (本サマリ + state/audit 更新) | ✅ |

## テスト結果

### Backend (Go)

```
ok  internal/adapters/bedrock
ok  internal/adapters/delivery
ok  internal/adapters/fallback
ok  internal/auth
ok  internal/handlers
ok  internal/logging
ok  internal/middleware
ok  internal/observability
ok  internal/order               (PBT P-1 + P-3 各 100 サンプル + 統合 9 シナリオ)
ok  internal/repo/order_history
```

### Frontend (Vitest)

```
Test Files  12 passed (12)
     Tests  44 passed (44)
```

Unit C 新規 26 テスト + Unit A 既存 18 テスト全てパス。

### Infrastructure (terraform-test)

```
modules/order_history    : 4 passed
modules/bedrock          : 3 passed
modules/observability    : 5 passed
```

**合計 12 tftest 全パス**、`mock_provider` でオフライン実行。

## NFR 達成根拠 (Comprehensive 深度)

| NFR | 実装 |
|---|---|
| NFRC-C01 (E2E p95 3.0s / p99 5.0s) | LatencyMiddleware + observability.Measure (P-OBS-01) |
| NFRC-C05 (冪等性 TTL 24h) | Wallet payload から OrderID 復元 (BR-C15)、PBT P-3 で検証 |
| NFRC-C06 (Bedrock リトライ 1 回) | adapter.go の maxAttempts=2 + RetryClassifier (P-RETRY-01) |
| NFRC-C07 (Bedrock タイムアウト 1.5s) | adapter.go の bedrockTimeoutPerCall = 1500 ms |
| NFRC-C08 (フォールバック閾値 5 件) | plan_builder.go の fallbackThreshold = 5 (P-PLAN-01) |
| NFRC-C09 (Insert 失敗時 200) | service.go の `_ = ierr` 握りつぶし + observability.Measure 経由のログ記録 |
| NFRC-C10 (Context Cancel 即時伝播) | adapter / plan_builder / service 全層で ctx.Err() 確認 |
| NFRC-C12 (構造化ログ 19 項目) | LogSummary 11 項目 + ContextAwareSlogHandler 8 項目 (P-OBS-02) |
| NFRC-C13 (CloudWatch アラーム 3 種) | modules/observability の metric filter ×3 + alarm ×3 |
| NFRC-C15 (PBT P-1 + P-3) | service_pbt_test.go (gopter、各 100 サンプル) |
| NFRC-C16 (Bedrock スタブマトリクス) | MockBedrockAdapter / FakeFallbackProvider / FakeDeliveryAdapter |
| NFRC-C17 (統合テスト 9 シナリオ) | service_test.go の 9 ケース |
| NFRC-C18 (Lambda 256MB / arm64 / 600ms cold) | init() で SDK 初期化 (P-INIT-01)、Unit A の Lambda 設定継承 |
| NFRC-C19 (TanStack Query 60s + invalidate) | useOrderHistory.ts (placeholderData: keepPreviousData) |
| NFRC-C20 (Bedrock 月 $5 アラート) | modules/observability の Budgets 80% + 100% |
| NFRC-C22 (エラー UX、自虐トースト 3 種) | mapOrderError + getRandomToast + useDisableLock |
| NFRC-C24 (Bedrock 本文ログ非記録) | LogSummary に Bedrock 本文フィールドなし (構造的防御) |

## 後続ステージへの引き継ぎ

1. **Unit B 連携**: `apps/api/wallet_stub.go` の `noopWalletService` を Unit B 実装に差し替え (main.go の DI 配線変更)
2. **Build & Test**: E2E (Playwright)、CI ワークフロー (go test / npm test / terraform test の自動化)
3. **本番化**:
   - DynamoDB PITR 有効化検討 (Q-I8 再検討)
   - Bedrock Provisioned Throughput 検討 (Q-I3 = C 昇格)
   - CloudWatch Dashboard 構築 (NFRC-C14 を超えた可視化)
4. **Unit D 連携**: 同じ `BedrockAdapter` / `FallbackSuggestProvider` / `modules/bedrock` を再利用

## 文書管理

- **承認**: ユーザ承認待ち (Construction フェーズ最終ゲート)
- **凍結契約への影響**: なし
- **次ステージ**: Build & Test (Construction Phase 最終ステージ、全 Unit 統合)
