# Auth Unit — NFR Requirements Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-21
**Unit**: A (`auth` / 認証)
**Construction Depth**: Standard
**Stage**: NFR Requirements (Construction Phase)
**Prerequisite**: Functional Design 承認済み (PR #61, #66 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit A（認証）の **NFR Requirements ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。Plan 承認後、回答内容を反映した NFR 成果物を生成する。

### 1.1 Functional Design からの引き継ぎ事項

Functional Design ステージで「NFR Requirements で扱う」とされた論点:
- Token Validity（IdToken / AccessToken / RefreshToken）の数値
- レート制限のしきい値
- ログイン応答時間目標
- フォーム入力の応答時間、`useAuth` 状態確定時間
- パスワード入力フィールドの autoComplete セキュリティ考慮（→ NFR Design へ）

### 1.2 全体要件書 (`requirements.md`) で Unit A に関連する NFR

| NFR ID | 内容 | Unit A への影響 |
|---|---|---|
| NFR-DEG-01 | 1 代行手配あたり最大 2 タップ以内 | Signup/Login 後のメイン画面到達速度に影響 |
| NFR-PERF-01 | ボタン押下→完了 3 秒以内 (目標) | 認証付き API 呼出のオーバーヘッドが影響 |
| NFR-PERF-02 | メイン画面初期表示 5 秒以内 (目標) | 認証状態確定 + サジェスト取得の合計時間 |
| NFR-PERF-03 | 同時利用者数 数人〜数十人（デモ規模） | スループット要件は緩い |
| NFR-AVAIL-01 | AWS マネージド標準可用性に準拠 | Cognito SLA に依拠 |
| NFR-AVAIL-02 | 本番相当 SLA は定めない（デモ用途） | Auth Unit も同方針 |
| NFR-SEC-01 | 認証は Cognito + JWT 検証必須 | 本 Unit が直接担う |
| NFR-SEC-02 | AWS マネージドのデフォルト暗号化 | KMS CMK は使わない |
| NFR-SEC-03 | API Gateway レート制限を最小限設定（任意） | Unit A の API 全体に影響 |
| NFR-SEC-04 | 本番考慮事項（HTTPS / WAF / CloudTrail / PCI DSS）は将来対応として明記 | 設計書で言及のみ |
| NFR-COMP-02 | 個人情報はメールアドレスのみ取得 | Cognito User Pool の属性スキーマ |
| NFR-OBS-01 | Lambda は CloudWatch Logs 構造化出力 | Auth middleware のログ方針 |
| NFR-A11Y-01 | 基本的なセマンティック HTML | フォームコンポーネントの記述 |

### 1.3 NFR Requirements で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| パフォーマンス目標値（Login 応答時間、JWT 検証オーバーヘッド等） | 実装手段（リトライ・バックオフ係数 → NFR Design へ） |
| 可用性方針の Unit A 向け明文化 | Cognito SLA 数値そのもの（Cognito 仕様参照） |
| セキュリティ要件（Token Validity、Password Policy 数値、レート制限のしきい値） | IAM ロール / KMS 鍵設計（→ Infrastructure Design） |
| スケーラビリティ前提（同時利用者数の Auth 観点） | DynamoDB / Lambda の物理パラメータ |
| 観測性要件（ログ項目・メトリクス項目） | CloudWatch ダッシュボード設計（→ NFR Design / Infra Design） |
| 信頼性要件（Lambda Trigger 失敗時の挙動方針） | 具体的なリトライポリシー実装 |
| Tech Stack の Unit A 向け確定（Amplify Auth バージョン、AWS SDK 等） | コードレベルの import 文（→ Code Generation） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む（Q-N6 / Q-N10 / Q-N12 / Q-N13 で追加説明実施）
- [x] `aidlc-docs/construction/auth/nfr-requirements/nfr-requirements.md` を作成
- [x] `aidlc-docs/construction/auth/nfr-requirements/tech-stack-decisions.md` を作成
- [ ] aidlc-state.md と audit.md を更新（承認後）
- [ ] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-N1: Token Validity の数値

Cognito Token の有効期限。Functional Design の R-Login-5（セッション 1 時間以上維持）と NFR-DEG-01（低摩擦）のバランス。

| 案 | IdToken / AccessToken | RefreshToken | 特徴 |
|---|---|---|---|
| A | 1 時間 | 30 日 | Cognito デフォルト。一般的・無難 |
| B | 8 時間 | 30 日 | 半日操作で再ログイン不要、ダメ化UXの「めんどくさい再ログインをさせない」に整合 |
| C | 24 時間 | 90 日 | 最大限の摩擦最小化。一方で漏洩時の影響範囲が広がる |
| D | 1 時間 | 7 日 | RefreshToken も短めにしたセキュリティ寄り設定 |

[Answer]: **B**（8h / 30d）。IdToken/AccessToken 8 時間、RefreshToken 30 日。半日作業中の再ログイン不要を保証し、NFR-DEG-01「めんどくさい再ログインをさせない」に整合。R-Login-5 の「最低 1 時間」を大きく上回る。

### Q-N2: ログイン応答時間目標 (P95)

`POST /api/auth/login`（実体は Cognito InitiateAuth）の Frontend 体感レイテンシ目標。

| 案 | P50 | P95 | 備考 |
|---|---|---|---|
| A | 500ms | 1.5s | 一般的な Web アプリ水準 |
| B | 800ms | 2.5s | デモ用途で緩め |
| C | 1.0s | 3.0s | NFR-PERF-01 (3秒以内) と同じ閾値、Auth で 3 秒使うとアプリ全体の体感が損なわれる |
| 数値指定 | — | — | ユーザ指定で確定 |

[Answer]: **A**（P50 500ms / P95 1.5s）。一般的な Web アプリ水準でダメ化UXの「さっとログイン」体験を担保。

### Q-N3: JWT 検証 middleware のオーバーヘッド許容値

API Gateway Cognito Authorizer の検証時間に Lambda middleware (`AttachUserID`) が claims を読み出すオーバーヘッドが追加される。

| 案 | Authorizer | middleware | 合計 |
|---|---|---|---|
| A | 50ms 以下 | 10ms 以下 | 60ms 以下（厳しめ目標） |
| B | 100ms 以下 | 20ms 以下 | 120ms 以下（一般的） |
| C | 目標値は定めず NFR-PERF-01 全体予算（3 秒）内に収まれば OK | — | — |

NFR-PERF-01 は 3 秒。OrderService の Bedrock 推論で大半を消費するため、Auth は軽量であるべき。

[Answer]: **C**（目標値を定めず、NFR-PERF-01 全体予算 3 秒内で OK）。Auth 単体で個別 SLA は設けず、E2E でアプリ全体の 3 秒目標が満たされていることを確認するアプローチ。MVP ではこれで十分。

### Q-N4: 同時利用者数 (Auth 観点)

NFR-PERF-03 は「数人〜数十人」だが、Cognito の TPS 上限と関連する。

| 案 | 想定値 | Cognito API quota |
|---|---|---|
| A | デモ最小: 同時 5 人、ピーク 10 req/s | デフォルト quota で十分 |
| B | デモ標準: 同時 30 人、ピーク 30 req/s（Sign-in 系 80 req/s 上限内） | デフォルトで OK |
| C | 余裕あり: 同時 100 人、ピーク 100 req/s | InitiateAuth 上限 (デフォルト 80 req/s) に注意 |

[Answer]: **A**（同時 5 人 / ピーク 10 req/s）。ハッカソン審査デモ + 友人レビュー想定。Cognito デフォルト quota（InitiateAuth 80 req/s）で十分余裕あり、quota 増額申請不要。

### Q-N5: 可用性目標と Cognito 障害時の挙動

Cognito 自体が落ちた場合の Auth Unit のフォールバック方針。

| 案 | 内容 |
|---|---|
| A | **可用性目標を立てない、Cognito の SLA に追従**（NFR-AVAIL-01 と整合）。障害時はエラー表示のみ、リトライさせない |
| B | **クライアント側のみリトライ**（Amplify SDK の標準リトライに任せる）、サーバ側追加対応なし |
| C | **障害告知ページを用意**（特定エラー時に「現在ログインできません」専用画面を表示） |

[Answer]: **A**（Cognito SLA 追従 + エラー表示のみ）。NFR-AVAIL-01 の方針通り AWS マネージドの可用性に依拠。`AuthError.NETWORK_ERROR` のダメ化トーン軽メッセージを表示するのみ、サーバ側 / 専用ページの追加対応はなし。

### Q-N6: API Gateway レート制限

NFR-SEC-03 に「最小限設定（任意）」とあるが、ブルートフォース防御として認証経路には設定すべき。

| 案 | 認証経路 | 一般 API | 備考 |
|---|---|---|---|
| A | 制限なし | 制限なし | デモのみ、運用しない前提 |
| B | 10 req/s/IP | 50 req/s/IP | 攻撃耐性ありつつデモ実演可能 |
| C | API Gateway Usage Plan で全体 100 req/s, Burst 200 | 同左 | テナント単位の細かい制限は MVP では不要 |
| D | Cognito 側の Advanced Security Features を使う | — | 追加コスト発生（Cognito 価格プラン変更） |

[Answer]: **C**（Stage 全体 100 req/s / Burst 200）。API Gateway Stage Throttling で全体上限を設定。Terraform 数行で実装可能、デモ規模には十分余裕（Q-N4=A: ピーク 10 req/s）。IP 単位の細かい制限・WAF は MVP 範囲外、本番化時の対応として NFR-SEC-04 に明記。なおユーザから Q-N6 の選択肢説明依頼あり、本 AI 応答で 4 案を実装コスト・攻撃耐性・デモ影響の観点で比較表提示済み。

### Q-N7: ログ・観測性の項目

NFR-OBS-01 に基づき、Auth Unit の Lambda が出力する構造化ログ項目。

| 案 | 内容 |
|---|---|
| A | 最小: `level`, `timestamp`, `userId`, `action` (例: `logout`) のみ |
| B | 標準: A + `traceId` (X-Ray ID 任意), `requestId`, `email_hash` (SHA256), `userAgent` |
| C | 詳細: B + `clientIp`, `cognitoApiResponseCode`, `latencyMs` |

ダメ化UX のデモ用途なので C は過剰、A は最低限すぎ、B が妥当か。

[Answer]: **B**（標準）。`level` / `timestamp` / `userId` / `action` / `traceId` / `requestId` / `email_hash` (SHA256) / `userAgent` の 8 項目。NFR-OBS-01 構造化ログ推奨と整合、トラブルシュート可能性とプライバシー（メール平文を出さず hash 化）のバランス。

### Q-N8: メトリクス収集の有無

CloudWatch Custom Metrics の取得範囲。

| 案 | 内容 |
|---|---|
| A | **取得しない**（NFR-OBS-02「高度な観測性は実装しない」と整合） |
| B | **基本のみ**: Cognito 標準メトリクス（SignInSuccesses / SignInFailures 等）の Dashboard 表示のみ |
| C | **カスタム取得**: `auth_login_duration_ms`, `auth_failure_count_by_reason` を EMF で出力 |

[Answer]: **A**（取得しない）。NFR-OBS-02「高度な観測性は本 MVP では実装しない」に厳密準拠。Cognito 標準メトリクスは AWS Console から確認可能だが、Dashboard 作成も省略。本番化時の対応として設計書に明記。

### Q-N9: Pre Sign-up Lambda Trigger の信頼性

auto-confirm を担う Pre Sign-up Lambda の失敗時の挙動。

| 案 | 内容 |
|---|---|
| A | **Cognito の標準挙動に任せる**: Trigger が失敗すれば SignUp 全体が失敗、ユーザに「登録に失敗しました」表示 |
| B | **A + 構造化ログに ERROR 出力**して運用者がアラートで気付ける状態にする |
| C | **A + B + DLQ 経路**を設ける（過剰、MVP 範囲外） |

[Answer]: **A**（Cognito 標準挙動 + ユーザにエラー表示）。Trigger 失敗 → SignUp 全体失敗 → `EMAIL_ALREADY_EXISTS` 等の Cognito エラーをユーザに表示。Trigger 内の例外はランタイムログに自動出力されるため明示的な構造化ログ追加もしない。シンプルに保つ。

### Q-N10: テスタビリティ要件

NFR-TEST-01〜04 の Unit A への適用範囲。

| 案 | 内容 |
|---|---|
| A | **PBT は適用しない**（Unit A に純粋関数の中核ロジックは少ない、適用対象は Unit B / C / E）。通常のユニット + 統合テスト + E2E |
| B | **メール正規化と email_hash 生成にだけ PBT**（ASCII / Unicode / 大文字混在の不変条件） |
| C | **JWT claims 抽出ロジック等にも PBT** を広げる |

[Answer]: **B**（メール正規化と email_hash にだけ PBT）。プロパティ例: `normalize` のべき等性・大文字小文字不問、`emailHash` の長さ不変・衝突なし。Q-N7=B で email_hash を採用するため、検証価値あり。Extension 設定 Partial と整合。なおユーザから Q-N10 の選択肢説明依頼あり、本 AI 応答で PBT の概念・各案で書くプロパティ例・工数比較を提示済み。

### Q-N11: ブラウザ互換性

NFR-A11Y-01 に「基本的なセマンティック HTML」とあるが、サポートブラウザ範囲は未定義。

| 案 | 内容 |
|---|---|
| A | モダンブラウザ最新版のみ（Chrome / Edge / Safari の最新 2 バージョン）、IE / 古い Safari は対象外 |
| B | A + iOS Safari 15+ / Android Chrome 100+ も明示サポート |
| C | 主要モバイルブラウザ全部（PWA インストール用途） |

[Answer]: **A**（モダンブラウザ最新 2 バージョンのみ）。Chrome / Edge / Safari の最新 2 バージョン。IE・古い Safari は対象外。MVP として最小スコープ、本番化時にモバイル最適化を検討する旨を設計書に明記。

### Q-N12: 言語・ロケール

`requirements.md` Q15 で「日本語のみ」と確定済みだが、Auth Unit の UI 文言と Cognito エラーメッセージの扱い。

| 案 | 内容 |
|---|---|
| A | **UI 文言だけ日本語化、Cognito エラーは英語のまま受け取って Frontend で日本語にマップ** |
| B | **A + Cognito Custom Message Lambda Trigger でメール本文も日本語化**（auto-confirm 採用なので影響軽微） |

Q-A1=B (auto-confirm) のため、メール送信は基本不要。A で十分か。

[Answer]: **A**（Frontend で日本語マッピング）。`authMessages.ts` (Functional Design §11) で Cognito 英語エラーコードを日本語ダメ化トーン軽メッセージにマップ。Cognito Custom Message Trigger は追加しない（auto-confirm でメール送信なし、実装コストゼロ）。なおユーザから「どっちが方針に沿ってる？」の確認あり、本 AI 応答で Q-A1/Q-A8/最小スコープ方針との整合性比較表を提示済み。

### Q-N13: 監査・コンプライアンス

NFR-COMP-02 「個人情報はメールアドレスのみ取得」「将来的な個人情報保護法 / 金融庁ガイドライン対応は明記」。

| 案 | 内容 |
|---|---|
| A | 設計書に「本 MVP は本番運用を想定しない、本番化時は別途レビュー」とのみ明記 |
| B | A + Cognito User Pool の削除手順（GDPR / 個人情報保護法対応の最低限）をドキュメント化 |

[Answer]: **A**（NFR-COMP-01〜03 を引用のみ）。Inception フェーズで既に確定済みの 3 項目を `nfr-requirements.md` に転記するのみ。削除手順の文書化は本番化時に対応する旨を明記、本 MVP では文書化しない。なおユーザから「Inception フェーズで定義しなかったっけ？」の確認あり、本 AI 応答で requirements.md §4.6 NFR-COMP-01〜03 を引用提示済み。

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `nfr-requirements.md` | Unit A 向け NFR の数値・しきい値・対応ポリシー一覧（性能 / 可用性 / セキュリティ / 観測性 / 信頼性 / テスト / アクセシビリティ / ロケール / コンプライアンス） |
| `tech-stack-decisions.md` | Unit A の Tech Stack 確定: Amplify Auth バージョン、AWS SDK バージョン、middleware ライブラリ、テストツール、Cognito 設定パラメータの上位概要 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- Cognito User Pool 詳細パラメータ（password policy 各項目、Lambda Trigger ARN 等）→ **Infrastructure Design**
- API Gateway Usage Plan の Terraform 実装 → **Infrastructure Design**
- middleware のリトライ・タイムアウト具体値 → **NFR Design**
- ログ出力の具体的フォーマット文字列 → **NFR Design / Code Generation**
- AuthGuard / SessionExpiredModal の実装 → **Code Generation**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-N1 〜 Q-N13 を対話形式で順にヒアリング開始
