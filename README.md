# ゴロゴロPay

> **ボタン 1 つで人生が回る → ボタンすら不要になる → 完全に人がダメになる**

「めんどくさい」を押すだけで、お金の力で全部解決してくれるアプリ。
ボタン 1 つで生活のあらゆる面倒を即時キャッシュで解消する、金融グループ発の **人をダメにするサービス**。

**AWS Summit Japan 2026 AI-DLC ハッカソン** 応募プロジェクト。

---

## コンセプト

### ペルソナ
一人暮らしの社会人「佐藤陽介」（27 歳・Web ディレクター）。実家では親が全部やってくれていたが、一人暮らしになり家事・自炊・各種手続き・体調管理を全部自分でやる羽目に。お金はそこそこあるが、何にどう使えば楽になるか考えるのすら面倒。詳細は [aidlc-docs/inception/requirements/requirements.md §2.2](./aidlc-docs/inception/requirements/requirements.md) を参照。

### 仕組み
1. 毎月の「ダメ予算」を設定（初期値: 月 30,000 円）
2. アプリを開くと画面中央にでっかいボタンが 1 つ
3. 「ご飯めんどくさい」を押す → 予算から即時引き落として代行サービスを自動手配
4. 完了

### 3 つのフェーズ — ユーザを退化させるループ

| フェーズ | 体験 | 実現機能 |
|---|---|---|
| **フェーズ 1** — ボタンを押すだけ | 決定疲れからの解放、即時解決の快感 | 1 タップで Bedrock が最適プランを推論、仮想ウォレットから引き落とし |
| **フェーズ 2** — ボタンすら不要 | AI が自分より自分を知り、依存が深まる | 起動時サジェスト（「そろそろご飯めんどくさいですよね？」）、1 タップ受諾 |
| **フェーズ 3** — 完全に人がダメになる | 無力感の可視化、翌月予算の増額誘導 | ダメ化メトリクス表示、予算 0 時の増額誘導モーダル |

---

## 技術スタック

### Backend (Application)
- **Go + Gin + Lambda Web Adapter**（モノリシック API Lambda + Scheduler Lambda の 2 Lambda 構成）

### Frontend
- **Next.js (App Router)**
- **Jotai** + **TanStack Query** — 状態管理
- **REST (JSON over HTTPS)** で Lambda + Gin と通信

### Infrastructure (AWS)
- **AWS Lambda** — API Lambda + Scheduler Lambda（コンテナイメージ）
- **Amazon API Gateway** (REST) + Cognito Authorizer
- **Amazon DynamoDB**（5 テーブル）— 仮想ウォレット・予算・履歴・冪等性・リセットログ
- **Amazon Cognito** User Pool — メール+パスワード認証
- **Amazon Bedrock**（Claude Converse API）— 注文プラン推論・先回り提案
- **Amazon EventBridge Scheduler** — 月初リセット（cron: 毎月 1 日 00:00 JST）
- **AWS Amplify Hosting** — PWA 配信

### Infrastructure as Code
- **Terraform** — モジュール構成（`lambda_api` / `lambda_scheduler` / `cognito` / `dynamodb` / `amplify` / `api_gateway` / `bedrock`）

---

## ライセンス

Unlicensed — ハッカソン提出用プロジェクト
