# Auth Unit — Domain Entities

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: A (`auth`)
**Stage**: Functional Design / Construction

本ドキュメントは Unit A の **ドメインエンティティ** を技術中立に定義する。具体的なデータ型・永続化技術・Cognito 属性のマッピング詳細は Infrastructure Design / Code Generation で確定する。

---

## 1. エンティティ一覧

| エンティティ | 種別 | 永続化 | 説明 |
|---|---|---|---|
| `UserIdentity` | Aggregate Root | Cognito User Pool | ゴロゴロPay を利用するユーザの不変アイデンティティ |
| `Credentials` | Value Object | Cognito 内部 | メールアドレスとパスワードの組（Cognito が秘匿管理） |
| `AuthSession` | Aggregate Root | クライアント側 + Cognito Token Service | ログイン後の認証コンテキスト |
| `AuthToken` | Value Object | クライアント側のメモリ + ストレージ | JWT 形式のトークン情報 |
| `AuthError` | Value Object | 永続化なし | 認証時に発生したエラーの分類 |

---

## 2. UserIdentity

### 2.1 概要
ゴロゴロPay の **ユーザを一意に識別する不変アイデンティティ**。本 Unit A の中核エンティティ。後続 Unit B/C/D/E はすべて `userId`（= Cognito `sub`）を外部キーとしてデータを管理する。

### 2.2 属性

| 属性 | 型 | 不変性 | 必須 | 説明 |
|---|---|---|---|---|
| `userId` | string (UUID v4) | **不変** | 必須 | Cognito `sub`。全 Unit のテーブル PK となる |
| `email` | string | 可変（将来対応） | 必須 | メールアドレス。MVP では変更不可とする |
| `status` | enum | 可変 | 必須 | `CONFIRMED` / `DISABLED`。本 MVP では基本 `CONFIRMED` 固定（auto-confirm 採用） |
| `createdAt` | timestamp | 不変 | 必須 | アカウント作成日時（Cognito 管理） |

### 2.3 不変条件 (Invariants)

- **I-User-1**: `userId` は一度発行されたら変更されない
- **I-User-2**: `email` は本 MVP の Cognito User Pool 内で一意（重複登録不可）
- **I-User-3**: 認証成功時の JWT `sub` クレームは必ず `userId` と一致する

### 2.4 ライフサイクル

```
[未登録] ──Signup──► [CONFIRMED]
                           │
                           ├──Login──► [認証セッション中]（AuthSession 生成）
                           │
                           └──Disable──► [DISABLED]（将来対応、本 MVP では未使用）
```

### 2.5 永続化責務
- 永続化主体: **Amazon Cognito User Pool**
- アプリケーション側に独立したユーザテーブルは持たない（要件 5.2 / NFR-COMP-02 整合）
- 後続 Unit のテーブルは `userId` のみを外部キーとして保持し、`email` は持たない

---

## 3. Credentials

### 3.1 概要
ユーザの認証情報。**Cognito の内部にのみ存在し、アプリケーション側は決して値を保持しない**。

### 3.2 属性

| 属性 | 型 | アクセス | 説明 |
|---|---|---|---|
| `email` | string | 入力時のみ | ログイン識別子（同時に `UserIdentity.email` でもある） |
| `password` | string (秘匿) | 入力時のみ | Cognito のパスワードポリシーに従う |

### 3.3 業務ルール
- パスワード強度: 詳細は [business-rules.md](./business-rules.md) §2 を参照
- パスワードはログ・Telemetry・エラーメッセージに**絶対に出力しない**

---

## 4. AuthSession

### 4.1 概要
ユーザがログイン中であることを表すコンテキスト。**JWT トークン群の保持・更新を責務とする**。

### 4.2 属性

| 属性 | 型 | 必須 | 説明 |
|---|---|---|---|
| `userId` | string | 必須 | このセッションが帰属するユーザ（Cognito `sub`） |
| `idToken` | AuthToken | 必須 | API 呼出時に `Authorization: Bearer` で送信 |
| `accessToken` | AuthToken | 必須 | Cognito API 呼出（GlobalSignOut 等）に使用 |
| `refreshToken` | AuthToken | 必須 | アクセストークン更新に使用 |
| `establishedAt` | timestamp | 必須 | ログイン成功時刻（クライアント時計） |

### 4.3 不変条件
- **I-Session-1**: `idToken.claims.sub === userId`
- **I-Session-2**: 全トークンの有効期限は `establishedAt` より未来
- **I-Session-3**: `userId` 変更不可

### 4.4 ライフサイクル

```
[未確立] ──Login成功──► [Active]
                            │
                            │ (idToken expires)
                            ├──Auto Refresh──► [Active]（透過更新）
                            │
                            │ (refreshToken expires)
                            ├──Re-Login要求──► [Expired]
                            │
                            └──Logout──► [Terminated]
```

### 4.5 永続化責務
- 主体: **Amplify Auth ライブラリ**（`@aws-amplify/auth`）
- ストレージ: ブラウザの localStorage（Amplify がデフォルトで管理）
- アプリケーション側で直接保存・取り出しはしない（`fetchAuthSession()` 経由でのみアクセス）

---

## 5. AuthToken

### 5.1 概要
JWT 形式のトークンを表す Value Object。

### 5.2 属性

| 属性 | 型 | 説明 |
|---|---|---|
| `value` | string | 生 JWT 文字列 |
| `claims` | TokenClaims | デコード済みクレーム |
| `expiresAt` | timestamp | `claims.exp` を時刻型に変換した値 |
| `tokenUse` | enum | `id` / `access` / `refresh` |

### 5.3 TokenClaims（IdToken の主要クレーム）

| クレーム | 説明 |
|---|---|
| `sub` | UserIdentity.userId と一致（UUID v4） |
| `email` | ユーザのメールアドレス |
| `email_verified` | true（auto-confirm のため常に true） |
| `cognito:username` | Cognito ユーザ名（本 MVP では `sub` と同値） |
| `iat` / `exp` | 発行時刻 / 有効期限 |
| `aud` | App Client ID |
| `iss` | Cognito User Pool の発行者 URL |

---

## 6. AuthError

### 6.1 概要
認証フロー上で発生するエラーの分類を表す Value Object。UI コピー (NFR-DEG-05) と HTTP エラーコードの紐付けに使う。

### 6.2 種別

| AuthError 種別 | 起因 | UI コピー（ダメ化トーン軽） | HTTP / Cognito |
|---|---|---|---|
| `INVALID_CREDENTIALS` | メール or パスワード不一致 | 「ログインすらめんどくさいですよね…もう一度お試しください」 | Cognito `NotAuthorizedException` |
| `USER_NOT_FOUND` | 未登録メールでログイン試行 | 同上（情報漏洩抑止のため `INVALID_CREDENTIALS` と同一文言） | Cognito `UserNotFoundException` |
| `EMAIL_ALREADY_EXISTS` | 既存メールで Signup | 「このメールはもう使われています。ログインしますか？」 | Cognito `UsernameExistsException` |
| `WEAK_PASSWORD` | パスワードポリシー違反 | 「パスワードは8文字以上、英大小と数字を含めてください」 | Cognito `InvalidPasswordException` |
| `INVALID_EMAIL_FORMAT` | メール形式エラー | 「メールアドレスの形式が正しくないようです」 | Cognito `InvalidParameterException`（client side でも検出） |
| `RATE_LIMIT_EXCEEDED` | 試行回数超過 | 「ちょっと頑張りすぎです。少し待ってから試してください」 | Cognito `TooManyRequestsException` |
| `SESSION_EXPIRED` | リフレッシュトークン失効 | 「お疲れ様でした。もう一度ログインしてください」（モーダル表示後 `/login`） | クライアント検出 |
| `NETWORK_ERROR` | 通信失敗 | 「通信が…ちょっと待ってもう一度」 | クライアント検出 |
| `UNKNOWN` | 未分類 | 「うまくいきませんでした。もう一度お試しください」 | フォールバック |

### 6.3 業務ルール
- **R-Err-1**: `INVALID_CREDENTIALS` と `USER_NOT_FOUND` はユーザ表示で同一の文言とする（メールアドレス存在判別の漏洩を防ぐ）
- **R-Err-2**: パスワードはエラーログに**絶対に含めない**

---

## 7. エンティティ関係図

```
┌──────────────────┐
│  UserIdentity    │  (1)
│   - userId (PK)  │ ◄────┐
│   - email        │      │ owns
│   - status       │      │
└──────────────────┘      │
                          │
                  ┌───────┴───────┐
                  │  AuthSession  │  (0..n、ただし MVP は同時 1 セッション)
                  │   - userId    │
                  │   - tokens    │
                  │   - establishedAt │
                  └───────┬───────┘
                          │ contains 3
                          ▼
                  ┌──────────────┐
                  │  AuthToken   │  (3 個: idToken, accessToken, refreshToken)
                  │   - value    │
                  │   - claims   │
                  │   - expiresAt│
                  └──────────────┘

  AuthError は永続化されず、エラー発生時に都度生成される Value Object
```

---

## 8. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| NFR Requirements | `AuthSession` の有効期限値 / Refresh 期限値の確定 |
| NFR Design | `AuthError.RATE_LIMIT_EXCEEDED` のしきい値、Token 更新失敗時のリトライ回数 |
| Infrastructure Design | Cognito User Pool 属性スキーマ、Password Policy 数値、Token Validity 数値 |
| Code Generation | エンティティの Go 型定義 / TypeScript 型定義の生成 |
