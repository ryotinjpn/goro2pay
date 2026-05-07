# ゴロゴロPay — Requirements Document

**Document Version**: 1.0
**Created**: 2026-05-07
**Project Type**: Greenfield
**Depth Level**: Comprehensive
**Deadline**: 2026-05-10 までに Inception フェーズ完了必須

---

## 1. Intent Analysis Summary

### 1.1 User Request
「AI-DLC を使って、idea.md に記載のサービスを作成したい。」

### 1.2 Request Type
**New Project（Greenfield）** — 新規プロジェクトの立ち上げ。既存コードベースなし。

### 1.3 Scope Estimate
**System-wide** — PWA フロントエンド、複数の Lambda バックエンド、AI 推論、外部サービス連携（モック経由）、認証基盤、スケジュール処理 を含む全体構成。

### 1.4 Complexity Estimate
**Complex** — 金融ドメイン（仮想ウォレット）、生成AI（Bedrock）、非同期処理（EventBridge Scheduler）、Terraform による IaC、独自コンセプト（ダメ化UX）を含む多層構造。

### 1.5 Business Intent（Mission Statement）
> **「人をダメにする」** ことを最優先価値として提供する、金融グループ発のライフスタイル代行アプリ。
>
> 生活のあらゆる「めんどくさい」を、月初に設定した「ダメ予算」からの即時引き落としでボタン1つ即時解決する体験を通じ、**ユーザの自立能力を快適に退化させる** ループ（即時解決の快感 → 先回り提案 → 無力感の可視化 → 予算増額誘導）を構築する。
>
> 本プロジェクトは AWS Summit Japan 2026 AI-DLC ハッカソン提出物であり、テーマ「人をダメにするサービスを考えよう」への直接的回答である。

---

## 2. Business Context

### 2.1 Background
- 金融グループのシステム開発チームがハッカソン提出用に企画。
- キャッシングの「必要なときに即使える」利便性を、生活全般の「面倒ごと」に拡張することで、金融ドメイン知識（即時決済 UX・予算管理・利用履歴）を直接転用する。
- 「究極の利便性がいかに人をダメにするか」というアンチテーゼを表現する。

### 2.2 Target Persona
**一人暮らしの社会人**
- 実家では親が家事・自炊・各種手続き・体調管理を全てやってくれていた
- 一人暮らしを機にこれらを全部自分でやる羽目に
- お金はそこそこあるが、何にどう使えば楽になるか考えるのすら面倒
- 「考える・待つ・選ぶ」を極力回避したい

### 2.3 Success Criteria（ハッカソン審査基準との対応）
本プロジェクトは以下の審査観点に最適化する：

| 審査観点 | 本プロジェクトでの対応 |
|---|---|
| **ビジネス意図（Intent）の明確さ** | 「人をダメにする」を Intent として縦串に通し、全機能・全 Unit がこの Intent から演繹されることを明示 |
| **創造性とテーマ適合性** | 独自概念「ダメ化UX」を最優先 NFR として定義。「ダメ予算」「先回り提案カード」「ダメ化メトリクス」「予算増額誘導」等の独自機能をテーマから導出 |
| **Unit 分解の適切さ** | 「ダメにする」物語のフェーズ1→2→3に呼応する形で Unit を分解（詳細は Units Generation で確定） |
| **ドキュメント品質** | Comprehensive depth で Inception フェーズを実施。全成果物を Intent の縦串で一貫させる |

### 2.4 Scope / Out of Scope

**In Scope**
- PWA フロントエンド（ブラウザベース）
- 仮想ウォレット（DynamoDB 上の数値残高管理）
- Amazon Bedrock (Claude 系) による「最適代行プラン提案」「先回り提案」
- 外部サービス連携のアダプタ層（モック実装のみ、本番実装のインターフェースは定義）
- 認証（Amazon Cognito メール+パスワード）
- 行動学習（利用履歴の蓄積と解析）
- 先回り提案（起動時サジェストカード表示）
- 月初の予算リセット（EventBridge Scheduler）
- Terraform による IaC
- 日本語・円建て・日本国内向け

**Out of Scope**
- 実外部サービス連携（UberEats / 出前館 / 家事代行 / 役所手続き代行等の実API接続）
- 実決済プロバイダ連携（Stripe / PAY.JP 等）
- プッシュ通知（Service Worker / Web Push / SNS / APNs / FCM）
- 音声入力（Amazon Transcribe / Web Speech API 等）
- ネイティブモバイルアプリ（iOS / Android）
- 多言語対応・多通貨対応
- PCI DSS / 個人情報保護法 / 金融庁ガイドラインへの厳格準拠（要点のみ記載）
- ソーシャルログイン（Google / Apple 等）
- マイクロサービス的な厳格な可用性設計（Saga / Circuit Breaker / 多段DLQ 等）

### 2.5 Assumptions
- AWS アカウントが利用可能で、Bedrock Claude モデルへのアクセスが有効化されている
- ハッカソン提出物としてのデモ環境を想定（プロダクション運用は想定しない）
- 審査員およびチームメンバー（数人〜数十人）が触るレベルの規模

### 2.6 Constraints
- **締切**: 2026-05-10 までに Inception フェーズを完了する必要あり
- **コンセプト**: テーマ「人をダメにするサービスを考えよう」への整合が最優先
- **技術**: 外部連携は AWS ネイティブサービスのみ実接続、それ以外はモック

---

## 3. Functional Requirements

機能要件は **ダメ化フェーズ1（ボタンを押すだけ）** → **フェーズ2（先回り提案）** → **フェーズ3（無力感の可視化・増額誘導）** の3層に対応させて記述する。

### 3.1 認証・ユーザー管理（FR-AUTH）

| ID | 要件 | 優先度 |
|---|---|---|
| FR-AUTH-01 | ユーザーは **メールアドレスとパスワード** で新規登録できる（Amazon Cognito User Pool） | MUST |
| FR-AUTH-02 | ユーザーはメール + パスワードでログインできる | MUST |
| FR-AUTH-03 | ログイン済みユーザーはセッションを維持したままアプリを利用できる（JWT） | MUST |
| FR-AUTH-04 | ユーザーはログアウトできる | MUST |
| FR-AUTH-05 | 未登録ユーザーには、ログイン前のランディング画面を表示する | SHOULD |

### 3.2 ダメ予算管理（FR-BUDGET） — フェーズ1基盤

| ID | 要件 | 優先度 |
|---|---|---|
| FR-BUDGET-01 | ユーザーは **月間のダメ予算（円）** を設定できる（初期値: 30,000円） | MUST |
| FR-BUDGET-02 | ユーザーはダメ予算を変更できる（増額・減額） | MUST |
| FR-BUDGET-03 | 仮想ウォレット残高は DynamoDB 上の数値として管理する（実決済は行わない） | MUST |
| FR-BUDGET-04 | 毎月1日 00:00 JST に、全ユーザーの残高を設定されたダメ予算値にリセットする（EventBridge Scheduler） | MUST |
| FR-BUDGET-05 | 残高は常時、メイン画面上部に **大きく表示** する（「残りダメ予算: ¥28,800」形式） | MUST |
| FR-BUDGET-06 | 残高不足時、代行手配の要求は失敗し、ユーザーに残高不足を通知する | MUST |
| FR-BUDGET-07 | 残高変動（減算・リセット）は履歴テーブルに記録する | MUST |

### 3.3 めんどくさい代行手配（FR-ORDER） — フェーズ1コア機能

本プロジェクトの最小ユースケース。

| ID | 要件 | 優先度 |
|---|---|---|
| FR-ORDER-01 | メイン画面中央に **「ご飯めんどくさい」ボタン** を大きく配置する | MUST |
| FR-ORDER-02 | ユーザーがボタンを押すと、フロントエンドはバックエンドの代行手配 API に注文要求を送る | MUST |
| FR-ORDER-03 | バックエンドは Bedrock（Claude）に対し、ユーザーの履歴・時刻・残予算を踏まえた最適なデリバリープラン（店舗・メニュー・金額想定）を問い合わせる | MUST |
| FR-ORDER-04 | バックエンドは `DeliveryAdapter` インターフェース経由で外部デリバリーサービスへ注文要求を送る。初期実装は `MockDeliveryAdapter` が固定応答を返す | MUST |
| FR-ORDER-05 | 注文応答受領時、ウォレットから注文金額を減算する（**二重引き落とし防止のため冪等性キーを利用**） | MUST |
| FR-ORDER-06 | 注文履歴に記録する（ユーザーID、カテゴリ、金額、店舗名、日時） | MUST |
| FR-ORDER-07 | ユーザーに完了画面を表示する（「注文完了: カレーハウスCoCo壱番屋 新宿店 ¥1,200 — 残りダメ予算 ¥28,800」） | MUST |
| FR-ORDER-08 | ボタン押下から完了画面表示までの **体感時間 3 秒以内** を目標とする（未達でも機能的には許容、将来改善） | SHOULD |
| FR-ORDER-09 | 将来拡張として、掃除・役所手続き等の他カテゴリを追加可能なアダプタ層設計とする（初期実装は「ご飯」カテゴリのみ） | SHOULD |

### 3.4 行動学習（FR-LEARNING） — フェーズ2基盤

| ID | 要件 | 優先度 |
|---|---|---|
| FR-LEARNING-01 | 代行手配が成功するたび、履歴テーブルに（ユーザーID、カテゴリ、時刻、曜日、金額、店舗名）を記録する | MUST |
| FR-LEARNING-02 | 履歴は少なくとも直近3ヶ月分を保持する（DynamoDB TTL で古いデータは自動削除） | SHOULD |

### 3.5 先回り提案（FR-SUGGEST） — フェーズ2コア

| ID | 要件 | 優先度 |
|---|---|---|
| FR-SUGGEST-01 | ユーザーがアプリを起動（メイン画面をロード）した際、バックエンドは履歴から「次にありうる代行要求」を Bedrock に推論させる | MUST |
| FR-SUGGEST-02 | 推論結果をメイン画面上部に **サジェストカード** として表示する（例: 「そろそろご飯めんどくさいですよね？」+ 1タップで注文ボタン） | MUST |
| FR-SUGGEST-03 | ユーザーはサジェストカードをタップすると、通常フローと同じ流れで代行手配が走る | MUST |
| FR-SUGGEST-04 | 履歴が不十分（例: 過去利用 0 回）な場合は、サジェストカードを表示せずデフォルトのカテゴリボタンのみ表示する | MUST |
| FR-SUGGEST-05 | プッシュ通知は実装しない（アプリ起動時にのみサジェストを提示） | MUST |

### 3.6 ダメ化メトリクス表示（FR-METRICS） — フェーズ3

| ID | 要件 | 優先度 |
|---|---|---|
| FR-METRICS-01 | メイン画面に「今月のダメ化回数」（代行手配回数）を表示する | MUST |
| FR-METRICS-02 | メイン画面に「ダメ予算消化率」（%）を表示する | MUST |
| FR-METRICS-03 | 予算消化率が 80% を超えた時、強調表示（色変化等）でユーザーの不安を煽る | SHOULD |
| FR-METRICS-04 | 残高 0 円到達時、「今月ダメになれません」画面とともに **翌月予算の増額提案モーダル** を表示する | MUST |
| FR-METRICS-05 | ユーザーは増額提案モーダルから翌月予算を変更できる（FR-BUDGET-02 との統合） | MUST |

### 3.7 一貫したユーザー体験（FR-UX）

「ダメ化UX」最優先方針の具体化。

| ID | 要件 | 優先度 |
|---|---|---|
| FR-UX-01 | ユーザーは「考える」「入力する」「選ぶ」を最小限とする。カテゴリボタン押下1回で代行手配を完結する | MUST |
| FR-UX-02 | 音声・テキスト自由入力は**実装しない**（全てボタンのみで完結） | MUST |
| FR-UX-03 | 代行手配の完了時、店舗・メニュー等の詳細決定は Bedrock がユーザーに確認することなく自動決定する | MUST |
| FR-UX-04 | メイン画面に戻る経路は **「完了 → 自動的にメインへ戻る」** を基本とし、ユーザーが明示的に操作する要素を最小化する | SHOULD |

---

## 4. Non-Functional Requirements

### 4.1 最優先 NFR: ダメ化UX（DegenerationQuality）

本プロジェクトのすべての NFR は、この **ダメ化UX** への貢献度で評価される。

| ID | 要件 | 根拠 |
|---|---|---|
| NFR-DEG-01 | ユーザーのアクション数（タップ/クリック）を 1 回の代行手配あたり最大 2 回以内とする | 認知負荷ゼロ |
| NFR-DEG-02 | アプリ起動時、ユーザーがボタンを押す前にサジェストカードが提示される | フェーズ2（先回り）の体現 |
| NFR-DEG-03 | 残高・消化率・今月のダメ化回数は常時可視化される | フェーズ3（無力感の可視化） |
| NFR-DEG-04 | 残高枯渇時、翌月予算の増額誘導を強めに提示する | 退化ループの完成 |
| NFR-DEG-05 | UI コピーは自虐的・依存促進的な文言を用いる（例:「残りダメ予算」「そろそろ◯◯めんどくさいですよね？」「今月ダメになれません」） | テーマ体現 |

### 4.2 パフォーマンス

| ID | 要件 | 目標値 |
|---|---|---|
| NFR-PERF-01 | ボタン押下から完了表示までの体感時間 | **3 秒以内**（目標） |
| NFR-PERF-02 | メイン画面の初期表示時間（サジェストカード含む） | 5 秒以内（目標） |
| NFR-PERF-03 | 想定同時利用者数 | **数人〜数十人**（デモ規模） |

### 4.3 可用性

| ID | 要件 |
|---|---|
| NFR-AVAIL-01 | AWS マネージドサービスの標準可用性に準拠（追加の冗長構成は実装しない） |
| NFR-AVAIL-02 | 本番相当の SLA は定めない（デモ用途のため） |

### 4.4 信頼性・データ整合性

| ID | 要件 |
|---|---|
| NFR-REL-01 | **仮想ウォレット残高が負の値にならない**ことをアプリケーションロジックで保証する（DynamoDB 条件付き書き込み） |
| NFR-REL-02 | **同一注文リクエストの二重処理（二重引き落とし）を防止する**（冪等性キーを利用） |
| NFR-REL-03 | 残高変動と注文履歴の記録は、可能な限りアトミックに実行する（DynamoDB TransactWriteItems の利用を検討） |
| NFR-REL-04 | 月初リセットが失敗した場合、手動リカバリ手順をドキュメント化する（自動リトライは実装しない） |

### 4.5 セキュリティ

**Security Extension は opt-out**（ブロッキング制約としての強制はしない）。以下は最低限の方針のみ記載する。

| ID | 要件 |
|---|---|
| NFR-SEC-01 | 認証は Amazon Cognito User Pool により管理し、バックエンド API は JWT 検証を必須とする |
| NFR-SEC-02 | データは AWS マネージドサービスのデフォルト暗号化を利用する（KMS 顧客管理キーは使用しない） |
| NFR-SEC-03 | API Gateway にレート制限を最小限設定する（任意） |
| NFR-SEC-04 | 本番運用時に考慮すべき事項（HTTPS 強制、WAF、CloudTrail 監査ログ、PCI DSS 準拠等）は将来対応として設計書に明記する |

### 4.6 コンプライアンス（要点のみ）

| ID | 要件 |
|---|---|
| NFR-COMP-01 | 本アプリは仮想ウォレット（数値管理のみ）を採用し、実カード情報・決済情報を扱わない。したがって PCI DSS の直接的対象外とする |
| NFR-COMP-02 | 個人情報は「メールアドレス」のみ取得する（Cognito User Pool 内で管理） |
| NFR-COMP-03 | 本番運用時は、個人情報保護法・金融庁ガイドラインへの適合性評価が別途必要である旨を設計書に注記する |

### 4.7 スケーラビリティ

| ID | 要件 |
|---|---|
| NFR-SCALE-01 | AWS サーバレス（Lambda + DynamoDB オンデマンド）の自動スケーリング特性に依存する |
| NFR-SCALE-02 | 本 MVP では明示的なスケール設計（プロビジョンドキャパシティ等）は行わない |

### 4.8 保守性・拡張性

| ID | 要件 |
|---|---|
| NFR-MAINT-01 | 外部サービス連携は **アダプタ層パターン** で抽象化し、`MockDeliveryAdapter` を `UberEatsDeliveryAdapter` 等の実装に差し替え可能とする |
| NFR-MAINT-02 | Terraform モジュール設計のプロジェクト規約（`terraform-plugin:terraform-module-design`）に準拠する |
| NFR-MAINT-03 | Terraform コーディング規約（`terraform-plugin:terraform-coding-rule`）に準拠する |

### 4.9 テスタビリティ

| ID | 要件 |
|---|---|
| NFR-TEST-01 | ビジネスロジックの純粋関数（残高計算・履歴集計等）に対し **Property-Based Testing** を適用する |
| NFR-TEST-02 | シリアライゼーション（JSON ラウンドトリップ）のテストには PBT を適用する |
| NFR-TEST-03 | 以下の不変条件を PBT で検証する: 「残高は負にならない」「同一冪等性キーの二重送信で残高は1回分しか減らない」「履歴合計引き落とし額 == 初期残高 - 現在残高」 |
| NFR-TEST-04 | 上記以外の IO・外部呼び出し部分は通常のユニットテスト/統合テストとする |

### 4.10 観測性

| ID | 要件 |
|---|---|
| NFR-OBS-01 | Lambda のログは CloudWatch Logs に出力する（構造化ログ推奨） |
| NFR-OBS-02 | 本 MVP では X-Ray・カスタムメトリクス等の高度な観測性は実装しない |

### 4.11 アクセシビリティ

| ID | 要件 |
|---|---|
| NFR-A11Y-01 | 基本的なセマンティック HTML（button, form 等）を利用する |
| NFR-A11Y-02 | 本 MVP では WCAG 準拠チェックは実施しない |

---

## 5. Technical Context

### 5.1 Frontend
- **形態**: レスポンシブ Web アプリ（PWA）
- **想定技術**: React / Next.js 等（詳細は Construction フェーズで確定）
- **配信**: Amazon S3 + CloudFront
- **入力手段**: ボタンのみ（音声・自由入力なし）

### 5.2 Backend
- **アーキテクチャ**: AWS サーバレス中心
- **API**: Amazon API Gateway + AWS Lambda
- **永続化**: Amazon DynamoDB
- **AI推論**: Amazon Bedrock（Claude 系モデル、Converse API を Lambda から直接呼び出し）
- **認証**: Amazon Cognito User Pool
- **スケジュール処理**: Amazon EventBridge Scheduler（月初の予算リセット）

### 5.3 Infrastructure as Code
- **ツール**: **Terraform**
- **準拠規約**:
  - `terraform-plugin:terraform-coding-rule`
  - `terraform-plugin:terraform-module-design`
  - `checkov-skip-rule-plugin:checkov-skip-rule`（Checkov のスキップルール運用）
  - `parameter-sheet-format-plugin:parameter-sheet-format`（パラメータシート作成）
  - `aws-cost-estimate-plugin:aws-cost-estimate`（コスト見積書作成）

### 5.4 External Integrations (Mocked)
以下はすべて **モック実装** とする。インターフェース（アダプタ層）だけ定義し、将来の差し替えポイントを残す。

- デリバリー: `DeliveryAdapter`（実装: `MockDeliveryAdapter`。将来: UberEats / 出前館 等）
- 清掃 / 手続き代行等: 初期スコープでは未実装。MVP は「ご飯」カテゴリのみ

### 5.5 Region
- プライマリ: `ap-northeast-1`（東京）
- マルチリージョン構成は行わない

---

## 6. User Scenarios（Happy Path と主要エッジケース）

### 6.1 Happy Path: 「ご飯めんどくさい」→ 代行手配完了

1. ユーザーが PWA にアクセスし、Cognito でログイン済み
2. メイン画面表示時、起動時サジェスト API が呼ばれ、Bedrock が履歴から「そろそろご飯めんどくさいですよね？」を生成
3. ユーザーはサジェストカードをタップ（または中央の「ご飯めんどくさい」ボタンを押す）
4. フロント → API Gateway → `orderHandler` Lambda
5. `orderHandler` は Bedrock で最適なデリバリープランを生成（例: カレーハウスCoCo壱番屋 新宿店、カレーライス、¥1,200）
6. `orderHandler` は `MockDeliveryAdapter.placeOrder()` を呼び出し、固定応答（`orderId`, `status=accepted` 等）を受領
7. `orderHandler` は DynamoDB で「ウォレット残高減算 + 注文履歴記録」を冪等性キー付きで実行
8. ユーザーに完了画面を表示（注文内容 + 残予算 + 今月のダメ化回数）
9. 数秒後、メイン画面に自動遷移

### 6.2 エッジケース: 残高不足

1. ユーザーが「ご飯めんどくさい」ボタンを押す
2. `orderHandler` が残高を検証し、推定注文額が残高を上回る
3. ユーザーに「今月ダメになれません」画面を表示し、翌月予算の増額提案モーダルを出す

### 6.3 エッジケース: 履歴なしの初回利用

1. 新規ユーザーが初ログイン
2. メイン画面のサジェスト API は「履歴なし」判定でサジェストカードを出さない
3. デフォルトの「ご飯めんどくさい」ボタンのみ表示

### 6.4 エッジケース: 月初リセット

1. 2026-06-01 00:00 JST、EventBridge Scheduler が `monthlyResetHandler` Lambda を起動
2. 全ユーザーの残高を設定された月間予算値にリセット
3. リセット操作を履歴テーブルに記録

### 6.5 エッジケース: 二重送信（ネットワーク不安定 / ボタン連打）

1. ユーザーが「ご飯めんどくさい」ボタンを連打
2. フロントは同一の冪等性キーで API 呼び出しを多重送信
3. `orderHandler` は冪等性キーをチェックし、2 回目以降は前回の結果を返す（残高は 1 回分しか減らない）

---

## 7. Quality Attributes Summary

| 品質特性 | 優先度 | 備考 |
|---|---|---|
| **ダメ化UX（独自）** | ★★★★★ | 本プロジェクトの魂。他のすべてがこれに奉仕 |
| 開発スピード | ★★★★ | 2026-05-10 締切（Inception 完了） |
| 信頼性（残高不変・冪等性） | ★★★ | 金融ドメインの最低限保証 |
| UX（即応性・シンプル） | ★★★ | ダメ化UXの表層 |
| 拡張性（アダプタ層） | ★★★ | 審査の「Unit分解」観点に寄与 |
| パフォーマンス | ★★ | デモ規模で十分 |
| セキュリティ | ★★ | 本番考慮は将来事項として記載のみ |
| 可用性 | ★★ | AWS 標準に依存 |
| コンプライアンス | ★ | 要点のみ記載 |
| 多言語対応 | - | スコープ外 |

---

## 8. Extension Configuration

| Extension | Enabled | Reason |
|---|---|---|
| Security Baseline | **No** | ハッカソン / PoC 方針。Q16 で opt-out |
| Property-Based Testing | **Partial** | 純粋関数・シリアライゼーションのみに適用。残高不変条件・冪等性・履歴整合を PBT で検証。Q17 で Partial |

---

## 9. Traceability: 審査観点への対応マップ

| 審査観点 | 関連する本ドキュメントのセクション |
|---|---|
| ビジネス意図の明確さ | §1.5 Business Intent、§2 Business Context、§4.1 最優先NFR |
| 創造性とテーマ適合性 | §4.1 ダメ化UX、§3.5 先回り提案、§3.6 ダメ化メトリクス表示、§3.7 FR-UX |
| Unit 分解の適切さ | §3 Functional Requirements のセクション分割（AUTH / BUDGET / ORDER / LEARNING / SUGGEST / METRICS / UX）、Units Generation ステージで確定 |
| ドキュメント品質 | 本ドキュメント全体（Comprehensive depth）、続く User Stories / Workflow Planning / Application Design / Units Generation のドキュメント一貫性 |

---

## 10. Open Questions / Future Considerations

以下は本 MVP のスコープ外だが、将来検討として記載する：

- Bedrock Agents への置き換えによる複数カテゴリ横断エージェント化
- プッシュ通知対応（Service Worker + Web Push + SNS）
- 実決済プロバイダ連携（Stripe Test モード等）
- 実外部サービス API 連携（UberEats / 出前館 等）
- 音声入力（Amazon Transcribe）
- ネイティブモバイルアプリ化（iOS / Android）
- ソーシャルログイン（Apple / Google）
- 多言語・多通貨対応
- PCI DSS / 個人情報保護法 / 金融庁ガイドラインへの厳格準拠
- Saga パターン / Circuit Breaker / 多段DLQ 等の高信頼構成
- CloudTrail 監査ログ / X-Ray トレーシング / カスタムメトリクス
- WCAG 準拠チェック

---

## 11. Requirements Summary

本 MVP は、**「ご飯めんどくさい」ボタン押下 → Bedrock で最適プラン提案 → モックアダプタ経由で代行手配 → 仮想ウォレットから減算 → 完了表示** のエンドツーエンド動線を中核に、**行動学習による先回り提案（起動時サジェストカード）** と **ダメ化メトリクス・予算増額誘導** を組み合わせることで、idea.md のビジョンとテーマ「人をダメにする」を忠実に体現する。

インフラは AWS サーバレス（Lambda + API Gateway + DynamoDB + Cognito + Bedrock + EventBridge Scheduler）、IaC は Terraform。PWA + S3 + CloudFront によるフロントエンド配信。

Inception フェーズのドキュメント群は 2026-05-10 までに完成させ、審査観点「ビジネス意図の明確さ」「創造性とテーマ適合性」「Unit 分解の適切さ」「ドキュメント品質」のすべてを満たすことを目指す。
