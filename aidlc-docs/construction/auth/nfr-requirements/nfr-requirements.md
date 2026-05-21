# Auth Unit — NFR Requirements

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: A (`auth`)
**Construction Depth**: Standard
**Stage**: NFR Requirements / Construction
**Predecessors**: Functional Design 完了 (PR #61, #66)

本ドキュメントは Unit A（認証）の **非機能要件の数値・しきい値・運用方針** を確定する。実装手段の詳細は NFR Design ステージ、AWS リソースパラメータの確定は Infrastructure Design ステージで扱う。

参照: [auth-nfr-requirements-plan.md](../../plans/auth-nfr-requirements-plan.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md), [requirements.md](../../../inception/requirements/requirements.md)

---

## 1. 性能要件 (Performance)

### A-NFR-PERF-01: ログイン応答時間 (Q-N2)
| 指標 | 目標値 | 測定対象 |
|---|---|---|
| P50 | **500ms 以下** | Frontend で `useAuth().login()` 呼出 → AuthSession 確立までの体感レイテンシ |
| P95 | **1500ms 以下** | 同上 |

- 測定範囲: ボタン押下 → `signIn` 完了 → `router.push("/")` 実行直前
- 測定環境: 本番デプロイ後、東京リージョン同一 AZ からの計測（Synthetic 計測手段は別途検討）

### A-NFR-PERF-02: Signup 応答時間
| 指標 | 目標値 |
|---|---|
| P50 | 1000ms 以下 |
| P95 | 2500ms 以下 |

- Signup は内部で SignUp API + InitiateAuth の 2 リクエストを直列で行うため Login より緩めの目標

### A-NFR-PERF-03: 認証付き API 呼出のオーバーヘッド (Q-N3)
**個別 SLA は設けない。** NFR-PERF-01（注文完了 3 秒以内）の全体予算内で Auth 関連処理 (Authorizer + middleware) が支障とならないことを E2E で確認する。

理由: OrderService の Bedrock 推論で 2-3 秒を消費する想定。Auth が個別 SLA を持つよりも E2E の合計を見るほうが意味がある。

### A-NFR-PERF-04: フォーム入力の応答性
- ユーザの input イベントから UI 反映まで **100ms 以内**（パスワード強度チェック・メール形式チェック等）

---

## 2. 可用性要件 (Availability)

### A-NFR-AVAIL-01: SLA 方針 (Q-N5)
- **AWS Cognito の標準 SLA に追従** する。Unit A 側に独立した可用性目標は設定しない（NFR-AVAIL-01, NFR-AVAIL-02 と整合）。
- 障害時はユーザに `AuthError.NETWORK_ERROR`（ダメ化トーン軽: 「通信が…ちょっと待ってもう一度」）を表示するのみ。
- 専用障害告知ページ・サーバ側追加対応・自動リトライは実装しない。

### A-NFR-AVAIL-02: 多重ログインの扱い
- 同一ユーザの複数デバイスからの同時ログインを許容する（Cognito デフォルト挙動）。
- セッション数の上限は設けない。MVP の Logout は `GlobalSignOut` を呼ぶため、明示的にログアウトすると全デバイスのセッションが切れる。

---

## 3. スケーラビリティ要件 (Scalability)

### A-NFR-SCALE-01: 想定同時利用者数 (Q-N4)
| 指標 | 値 |
|---|---|
| 同時利用者数 | **5 人**（ハッカソン審査デモ + 友人レビュー想定） |
| Auth 系 API ピーク | **10 req/s** |
| Cognito API quota への余裕 | InitiateAuth デフォルト 80 req/s に対して 8 倍以上の余裕 |

### A-NFR-SCALE-02: スケール戦略
- Cognito User Pool / API Gateway / Lambda はマネージドスケーリングに依拠（NFR-SCALE-01）。
- プロビジョンドキャパシティ・予約済み容量等の明示的なスケール設計は本 MVP では行わない（NFR-SCALE-02）。

---

## 4. セキュリティ要件 (Security)

### A-NFR-SEC-01: 認証方式 (NFR-SEC-01 を Unit A に射影)
- Amazon Cognito User Pool（メールアドレス + パスワード）。
- API Gateway Cognito Authorizer による JWT 必須検証。詳細は [business-rules.md §6](../functional-design/business-rules.md)。

### A-NFR-SEC-02: パスワードポリシー (Q-A2 + Q-N1 確定値)
| 項目 | 値 |
|---|---|
| 最小長 | 8 文字 |
| 数字 | 必須 |
| 英大文字 | 必須 |
| 英小文字 | 必須 |
| 記号 | 任意 |
| 最大長 | 256 文字 |

### A-NFR-SEC-03: Token Validity (Q-N1)
| Token | 有効期限 |
|---|---|
| **IdToken** | **8 時間** |
| **AccessToken** | **8 時間** |
| **RefreshToken** | **30 日** |

- R-Login-5（セッション 1 時間以上維持）の超過達成。
- 漏洩時の被害最小化と利便性のバランスをダメ化UX 寄りに調整。

### A-NFR-SEC-04: API Gateway レート制限 (Q-N6)
| 設定箇所 | 値 |
|---|---|
| API Gateway Stage Throttling | **Rate: 100 req/s, Burst: 200** |

- Stage 全体に一括適用。IP 単位の細かい制限・WAF・Cognito Advanced Security Features は本 MVP では導入しない。
- 同時利用者数 5 人想定（A-NFR-SCALE-01）に対して十分な余裕。
- 本番化時の WAF / Cognito Advanced Security 導入は A-NFR-SEC-09 に明記。

### A-NFR-SEC-05: 暗号化 (NFR-SEC-02 を Unit A に射影)
- Cognito User Pool / DynamoDB（Unit B/C 所有）/ CloudWatch Logs はすべて AWS マネージドのデフォルト暗号化に依拠。
- KMS 顧客管理キー (CMK) は使用しない。

### A-NFR-SEC-06: 個人情報の取扱 (NFR-COMP-02 整合)
- Cognito User Pool に保持する属性は **メールアドレスのみ**。
- アプリ側 DynamoDB には `userId` (Cognito `sub` UUID v4) のみ保存し、メールアドレスは保持しない。

### A-NFR-SEC-07: パスワードの取扱
- Functional Design [business-rules.md §2 R-Pwd-3](../functional-design/business-rules.md) に従う:
  - localStorage / sessionStorage / Cookie に保存しない
  - ログ・Telemetry・例外メッセージに出力しない
  - 送信完了次第メモリから速やかに参照を断つ

### A-NFR-SEC-08: メールアドレスの取扱
- Frontend での `toLowerCase()` 正規化（R-Email-3）。
- Backend ログには **`email_hash` (SHA256)** のみを出力（A-NFR-OBS-01）。平文メールはログに残さない。

### A-NFR-SEC-09: 本番化時の追加考慮（明記のみ）
本 MVP では実装しないが、本番運用時には以下を再評価する旨を明記する（NFR-SEC-04 と整合）:
- AWS WAF による IP レピュテーション・レート制限
- CloudTrail Data Events による監査ログ
- Cognito Advanced Security Features（適応的脅威検知）
- HTTPS 強制（API Gateway / Amplify Hosting で標準有効、設定確認のみ）
- パスワードリセット機能の追加
- MFA の導入（現状 OFF）

---

## 5. 信頼性要件 (Reliability)

### A-NFR-REL-01: Pre Sign-up Lambda Trigger 失敗時 (Q-N9)
- **Cognito 標準挙動に追従**: Trigger が失敗した場合 SignUp 全体が失敗。
- ユーザには `EMAIL_ALREADY_EXISTS` 等の Cognito 由来エラー（または `UNKNOWN`）をダメ化トーン軽でマッピングして表示。
- Trigger 内例外は Lambda ランタイムが自動的に CloudWatch Logs に記録。明示的な構造化ログ追加・DLQ 経路は本 MVP では設けない。

### A-NFR-REL-02: 不正な claims のフォールバック (Functional Design R-JWT-4 と整合)
- middleware が `claims["sub"]` 取得に失敗した場合、Lambda は `500 INTERNAL_ERROR` を返す。
- 構造化ログに `level=ERROR`, `event=auth_claims_missing` を出力。

### A-NFR-REL-03: Token 自動更新失敗時 (Functional Design R-Session-1/2 と整合)
- Amplify Auth Hub `tokenRefresh_failure` 検出 → ダメ化風モーダル表示 → 1.5 秒後 `/login?from=session_expired`。
- バックエンド側は 401 を返すのみ、独自リトライ実装はしない。

---

## 6. 観測性要件 (Observability)

### A-NFR-OBS-01: 構造化ログ項目 (Q-N7)
Auth Unit の Lambda（API Lambda の auth handler 経路 + Pre Sign-up Lambda Trigger）が出力する JSON ログのフィールド:

| フィールド | 型 | 説明 |
|---|---|---|
| `level` | string | `info` / `warn` / `error` |
| `timestamp` | string | ISO 8601 |
| `userId` | string \| null | Cognito `sub`、認証前のリクエストでは null |
| `action` | string | `signup` / `login` / `logout` / `token_refresh` 等 |
| `traceId` | string \| null | X-Ray Trace ID（API Gateway / Lambda が自動付与する場合のみ） |
| `requestId` | string | API Gateway Request ID または Lambda Request ID |
| `email_hash` | string \| null | SHA256(`normalize(email)`)、認証前で email がリクエストに含まれる場合のみ |
| `userAgent` | string \| null | リクエストヘッダから取得 |

NFR-OBS-01「構造化ログ推奨」と整合。`email` 平文の出力は禁止（A-NFR-SEC-08）。

### A-NFR-OBS-02: メトリクス収集 (Q-N8)
- **CloudWatch Custom Metrics は本 MVP では収集しない**（NFR-OBS-02 厳密準拠）。
- Cognito 標準メトリクス（SignInSuccesses / SignInFailures 等）は AWS Console から確認可能だが、本 MVP では Dashboard 構築も省略。
- 本番化時に EMF + Custom Metrics + Dashboard 構築を再評価する旨を明記。

### A-NFR-OBS-03: トレーシング
- X-Ray は本 MVP では有効化しない（NFR-OBS-02）。
- API Gateway / Lambda が自動で生成する Request ID をログに含めることでトラブルシュート可能性を最小限担保。

---

## 7. テスタビリティ要件 (Testability)

### A-NFR-TEST-01: 単体テスト
- Functional Design `frontend-components.md §13` のテスト計画に基づき、Go パッケージ・TypeScript フックを Example-Based でカバー。
- 目標カバレッジ: 主要パス 80% 以上（厳密な数値目標は設けず、レビュー時に判断）。

### A-NFR-TEST-02: Property-Based Testing (Q-N10)
**適用範囲**: メール正規化と email_hash 生成にのみ PBT を適用する。

| 関数 | プロパティ |
|---|---|
| `normalize(email)` | べき等性: `normalize(normalize(x)) == normalize(x)` |
| `normalize(email)` | 大文字小文字不問: `normalize(toUpper(x)) == normalize(x)` |
| `emailHash(email)` | 長さ不変: `len(emailHash(x)) == 64`（SHA256 hex） |
| `emailHash(email)` | 正規化整合: `emailHash(x) == emailHash(normalize(x))` |
| `emailHash(email)` | 衝突なし: `x != y` のとき `emailHash(x) != emailHash(y)`（高確率） |

PBT Extension 設定 (Partial: 純粋関数とシリアライゼーションのみ) と整合。NFR-TEST-04（IO は通常テスト）に従い、Cognito 連携・middleware は Example-Based のままとする。

### A-NFR-TEST-03: 統合テスト
- Cognito ローカルモック（または LocalStack の Cognito モック）を使った Signup / Login の統合テスト。
- API Gateway + Lambda + Cognito Authorizer の連携は本 MVP では E2E のみで担保。

### A-NFR-TEST-04: E2E テスト
- Functional Design `frontend-components.md §13` の E2E テスト計画に従う:
  - Signup → Login → Logout のハッピーパス
  - US-0-01 受入基準（総タップ数 5 回以下）
  - 認証失敗 → エラー表示 → 再試行のリカバリパス

---

## 8. アクセシビリティ要件 (Accessibility)

### A-NFR-A11Y-01: ブラウザ対応 (Q-N11)
| ブラウザ | 対応バージョン |
|---|---|
| Chrome | 最新 2 バージョン |
| Edge | 最新 2 バージョン |
| Safari (macOS) | 最新 2 バージョン |
| iOS Safari / Android Chrome | 明示サポートしない（動作した場合のベストエフォート） |
| IE / 古い Safari | 対象外 |

NFR-A11Y-01「基本的なセマンティック HTML」と整合し、最小スコープに絞る。

### A-NFR-A11Y-02: HTML 構造
Functional Design `frontend-components.md §5.6 / §6.7` に従う:
- `<form>`, `<label htmlFor>`, `<input type="email">`, `<input type="password" autoComplete>`
- エラーメッセージは `aria-live="polite"`
- WCAG 準拠チェック（NFR-A11Y-02）は本 MVP では実施しない。

---

## 9. 国際化・ロケール要件 (Internationalization)

### A-NFR-I18N-01: 言語 (Q-N12)
- **日本語のみ**（requirements.md Q15 と整合）。
- Cognito からの英語エラーレスポンスは Frontend の `authMessages.ts` (Functional Design §11) で日本語ダメ化トーン軽メッセージにマッピングする。
- Cognito Custom Message Lambda Trigger は導入しない（auto-confirm 採用でメール送信が発生しないため）。

### A-NFR-I18N-02: タイムゾーン
- 表示は JST 固定。Token の `iat` / `exp` 等の比較は UTC で行うが、UI では JST に変換して表示する場面があれば JST に揃える。

---

## 10. コンプライアンス要件 (Compliance)

### A-NFR-COMP-01: 既存コンプライアンス要件の引用 (Q-N13)
本 Unit A は以下の Inception フェーズ確定要件に従う:

| ID | 内容 | Unit A への影響 |
|---|---|---|
| NFR-COMP-01 | 仮想ウォレット採用、PCI DSS 直接対象外 | Unit A は決済情報を扱わないため影響軽微 |
| NFR-COMP-02 | 個人情報はメールアドレスのみ | Cognito 属性スキーマを email 単独に絞る |
| NFR-COMP-03 | 本番運用時の個人情報保護法・金融庁ガイドライン適合性評価が必要 | A-NFR-SEC-09 と合わせて本番化時のレビュー必要性を明記 |

### A-NFR-COMP-02: 本 MVP での対応範囲
- 本 MVP は本番運用を想定しない（requirements.md Assumptions と整合）。
- ユーザ削除手順・GDPR 削除権対応・データポータビリティ等は本 MVP では文書化しない。
- 本番化時に再評価する旨を引用するのみ。

---

## 11. 保守性要件 (Maintainability)

### A-NFR-MAINT-01: コーディング規約準拠
- Backend (Go): プロジェクト共通の Go 規約に準拠。
- Frontend (TypeScript): プロジェクト共通の TypeScript / React 規約に準拠。
- Terraform: terraform-coding-rule / terraform-module-design プラグイン規約に準拠（NFR-MAINT-02, NFR-MAINT-03）。

### A-NFR-MAINT-02: 設定値の外部化
[unit-interfaces.md §10](../../interfaces/unit-interfaces.md) の環境変数規約に従う:
- `AWS_REGION` (固定: `ap-northeast-1`)
- `COGNITO_USER_POOL_ID` (Infrastructure Design で確定)
- `COGNITO_APP_CLIENT_ID` (同上)
- `LOG_LEVEL` (`info` / `debug`)

### A-NFR-MAINT-03: 凍結 Interface 整合
- [unit-interfaces.md §2](../../interfaces/unit-interfaces.md) (Unit A) の公開 Go API（`AttachUserID`, `UserIDFromContext`, `ErrUnauthorized`）と §9 の Frontend Hook (`useAuth().signup` / `login` / `logout`) を変更しない。

---

## 12. NFR トレーサビリティマトリクス

Inception の NFR / Unit A の A-NFR の対応関係:

| Inception NFR ID | A-NFR ID（Unit A 射影） |
|---|---|
| NFR-DEG-01（低摩擦） | A-NFR-SEC-03 (8h Token), A-NFR-PERF-04 |
| NFR-DEG-05（コピー） | Functional Design `business-rules.md §9` で実装、本書 §9 で言語固定 |
| NFR-PERF-01 | A-NFR-PERF-03 (E2E 内で確認) |
| NFR-PERF-02 | A-NFR-PERF-01, A-NFR-PERF-02 |
| NFR-PERF-03 | A-NFR-SCALE-01 |
| NFR-AVAIL-01 | A-NFR-AVAIL-01 |
| NFR-AVAIL-02 | A-NFR-AVAIL-01 |
| NFR-SEC-01 | A-NFR-SEC-01 |
| NFR-SEC-02 | A-NFR-SEC-05 |
| NFR-SEC-03 | A-NFR-SEC-04 |
| NFR-SEC-04 | A-NFR-SEC-09 |
| NFR-COMP-01 | A-NFR-COMP-01 |
| NFR-COMP-02 | A-NFR-SEC-06, A-NFR-COMP-01 |
| NFR-COMP-03 | A-NFR-COMP-01, A-NFR-SEC-09 |
| NFR-SCALE-01 | A-NFR-SCALE-02 |
| NFR-SCALE-02 | A-NFR-SCALE-02 |
| NFR-MAINT-01 | （Adapter 層は Unit C 主、Unit A は対象外） |
| NFR-MAINT-02 | A-NFR-MAINT-01 |
| NFR-MAINT-03 | A-NFR-MAINT-01 |
| NFR-TEST-01 | A-NFR-TEST-02 |
| NFR-TEST-02 | A-NFR-TEST-02（PBT 対象が Unit B/C 主、A はメール正規化のみ） |
| NFR-TEST-03 | （Unit B/C の不変条件、Unit A は範囲外） |
| NFR-TEST-04 | A-NFR-TEST-01, A-NFR-TEST-03 |
| NFR-OBS-01 | A-NFR-OBS-01 |
| NFR-OBS-02 | A-NFR-OBS-02, A-NFR-OBS-03 |
| NFR-A11Y-01 | A-NFR-A11Y-02 |
| NFR-A11Y-02 | A-NFR-A11Y-02 |

---

## 13. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | Token Validity (8h/30d) を Cognito App Client にどう設定するか、Stage Throttling の Terraform 実装、ダメ化風 Modal の表示遅延 1.5 秒の実装方法、PBT セットアップ |
| **Infrastructure Design** | Cognito User Pool / App Client / Pre Sign-up Lambda Trigger の Terraform 構成、API Gateway Stage Throttling のパラメータ |
| **Code Generation** | `email_hash` 生成ロジック、ログ出力ライブラリ選定、PBT テストコード、`authMessages.ts` の Cognito エラー → 日本語マッピング |
