---
marp: true
theme: default
paginate: true
style: |
  :root {
    --bg-deep: #050402;
    --bg-mid: #0e0a06;
    --gold-100: #f4d990;
    --gold-500: #c9a96b;
    --cream: #f4ecd8;
    --mute: #6a5a3a;
    --danger: #ff4f4f;
  }
  section {
    background: var(--bg-deep);
    color: var(--cream);
    font-family: 'Hiragino Kaku Gothic ProN', 'Noto Sans JP', sans-serif;
    padding: 48px 64px;
  }
  section::after {
    color: var(--mute);
    font-size: 14px;
  }
  h1 {
    color: var(--gold-100);
    font-size: 2.4em;
    font-weight: 900;
    border-bottom: 2px solid var(--gold-500);
    padding-bottom: 12px;
    margin-bottom: 0.6em;
  }
  h2 {
    color: var(--gold-500);
    font-size: 1.5em;
    font-weight: 700;
    margin-bottom: 0.8em;
  }
  h3 {
    color: var(--gold-100);
    font-size: 1.2em;
  }
  p {
    font-size: 1.05em;
    line-height: 1.8;
  }
  ul, ol {
    line-height: 2;
    font-size: 1.05em;
  }
  li::marker {
    color: var(--gold-500);
  }
  strong {
    color: var(--gold-100);
  }
  code {
    background: #1a1208;
    color: var(--gold-100);
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.9em;
  }
  table {
    border-collapse: collapse;
    width: 100%;
    font-size: 0.95em;
  }
  th {
    background: #2a1e08;
    color: var(--gold-100);
    padding: 10px 16px;
    border: 1px solid var(--gold-500);
    font-weight: 700;
  }
  td {
    padding: 10px 16px;
    border: 1px solid #3a2a14;
    color: var(--cream);
    background: #0e0a06;
  }
  tr:nth-child(even) td {
    background: #1a1208;
  }
  tr:has(th:empty) {
    display: none;
  }
  .lead h1 {
    font-size: 3.2em;
    border: none;
    text-align: center;
    margin-bottom: 0.2em;
  }
  .lead p {
    text-align: center;
    color: var(--mute);
    font-size: 1em;
  }
  .impact h1 {
    font-size: 2.8em;
    border: none;
    text-align: center;
    color: var(--cream);
    margin-top: 0.4em;
  }
  .impact h2 {
    text-align: center;
    font-size: 1.1em;
    color: var(--mute);
    font-weight: 400;
  }
  .phase h2 {
    font-size: 1.1em;
    color: var(--mute);
    font-weight: 400;
    margin-bottom: 0.4em;
  }
  .phase h1 {
    font-size: 2em;
    border: none;
    margin-bottom: 0.6em;
  }
  blockquote {
    border-left: 3px solid var(--gold-500);
    padding-left: 20px;
    color: #a89070;
    font-style: normal;
    font-size: 1em;
    line-height: 1.9;
    margin: 0.5em 0;
    background: none;
  }
  blockquote p {
    color: #a89070;
  }
  .arch p {
    font-size: 0.9em;
    line-height: 1.7;
  }
  .final h1 {
    font-size: 3em;
    border: none;
    text-align: center;
    color: var(--gold-100);
    margin-bottom: 0.2em;
  }
  .final p {
    text-align: center;
    color: var(--mute);
  }
---

<!-- _class: lead -->

# ゴロゴロPay

AWS Summit Japan 2026 AI-DLC ハッカソン 予選会

---

<!-- _class: impact -->

## Opening

# 人がダメになるとは、<br>どういうことでしょうか。

---

# 考えるという行為を<br>やめた時です。

食べたいものを考えない。
今日着ていく服を考えない。
どう動けばいいかを考えない。

---

# 35,000回

現代人が1日に行う意思決定の数

---

## たとえば

- 今夜の献立
- 今日着ていく服
- 目的地までの交通手段
- 週末の清掃

これらが、毎日積み重なる。

---

<!-- _class: impact -->

## 意思決定は

# 人を育てる。<br>そして、削る。

---

## 私たちの思想

**不要な意思決定は、すべてAIへ。**

空いた時間と気力は、
自己投資や趣味など**大切なことへ。**

---

<!-- _class: lead -->

# ゴロゴロPay

*「めんどくさい」をお金で即時解決する*

---

## どう動くか

1. 毎月「ダメになる予算」を設定
2. ボタンを押す
3. **3秒後、完了通知**

それだけ。

---

## 使うほど、考えなくなる

| フェーズ | 状態 |
|---|---|
| 使い始め | 何を頼むかを、AIが決める |
| 2ヶ月後 | ボタンすら不要。AIが先回り |
| 月末 | 何も思い浮かばない。増額する |

**ボタン1つで人生が回る。人がダメになる。**

---

## ゴロゴロ太郎

| | |
|---|---|
| 年齢 | 27歳・一人暮らし |
| 職業 | 都内IT企業 Webディレクター |
| 年収 | 550万円 |
| 特徴 | 今自分が何をしたいか、思い浮かばない |

> 「趣味は？」と聞かれて、答えられない。

---

<!-- _class: phase -->

## Phase 1

# 何を食べるか、をAIに委ねる

> 「考えなくていい。選ばなくていい。
> ボタンを押すだけで全部終わる。
> 1,200円でこれが買えるなら——安い。」

**→ 意思決定疲れからの解放**

---

<!-- _class: phase -->

## Phase 2

# いつ食べるのか、もAIに委ねる

> 「毎日夕飯の時間になるとご飯が出てくる。
> まるで実家で生活しているようだ。」

サジェストが出ない日に、**不安を感じる。**

**→ 自己決定権の放棄**

---

<!-- _class: phase -->

## Phase 3

# 何も考えられなくなる

> 「あれ……夕飯ってどうやって
> 準備したらいいんだっけ……？」

何も思い浮かばない。
**迷わず、増額ボタンを押す。**

---

<!-- _class: impact -->

## Demo

# 実際に動くものを<br>ご覧ください。

---

## 画面の流れ

| 状態 | 概要 |
|---|---|
| IDLE | ボタン1つ。「押せ。考えるな。」 |
| SUGGESTED | アプリが先回り。「そろそろだろ。」 |
| SLOT | リールが回り、Bedrockが選ぶ |
| COMPLETE | 「いい判断だ。」 |
| DEAD | 画面が脱色。「今月は、終わりだ。」 |

---

## AI-DLC プロセス

5 Unit に分解し、全フェーズを実践

`Auth` / `Budget` / `Order` / `Suggest` / `Metrics`

各 Unit を独立 PR で管理。
ドキュメントから実装まで**一気通貫。**

---

## 工夫① ペルソナと設計の1対1対応

太郎の3フェーズを先に深く設計。

各 Unit・各画面・各コピーが
どのフェーズに対応するかを**トレーサブルに。**

> 「なぜこのUI、なぜこのコピーか」が
> 全て太郎の感情変遷から導かれる。

---

## 工夫② デザインシステムを横串成果物として独立

`_design-system/design-spec.md` に一本化。

- フォント / カラートークン
- コピーシステム
- 状態モデル（idle → suggested → slot → dead）

**実装担当者が迷わず動ける構造。**

---

## 工夫③ NFRに「ダメ化UX」を独自定義

> 一般的なNFRは性能・セキュリティ。
> 私たちは**ビジネスアイデアそのものをNFRに。**

| コード | 内容 |
|---|---|
| NFR-DEG-02 | 起動時サジェスト必須 |
| NFR-DEG-05 | 依存促進的コピー必須 |

---

<!-- _class: arch -->

## 技術アーキテクチャ（フルサーバーレス）

**Frontend**: Next.js PWA — AWS Amplify Hosting
**Backend**: Go + Gin — Lambda Web Adapter
**AI**: Amazon Bedrock（Lambda から直接呼出）
**DB**: DynamoDB
**Auth**: Amazon Cognito
**Scheduler**: EventBridge Scheduler（月初リセット）
**IaC**: Terraform

AIの使い所：**注文時の Delivery Plan 生成** と **先回りサジェスト**

---

<!-- _class: final -->

# ゴロゴロPay

あなたの代わりに、考えます。

ご清聴ありがとうございました。
