# Auth Unit — Business Rules

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: A (`auth`)
**Stage**: Functional Design / Construction

本ドキュメントは Unit A の **業務ルール、検証ロジック、制約** を定義する。実装言語に依存しない論理仕様として記述する。

---

## 1. メールアドレス検証

### R-Email-1: 形式検証
- **何**: メールアドレスは RFC 5321 / 5322 に準拠した形式であること
- **どこで**: クライアント（Signup / Login フォーム）と Cognito の両方
- **クライアント側ルール**: 簡易な正規表現でユーザに即時フィードバック（送信ボタン押下不要）
  - パターン: `/^[^\s@]+@[^\s@]+\.[^\s@]+$/` 程度の簡易検証
- **サーバ側ルール**: Cognito User Pool の `InvalidParameterException` で最終判定

### R-Email-2: 一意性
- **何**: 同一 User Pool 内に同じメールアドレスは登録できない
- **どこで**: Cognito User Pool（`UsernameAttributes: ["email"]` 設定）
- **エラー**: `EMAIL_ALREADY_EXISTS`（[domain-entities.md §6.2](./domain-entities.md)）

### R-Email-3: 大文字小文字
- **何**: メールアドレスは大文字小文字を区別せず扱う
- **どこで**: Login 入力時、クライアント側で `toLowerCase()` してから Cognito に送信
- **理由**: ユーザの入力ぶれによる「USER_NOT_FOUND」誤発火を防ぐ

---

## 2. パスワードポリシー

Q-A2 の決定（標準）に従う。

### R-Pwd-1: 強度
| 項目 | 値 |
|---|---|
| 最小長 | **8 文字** |
| 数字を含む | **必須** |
| 英大文字を含む | **必須** |
| 英小文字を含む | **必須** |
| 記号を含む | **任意**（要求しない） |
| 最大長 | 256 文字（Cognito 上限） |

### R-Pwd-2: クライアント側事前検証
- 入力中にリアルタイムで強度を判定し、未充足要件をフォーム下に表示
- 表示テキストは肯定形（例: 「8文字以上にしてください」「数字を1文字以上含めてください」）

### R-Pwd-3: 取り扱い
- **R-Pwd-3-a**: パスワードは **localStorage / sessionStorage / Cookie に保存しない**
- **R-Pwd-3-b**: パスワードはアプリのログ・Telemetry・例外メッセージに**一切出力しない**
- **R-Pwd-3-c**: パスワードは入力後 Cognito API への送信完了次第、**メモリから速やかに参照を断つ**（フォーム state を空にする）

---

## 3. 新規登録（Signup）

### R-Signup-1: 必須入力
- 入力項目はメールアドレスとパスワードの 2 つのみ
- 「登録」ボタンは両項目が R-Email-1 / R-Pwd-1 を満たすまで disabled

### R-Signup-2: 自動確認（auto-confirm）
- Q-A1 の決定に従い、メール検証コードフローは採用しない
- Cognito の **Pre Sign-up Lambda Trigger** を導入し、`autoConfirmUser = true` および `autoVerifyEmail = true` を設定
- Signup レスポンスには `userConfirmed: true` が返り、その後即座に `signIn` を呼び出してセッション確立

### R-Signup-3: 登録後の遷移
- Signup 成功 → `signIn` 自動実行 → AuthSession 確立 → `router.push("/")`
- `/` 側（MainScreen, Unit B/C 横断）が予算未設定なら `/budget` へ遷移する責務を持つ（Q-A10）

### R-Signup-4: 総タップ数の上限
- US-0-01 受入基準を満たすため、登録から最初のメイン画面ボタン操作までの総タップ数は **5 回以下**:
  1. ランディング画面の「はじめる」ボタン
  2. メールアドレス入力フィールドへフォーカス（Tap 1）
  3. パスワード入力フィールドへフォーカス（Tap 2）
  4. 「登録」ボタン（Tap 3）
  5. 予算設定画面の「決定」ボタン（Tap 4 — 予算未設定の初回のみ）
  6. メイン画面の「ご飯めんどくさい」ボタン（Tap 5）
- ※ ソフトウェアキーボードでの文字入力タップは「総タップ数」に含めない

---

## 4. ログイン（Login）

### R-Login-1: 必須入力
- メールアドレス + パスワード
- 「ログイン」ボタンは形式検証を満たすまで disabled

### R-Login-2: エラー時の文言統一
- `INVALID_CREDENTIALS` と `USER_NOT_FOUND` は同一文言（[domain-entities.md R-Err-1](./domain-entities.md)）
- ダメ化トーン軽（Q-A8）: 「ログインすらめんどくさいですよね…もう一度お試しください」

### R-Login-3: ログイン後の遷移
- 成功 → AuthSession 確立 → `router.push("/")`
- 分岐責務は `/` 側（Q-A10、Signup と同じ）

### R-Login-4: パスワードを間違える（US-0-02 シナリオ 2）
- 受入基準: ユーザは「メールアドレスまたはパスワードが違います」とエラーを見る
- 実装: R-Login-2 のダメ化トーンで実現する（受入基準のニュアンスは満たしつつ NFR-DEG-05 を優先）

### R-Login-5: セッション維持
- 受入基準: 「セッションは少なくとも 1 時間維持される」
- 実装: Cognito の Token Validity を IdToken / AccessToken **1 時間以上**、RefreshToken **30 日**（Infrastructure Design で確定）

---

## 5. ログアウト（Logout）

### R-Logout-1: 提供範囲（Q-A3 = A）
- バックエンド API: `POST /auth/logout`
- フロントエンド: ヘッダ右上の「⏏︎ ログアウト」ボタン or 設定画面のメニュー項目

### R-Logout-2: 実装フロー

```
1. クライアント: ボタン押下
2. クライアント: Amplify Auth.signOut({ global: true }) を呼ぶ
   → 内部で Cognito GlobalSignOut が走り、Refresh Token が無効化
3. クライアント: ローカルストレージのトークンが Amplify によって除去
4. クライアント: POST /auth/logout を呼ぶ（任意、サーバ側ログ記録のため）
5. クライアント: router.push("/") で Landing 表示に戻る
```

### R-Logout-3: API 仕様

```
POST /auth/logout
Authorization: Bearer <IdToken>
Body: なし

Response:
  204 No Content        — 成功
  401 Unauthorized      — JWT 不正
```

サーバ側は GlobalSignOut を再実行する義務はない（Amplify が既に呼出済）。本 API は監査ログ目的のみ。

### R-Logout-4: 確認ダイアログ
- ログアウト誤タップを防ぐため、確認モーダルを 1 段挟む
- 文言（ダメ化トーン軽）: 「もうダメ化を終わらせますか？」「キャンセル / ログアウト」

---

## 6. JWT 検証と userId 注入

### R-JWT-1: 検証主体（Q-A4 = A）
- API Gateway の **Cognito Authorizer** が JWT 署名・有効期限・発行者を検証
- Lambda は再検証しない

### R-JWT-2: claims 抽出
- Lambda は `event.requestContext.authorizer.claims` から下記を取得:
  - `sub` → `userId` として Gin Context に注入
  - `email` → 監査ログ用にメモリ上のみ保持（永続化禁止）
- Authorizer が許可しない場合は API Gateway が 401 を返し、Lambda は呼ばれない

### R-JWT-3: userId フォーマット検証
- `sub` は UUID v4 形式（Cognito 仕様）
- middleware は形式チェックをパススルー（Cognito を信頼）

### R-JWT-4: 不正な claims のフォールバック
- 万が一 `claims` が欠落していた場合（バグ・設定ミス）、Lambda は **500 INTERNAL_ERROR** を返し、構造化ログに `level=ERROR` で記録
- ユーザに認証エラーは出さない（バグの可能性が高いため）

---

## 7. セッション失効処理

### R-Session-1: 失効検出（Q-A9）
- API 呼出が **401 を 1 回でも返した時点で失効**と判定
- Amplify Auth Hub の `tokenRefresh_failure` イベントもトリガーとして使用

### R-Session-2: ユーザ通知 + 遷移
1. アプリ全体で開いているモーダルがあれば閉じる
2. ダメ化トーン軽のモーダル表示: 「お疲れ様でした。もう一度ログインしてください」
3. 1.5 秒後（または OK タップ）で `router.push("/login")`
4. URL クエリ `?from=session_expired` を付与（Login 画面で初期状態のヒント表示用）

---

## 8. 未認証アクセスの挙動（Q-A6）

### R-Unauth-1: ルート別ポリシー

| ルート | 未認証時の挙動 |
|---|---|
| `/` | LandingScreen を描画（同一 URL 内で UI 切替） |
| `/login` | LoginScreen を描画（誰でもアクセス可） |
| `/(auth)/signup` | SignupScreen を描画（誰でもアクセス可） |
| `/budget` | `/login` にリダイレクト |
| `/budget-empty` | `/login` にリダイレクト |
| `/order/[id]/complete` | `/login` にリダイレクト |
| 他全て | `/login` にリダイレクト |

### R-Unauth-2: ガード機構
- Next.js の各ページで `useAuth()` の `isAuthenticated` を判定
- 共通の `<AuthGuard>` コンポーネントを `app/(authenticated)` レイアウトに配置することを推奨（Code Generation で実装方法を最終決定）

---

## 9. ダメ化UX 制約の織り込み

| NFR ID | 適用先ルール | 実装方法 |
|---|---|---|
| NFR-DEG-01（低摩擦） | R-Signup-4 | 総タップ数 5 回以下を受入基準とする |
| NFR-DEG-05（コピー） | R-Login-2, R-Logout-4, R-Session-2 | エラーメッセージはダメ化トーン軽で統一 |

---

## 10. ガード句一覧（実装漏れチェック用）

| ガード | 場所 | 違反時の挙動 |
|---|---|---|
| メール形式 | クライアント (Signup/Login) | 送信ボタン disabled |
| パスワード強度 | クライアント (Signup) | 送信ボタン disabled、強度ヒント表示 |
| パスワード非保存 | クライアント全体 | コードレビュー / ESLint 規約で保証 |
| JWT claims 存在 | API middleware | 500 + ERROR ログ |
| 未認証ルート守備 | Next.js Layout / Guard | `/login` にリダイレクト |
| ログアウト確認 | クライアント | 誤タップ防止モーダル |

---

## 11. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| NFR Requirements | Token Validity の数値、レート制限のしきい値、ログイン応答時間目標 |
| NFR Design | Pre Sign-up Lambda Trigger の実装パターン、CloudFront / WAF レイヤでのレート制限 |
| Infrastructure Design | Cognito Password Policy / Token Validity / App Client 設定の Terraform 実装 |
| Code Generation | バリデーションロジックの実装、エラーメッセージリソース、AuthGuard コンポーネント |
