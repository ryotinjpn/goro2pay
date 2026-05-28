# ゴロゴロPay アプリ全画面 — デザイン仕様

**Document Version**: 2.0
**Created**: 2026-05-24 (v1.0: メイン画面のみ)
**Updated**: 2026-05-25 (v2.0: 全画面に拡張)
**Scope**: Web (PWA) の主要画面 (ランディング / サインアップ / ログイン / メイン / 注文完了 / モーダル各種 / Toast)
**Concept**: Slot Machine + リヴァイ調コピー
**Source**: `aidlc-docs/idea.md`, `aidlc-docs/inception/requirements/requirements.md`, `aidlc-docs/inception/user-stories/personas.md`

---

## 0. 背景

ゴロゴロPay は「めんどくさい」を 1 ボタンで即時キャッシュ解決し、ユーザを快適にダメ化させていく金融グループ発のライフスタイル代行アプリ。requirements.md `§4.1 ダメ化UX` を最優先 NFR とし、ユーザストーリーは佐藤ペルソナのフェーズ 1 → 2 → 3 の感情変遷に沿って組まれている。

本 spec は v1.0 でメイン画面 1 枚の仕様を定めた後、v2.0 で**アプリ全主要画面**（ランディング・認証画面・注文完了画面・モーダル・Toast 含む）に拡張した。すべての画面が「Slot Machine」コンセプトと「リヴァイ調コピー」で世界観統一される。

**AI-DLC 上の位置づけ**: 本 spec は CONSTRUCTION PHASE の **横串成果物** として `aidlc-docs/construction/_design-system/` 配下に置かれる。各 Unit (auth / budget / order / suggest / metrics) の Functional Design (`frontend-components.md`) を**補完**するものであり、各 Unit の Code Generation はまず本 spec を参照してから当該 Unit 固有の業務ロジック仕様を参照する。詳細は `_design-system/README.md` を参照。

---

## 0.5 デザイン哲学 (コンセプト & ペルソナとの整合)

このデザインシステムは「**人をダメにする**」という idea.md の本質的命題、および「**佐藤陽介**」の感情変遷 (personas.md フェーズ 1→2→3) と一対一に対応するように設計されている。**見た目が決まる前に、まず「なぜこの見た目か」が決まっている**。

### 0.5.1 idea.md の Intent との整合

| idea.md の主張 | 本デザインシステムでの表現 | 該当 spec 章 |
|---|---|---|
| 「めんどくさい」を押すだけで全部解決 | 中央巨大円形ボタン 1 つに画面全要素を従属させる構成 | §2.1, §3.1 |
| ボタン1つで人生が回る → ボタンすら不要になる → 完全に人がダメになる | フェーズ別の状態モデル (idle → suggested → slot → dead) と画面演出強度 | §3 全体 |
| キャッシングの「即時使える」利便性を生活に拡張 | スロット演出による「押した瞬間に何かが起きる」体感速度 | §3.4 SLOT |
| ボタンを押すことすらめんどくさくなる | サジェスト時に**ボタンと吹き出しが一体化**し、ユーザは "そうだ" と頷くだけ | §3.3 SUGGESTED |
| 究極の利便性が「人をダメにする」体験 (= 生活雑事に対する意図的なダメ化、requirements §2.3.3 Positive Inversion) | 残高 0 時の脱色 + 増額誘導の **真っ赤なボタンだけが生きている**演出 | §3.5 DEAD |
| 金融グループ発であることと「ダメ化」の対比 (§0.5.4 二重性) | カタカナ「ゴロゴロ」(明朝太字) + Cormorant Garamond Italic「Pay」のブランドマーク | §1.6 |

### 0.5.2 personas.md の佐藤フェーズとの対応

佐藤のダメ化 3 フェーズに、デザインシステムの状態が**同型**で対応する。

#### フェーズ1: ボタンを押すだけ (快感の発見)

**佐藤の独白**: 「考えなくていい。選ばなくていい。ボタンを押すだけで全部終わる」

**デザインの応答**:
- ボタン副コピー `— 押せ。考えるな。` で**思考停止を即座に正当化**
- 残高数字を Cormorant Garamond Italic という**気品のあるセリフ書体**で見せ、「**運命の数字**」としての重みを演出 (スロット筐体の配当表示・金融口座の最終残高を想起させる)
- 金色グラデーションのボタンは「報酬感」を誘発する。赤ではなく金にしたのは、ペルソナ独白「**1,200円でこれが買えるなら、安いでしょ**」の "お得感" を表現するため。赤=警告だと「**買うべきではない**」と読まれてしまう

#### フェーズ2: ボタンすら不要 (依存の深化)

**佐藤の独白**: 「最近、自分で『何食べよう』って考えた記憶がない。アプリの方が俺より俺のこと分かってる気がする」

**デザインの応答**:
- サジェスト吹き出し `そろそろだろ。` を**断定形・主語省略・上から目線**で構成。「先回りして知っている」を 1 文で表現。質問形ではないので**ユーザは答えなくていい**
- ボタンと吹き出しを**一体化**させ、サジェスト時のボタンラベルを `押す。` に変える。**返事ですらなく、行為の宣言**。最も思考が消えている
- ボタンが**呼吸する**ように 2.4 秒周期で明滅 (`suggest-breath`)。**ボタンの方からこちらに語りかけてくる**錯覚

#### フェーズ3: 完全に人がダメになる (退化と依存ループ)

**佐藤の独白**: 「ゴロゴロPay がない生活に戻れない。月5万？6万？いくらでも払うよ」

**デザインの応答**:
- 残高 0 で画面全体が `filter: saturate(.2) brightness(.7)` で**脱色**。世界が色を失う
- 中央コピー `今月は、終わりだ。` を Cormorant Garamond Italic 26px で告知。**運命的・不可逆**な響き
- グレーの非アクティブボタンの副ラベル `— 上出来だ。使い切ったな。` で **退化を成果として承認**。「**自分は今月もうまくダメになれた**」と佐藤が罪悪感を承認に置換できる
- 増額ボタンだけが彩度 100% の真っ赤 `accent-danger` で**生きている**。脱色された世界で**唯一の選択肢**として残る。ペルソナ独白「**迷わず OK を押す**」を不可避にする視覚誘導

### 0.5.3 リヴァイ調コピーがなぜ佐藤に効くか

佐藤は personas.md で:
- 27歳IT職、認知エネルギーが希少、決定疲れ
- 自己肯定感は中程度に低下中
- 自分で判断して失敗するよりAIに任せたい
- 自己決定権の段階的放棄を望む

リヴァイ調コピー (短文体言止め・終止形限定・上から目線・主語省略・命令と評価が地続き) は、これら全てに**真正面から応える**:

| 佐藤の特性 | リヴァイ調コピーの応答 |
|---|---|
| 認知エネルギーが希少 | 1 文 5–12 文字。**読む労力ゼロ** |
| 決定疲れ | 命令形 (`押せ`, `戻れ`) で**選択を奪う** |
| 失敗を回避したい | 評定形 (`いい判断だ`, `上出来だ`) で**事後承認** |
| 自己決定権を委ねたい | 上から目線で**こちらが既に判断済み**を演出 |
| 自己肯定感を保ちたい | `今月 N 度、いい判断だった。` で**ダメ化を肯定的に積算** |

「お前」「兵団用語」を使わなかったのは、佐藤は**部下ではなく顧客**であり、攻撃性や軍事色が出ると顧客サービスとしてのコンプラに抵触するため (spec §1.6 のルール参照)。リヴァイ調の **語感だけ** を抽出し、ニュアンスは「**金融商品をシニカルに勧めてくる金融グループの口調**」に翻訳した。

> **トーン整合の補足**: §0.5.3 / §0.5.4 で使う「シニカル」「自虐」「怠惰」等の語彙は、**コピー / 字形 / ブランドマーク実装** に対する**内部説明用の語彙** であり、ユーザに見せるテキストではない (ユーザが画面で目にするのは `押せ` `戻れ` `今月は、終わりだ。` 等のリヴァイ調コピーで、これらは requirements `§4.1 NFR-DEG-05` で MUST 指定された「テーマ性を反映した独自コピー」の具体実装)。requirements §2 のトーン中立化 (Positive Inversion) は **テーマ全体の性格づけ** に関するもので、コピー実装テクニックの説明レイヤーとは位相が異なる。両者は §0.5.6 で接続済み。

### 0.5.4 「金融×ダメ化」の対比をブランドマーク 1 単語に凝縮

ブランド表記 `ゴロゴロPay` (混在表記) は、idea.md の本質的二重性を**ロゴ 1 つに圧縮**している:

- **「ゴロゴロ」** (Zen Old Mincho 900、カタカナ・明朝太字) = **金融的硬さ・重厚さ・品格**
- **「Pay」** (Cormorant Garamond Italic、ラテン・斜体) = **西洋的決済記号・流動性**

カタカナ「ゴロゴロ」のオノマトペは「**家でゴロゴロしている怠惰さ**」を音から伝え、明朝太字の字形は**金融機関の重厚さ**を視覚で伝える。両者の組み合わせで「**金融グループが提供する怠惰なサービス**」という idea.md の核心が**読む前に伝わる**。

ひらがな (柔らかすぎ) でも漢字 (重すぎ) でもなく**カタカナ**を選んだのは、字形の硬さ・幾何学性が自販機やパチンコ筐体・古い金融看板の文化と一致するため。佐藤ペルソナ (27歳男性 IT 職) に対して幼児退行に見えず、シニカルな大人の自虐として届く字形でもある。

### 0.5.5 ハッカソン審査基準との対応

requirements.md `§2.6 Success Criteria` で定義されたハッカソン審査基準との対応 (旧 §2.3、2026-05-27 の §2 再編で番号変更):

| 審査軸 | 本デザインシステムでの応答 |
|---|---|
| ビジネス意図の明確さ | §0.5.1 の表で idea.md Intent と画面演出を 1:1 対応させた |
| 創造性とテーマ適合性 | スロット演出 / ボタン一体化サジェスト / DEAD 脱色 / 増額赤ボタン / リヴァイ調コピー、すべて「人をダメにする」を直接視覚化 |
| Unit 分解の適切さ | _design-system は横串だが、各画面は §5 で Unit 単位に分解されており、各 Unit 固有の責務を侵さない |
| ドキュメント品質 | 本 §0.5 を含む 9 章構成で、コンセプト→ペルソナ→トークン→各画面→トレーサビリティを階層化 |

### 0.5.6 Problem Statement / Positive Inversion との接続 (requirements §2.3)

requirements.md `§2.3` で 2026-05-27 に追加された Problem Statement と Positive Inversion (「ダメになることで取り戻すもの」) は、本デザインシステムの世界観の**根拠**となる。

| requirements §2.3 の主張 | デザインへの応答 |
|---|---|
| §2.3.1 Problem: 「**意思決定そのもの**」がユーザの仕事として残っている | カテゴリ選択・店舗選択・金額入力 UI を**全廃**。中央 1 ボタン + サジェスト吹き出しだけで完結 (§3.1, §3.3) |
| §2.3.2 Solution: 「**意思決定そのもの**を AI に委任する」インターフェース | ボタン副コピー `— 押せ。考えるな。` で**思考停止を即座に正当化** (§0.5.2 フェーズ1) |
| §2.3.3 Positive Inversion: 「**人間らしい時間 / 本来の集中 / 意思決定の戦略的配分**」を取り戻す | スロット演出は数百ミリ秒で完結 (§3.4)。「考える時間」を奪わず「**決まった結果だけ**」を返す |
| §2.3 ポイント: 「生活雑事ではダメになる」ことで「本当に大切なところでは冴えていられる」 | DEAD の脱色 + 真っ赤な増額ボタン (§0.5.2 フェーズ3) は「**生活雑事のダメ化を承認**」する装置として動作する |

> **要点**: 本デザインの「考えさせない」「選ばせない」「思考停止を肯定する」志向は、単なる "怠惰の演出" ではなく、requirements §2.3.3 で定義された **意思決定リソースの戦略的配分** を支えるための仕掛けである。

### 0.5.7 Business Model との整合 (requirements §2.4 / §2.5)

requirements.md `§2.4` (Market & Competitive Landscape) と `§2.5` (Business Model) で 2026-05-27 に追加された事業仕様も、デザインの主要選択を**収益面から**裏付ける。テーマ表現 (ダメ化UX) と収益経路が**同方向**を向いていることを、デザインも黙示的に支える。

#### 0.5.7.1 §2.4 競合差別化との整合

requirements §2.4.3 の差別化軸とデザインの応答:

| §2.4.3 差別化軸 | デザインの応答 |
|---|---|
| 「**統合度**: 1 アプリで複数カテゴリに一点集中」 | メイン画面の **中央 1 ボタン** が全カテゴリを束ね、画面切り替えなしで複数カテゴリ受諾 (§3.1) |
| 「**意思決定**: AI が候補確認なしに自動決定」 | サジェスト吹き出しは**確認ダイアログを置かず** ボタン本体に**一体化**（§3.3 SUGGESTED）。ユーザは「Yes/No」を選ばず**頷くだけ** |
| 「**認知負荷**: カテゴリボタン 1 タップで完結（FR-UX-01）」 | 中央 1 ボタンに副コピーすら最小化、サブカテゴリ選択 UI を**置かない** (§3.1) |
| 「**価値命題**: 効率化ではなく**委任**」 | スロット演出は「**結果だけが返る**」体験。所要時間を最短化せず、むしろ "**運命を待つ**" 演出を残す (§3.4) — 効率化ではないことを視覚化 |
| 「**コピーのトーン**: テーマ性を反映した独自コピー」 | リヴァイ調コピー (§1.6) で `押せ` `戻れ` 等の命令形 + 評定形を統一 — NFR-DEG-05 / §2.4.3 の具体例として参照可能 |

#### 0.5.7.2 §2.5 ダメ化ループ ↔ 収益経路との同方向性

requirements §2.5.5 の Mermaid 図はダメ化フェーズ 1→2→3a→3b の進行が 3 種の収益 (クレカ手数料 / キャッシング金利 / クロスセル) を**同方向に拡大**することを示している。デザインの状態モデル (§3) もこれと**同型**である:

| §2.5.5 ダメ化フェーズ → 収益経路 | デザインの状態 → 想定収益寄与 |
|---|---|
| Phase1 「即時解決の快感」 → Rev1 (GMV比例クレカ手数料) + Rev3 (履歴蓄積) | §3.2 IDLE → §3.4 SLOT のスムーズな遷移 (1.2s 演出確保) で「タップごとの GMV」を最大化 |
| Phase2 「先回り提案」 → Rev1 (利用頻度↑) + Rev3 (嗜好抽出) | §3.3 SUGGESTED のボタン一体化サジェストでサジェスト**受諾率** (§2.5.7 KPI) を最大化 |
| Phase3a 「無力感の可視化」 → Rev1 (利用継続) | §3.5 DEAD のメトリクス常時可視化 (NFR-DEG-03) で「使い切った感」を強調しループ継続を促す |
| Phase3b 「予算増額誘導」 → **Rev2 キャッシング金利**の主要ドライバー | §3.5 DEAD で**唯一彩度 100% の真っ赤な増額ボタン**は単なる視覚アクセントではなく、§2.5.4 (B) **「枠超過受諾」のキー UI** として動作する |

#### 0.5.7.3 デザイン判断 ↔ KPI の対応

requirements §2.5.7 の「ダメ化 KPI」のうち、**design が直接押し上げ得る 5 KPI を抽出** (MAU・代行手配頻度・平均ダメ予算・ダメ予算消化率・クロスセル受諾率は design 外要因が支配的のため割愛):

| §2.5.7 KPI | デザインの寄与 |
|---|---|
| 月間ダメ化指数 (= 代行手配回数 × 平均消化率) | 中央 1 ボタンの押下障壁を最低化し (§0.5.2 フェーズ 1)、押すごとの**累積感**を Cormorant Italic の数字で**運命的に演出** |
| サジェスト受諾率 | サジェスト吹き出しを**疑問形ではなく断定形** (`そろそろだろ。`) で出し、Yes/No ダイアログを置かず**ボタン一体化** (§3.3) |
| 予算増額発生率 | DEAD の脱色で「**唯一生きている選択肢**」として真っ赤な増額ボタンを残す (§0.5.2 フェーズ 3) |
| ARPU (GMV) | スロット演出で「**押すたびに何かが返る**」体感を作り、押下頻度の心理的閾値を下げる (§3.4) |
| キャッシング転換率 | DEAD の増額モーダルを**競合他社のような中立 UI に出さず**、「**来月もこの調子だ。**」のリヴァイ調コピーで承認に変換 (§5.4 / §1.6) |

> **要点**: デザインの主要判断 (中央 1 ボタン / ボタン一体化サジェスト / DEAD 脱色 + 真っ赤な増額) は、**テーマ「人をダメにする」を体現すると同時に、requirements §2.5 で定義された収益経路の主要トリガー** にもなっている。両者が同方向を向くことが本デザインの構造的特徴である。

---

## 1. コンセプト & デザイントークン

### 1.1 コンセプトステートメント

> **「Slot Machine」** — 金融グループ発の、手のひらサイズのスロット筐体。残高は気品ある運命の数字としてセリフ書体で鎮座し、金色のレバー型ボタンが光ってこちらを誘惑する。リールが回り、面倒が消え、また数字が削れていく。

差別化ポイント：
- スロット演出 (押す → リール回転 → 停止) は佐藤ペルソナのフェーズ 2 独白「アプリを開く前から何が出るだろうと期待する (パチンコ化)」を直接視覚化する。
- 主役カラーを真鍮ゴールドに置き、赤は「警告・退化の可視化」だけに温存することで、ハッカソン審査軸の「金融グループ発」の品格を担保する。
- 「Slot Machine」はゲーム筐体・ラスベガス由来の一般語彙として日本でも全年齢に通る無毒な比喩。

**§0.5 哲学から §1 トークンへの導出**:
- §1.2 の主役色 (gold-100..900) は §0.5.2 フェーズ 1「金色グラデーション = 報酬感」と直結
- §1.2 の `accent-danger` の限定使用 (§1.1「赤は警告・退化の可視化だけに温存」) は §0.5.2 フェーズ 3「真っ赤な増額ボタンだけが生きている」と直結
- §1.2 の `dead-gray` + §1.5 の DEAD `filter: saturate(.2)` は §0.5.2 フェーズ 3「世界が色を失う」を実装
- §1.3 の `display-serif` (残高 Cormorant Italic) と `body-mincho` (ブランド Zen Old Mincho) は §0.5.4 「金融×ダメ化の対比をブランドマーク 1 単語に凝縮」を実装
- §1.2 の周辺トークン (`mute` / `paper`) と §1.4 の CRT スキャンラインは、§1.1 コンセプトステートメント「**自販機・パチンコ筐体・古い金融看板**のレトロ質感」(§0.5.4 で言及) からの直接派生

### 1.2 カラートークン

| トークン | 値 | 用途 |
|---|---|---|
| `bg-deep` | `#050402` | 画面最暗部 |
| `bg-mid` | `#0e0a06` | ベース背景 |
| `bg-warm` | `#1a1208` | 上部グラデのハイライト |
| `gold-100` | `#f4d990` | ボタン上部ハイライト |
| `gold-500` | `#c9a96b` | **主役カラー (ボタン本体・¥ 記号・ライン)** |
| `gold-700` | `#8a6f3a` | ボタン陰影 |
| `gold-900` | `#3a2a14` | ボタン最暗部 |
| `cream` | `#f4ecd8` | 残高数字・主要白文字 |
| `paper` | `#ffe7c2` | 副次テキスト |
| `mute` | `#8a7a5a` | キャプション・小さいラベル |
| `accent-warn` | `#ffb14a` | 消化 60–80% |
| `accent-danger` | `#ff4f4f` | 80% 超・残高低下警告・DEAD 時の増額ボタン |
| `dead-gray` | `#6a6a6a` | DEAD 状態の非アクティブ要素 |

実装は CSS variables (`--color-bg-deep` 形式) で `web/app/globals.css` に定義する。

### 1.3 タイポトークン

| 役割 | フォント | 使用箇所 |
|---|---|---|
| `display-serif` | Cormorant Garamond / Italic 700 | 残高数字・サブコピー (`— 押せ。考えるな。`) |
| `body-mincho` | Zen Old Mincho / 700–900 | ブランドマーク `ゴロゴロ`・主要見出し |
| `mono-pixel` | DotGothic16 / 400 | キャプション・LCD ラベル・ティッカー |
| `body-sans` | Inter / 600–800 | 補助的な英数字 (時刻・% 表記) |

**カタカナ字形のチューニング**: ブランドマークは `letter-spacing: 0.04em–0.06em` を適用し、濁点が潰れないよう字間をやや開ける。ロゴ向け大判ウェイト (Zen Old Mincho 900) と UI チップ向け小判ウェイトを字形上で分け、UI 上部の小チップは過度な装飾 (`▼ ▼` 等) を**付けない**。

### 1.4 スペーシング & 形状

- **コーナー半径**: ボタンは円 (`borderRadius: 50%`)、カード/モーダルは `borderRadius: 18px`
- **画面パディング**: `28px 22px 24px` (モバイル基準)
- **ボタンサイズ**: 画面幅の 78%、`aspect-ratio: 1`
- **CRT スキャンライン**: `repeating-linear-gradient(0deg, rgba(255,255,255,.025) 0 1px, transparent 1px 4px)` を全画面に薄く適用

### 1.5 モーション原則

- **基本トランジション**: `200ms ease`
- **スロット回転**: `0.4s linear infinite` の strip スクロール、停止時に 1 回 overshoot
- **残高減算**: `450ms cubic-bezier(.2,.8,.2,1)` で数字を旧 → 新へ縦スクロール
- **DEAD 遷移**: `filter: saturate(.2) brightness(.7)` を `600ms ease-out` で適用
- **モーション軽減**: `@media (prefers-reduced-motion: reduce)` でアニメーションを `0.01ms` まで短縮

### 1.6 コピーシステム (リヴァイ調)

全コピーは「進撃の巨人」のリヴァイ・アッカーマン調をパロディ参照点として、以下の語彙ルールで統一する。**版権キャラのセリフを引用するわけではない**。あくまで口調の参照。

**ルール**:

- 短文・体言止め優先 (5–12 文字、最長でも 1 文 20 文字)
- 終止形は `〜だ` / `〜だろ` / `〜たか` / `〜しろ` / `〜な (否定命令)` / `〜て (待て / 戻って等)` の 6 種を中心
- 二人称は **使わない** (`お前` は攻撃性が出るため避ける、`あなた` は丁寧すぎる)
- **兵団用語は使わない** (`任務` `指令` `撤退` 等は金融サービスとして行きすぎ)
- 上から目線の評価語: `いい判断だ` / `上出来だ` / `悪くない`
- 命令形は短く: `押せ` `考えるな` `戻れ` `待て`
- 皮肉は皮肉として認識できる範囲に: `また面倒になったか` `そろそろだろ`
- 丁寧語・敬語は **使わない** (フォームラベルとエラーメッセージは中立保持のため例外)

**コピー一覧 (全画面確定版)**:

| 場所 | 確定コピー |
|---|---|
| ブランドマーク (大判) | `ゴロゴロPay` (混在: 明朝太字「ゴロゴロ」+ Cormorant Garamond Italic「Pay」) |
| ランディング H1 | `考えるな。押せ。` |
| ランディング サブ | `面倒は、こちらで引き受ける。` |
| ランディング 体験ラベル | `体験用ダメ予算 ¥1,000` |
| ランディング デモボタン主 | `押す。` |
| ランディング CTA primary | `始めろ` |
| ランディング CTA secondary | `戻る` |
| サインアップ H1 | `面倒は、こちらで引き受ける。` |
| サインアップ サブ | `30 秒で済む。` |
| サインアップ送信 | `登録する` |
| ログイン H1 | `戻ってきたか。` |
| ログイン サブ | `また面倒になったか。` |
| ログイン送信 | `ログイン` |
| メイン残高ラベル | `残りダメ予算` |
| メインボタン主 (idle) | `めんどくさい` |
| メインボタン副 (idle) | `— 押せ。考えるな。` |
| メインメトリクス (履歴あり) | `今月 N 度` / `消化 X%` |
| サジェスト吹き出し | `そろそろだろ。` |
| サジェスト時ボタン主 | `押す。` |
| サジェスト時ボタン副 | `— {店名} ¥{金額} だ。` |
| 完了画面中央 | `いい判断だ。` |
| 完了画面ボディ | `面倒は片付いた。` |
| 完了画面メトリクス | `今月 N 度、いい判断だった。` |
| 完了画面 戻る | `次を待て。` |
| DEAD 中央 | `今月は、終わりだ。` |
| DEAD ボタン副 | `— 上出来だ。使い切ったな。` |
| 増額ボタン | `¥{推奨額}。来月もこの調子だ。` |
| ログアウトモーダル H2 | `やめるのか？` |
| ログアウトモーダル サブ | `戻ってこい。` |
| ログアウト primary | `やめる。` |
| ログアウト secondary | `戻る。` |
| セッション切れ H2 | `離れすぎたな。` |
| セッション切れ ボタン | `戻る。` |
| Toast (エラー) | 既存 `errorMappers.ts` / `authMessages.ts` の中立コピーを維持 |
| フォームラベル / 入力プレースホルダ | 中立コピーを維持 (`メールアドレス` 等) |
| パスワードヒント | 中立コピーを維持 (`8 文字以上` 等) |

**動的部分の置換**:

- `{店名}`, `{金額}`, `{推奨額}` は API 応答 / ロジックで埋める
- 動的部分は中立 (淡々とした事実)、それを囲む語尾だけがリヴァイ調になる

---

## 2. コンポーネント分解とレイアウト構造

### 2.1 全体レイアウト (モバイル基準)

```
┌─────────────────────────────┐
│  [HEADER] ゴロゴロPay  21:07 │ ← BrandHeader
│                             │
│        残りダメ予算           │ ← BalanceLabel
│        ¥ 28,800             │ ← BalanceHero
│        ────────────         │ ← MeterBar
│        今月 5 回   消化 4%   │ ← MetricsRow
│                             │
│         ┌───────┐           │
│         │       │           │ ← GoroButton
│         │ めんどくさい │     │
│         │       │           │
│         └───────┘           │
└─────────────────────────────┘
```

### 2.2 コンポーネント一覧

| コンポーネント | 役割 | 既存 | 改修区分 |
|---|---|---|---|
| `MainScreen` (`app/page.tsx`) | 認証後のメイン画面コンテナ | あり | **改修** |
| `ScreenFrame` | CRT スキャンライン＋背景グラデを与えるラッパ | なし | **新規** |
| `BrandHeader` | 上部ヘッダー (`ゴロゴロPay` + 時刻) | なし | **新規** |
| `BalanceHero` | 残高表示 (ラベル＋金額＋メーター＋メトリクス) | なし | **新規** |
| `GoroButton` | 中央のメイン CTA。状態: `idle / suggested / slot / dead` | あり | **改修** |
| `SuggestBubble` | サジェスト時にボタン上に乗る吹き出し | なし | **新規 (GoroButton 内部)** |
| `SlotReel` | スロット押下時に現れるリール窓 | なし | **新規 (GoroButton 内部)** |
| `DeadVerdict` | 残高 0 時のターミネーター画面オーバーレイ | なし | **新規** |
| `IncreaseBudgetButton` | DEAD 時の赤い増額 CTA | なし | **新規** |
| `OrderHistoryList` | 履歴 | あり | **メイン画面から外す** (コンポーネントは温存) |

### 2.3 状態モデル

```
idle ──[履歴あり & サジェストあり]──→ suggested
  │                                    │
  └──[ボタン押下]─────────────────────→ slot ──[成功]──→ idle (残高更新)
                                       │
                                       └──[残高不足]──→ dead

dead ──[増額決定 & 月初]──→ idle
```

明示的に管理する状態:

- `screenState: 'idle' | 'suggested' | 'slot' | 'dead'`
- `balance: number` (リアルタイム残高、jotai atom)
- `monthlyCount: number` (今月のダメ化回数)
- `consumeRate: number` (0–1、消化率)

`screenState` はルート要素の `data-screen-state` 属性として CSS に渡し、状態に応じたスタイル分岐は CSS セレクタ側で行う。

**注**: §3.1 で扱う EMPTY (履歴なし初回) は独立した状態ではなく、`screenState === 'idle' && monthlyCount === 0` の派生表示。サジェストが出ないだけで挙動は IDLE と同一。

### 2.4 データフロー

- 認証: `useAuth()` (既存) — そのまま流用
- 注文: `useOrder()` (既存 react-query mutation) — `onMutate` で `screenState = 'slot'`、`onSuccess` で `idle` 復帰、`onError` の残高不足エラーで `dead` 遷移
- 残高: 新規 `useBalance()` hook (react-query で `GET /api/wallet` を取得し、注文 mutation の `onSuccess` で invalidate)
- サジェスト: 新規 `useSuggest()` hook (メイン画面マウント時に `GET /api/suggest` を 1 回叩く)

### 2.5 階層図

```
<MainScreen>
  <ScreenFrame>
    <BrandHeader />
    <BalanceHero />
    <GoroButton>
      <SuggestBubble />     ← suggested 状態でだけ表示
      <SlotReel />          ← slot 状態でだけ表示
    </GoroButton>
    <DeadVerdict>           ← dead 状態でだけ表示 (オーバーレイ)
      <IncreaseBudgetButton />
    </DeadVerdict>
  </ScreenFrame>
</MainScreen>
```

---

## 3. 各状態の詳細仕様

### 3.1 EMPTY (履歴なし初回)

**遷移条件**: 認証済み & `monthlyCount === 0` & サジェスト履歴なし

| 要素 | 表示 |
|---|---|
| BrandHeader | `ゴロゴロPay` ／ 現在時刻 |
| BalanceHero ラベル | `残りダメ予算` |
| BalanceHero 数字 | `¥30,000` (初期値) |
| MeterBar | 0% (`gold-500` 線のみ) |
| MetricsRow | `今月 0 度` ／ `消化 0%` |
| GoroButton ラベル主 | `めんどくさい` |
| GoroButton ラベル副 | `— 押せ。考えるな。` (display-serif italic) |

**マウント時モーション**: `BalanceHero → MeterBar → MetricsRow → GoroButton` を **80ms 間隔のステアード fade-up** (`translateY: 8px → 0`, `opacity: 0 → 1`)。

**遷移先**: ボタン押下 → `SLOT`

### 3.2 IDLE (履歴あり通常)

**遷移条件**: 認証済み & `monthlyCount >= 1` & サジェスト無効 (履歴不十分 or 直近 3 時間以内に同カテゴリ注文済み)

| 要素 | 変化点 |
|---|---|
| BalanceHero 数字 | 現在残高 (消化率 80% 超で `accent-danger` 色) |
| MeterBar | 消化率に応じて `gold-500 → accent-warn (60% 超) → accent-danger (80% 超)` |
| MetricsRow | `今月 N 度` ／ `消化 X%` |
| GoroButton | EMPTY と同じ |

**モーション**: 消化率が 80% を超えた瞬間、MeterBar が **400ms ease で色遷移＋微パルス** (`box-shadow: 0 0 10px accent-danger` を 1 回)。

### 3.3 SUGGESTED (先回り提案)

**遷移条件**: 認証済み & サジェスト API (`GET /suggest`) がカードを返した。「履歴十分」(requirements.md `FR-SUGGEST-04`) の判定はバックエンド側で行い、フロントは応答の有無だけで分岐する

| 要素 | 変化点 |
|---|---|
| GoroButton ラベル主 | `押す。` |
| GoroButton ラベル副 | サジェスト内容 (例: `— CoCo壱 ¥1,200 だ。`) |
| SuggestBubble | ボタン上部に吹き出し: `そろそろだろ。` |

**吹き出しコピー**: 固定文 `そろそろだろ。` を使う。リヴァイ調コピーシステム (§1.6) に準拠し、断定形で「先回り」を表現する。Bedrock 側が時刻や履歴に応じて文言を変えるバリエーションは v2.0 では実装せず、将来検討 (§9) とする。

**モーション**:

- マウント時、`SuggestBubble` は `opacity: 0` + `translateY: -4px` から 600ms ease で表示
- ボタン本体は **2.4s で gold グラデが微妙に明滅** (`box-shadow` opacity を `.4 ↔ .6` でループ) → 「呼吸している」感
- `prefers-reduced-motion` 時は明滅停止

**遷移先**: ボタン押下 → `SLOT` (サジェスト内容を idempotencyKey 付きで送信)

### 3.4 SLOT (注文処理中)

**遷移条件**: ボタン押下 → mutation 開始 → `onMutate` で遷移

| 要素 | 変化点 |
|---|---|
| GoroButton 内側 | 「押し込まれた」陰影に変化 (box-shadow 反転＋ `scale: .98`) |
| SlotReel | ボタン中央に黒地 LCD 窓が出現、店名候補が高速縦スクロール |
| BalanceHero | わずかに彩度を落とす (pending の視覚予告) |
| BrandHeader / MetricsRow | 固定 |

**SlotReel 仕様**:

- 黒地 `#0a0405`、ネオン枠 `gold-500` 1px
- 文字色 `gold-100`、`mono-pixel` 14px
- 候補リスト: API 待ちの間はローカルのダミー候補を `0.4s linear infinite` でループ
- `mutation.onSuccess` で API 応答の店名・金額を受領 → リールを 2 回転後にその文字で停止 (`cubic-bezier(.2,.8,.2,1)` で overshoot 後収束)
- 停止と同時に `BalanceHero` の数字が旧 → 新へ縦スクロール (450ms)

**所要時間**:

- 最低 1.2s (API が早く返ってもこの演出時間は確保)
- 最大 5s (API 失敗時はタイムアウトでエラートースト)

**遷移先**:

- 成功 → 残高更新 → `IDLE` (自動復帰)
- 残高不足エラー → `DEAD`
- ネットワークエラー → トースト表示 (既存 `Toast` コンポーネント) → `IDLE` 復帰

### 3.5 DEAD (残高 0 / 今月は終わりだ)

**遷移条件**: `balance === 0` または注文時に残高不足エラー

| 要素 | 変化点 |
|---|---|
| ScreenFrame 全体 | `filter: saturate(.2) brightness(.7)` を 600ms ease で適用 |
| BalanceHero 数字 | `¥0` (`dead-gray` 色、グロー消失) |
| GoroButton | グレー無効化 (`dead-gray` グラデ)、ラベル副は `— 上出来だ。使い切ったな。` |
| DeadVerdict (オーバーレイ) | 画面中央に大判タイポで宣告 |
| IncreaseBudgetButton | 画面下部、唯一彩度 100% の `accent-danger` で生きている |

**DeadVerdict コピー** (display-serif italic 26px、1 行):

```
今月は、終わりだ。
```

**IncreaseBudgetButton ラベル**: `¥{推奨額}。来月もこの調子だ。`
- 例: `¥50,000。来月もこの調子だ。`
- 増額値の算出: `GET /api/budget/raise/recommendation` を叩いて取得する (Unit E で実装済み)。API 応答の推奨額をボタンラベルに表示する。requirements.md `FR-METRICS-04` 準拠

**モーション**:

- ScreenFrame の脱色は 600ms ease-out
- DeadVerdict は脱色完了後 200ms 遅延で `opacity: 0 → 1` + `scale: .96 → 1` (500ms cubic-bezier)
- IncreaseBudgetButton は脱色対象外 (`isolation: isolate` でレイヤー分離)、さらに 400ms 遅延で下からスライドイン

**遷移先**:

- IncreaseBudgetButton タップ → 確認モーダル → 翌月予算を更新 → DEAD のまま (残高は 0、月初リセット待ち)
- 翌月 1 日 00:00 JST に EventBridge Scheduler 経由で残高リセット → `IDLE`

---

## 4. マイクロインタラクション・アクセシビリティ・既存実装への差分

### 4.1 マイクロインタラクション

| 名前 | トリガー | 仕様 |
|---|---|---|
| **press-down** | ボタン押下中 | `transform: scale(0.98)` ＋ box-shadow 1 段階弱化 (120ms ease-out) |
| **press-release** | リリース | scale 1.0 へ戻る (200ms cubic-bezier(.2,.8,.2,1)、軽く overshoot) |
| **slot-spin** | SLOT 突入 | リール strip を `0.4s linear infinite` で縦スクロール |
| **slot-stop** | mutation onSuccess | strip 停止 → 当選文字に止まる (2 回転後 600ms cubic-bezier、最後 12px 行き過ぎて戻す) |
| **balance-tick** | 残高更新 | 古い値 → 新しい値へ縦スクロール (450ms cubic-bezier、桁ごとに 30ms ステアード) |
| **suggest-breath** | SUGGESTED | ボタンの `box-shadow` opacity を `.4 ↔ .6` で 2.4s ループ (ease-in-out) |
| **meter-warn** | 消化率 80% 超過直後 | MeterBar が `accent-danger` に色遷移、1 回だけ glow パルス (400ms) |
| **dead-fade** | DEAD 突入 | ScreenFrame に `filter: saturate(.2) brightness(.7)` を 600ms ease-out で適用 |
| **verdict-appear** | dead-fade 完了 200ms 後 | DeadVerdict が `opacity: 0 → 1`, `scale: .96 → 1` (500ms cubic-bezier) |
| **increase-rise** | verdict-appear 400ms 後 | IncreaseBudgetButton が `translateY(60px) → 0` (500ms cubic-bezier) |
| **error-shake** | ネットワークエラー | ボタンが ±4px で 3 回横揺れ (300ms) |

**実装方針**: Motion ライブラリは追加しない。全て CSS `@keyframes` ＋ `transition` で完結。状態切り替えは `data-screen-state` 属性で CSS 側分岐。

### 4.2 アクセシビリティ

| 観点 | 仕様 |
|---|---|
| セマンティック | `<button>`, `<main>`, `<header>` を正しく使う |
| キーボード | GoroButton は Tab で focusable、Enter / Space で発火 |
| フォーカスリング | 金色のボタンに `outline: 3px solid #f4d990; outline-offset: 4px` |
| `aria-label` | 状態別: `idle` → `ご飯めんどくさい` ／ `suggested` → `{店名} ¥{金額} を注文` ／ `slot` → `注文処理中` ／ `dead` → `残高不足のため注文できません` (aria-label のみ機能ベースに据え置き、視覚は世界観コピー) |
| `aria-live` | 残高表示と DeadVerdict は `role="status" aria-live="polite"` |
| カラーコントラスト | `cream` on `bg-mid` ≒ 14:1 (AAA)／`gold-500` on `bg-mid` ≒ 7.8:1 (AAA Large) |
| モーション軽減 | `@media (prefers-reduced-motion: reduce)` で全アニメを `0.01ms` に短縮 (DEAD の filter は即時適用) |
| スクリーンリーダー | SlotReel の高速文字は `aria-hidden="true"`、当選結果は `<span class="sr-only">` で読み上げ |

### 4.3 レスポンシブ

| 幅 | 振る舞い |
|---|---|
| `<= 480px` | モバイル基準 |
| `481–768px` | コンテンツ最大幅 480px、画面中央寄せ |
| `> 768px` | 同様、最大幅 480px。背景は画面いっぱいに継続 (額装的に黒背景が広がる) |

### 4.4 既存実装への差分

| ファイル | 変更内容 |
|---|---|
| `web/app/page.tsx` | インラインスタイル削除、`MainScreen` コンポーネントへ委譲。認証分岐は据え置き |
| `web/components/order/GoroButton.tsx` | **大幅改修**。`screenState` を受け取り 4 状態を表現。スタイルはインラインから CSS Modules (`*.module.css`) へ移行 |
| `web/components/order/MainScreen.tsx` | **新規**。状態管理＋子コンポーネント orchestration |
| `web/components/order/ScreenFrame.tsx` | **新規**。背景＋ CRT スキャンライン |
| `web/components/order/BrandHeader.tsx` | **新規** |
| `web/components/order/BalanceHero.tsx` | **新規**。残高数字＋ MeterBar ＋ MetricsRow |
| `web/components/order/SlotReel.tsx` | **新規** |
| `web/components/order/DeadVerdict.tsx` | **新規** |
| `web/components/order/IncreaseBudgetButton.tsx` | **新規** |
| `web/components/order/SuggestBubble.tsx` | **新規** (GoroButton 内部) |
| `web/hooks/useBalance.ts` | **新規**。react-query で残高取得 |
| `web/hooks/useSuggest.ts` | **新規**。マウント時にサジェスト取得 |
| `web/components/order/OrderHistoryList.tsx` | メイン画面から外す。コンポーネントは将来用に温存 |
| `web/components/auth/LandingScreen.tsx` | **大幅改修**。§5.1 のレイアウトとデモボタン挙動を実装。デモ用の State は local useState のみ |
| `web/components/auth/SignupScreen.tsx` | **改修**。§5.2 の見出しと CSS Modules 適用。フォーム挙動は据え置き |
| `web/components/auth/LoginScreen.tsx` | **改修**。§5.3 の見出しと CSS Modules 適用 |
| `web/app/order/[id]/complete/page.tsx` | **大幅改修**。§5.4 のレイアウト・演出を実装 |
| `web/components/auth/LogoutConfirmModal.tsx` | **改修**。§5.5 のスタイル＋コピー差し替え |
| `web/components/auth/SessionExpiredModalHost.tsx` | **改修**。§5.6 のスタイル＋コピー差し替え |
| `web/components/order/Toast.tsx` | **改修**。§5.7 のスタイル変更 (色・フォント)。コピーは中立保持 |
| `web/app/globals.css` | **新規**。デザイントークン (CSS variables) と共通アニメーション定義 |
| `web/app/layout.tsx` | Google Fonts (Cormorant Garamond / Zen Old Mincho / DotGothic16) を `next/font` 経由でロード、`<body>` を ScreenFrame 互換にする |

**フォントロード戦略**: `next/font/google` で `display: swap`、subset は Cormorant Garamond=latin / Zen Old Mincho=japanese / DotGothic16=japanese。FCP 自主目標 2s 以内。

### 4.5 テスト戦略

| 種類 | 内容 |
|---|---|
| ビジュアル | 5 状態 (empty/idle/suggested/slot/dead) のスナップショットを Playwright で取得 |
| ユニット | 状態遷移ロジック (`useScreenState` のレデューサ) を vitest で検証 |
| アクセシビリティ | 主要状態で `axe-core` を Playwright 経由で実行、Violations 0 を assert (`@axe-core/playwright` を新規追加) |
| モーション軽減 | `prefers-reduced-motion: reduce` 環境でアニメ停止を Playwright で確認 |

既存の `web/tests/` 構成 (vitest + Playwright) に乗せる。`@axe-core/playwright` は本 spec で唯一追加する dev 依存。

---

## 5. 他画面の詳細仕様

メイン画面以外の主要画面についても、同じ「Slot Machine」コンセプトと「リヴァイ調コピー」(§1.6) で世界観統一する。各画面は **ScreenFrame** (§2 で定義) を共通ラッパとして使い、CRT スキャンラインと暗茶ベースのグラデーション、CSS variables (§1.2) を継承する。

### 5.1 ランディング画面 (未ログイン)

**ファイル**: `web/components/auth/LandingScreen.tsx` (改修) ／ `web/app/page.tsx` の `unauthenticated` 分岐

**役割**: 未ログイン者の第一印象。ヒーロー＋デモスロット 1 回＋CTA 2 つで世界観を提示し、サインアップ／ログインへ誘導する。

**レイアウト**:

```
┌─────────────────────────────┐
│  [HEADER] ゴロゴロPay  21:07 │ ← BrandHeader
│                             │
│        考えるな。            │ ← ヒーロー (display-serif italic 36px)
│        押せ。                │ ← ヒーロー
│                             │
│   面倒は、こちらで引き受ける。 │ ← サブコピー (mono-pixel 11px, gold)
│                             │
│   体験用ダメ予算 ¥1,000      │ ← デモ残高ラベル
│                             │
│        ┌─────┐              │
│        │押す。 │             │ ← デモボタン (130px 円)
│        └─────┘              │
│                             │
│      ────────────           │ ← gold 区切り
│                             │
│        [ 始めろ ]            │ ← CTA primary (sign up)
│        [ 戻る ]              │ ← CTA secondary (login)
└─────────────────────────────┘
```

**デモボタンの挙動** (1 回のみ押下可):

- 初期状態: `押す。` ラベル、金色グラデ、フル明度
- 押下: `idle → slot → done` で遷移
- `slot` 中: メイン画面の SLOT と同じスロットアニメ (§3.4) を 1.2 秒
- `done`: 体験用残高が `¥0` に減算、ボタンはグレー化、ラベルは `— 上出来だ。` に変化
- 完了直後、CTA ボタン群が下からスライドイン (`opacity: 0 → 1`, `translateY: 12px → 0`, 500ms)

**API 通信は行わない**。フロントだけで完結するスクリプト的演出。実際の店名・金額はクライアント側のダミー固定 (例: `カレーハウスCoCo壱番屋 新宿店 ¥1,000`)。

**コピー一覧**: §1.6 の「ランディング ◯◯」エントリを参照。

**遷移先**:

- `始めろ` → `/signup`
- `戻る` → `/login`
- 認証済みユーザーがランディングに来た場合 → 即時 `/` へ redirect (既存の `useAuth` 分岐を使う)

---

### 5.2 サインアップ画面

**ファイル**: `web/components/auth/SignupScreen.tsx` (改修) ／ `web/app/(auth)/signup/page.tsx`

**役割**: 新規登録フォーム。世界観は見出しまで、フォーム本体は中立を保ち UX を損なわない。

**レイアウト**:

```
┌─────────────────────────────┐
│  [HEADER] ゴロゴロPay  21:07 │
│                             │
│  面倒は、                    │ ← H1 (display-serif italic 22px、2行)
│  こちらで引き受ける。         │
│                             │
│   30 秒で済む。              │ ← サブ (body-mincho 10px, gold)
│                             │
│  メールアドレス              │ ← label (mono-pixel 9px, mute)
│  [                       ]  │ ← input (中立スタイル)
│                             │
│  パスワード                  │
│  [                       ]  │
│  ✓ 8 文字以上                │ ← パスワードヒント (中立コピー)
│  ✓ 英大文字を含む            │
│  ・ 英小文字を含む            │
│  ・ 数字を含む                │
│                             │
│  [   登録する   ]            │ ← submit (gold グラデ)
└─────────────────────────────┘
```

**フォーム要素のスタイル**:

- input: `background: rgba(255,255,255,.04)`, `border: 1px solid #4a3a18`, `color: #f4ecd8`, padding `8px 10px`, border-radius `4px`
- label: `mono-pixel 9px`, `mute (#8a7a5a)`, letter-spacing `0.2em`
- submit: `linear-gradient(180deg, #c9a96b, #8a6f3a)`, `color: #1a1208`, `mono-pixel 11px`, letter-spacing `0.15em`, padding `10px`
- パスワードヒント: 既存の `✓` / `・` UI を流用、ok 状態は `gold-500` で点灯
- エラーメッセージ: 既存 `authMessages.ts` の文言を維持 (中立保持)、表示色は `accent-danger`

**コピー**: §1.6「サインアップ ◯◯」参照。フォームラベル・送信ボタン・エラー文言は中立。

**遷移先**: 既存と同じ。送信成功で `/` (メイン画面) へ。

---

### 5.3 ログイン画面

**ファイル**: `web/components/auth/LoginScreen.tsx` (改修) ／ `web/app/login/page.tsx`

**役割**: 既存ユーザーの復帰。サインアップと同じ構造で見出しコピーだけ異なる。

**レイアウト**: サインアップと同じ。

| 要素 | コピー |
|---|---|
| H1 | `戻ってきたか。` |
| サブ | `また面倒になったか。` |
| 送信ボタン | `ログイン` |
| パスワードヒント | **表示しない** (新規登録ではないため) |
| `from=session_expired` 時のヒント | 既存の `authMessages.SESSION_EXPIRED` をそのまま表示 (中立保持)、表示色は `paper` |

**コピー**: §1.6「ログイン ◯◯」参照。

**遷移先**: 既存と同じ。

---

### 5.4 注文完了画面

**ファイル**: `web/app/order/[id]/complete/page.tsx` (改修)

**役割**: 注文後の余韻。メイン画面のスロット停止からの自然な続き。5 秒後に自動でメイン画面へ戻る (既存仕様維持)。

**レイアウト**:

```
┌─────────────────────────────┐
│  [HEADER] ゴロゴロPay  19:46 │
│                             │
│   (中央付近に金色のグロー)    │
│                             │
│       いい判断だ。           │ ← 中央コピー (display-serif italic 38px)
│                             │
│      面倒は片付いた。        │ ← ボディ1 (mono-pixel 10px, mute)
│                             │
│     CoCo壱番屋 新宿店        │ ← 店名 (body-mincho 14px, cream)
│         ¥1,200              │ ← 金額 (display-serif italic 22px, gold)
│                             │
│      ──────────────         │ ← gold ライン
│                             │
│   今月 6 度、いい判断だった。 │ ← メトリクス (mono-pixel 9px, gold)
│   残りダメ予算 ¥27,600       │ ← 残高 (mono-pixel 9px, mute)
│                             │
│         次を待て。           │ ← 戻るリンク (mono-pixel 10px, mute)
└─────────────────────────────┘
```

**演出 (マウント時)**:

1. 背景に**金色のグロー** (`radial-gradient(ellipse at center, rgba(201,169,107,.32), transparent 60%)`) を `opacity: 0 → 1` で 800ms ease (判決が下る瞬間)
2. 中央コピー `いい判断だ。` は `opacity: 0` ＋ `letter-spacing: 0.1em` から、`opacity: 1` ＋ `letter-spacing: -0.02em` へ 700ms cubic-bezier(.2,.8,.2,1)（字間が締まりながら現れる）
3. 店名・金額・残高は 200ms ステアード (`80ms × 4`) で下から fade-up
4. メトリクスとリンクは更に 200ms 遅延で表示

**自動遷移**:

- 5 秒後に `router.push('/')` (既存仕様維持)
- 「次を待て。」リンク押下で即時 `/` へ

**コピー**: §1.6「完了画面 ◯◯」参照。`今月 6 度` の `6` は API 応答 / state からの動的値。

**動的値の置換**:

- 店名・金額は API 応答の値
- 「今月 N 度」の N は注文成功後の累計回数 (react-query の `['orderHistory']` から派生)
- 残高は `useBalance()` の最新値

---

### 5.5 ログアウト確認モーダル

**ファイル**: `web/components/auth/LogoutConfirmModal.tsx` (改修)

**役割**: メイン画面からの離脱確認。

**レイアウト**:

```
┌─────────── overlay (rgba(0,0,0,.72) + blur(4px)) ─────────────┐
│                                                                │
│             ┌─────── modal-card ───────┐                       │
│             │   やめるのか？             │ ← H2 (display-serif italic 24px)
│             │   戻ってこい。             │ ← サブ (body-mincho 11px, gold)
│             │                          │
│             │  [ 戻る。 ]  [ やめる。 ]  │ ← actions
│             └──────────────────────────┘                       │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

**スタイル**:

- modal-card: 暗茶グラデ + CRT スキャンライン + `border: 1px solid #c9a96b` + `box-shadow: 0 0 30px rgba(201,169,107,.3)`
- 幅: モバイル `86%`、デスクトップ `min(420px, 86%)`
- パディング: `22px 18px`
- 「戻る。」(secondary): 透明背景 + gold ボーダー + gold 文字
- 「やめる。」(primary): gold グラデ + 暗文字

**演出**:

- overlay は 200ms ease で fade-in
- modal-card は `scale: .94 → 1` + `opacity: 0 → 1` (300ms cubic-bezier)

**コピー**: §1.6「ログアウト ◯◯」参照。

**遷移先**: 既存と同じ。`onConfirm` で logout 実行 → `/` へ push (既存ロジック維持)。

---

### 5.6 セッション切れモーダル

**ファイル**: `web/components/auth/SessionExpiredModalHost.tsx` (改修)

**役割**: トークン失効時の通知。1.5 秒後に自動で `/login?from=session_expired` へ遷移する (既存仕様維持)。

**レイアウト**: ログアウトモーダルと同じ構造で、ボタンは 1 個のみ。

| 要素 | コピー |
|---|---|
| H2 | `離れすぎたな。` |
| サブ | 既存 `authMessages.SESSION_EXPIRED` を維持 (中立保持) |
| ボタン | `戻る。` |

**演出**: ログアウトモーダルと同じ。

**遷移先**: 既存と同じ (1.5 秒後または OK 押下で `/login?from=session_expired`)。

---

### 5.7 Toast 通知

**ファイル**: `web/components/order/Toast.tsx` (改修) ／ `web/components/order/ToastHost.tsx` (改修)

**役割**: ネットワークエラーや冪等性衝突時の通知。世界観コピーは入れず**機能優先で中立保持**。見た目だけ世界観に合わせる。

**スタイル変更**:

- 旧: `background: rgba(40, 40, 40, 0.92)` ／ `color: #fff`
- 新: `background: rgba(26, 18, 8, 0.92)` (暗茶) ／ `color: #f4ecd8` (cream) ／ `border: 1px solid rgba(201,169,107,.3)` ／ `box-shadow: 0 4px 14px rgba(201,169,107,.15)`
- フォント: `mono-pixel 13px` (DotGothic16)、letter-spacing `0.04em`

**コピー**: 既存 `errorMappers.ts` / `authMessages.ts` のメッセージをそのまま使う。リヴァイ調にしない理由は、エラーメッセージは事実伝達と復帰指示が最優先であり、世界観で曖昧にすると UX を損なうため。

**配置**: 既存仕様維持 (`position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%);`、最大 3 件)。

---

## 6. 共通アクセシビリティ・レスポンシブ

メイン画面以外の画面についても、§4.2 / §4.3 のアクセシビリティ・レスポンシブ規則を適用する。各画面の追加 a11y 配慮:

| 画面 | 追加 a11y |
|---|---|
| ランディング | デモボタンは `<button>`、押下後は `disabled` で再押下不可、押下完了後は `aria-label="体験用デモ完了"` |
| サインアップ／ログイン | フォームラベル・エラーは中立保持により WCAG 適合、`aria-describedby` でパスワードヒントを送信ボタンに紐付け |
| 完了画面 | 中央コピー・メトリクスは `role="status" aria-live="polite"`、自動遷移までの残り秒数は `<span class="sr-only">` で読み上げ |
| ログアウトモーダル | 既存 `role="dialog" aria-modal="true" aria-labelledby` を維持、focus trap は将来課題 |
| セッション切れ | 既存実装の `role="dialog" aria-modal="true"` を維持 |
| Toast | 既存 `role="status"` 維持、`aria-live="polite"` |

---

## 7. 要件トレーサビリティ

| 要件 ID | 充足箇所 |
|---|---|
| FR-AUTH-01 (新規登録) | §5.2 サインアップ |
| FR-AUTH-02 (ログイン) | §5.3 ログイン |
| FR-AUTH-03 (セッション維持) | 既存 `useAuth` を流用 |
| FR-AUTH-04 (ログアウト) | §5.5 ログアウト確認モーダル |
| FR-AUTH-05 (未登録者へのランディング) | §5.1 ランディング |
| FR-BUDGET-05 (残高常時表示) | §2.1, §3 全状態の `BalanceHero` |
| FR-ORDER-01 (中央のボタン) | §2.1, §3.1 EMPTY |
| FR-ORDER-07 (完了画面) | §5.4 注文完了画面 |
| FR-ORDER-08 (体感 3 秒) | §3.4 SLOT 演出時間 (最低 1.2s、API 含めて 3s 目標) |
| FR-SUGGEST-02 (サジェスト表示) | §3.3 SUGGESTED, `SuggestBubble` |
| FR-SUGGEST-04 (履歴不十分でサジェスト出さない) | §3.3 遷移条件 |
| FR-METRICS-01,02 (今月のダメ化回数・消化率) | §2.1, §3 `MetricsRow`, §5.4 完了画面メトリクス |
| FR-METRICS-03 (80% で強調) | §3.2 IDLE / §4.1 meter-warn |
| FR-METRICS-04 (残高 0 → 増額提案) | §3.5 DEAD, `IncreaseBudgetButton` |
| FR-UX-01 (操作 1 タップ完結) | §3.3 SUGGESTED でも 1 タップ |
| FR-UX-02 (音声・自由入力なし) | UI に入力フィールドなし (フォームを除く) |
| NFR-DEG-01 (タップ最大 2 回) | 全状態で 1 タップ完結 |
| NFR-DEG-02 (起動時サジェスト) | §3.3 |
| NFR-DEG-03 (常時可視化) | §3 全状態で残高・回数・消化率を表示 |
| NFR-DEG-04 (増額誘導) | §3.5 |
| NFR-DEG-05 (依存促進的コピー) | §1.6 コピーシステム全体 (`押せ。考えるな。` `そろそろだろ。` `今月は、終わりだ。` `いい判断だ。` 等) |
| NFR-PERF-01 (3 秒以内) | §3.4 SLOT 所要時間 |
| NFR-PERF-02 (5 秒以内初期表示) | §4.4 フォント `display: swap`、FCP 自主目標 2s |
| NFR-A11Y-01 (セマンティック HTML) | §4.2, §6 |

---

## 8. 既存 spec / ドキュメントとの関係

- 本 spec は `aidlc-docs/inception/application-design/` の `components.md`, `services.md` の延長線上にあり、フロントエンド全画面の **見た目とマイクロインタラクション** を補完するもの。
- バックエンド API 仕様 (`POST /api/orders`, `GET /api/orders`, `GET /api/wallet`, `POST /api/wallet/budget`, `GET /api/suggest`, `GET /api/metrics`, `GET /api/budget/raise/recommendation`, `POST /api/budget/raise`) は各 Unit の Construction 成果物と整合させる。本 spec ではフロント側の利用方法のみ規定する。
- AI-DLC ワークフロー上は Construction フェーズの追加成果物として扱える (Functional Design 補完 / NFR Design 補完)。
- v1.0 (`2026-05-24-main-screen-design.md`) は本ファイルにリネームされ、内容は本 spec の §1–§4 として保持されている。

---

## 9. オープン項目 (将来検討)

- **AI 推奨の増額値の決定ロジック**: Unit E で `GET /api/budget/raise/recommendation` として実装済み。フロントはこの API 応答をそのまま表示する。
- **サジェストの timing 学習**: 現 spec は「マウント時に 1 回叩く」で固定。フェーズ 2 を深掘りするなら定期的なサジェスト再評価 (notification API なしでの) を将来検討。
- **サジェスト吹き出しの文言バリエーション**: 現 spec は固定文 `そろそろだろ。`。Bedrock 側で時刻・履歴に応じた文言切替を実装するなら別途プロンプト設計が必要。
- **多カテゴリ対応**: 本 spec は「ご飯」カテゴリのみ。requirements.md `FR-ORDER-09` の SHOULD 要件である清掃・役所手続き等は、ボタンを縦に並べる or サジェストカードで掘り下げる将来案あり。
- **focus trap (モーダル)**: ログアウトモーダル・セッション切れモーダル共に focus trap は未実装。a11y 強化として将来対応。
- **エラー Toast の世界観統一**: 現 spec は中立コピー保持の方針。世界観コピーへの段階的置換は将来検討するが、UX 損失リスクが高いので慎重に。
