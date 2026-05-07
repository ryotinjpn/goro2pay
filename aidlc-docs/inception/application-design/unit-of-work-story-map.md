# Unit of Work Story Map — ゴロゴロPay

**Document Version**: 1.0
**Created**: 2026-05-07
**Related**: [stories.md](../user-stories/stories.md), [unit-of-work.md](./unit-of-work.md), [unit-of-work-dependency.md](./unit-of-work-dependency.md)

本ドキュメントは、[stories.md](../user-stories/stories.md) の **全 23 ユーザーストーリー** を **5 Unit of Work** にマッピングする。どのストーリーが Construction フェーズでどの Unit の per-unit ステージで扱われるかを明示する。

---

## 1. マッピングサマリ

| Unit | 主要担当ストーリー数 | 共同担当ストーリー数 | Unit 内主要 Epic |
|---|---|---|---|
| Unit A (auth) | 3 | 0 | Epic 0（部分） |
| Unit B (budget) | 7 | 0 | Epic 0（部分） + Epic 1（基盤） + Epic 3（リセット） |
| **Unit C (order)** ★ | 7 | 1 | **Epic 1（コア）** + Epic X-01 |
| Unit D (suggest) | 4 | 1 | Epic 2 + Epic X-02 |
| Unit E (metrics) | 5 | 0 | Epic 3 + Epic X-03 |
| **合計** | **26（再掲含む）** | **2** | 全 23 ストーリー |

**注**: 「再掲含む」とは、1 ストーリーが複数 Unit にまたがる場合に両方でカウントしていることを意味する（例: US-2-03 は C と D が共同担当）。純粋な 1 対 1 対応ストーリー数では 23 件すべてが網羅される。

---

## 2. Unit A: `auth`（認証）担当ストーリー

### 2.1 主要担当

| ID | ストーリー要約 | 対応 Epic | 主要機能要件 | 受入基準の要点 |
|---|---|---|---|---|
| US-0-01 | 新規登録でゴロゴロPay を始める | Epic 0 | FR-AUTH-01 | Cognito User Pool 登録、登録 5 タップ以下 |
| US-0-02 | ログインで前回の続きから始める | Epic 0 | FR-AUTH-02, FR-AUTH-03 | JWT 取得、失敗時のエラー表示、セッション 1h |
| US-0-04 | 初回メイン画面を表示する（認証部分） | Epic 0 | FR-AUTH-03, FR-UX-01 | ログイン済状態の維持、認証後の画面遷移 |

### 2.2 共同担当
なし

### 2.3 Unit A 不担当だが間接関係あり
- 全ストーリー（US-1-* / US-2-* / US-3-*）は認証済みを前提として動作するため、間接的に Unit A の健全性に依存

---

## 3. Unit B: `budget`（ダメ予算）担当ストーリー

### 3.1 主要担当

| ID | ストーリー要約 | 対応 Epic | 主要機能要件 | 受入基準の要点 |
|---|---|---|---|---|
| US-0-03 | ダメ予算を設定する | Epic 0 | FR-BUDGET-01, FR-BUDGET-02 | 30,000 円初期化、翌月から増額適用 |
| US-0-04 | 初回メイン画面を表示する（残高初期表示） | Epic 0 | FR-BUDGET-05 | 「残りダメ予算: ¥30,000」表示 |
| US-1-02 | 残高を常に視認できる | Epic 1 | FR-BUDGET-05 | 大きな残高表示、stale state 禁止 |
| US-1-04 | 残高不足時に今月ダメになれないことを知る | Epic 1 | FR-BUDGET-06 | 注文拒否、画面遷移 |
| US-1-05 | 二重引き落としを防ぐ | Epic 1 | NFR-REL-02 | 冪等性キーによる多重送信の無害化 |
| US-1-06 | 残高が負にならない | Epic 1 | NFR-REL-01 | DynamoDB 条件付き書き込み、ConditionalCheckFailed |
| US-3-05 | 月初にダメ予算が自動リセットされる | Epic 3 | FR-BUDGET-04 | EventBridge Scheduler、BudgetResetLog 記録 |

### 3.2 共同担当
なし

### 3.3 Unit B が他 Unit から呼ばれるパターン
- Unit C: `WalletService.Deduct` 呼び出し（中核依存）
- Unit E: `WalletRepository` / `BudgetSettingsRepository` の読取参照

---

## 4. Unit C: `order`（代行手配コア）★ 担当ストーリー

本 MVP の **コア Unit**。Comprehensive 深度。

### 4.1 主要担当

| ID | ストーリー要約 | 対応 Epic | 主要機能要件 | 受入基準の要点 |
|---|---|---|---|---|
| **US-1-01** | **「ご飯めんどくさい」ボタンで代行手配する** | **Epic 1** | **FR-ORDER-01~08** | **本 MVP の心臓、体感 3 秒以内** |
| US-1-03 | 注文履歴に代行手配が記録される | Epic 1 | FR-ORDER-06, FR-LEARNING-01 | OrderHistoryRepository.Insert、TTL 90 日 |
| US-1-07 | 注文完了画面から自動でメインへ戻る | Epic 1 | FR-UX-04 | 5 秒後に自動遷移 |
| US-X-01 | 決定疲れから解放される快感を得る（フェーズ 1 体験） | Epic X | NFR-DEG-01 | 1 タップ完結、意思決定回避の体験 |

### 4.2 共同担当

| ID | ストーリー要約 | 共同 Unit | Unit C の担当範囲 |
|---|---|---|---|
| US-2-03 | サジェストカードから 1 タップで注文する | Unit D | OrderService が SuggestService.ResolveSuggestion を呼び出し、通常フローと合流 |

### 4.3 Unit C の技術的見せ場（Comprehensive で詳細化）

- **Bedrock 推論フロー**: プロンプトテンプレート / 入出力例 / リトライ戦略 / フォールバック
- **冪等性処理**: IdempotencyRepository.TryAcquire の詳細、レースコンディション対策
- **条件付き書き込み**: DynamoDB UpdateItem の CAS パターン、失敗時の制御フロー
- **アダプタ層**: DeliveryAdapter の interface 設計、将来の実 API 差し替え戦略
- **体感 3 秒以内**: 各フェーズ（Bedrock / Wallet / Adapter）のレイテンシ予算配分

### 4.4 横串コンポーネントの利用
- `BedrockAdapter.InferOrderPlan`（主要）
- `DeliveryAdapter.PlaceOrder`（Mock 実装）
- `FallbackSuggestProvider.BuildFromHistory` / `Default`（フォールバック時）

---

## 5. Unit D: `suggest`（学習・先回り）担当ストーリー

### 5.1 主要担当

| ID | ストーリー要約 | 対応 Epic | 主要機能要件 | 受入基準の要点 |
|---|---|---|---|---|
| US-2-01 | アプリを開いた瞬間に先回りサジェストが表示される | Epic 2 | FR-SUGGEST-01, FR-SUGGEST-02 | Bedrock 推論、サジェストカード表示 |
| US-2-02 | 履歴が不十分なときはサジェストを出さない | Epic 2 | FR-SUGGEST-04 | 履歴 5 件未満で非表示 |
| US-2-04 | 行動履歴が蓄積されて学習される | Epic 2 | FR-LEARNING-01, FR-LEARNING-02 | 履歴テーブル参照、TTL 3 ヶ月 |
| US-X-02 | 自分より自分を知る AI に委ねる心地よさを得る（フェーズ 2 体験） | Epic X | NFR-DEG-02 | サジェスト精度と受諾速度 |

### 5.2 共同担当

| ID | ストーリー要約 | 共同 Unit | Unit D の担当範囲 |
|---|---|---|---|
| US-2-03 | サジェストカードから 1 タップで注文する | Unit C | `SuggestService.ResolveSuggestion` で保存済み plan を返す |

### 5.3 横串コンポーネントの利用
- `BedrockAdapter.InferSuggestion`（主要）
- `FallbackSuggestProvider.BuildFromHistory`（Bedrock 失敗時）

---

## 6. Unit E: `metrics`（ダメ化メトリクス）担当ストーリー

### 6.1 主要担当

| ID | ストーリー要約 | 対応 Epic | 主要機能要件 | 受入基準の要点 |
|---|---|---|---|---|
| US-3-01 | 今月のダメ化回数を見て優越感と空虚感を同時に感じる | Epic 3 | FR-METRICS-01 | MetricsService.GetMetrics |
| US-3-02 | ダメ予算消化率を見る | Epic 3 | FR-METRICS-02, FR-METRICS-03 | 80% 超で警告色 |
| US-3-03 | 残高 0 円時に「今月ダメになれません」画面を見る | Epic 3 | FR-METRICS-04 | BudgetEmptyScreen 遷移 |
| US-3-04 | 翌月予算の増額誘導モーダルで意思薄弱になる | Epic 3 | FR-METRICS-04, FR-METRICS-05 | BudgetRaiseService、+50% 推奨 |
| US-X-03 | 自力で何もできない自分に気づき、それでも戻れないループを完成する（フェーズ 3 体験） | Epic X | NFR-DEG-04 | 退化ループの完成、増額受諾 |

### 6.2 共同担当
なし

### 6.3 Unit E の集計対象
- Unit B: 残高、月間予算（読取）
- Unit C: 注文履歴（読取、件数・合計額集計）

---

## 7. 全 23 ストーリー → Unit の網羅マッピング表

すべてのストーリーが少なくとも 1 Unit に割り当てられていることを一覧で確認。

| Story ID | Title | Primary Unit | Secondary Unit |
|---|---|---|---|
| US-0-01 | 新規登録でゴロゴロPay を始める | **A** | — |
| US-0-02 | ログインで前回の続きから始める | **A** | — |
| US-0-03 | ダメ予算を設定する | **B** | — |
| US-0-04 | 初回メイン画面を表示する | **A + B** | — |
| US-1-01 | 「ご飯めんどくさい」ボタンで代行手配する ★core | **C** | — |
| US-1-02 | 残高を常に視認できる | **B** | — |
| US-1-03 | 注文履歴に代行手配が記録される | **C** | — |
| US-1-04 | 残高不足時に今月ダメになれないことを知る | **B** | C（注文時の分岐） |
| US-1-05 | 二重引き落としを防ぐ | **B** | C（idempotencyKey 発行元） |
| US-1-06 | 残高が負にならない | **B** | — |
| US-1-07 | 注文完了画面から自動でメインへ戻る | **C** | — |
| US-2-01 | アプリを開いた瞬間に先回りサジェストが表示される | **D** | — |
| US-2-02 | 履歴が不十分なときはサジェストを出さない | **D** | — |
| US-2-03 | サジェストカードから 1 タップで注文する | **C + D** | — |
| US-2-04 | 行動履歴が蓄積されて学習される | **D** | C（履歴 Insert 側） |
| US-3-01 | 今月のダメ化回数を見て優越感と空虚感を同時に感じる | **E** | — |
| US-3-02 | ダメ予算消化率を見る | **E** | — |
| US-3-03 | 残高 0 円時に「今月ダメになれません」画面を見る | **E** | B（残高判定元） |
| US-3-04 | 翌月予算の増額誘導モーダルで意思薄弱になる | **E** | B（予算更新） |
| US-3-05 | 月初にダメ予算が自動リセットされる | **B** | — |
| US-X-01 | 決定疲れから解放される快感を得る（フェーズ 1 体験） | **C** | — |
| US-X-02 | 自分より自分を知る AI に委ねる心地よさを得る（フェーズ 2 体験） | **D** | — |
| US-X-03 | 自力で何もできない自分に気づき、それでも戻れないループを完成する（フェーズ 3 体験） | **E** | — |

**カバレッジ確認**:
- ✅ 全 23 ストーリーに Primary Unit が割り当てられている
- ✅ Epic 0（US-0-*）: A + B に分散、整合
- ✅ Epic 1（US-1-*）: B + C に分散（B が基盤、C がコア）
- ✅ Epic 2（US-2-*）: D 主、US-2-03 のみ C と共同
- ✅ Epic 3（US-3-*）: B + E に分散（B がリセット、E が表示）
- ✅ Epic X（US-X-*）: 各フェーズに対応した Unit（C/D/E）へ割り当て

---

## 8. per-unit Construction での扱い

各 Unit の Construction フェーズでは、担当ストーリーを以下のステージ成果物に展開する。

### 8.1 Functional Design で扱うストーリー（per-unit）

| Unit | Functional Design で展開するストーリー |
|---|---|
| A | US-0-01, US-0-02, US-0-04（認証部分） |
| B | US-0-03, US-1-02, US-1-04, US-1-05, US-1-06, US-3-05 |
| C | **US-1-01（最重要）**, US-1-03, US-1-07, US-2-03（Resolve 側）, US-X-01 |
| D | US-2-01, US-2-02, US-2-03（主）, US-2-04, US-X-02 |
| E | US-3-01, US-3-02, US-3-03, US-3-04, US-X-03 |

### 8.2 NFR Requirements で扱う NFR（per-unit）

| Unit | NFR 項目 |
|---|---|
| A | NFR-DEG-01（登録 5 タップ以下） |
| B | NFR-REL-01（残高不変）, NFR-REL-02（冪等性）, NFR-DEG-03（可視化） |
| C | NFR-PERF-01（3 秒以内）, NFR-DEG-01（1 タップ）, NFR-REL-02（C 側の冪等性キー発行） |
| D | NFR-DEG-02（サジェスト提示） |
| E | NFR-DEG-03（消化率表示）, NFR-DEG-04（増額誘導）, NFR-DEG-05（自虐コピー） |

### 8.3 Build and Test フェーズの統合試験対象

全 Unit 統合後、以下のシナリオで E2E 試験を実施:
- **ゴールデンパス**: 登録 → 予算設定 → 「ご飯めんどくさい」押下 → 完了画面（US-1-01）
- **冪等性試験**: 同一 idempotencyKey で 3 回送信 → 1 回のみ減算（US-1-05）
- **残高不足試験**: 残高 500 円で注文試行 → 402 応答 → BudgetEmpty 遷移（US-1-04, US-3-03）
- **サジェスト受諾**: 履歴 5 件蓄積後、サジェスト → 1 タップ注文（US-2-01, US-2-03）
- **月初リセット**: 手動 Scheduler 実行 → 全ユーザ残高リセット確認（US-3-05）

---

## 9. 審査観点へのトレーサビリティ

| 審査観点 | 対応 |
|---|---|
| ビジネス意図の明確さ | §7 網羅マッピングでペルソナのダメ化フェーズが Unit 構成に確実に写像されることを確認 |
| 創造性とテーマ適合性 | Epic X（ダメ化UX 体験ストーリー）が C/D/E に 1 本ずつ対応し、テーマの 3 フェーズが Unit 群に縦貫 |
| Unit 分解の適切さ | §1 マッピングサマリと §7 網羅表で 23 ストーリー全てが Unit に所属することを保証、カバレッジ 100% |
| ドキュメント品質 | Unit ごと詳細節 + §7 一覧表 + §8 Construction 引き継ぎ情報で立体構造 |
