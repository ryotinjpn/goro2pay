# Unit D (`suggest`) — NFR Requirements Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Requirements
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Prerequisite**: Functional Design 承認済み（PR #100 マージ済み）、Unit C NFR Requirements（PR #78 マージ済み）= Bedrock 横串の先行決定
**Related**: 凍結契約 [unit-interfaces.md](../interfaces/unit-interfaces.md) §5・§7、[suggest/functional-design/](../suggest/functional-design/)、先行 [order/nfr-requirements/](../order/nfr-requirements/)（Bedrock 継承元）、[budget/nfr-requirements/](../budget/nfr-requirements/)

---

## 1. Plan の目的

Unit D `suggest` の NFR を数値・しきい値・ポリシーとして確定する。Unit D は **Bedrock を Unit C と共有**するため、Bedrock 関連 NFR（モデル / リトライ / タイムアウト / スロットリング / mock / コスト）は **Unit C NFR-R（NFRC-Cxx）を継承**し、Unit D 固有の論点（サジェスト固有のレイテンシ・Suggestion テーブル・useSuggestion・PBT 範囲）に絞って確定する。

## 2. 担当ストーリー / FD 引き継ぎ

| Story | NFR 観点 |
|---|---|
| US-2-01 起動時サジェスト | GetSuggestion レイテンシ（NFR-PERF-02 メイン画面 5s 内） |
| US-2-02 履歴不足非表示 | 履歴判定の応答時間 |
| US-2-03 1タップ注文（Resolve） | ResolveSuggestion レイテンシ（Unit C 3s パス内、DynamoDB Get のみ） |
| US-2-04 学習 | OrderHistory 読取（Unit C 所有） |
| US-X-02 自己委譲 | フォールバック透過性 |

FD 確定（再ヒアリング不要）: 履歴十分5件 / 直近3h抑制 / Bedrock 1.5s×1リトライ→BuildFromHistory / TTL30分 / Resolve失効時nil / マウント時1回取得（BR-D01〜D20）。

## 3. Unit C/B からの継承方針（確認のみ、原則そのまま踏襲）

| 項目 | 継承元 | Unit D の方針 |
|---|---|---|
| Bedrock モデル | NFRC-C20 | `jp.anthropic.claude-haiku-4-5-20251001-v1:0`（ap-northeast-1 IP）/ Converse API |
| Bedrock リトライ/タイムアウト | NFRC-C06/C07 | 1.5s/回・1リトライ・待機0（FD Q-DF4 と一致） |
| Bedrock スロットリング | NFRC-C11 | リトライ→フォールバック、ThrottlingException はメトリクスフィルタ検知 |
| Bedrock テスト | NFRC-C16 | go test/CI は mock 必須、dev/stg/prd は実呼出（IS_TEST 分岐） |
| 観測性 | NFRC-C14 | **カスタムメトリクス不実装**（X-Ray なし / PutMetricData なし）、Logs メトリクスフィルタのみ |
| 構造化ログ | NFRC-C12 | 共通 8 項目 + suggest 固有項目 |
| Lambda | NFRC-C18 | 共有 API Lambda 256MB / arm64 / 10s（Bedrock 用、Unit C が設定済み） |
| Bedrock PII | NFRC-C24 | 履歴要約のみ送信、プロンプト/レスポンス本文はログ非記録 |
| 認証 | NFRC-C25 | Unit A middleware 委譲 |
| 同時利用/スケール | NFRC-C11 | デモ数十人、オンデマンド |

## 4. 生成する Artifacts（Plan 承認 + 回答後）

1. `aidlc-docs/construction/suggest/nfr-requirements/nfr-requirements.md`（NFRD-Dxx）
2. `aidlc-docs/construction/suggest/nfr-requirements/tech-stack-decisions.md`（Unit A/B/C スタック継承 + suggest 固有の追認）

## 5. 作業手順（Checkboxes）

- [x] §6 の確認質問（Q-ND1〜Q-ND8）にユーザが回答（1 問ずつ対話、全 A）
- [x] 曖昧さ・矛盾を点検（Unit C 継承との整合、矛盾なし）
- [x] ユーザによる Plan 承認（2026-05-25「ok」）
- [x] `nfr-requirements.md` 生成
- [x] `tech-stack-decisions.md` 生成
- [ ] 完了メッセージ（2-option ゲート）
- [ ] aidlc-state.md / audit.md 更新、コミット

---

## 6. 確認質問（Q-ND1 〜 Q-ND8）

> 1 問ずつ対話形式で提示。**(推奨)** は調査に基づく既定案。「全部推奨で」で一括採用可。

#### Q-ND1 — GetSuggestion の E2E レイテンシ SLO
起動時に `GET /api/suggest`。NFR-PERF-02（メイン画面初期表示 5 秒）の予算内。Bedrock 1.5s×2（リトライ時最悪 3.0s）+ 履歴読取。

- A) **p95 ≤ 2.5s（Bedrock 1 回）/ リトライ時最悪 ~3.0s、NFR-PERF-02 の 5s 予算内。単体ハードSLOは設けず Logs で監視** **(推奨)**
- B) p95 ≤ 2.0s に厳しめ設定
- C) 単体 SLO なし、NFR-PERF-02 全体で管理
- D) Other

[Answer]: **A**（p95 ≤ 2.5s（Bedrock 1回）/ リトライ時最悪 ~3.0s、NFR-PERF-02 の 5s 予算内。単体ハード SLO は設けず Logs 監視）

#### Q-ND2 — Bedrock 関連 NFR の Unit C 継承
モデル / リトライ / タイムアウト / スロットリング / mock を Unit C（NFRC-C06/07/11/16/20）からそのまま継承するか。

- A) **すべて継承**（claude-haiku-4-5 jp IP / Converse / 1.5s×1 / throttle→fallback / mock）。InferSuggestion は InferOrderPlan と同じ Adapter ポリシー **(推奨)**
- B) 一部変更（[Answer] に指定）
- C) Other

[Answer]: **A**（Bedrock NFR をすべて Unit C から継承: claude-haiku-4-5 jp IP / Converse / 1.5s×1リトライ / throttle→fallback / mock。InferSuggestion は InferOrderPlan と同ポリシー）

#### Q-ND3 — GoroPay_Suggestion の DynamoDB キャパシティ
凍結契約: PK `suggestionId` / TTL `expiresAt`(30分)。Unit B/C は無料枠前提のプロビジョンド 1RCU/1WCU。

- A) **プロビジョンド 1 RCU / 1 WCU（Unit B/C と統一、無料枠、バースト吸収）** **(推奨)**
- B) On-Demand（PAY_PER_REQUEST）
- C) Other

[Answer]: **A**（プロビジョンド 1 RCU / 1 WCU、Unit B/C と統一、無料枠・バースト吸収）

#### Q-ND4 — 観測性（カスタムメトリクス方針 + ログ項目 + アラーム）
Unit C は NFRC-C14 で「カスタムメトリクス不実装、Logs メトリクスフィルタのみ」。Unit D も揃えるか。suggest 固有ログ項目（`suggestionId` / `fallbackUsed` / `historyCount` / `bedrockLatencyMs` / `bedrockAttempt` / `suppressedRecentOrder`）。

- A) **Unit C と完全統一：カスタムメトリクスなし、構造化ログ + suggest 固有項目。フォールバック率の監視は Unit C NFRC-C13-3 と同様メトリクスフィルタで（Unit D 分のアラームは Infra Design で要否判断）** **(推奨)**
- B) Unit D 専用のカスタムメトリクスを出す（NFR-OBS-02 と差）
- C) ログのみ、アラームも設けない
- D) Other

[Answer]: **A**（Unit C と完全統一: カスタムメトリクス不実装、構造化ログ + suggest 固有項目（suggestionId/fallbackUsed/historyCount/bedrockLatencyMs/bedrockAttempt/suppressedRecentOrder）。フォールバック率はメトリクスフィルタ、Unit D 用アラーム要否は Infra Design で判断）

#### Q-ND5 — Lambda 設定
`GET /api/suggest` は共有 API Lambda 上。Unit C が Bedrock 用に 256MB / arm64 / 10s に設定済み。

- A) **共有 API Lambda の設定（256MB / arm64 / 10s）をそのまま利用（GetSuggestion も Bedrock を使うため 256MB が妥当）** **(推奨)**
- B) suggest 用に別設定（[Answer] に指定）
- C) Other

[Answer]: **A**（共有 API Lambda の 256MB / arm64 / 10s をそのまま利用。GetSuggestion も Bedrock 使用のため 256MB 妥当）

#### Q-ND6 — useSuggestion の TanStack Query 設定
FD Q-DF10=A「マウント時 1 回、再評価なし」。Unit C は `useOrderHistory` staleTime 60s。

- A) **`staleTime: Infinity` 相当（または十分長く）、`refetchOnWindowFocus: false`、`retry: 0`、queryKey `['suggestion']`。マウント時 1 回のみ取得（BR-D14）。取得失敗は hasSuggestion=false 相当に丸めカード非表示** **(推奨)**
- B) 別設定（[Answer] に指定）
- C) Other

[Answer]: **A**（staleTime 実質無限 / refetchOnWindowFocus:false / retry:0 / queryKey ['suggestion']。マウント時 1 回のみ、失敗は hasSuggestion=false 相当に丸めカード非表示。BR-D14 と整合）

#### Q-ND7 — Property-Based Testing の適用範囲
Extension: Partial。Unit C は P-1/P-3（冪等性）を gopter で。Unit D は冪等性のような強い不変条件が薄い（提案生成は副作用が小さい）。

- A) **PBT は軽量適用：`BuildFromHistory` の最頻判定の決定性 / レスポンス→props 変換のラウンドトリップを gopter（backend）/ fast-check（frontend）で 1〜2 プロパティ。主要ロジックは通常ユニットテストで担保** **(推奨)**
- B) PBT は適用せず通常ユニットテストのみ（Unit D は PBT 対象外と明記）
- C) Other

[Answer]: **A**（軽量適用: BuildFromHistory 最頻判定の決定性 / レスポンス→props ラウンドトリップを gopter・fast-check で 1〜2 プロパティ。他は通常ユニットテスト）

#### Q-ND8 — Bedrock コスト
GetSuggestion は起動時に 1 回 Bedrock を呼ぶ。Unit C と同じモデル・予算枠。

- A) **Unit C と Bedrock 月次予算を共有（合算 $10/月、AWS Budgets は Unit C が設定済み）。suggest 分は「起動毎 1 回・入力~500tok/出力~200tok」想定で増分は小さい。超過監視は既存 Budgets に相乗り** **(推奨)**
- B) suggest 専用の予算枠・アラートを別途設ける
- C) Other

[Answer]: **A**（Unit C と Bedrock 月次予算を共有（合算 $10/月、AWS Budgets は Unit C 設定済みに相乗り）。suggest 増分は小さい）

---

## 7. 矛盾チェック観点

1. Q-ND1 の SLO が NFR-PERF-02（5s）と FD の Bedrock タイムアウト（1.5s×2）に整合するか
2. Q-ND2/Q-ND4/Q-ND5 が Unit C NFRC-C06/07/11/14/16/18/20 と矛盾しないか
3. Q-ND6 が FD Q-DF10（マウント時1回）と整合するか
4. 「depends / たぶん / 標準で」等の曖昧回答がないか

## 8. 完了条件

- Q-ND1〜Q-ND8 回答 + 矛盾なし確認 + Plan 承認
- `nfr-requirements.md` / `tech-stack-decisions.md` 生成
- 2-option 完了ゲート確認

## 9. 次ステージ（NFR Design）への引き継ぎ

Q-ND4 のログ項目・アラーム方針は NFR Design の logical-components / パターンに、Q-ND1 の SLO は CloudWatch 監視条件に直結。Bedrock 継承事項は Unit C の nfr-design-patterns（P-RETRY-01 等）を参照して Unit D 版に展開。
