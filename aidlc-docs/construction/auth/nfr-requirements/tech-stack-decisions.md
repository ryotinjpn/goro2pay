# Auth Unit — Tech Stack Decisions

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: A (`auth`)
**Stage**: NFR Requirements / Construction
**Predecessors**: Application Design / Functional Design / Q-A1〜Q-A10, Q-N1〜Q-N13

本ドキュメントは Unit A の Tech Stack 上位概要を確定する。バージョンの厳密な pin（go.mod / package.json の数値）は Code Generation で行うが、本書では「採用する技術と大枠のバージョン要件」を凍結する。

参照: [unit-interfaces.md](../../interfaces/unit-interfaces.md), [nfr-requirements.md](./nfr-requirements.md)

---

## 1. Backend (Go on Lambda)

### 1.1 ランタイム

| 項目 | 採用 | 出典 |
|---|---|---|
| Lambda ランタイム | `provided.al2023` (custom runtime) | Application Design Q-A: Go + Gin + Lambda Web Adapter |
| Go バージョン | **1.22 系**（2026-05 時点 LTS 相当）minimum | プロジェクト全体共通 |
| デプロイ形態 | コンテナイメージ（LWA + Go バイナリ） | Application Design |

### 1.2 主要ライブラリ

| 役割 | ライブラリ | 採用理由 |
|---|---|---|
| HTTP framework | **`github.com/gin-gonic/gin`** | Application Design 確定 |
| Lambda → HTTP adapter | **AWS Lambda Web Adapter (LWA)** | コンテナイメージのサイドカー extension |
| Logger | **`log/slog`**（Go 標準）+ JSON Handler | A-NFR-OBS-01 構造化ログを最小依存で実現、外部依存追加なし |
| AWS SDK | **AWS SDK for Go v2 (`aws-sdk-go-v2`)** | プロジェクト全体共通、本 Unit では Cognito SDK 直接利用は基本なし（Authorizer 任せ） |
| エラー型 | 標準 `errors`（`errors.New` / `errors.Is`） | sentinel error は既に [unit-interfaces.md §2.1](../../interfaces/unit-interfaces.md) で定義 |

### 1.3 テスト

| 役割 | ライブラリ | 採用理由 |
|---|---|---|
| 単体テスト | **`testing`**（Go 標準）+ `github.com/stretchr/testify` | プロジェクト共通 |
| Property-Based Testing | **`github.com/leanovate/gopter`** | A-NFR-TEST-02 で `normalize` / `emailHash` に適用 |
| HTTP テスト | `net/http/httptest` + Gin の test helpers | middleware 単体テスト用 |
| Cognito モック | LocalStack の Cognito（A-NFR-TEST-03） | 統合テスト用、必要に応じて |

### 1.4 採用しないもの

- **`go-jwt` 系ライブラリでの JWT 再検証** — Q-A4=A により Authorizer のみで完結、Lambda 内再検証は実施しない
- **AWS X-Ray SDK** — A-NFR-OBS-03 によりトレース未導入

---

## 2. Frontend (Next.js / TypeScript)

### 2.1 ランタイム

| 項目 | 採用 | 出典 |
|---|---|---|
| Node.js | **20 LTS 以上** | Amplify Hosting の Build 環境準拠 |
| Next.js | **15 系** (App Router) | Application Design |
| TypeScript | **5.x** | Next.js 15 の標準 |

### 2.2 主要ライブラリ

| 役割 | ライブラリ | バージョン要件 | 採用理由 |
|---|---|---|---|
| 認証 SDK | **`aws-amplify`**, **`@aws-amplify/auth`** | **Amplify v6 系** | Q-A7=A 確定、Amplify Hosting と統合 |
| State Management | `jotai` | （Application Design で確定） | プロジェクト共通 |
| データフェッチ | `@tanstack/react-query` | （同上） | プロジェクト共通 |
| HTTP クライアント | `fetch` (web 標準) | — | apiClient ラッパで使用 |
| フォーム検証 | **手書き**（軽量、フォームが 2 つのみ） | — | `react-hook-form` 等は MVP には過剰 |
| スタイル | プロジェクト共通方針に準拠（未確定） | — | Functional Design 範囲外、UI 実装で確定 |

### 2.3 テスト

| 役割 | ライブラリ | 採用理由 |
|---|---|---|
| 単体テスト | **Vitest** または **Jest**（プロジェクト方針に追従） | A-NFR-TEST-01 |
| コンポーネントテスト | **`@testing-library/react`** | A-NFR-TEST-01 |
| Property-Based Testing | **`fast-check`** | A-NFR-TEST-02、Frontend 側で同じ `normalize` / `emailHash` を実装する場合に適用 |
| E2E | **Playwright** | A-NFR-TEST-04、ハッピーパス・タップ数検証 |

### 2.4 採用しないもの

- **`amazon-cognito-identity-js` 直接利用** — Q-A7=A により Amplify Auth 経由に統一
- **Cognito Hosted UI** — Q-A7=A により独自 UI（ダメ化UX コピー保持のため）
- **`react-hook-form`** — フォームが 2 つのみで自前検証で十分
- **国際化ライブラリ (`next-intl` 等)** — A-NFR-I18N-01 により日本語固定、不要

---

## 3. AWS マネージドサービス

| サービス | 用途 | 主要設定（Infrastructure Design で詳細確定） |
|---|---|---|
| **Amazon Cognito User Pool** | 認証主体 | Sign-in attribute = email, MFA OFF, password policy = A-NFR-SEC-02 準拠, Token Validity = A-NFR-SEC-03 準拠, Pre Sign-up Lambda Trigger = auto-confirm |
| **Amazon Cognito App Client** | Frontend 認証クライアント | Auth Flow = USER_SRP_AUTH（平文パスワードを送らない / Amplify v6 default）, Refresh Token 30 日 |
| **AWS Lambda** | Pre Sign-up Trigger（auto-confirm） | Go ランタイム or Node.js ランタイム、5 行程度のシンプル実装 |
| **Amazon API Gateway (REST)** | API ゲート | Cognito Authorizer を全認証必須エンドポイントに適用、Stage Throttling = 100 req/s / Burst 200 |
| **AWS Amplify Hosting** | Frontend 配信 | Application Design 確定（PR #10）、Amplify Auth 統合の利点を活用 |
| **Amazon CloudWatch Logs** | Lambda ログ | A-NFR-OBS-01 構造化ログ出力先 |

### 3.1 採用しないマネージドサービス

| サービス | 理由 |
|---|---|
| AWS WAF | A-NFR-SEC-09 本番化時に再評価 |
| AWS X-Ray | A-NFR-OBS-03 本 MVP では未導入 |
| Cognito Identity Pool | 本 MVP は User Pool のみ。AWS リソース直接アクセスは Lambda 経由のみ |
| Cognito Hosted UI | Q-A7=A により独自 UI |
| AWS Secrets Manager / Parameter Store | Auth Unit が扱う秘匿情報がない（パスワードは Cognito 内で管理） |
| AWS KMS（顧客管理キー） | A-NFR-SEC-05 マネージド標準暗号化のみ |

---

## 4. インフラ管理

| 項目 | 採用 |
|---|---|
| IaC | **Terraform**（プロジェクト共通、`requirements.md` Q12=C） |
| モジュール構成 | terraform-module-design プラグイン規約準拠 |
| コーディング規約 | terraform-coding-rule プラグイン規約準拠 |
| テスト | terraform-test プラグイン規約準拠（mock_provider 利用） |

詳細な Terraform モジュール設計は Infrastructure Design ステージで確定。

---

## 5. CI/CD・開発ワークフロー

| 項目 | 採用 | 詳細 |
|---|---|---|
| バージョン管理 | Git | branch-convention プラグイン規約準拠 |
| コミット規約 | commit-convention プラグイン規約準拠 | 絵文字 prefix 必須 |
| Secret スキャン | secretlint + pre-commit hook | team-baseline プラグイン適用済み |
| AWS リソース命名 | aws-naming-convention プラグイン規約準拠 | Infrastructure Design で命名値確定 |
| pre-commit hook | プロジェクト共通 | secretlint 等 |

---

## 6. バージョンマトリクス（Code Generation での pin 対象）

下記は Code Generation で `go.mod` / `package.json` に pin される最終バージョンの目安。実際の値は Code Generation 時に最新安定版を選定する。

| 領域 | パッケージ | 目安バージョン |
|---|---|---|
| Backend Go | `github.com/gin-gonic/gin` | v1.10.x |
| Backend Go | `github.com/stretchr/testify` | v1.10.x |
| Backend Go | `github.com/leanovate/gopter` | v0.2.x |
| Frontend | `next` | 15.x |
| Frontend | `react` | 19.x |
| Frontend | `aws-amplify` | 6.x |
| Frontend | `@aws-amplify/auth` | 6.x |
| Frontend | `@tanstack/react-query` | 5.x |
| Frontend | `jotai` | 2.x |
| Frontend | `fast-check` | 3.x |
| Frontend | `@playwright/test` | 1.50.x |

---

## 7. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | Cognito 設定の Terraform 表現、Stage Throttling のパラメータ、ログ出力フォーマットの詳細実装 |
| **Infrastructure Design** | Terraform モジュール分割、env 変数の Lambda への注入経路、Pre Sign-up Lambda Trigger の Go コード本体 + Terraform 連携 |
| **Code Generation** | go.mod / package.json の最終 version pin、`normalize` / `emailHash` 実装、authMessages.ts、`useAuth` 実装、middleware 実装、PBT テストコード |
