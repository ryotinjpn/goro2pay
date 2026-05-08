# ゴロゴロPay — Requirements Verification Questions

以下の質問に、各設問の `[Answer]:` タグの右に **A / B / C / D / E / X** の文字で回答してください。  
`X) Other` を選んだ場合は `[Answer]: X` の後に説明を追記してください。  
回答が終わったら「完了しました」などと教えてください。

> **前提**: 本プロジェクトは AWS Summit Japan 2026 の AI-DLC ハッカソン向けです。この前提が変わる場合はコメントしてください。

---

## 1. プロジェクトのゴール / 成果物

### Question 1
本プロジェクトで最終的に提出・披露したい成果物はどれですか？

A) 動作するデモ可能な MVP（エンドツーエンドで1ユースケースが動く）
B) 概念実証 PoC（UIと簡易バックエンドのみ、外部連携はすべてモック）
C) 本番運用を想定したプロダクション品質の初版
D) 企画・設計資料のみ（コード生成は不要）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — 方針: AWS ネイティブサービス（Bedrock, Cognito, Lambda, DynamoDB, API Gateway 等）は実連携。UberEats / Stripe / 出前館 / 家事代行 / 役所手続き代行など非AWSの外部サービスはすべてモック。決済は外部プロバイダを使わず、DynamoDB上の仮想ウォレット残高で「ダメ予算」を表現する。

---

### Question 2
ハッカソン提出までに実装すべき「最小ユースケース」はどれですか？（1つだけ選ぶなら）

A) 「ご飯めんどくさい」→ デリバリー注文（モックでも可）+ 予算引き落とし
B) 「掃除めんどくさい」→ 清掃サービス手配 + 予算引き落とし
C) 「何でも」→ AI が内容を解釈して複数カテゴリから最適代行手配
D) ボタンを押すだけで固定の1アクションが走る最小導線
X) Other (please describe after [Answer]: tag below)

[Answer]: A

---

## 2. ターゲットプラットフォーム

### Question 3
主たるフロントエンド形態はどれですか？

A) iOS ネイティブアプリ（Swift / SwiftUI）
B) Android ネイティブアプリ（Kotlin）
C) クロスプラットフォームモバイル（React Native / Flutter）
D) レスポンシブ Web アプリ（PWA）
X) Other (please describe after [Answer]: tag below)

[Answer]: D

---

### Question 4
音声入力（「ご飯めんどくさい…」を声で伝える）はスコープに含めますか？

A) 必須 — 初版から音声入力をサポート
B) 任意 — テキスト入力を主とし、時間があれば音声も追加
C) 不要 — テキスト入力のみ
X) Other (please describe after [Answer]: tag below)

[Answer]: X — 音声もテキスト入力も不要。操作はすべてボタンのみ（カテゴリボタンを押す方式）。

---

## 3. AI / エージェント

### Question 5
「めんどくさい」の内容を解釈して最適なサービスを選ぶAIエンジンは何を使いますか？

A) Amazon Bedrock（Claude 系モデル）
B) Amazon Bedrock Agents（ツール呼び出しを含むエージェント）
C) 自前ルールベース + 簡易分類（AIは最小限）
D) 選定はAI-DLC側に任せる（推奨を提示してほしい）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — Amazon Bedrock（Claude系モデル）の Converse API を Lambda から直接呼び出す方式。

---

### Question 6
「先回り提案（フェーズ2: ボタンすら押さない）」機能の扱いは？

A) 必須 — 初版で行動学習と通知プッシュまで実装
B) 設計だけ含める — 実装は今回スコープ外、将来対応と明記
C) 今回は考慮しない（ボタンを押す体験のみ）
X) Other (please describe after [Answer]: tag below)

[Answer]: X — 行動学習と「先回り提案の表示」は実装する（履歴を DynamoDB に保存、Bedrock で解析、アプリ起動時にメイン画面にサジェストカードを表示）。ただしプッシュ通知（Service Worker / Web Push / SNS）は実装しない。ユーザがアプリを開いたタイミングでのみサジェストを提示する。

---

## 4. 決済・金融機能

### Question 7
「ダメ予算」からの引き落としは、実際の決済を実行しますか？

A) 実決済 — Stripe / PAY.JP など本物の決済APIに接続
B) 仮想ウォレット — アプリ内の仮想残高を減らすだけ（DB上の数値）
C) ハイブリッド — UIは決済風だが裏はモック、決済接続は設計のみ
X) Other (please describe after [Answer]: tag below)

[Answer]: B — DynamoDB上の仮想ウォレット残高を減算する方式。月初(毎月1日)にEventBridge Schedulerで全ユーザの残高を「ダメ予算」設定値にリセット。

---

### Question 8
金融グループのシステム開発チームという設定上、どの規制・ガイドライン準拠を想定しますか？

A) 準拠を設計に明記する（PCI DSS / 個人情報保護法 / 金融庁ガイドライン）
B) 触れる程度で可（ハッカソン向けなので要点のみ記載）
C) 今回は対象外（演出のためのストーリーに過ぎない）
X) Other (please describe after [Answer]: tag below)

[Answer]: B — PCI DSS / 個人情報保護法 / 金融庁ガイドラインへの厳格な対応は定めない。要件書・設計書には「本番運用時はここを考慮すべき」という将来留意事項として要点のみ記載する。

---

## 5. 外部サービス連携

### Question 9
デリバリー / 清掃 / 手続き代行などの外部サービス連携はどう扱いますか？

A) 実APIに接続（UberEats API 等、実在するものだけ使う）
B) すべてモック（外部連携は「発注しました」のスタブ応答）
C) アダプタ層を作り、実装は一部モック/一部実API（混在）
X) Other (please describe after [Answer]: tag below)

[Answer]: A(方針変更) — アダプタ層パターンを採用。DeliveryAdapter 等のインターフェースを定義し、MockDeliveryAdapter が固定応答を返す実装を用意する。実API接続用の実装（UberEatsDeliveryAdapter 等）は未実装のままインターフェースの差し替えポイントだけ設計に残す。

---

## 6. ユーザー認証

### Question 10
ユーザー認証はどう構成しますか？

A) Amazon Cognito（メール + パスワード）
B) Amazon Cognito + ソーシャルログイン（Google / Apple）
C) 認証なし（デモ用固定ユーザー）
D) 選定はAI-DLC側に任せる
X) Other (please describe after [Answer]: tag below)

[Answer]: A — Amazon Cognito User Pool でメール + パスワード認証を提供する。

---

## 7. インフラ / アーキテクチャ

### Question 11
インフラの基本方針はどれですか？

A) AWS サーバレス中心（Lambda + API Gateway + DynamoDB + Cognito + Bedrock）
B) AWS コンテナ中心（ECS / Fargate + RDS）
C) AWS Amplify Gen2 ベースのフルスタック
D) 選定はAI-DLC側に任せる（コスト・スピード重視で推奨）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — AWS サーバレス中心構成。API Gateway + Lambda + DynamoDB + Cognito + Bedrock + EventBridge Scheduler。PWA フロントエンドは AWS Amplify Hosting に統一してホストする。

---

### Question 12
IaC（Infrastructure as Code）の方針は？

A) AWS CDK (TypeScript)
B) AWS SAM
C) Terraform
D) 選定はAI-DLC側に任せる
X) Other (please describe after [Answer]: tag below)

[Answer]: C — Terraform を採用する。

---

## 8. 非機能要件の重み

### Question 13
以下のうち、**最優先**したい品質特性は？

A) 開発スピード（ハッカソンで動くことが最優先）
B) UX（「ボタン1つ」のシンプルさと即応性）
C) 信頼性（金融アプリらしい堅牢さ）
D) 拡張性（将来のフェーズ2・3への発展余地）
X) Other (please describe after [Answer]: tag below)

[Answer]: X — 「ダメ化UX」を最優先品質特性として定義する。ユーザが「考える」「待つ」「選ぶ」行為を最小化し、即時解決の快感 → 先回り提案 → 無力感の可視化・予算増額誘導 という "快適さによる自立退化" のループを最優先する。スピード・UX・信頼性・拡張性はすべてこのダメ化UXに奉仕する手段として扱う。idea.md のビジネス意図「人をダメにする」とテーマ「人をダメにするサービスを考えよう」に直接整合する。

---

### Question 14
想定する初期ユーザー規模・レスポンス要件は？

A) デモ用（数人〜数十人、パフォーマンス要件ほぼ無視）
B) 小規模運用想定（〜1,000MAU、API応答 < 2秒）
C) 中規模想定（〜10万MAU、API応答 < 500ms、スケール設計必要）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — デモ用途（数人〜数十人）。厳密なパフォーマンス要件は設けない。ただし「ダメ化UX」の体感を守るため、ボタン押下から完了表示までの体感速度目標として「3秒以内」を非機能要件に明記する（未達でも機能的には許容、将来改善）。

---

## 9. 地域 / 言語

### Question 15
対象地域・言語は？

A) 日本のみ / 日本語のみ（円建て、国内サービス前提）
B) 日本優先 + 多言語対応の余地を残す
C) グローバル展開を想定
X) Other (please describe after [Answer]: tag below)

[Answer]: A — 日本国内のみ。UI・データ・通貨はすべて日本語・円建てで統一する。

---

## 10. 拡張機能（Extensions）

### Question 16: Security Extensions
Should security extension rules be enforced for this project?

A) Yes — enforce all SECURITY rules as blocking constraints (recommended for production-grade applications)
B) No — skip all SECURITY rules (suitable for PoCs, prototypes, and experimental projects)
X) Other (please describe after [Answer]: tag below)

[Answer]: B — Security Extension ルールは強制しない（PoC・ハッカソン向け）。AWSデフォルトのセキュリティ設定は踏襲するが、ブロッキング制約としての強制は行わない。

---

### Question 17: Property-Based Testing Extension
Should property-based testing (PBT) rules be enforced for this project?

A) Yes — enforce all PBT rules as blocking constraints (recommended for projects with business logic, data transformations, serialization, or stateful components)
B) Partial — enforce PBT rules only for pure functions and serialization round-trips (suitable for projects with limited algorithmic complexity)
C) No — skip all PBT rules (suitable for simple CRUD applications, UI-only projects, or thin integration layers with no significant business logic)
X) Other (please describe after [Answer]: tag below)

[Answer]: B — 純粋関数（残高計算・履歴集計等）とシリアライゼーション（JSONラウンドトリップ）に対してのみ PBT を適用する。特に「残高は負にならない」「二重引き落とし防止（冪等性）」「履歴合計と残高差の整合」等の不変条件をプロパティベースで検証する。それ以外のI/O・外部呼び出し部分はPBTを強制しない。

---

## 11. その他の制約

### Question 18
締め切り・制約で AI-DLC 側が知っておくべきことは？

A) 特になし、通常通り進めてよい
B) 締め切りが近く、スピード最優先（詳細設計は最小限）
C) チーム開発のため、ドキュメント重視（詳細な要件・設計が必要）
X) Other (please describe after [Answer]: tag below)

[Answer]: X — 締め切り制約: 2026-05-10 までに Inception フェーズ完了が必須（ハッカソン応募要件: aidlc-state.md で Inception フェーズ完了、かつ Inception フェーズで実施すべき成果物が揃っていること）。審査観点「ドキュメントの品質」を踏まえ、Inception フェーズのドキュメント（要件書・ユーザーストーリー・ワークフロー計画・アプリケーション設計・Unit 分解）は丁寧に整備する。Construction フェーズ以降は締め切り後に進めるため、本質問の時点では Inception 完了を最優先とする。
