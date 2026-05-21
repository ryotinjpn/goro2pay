# Unit B `budget` — Business Rules

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: B (`budget` / ダメ予算)
**Stage**: Functional Design (Construction Phase)
**Depth**: Standard

---

## 0. このドキュメントの目的

Unit B `budget` のビジネスルール（Decision Rules / Validation Rules / Constraint Rules / Policy Rules）を分類して列挙する。各ルールは Plan の対話ヒアリング結果を根拠とし、対応するストーリー / NFR にトレース可能な形で記述する。

ルールは技術非依存で記述し、実装上の表現（Go のエラー型・DynamoDB Condition Expression 等）は Code Generation で扱う。

---

## 1. ルール分類サマリ

| 分類 | ルール数 | 主な対象 |
|---|---|---|
| Validation Rules（入力検証） | 6 | `SetBudget`, `Deduct` のリクエスト |
| Decision Rules（判断分岐） | 5 | 初回判定、冪等性ヒット、残高不足 |
| Constraint Rules（不変条件） | 5 | 残高、予算、リセット冪等性 |
| Policy Rules（運用方針） | 6 | リセット、保持期間、ダメ化UX |
| **合計** | **22** | |

---

## 2. Validation Rules（入力検証）

### VR-B-01: 月間予算の範囲

- **対象**: `SetBudget(userID, monthlyBudget)`
- **ルール**: `1 ≤ monthlyBudget ≤ 100,000`
- **違反時**: `ErrValidation` を返す（HTTP 400 `VALIDATION_FAILED`）
- **根拠**: Q-B1 = A、Application Design / requirements.md
- **対応ストーリー**: US-0-03
- **テスト境界値**: `0`, `1`, `100000`, `100001`, `-1`

### VR-B-02: 月間予算の刻み

- **対象**: `SetBudget(userID, monthlyBudget)`
- **ルール**: `monthlyBudget % 1000 == 0`（1,000 円刻み）
- **違反時**: `ErrValidation`
- **根拠**: Q-B11 補足 = γ
- **対応ストーリー**: US-0-03
- **テスト境界値**: `1000`（OK）, `1500`（NG）, `30000`（OK）, `30001`（NG）

### VR-B-03: `Deduct` の amount は正値

- **対象**: `Deduct(userID, amount, idempotencyKey)`
- **ルール**: `amount > 0`
- **違反時**: `ErrValidation`
- **根拠**: ドメイン上の自明な制約（負の減算 = 加算は許容しない）
- **対応ストーリー**: US-1-05, US-1-06
- **テスト境界値**: `0`（NG）, `1`（OK）, `-1`（NG）

### VR-B-04: `idempotencyKey` の形式

- **対象**: `Deduct(userID, amount, idempotencyKey)`
- **ルール**: `idempotencyKey` は `{userID}:{ulid}` 形式（ULID は 26 文字の Crockford's Base32）
- **違反時**: `ErrValidation`（HTTP 400 `VALIDATION_FAILED`）
- **根拠**: Q-B4 = A
- **対応ストーリー**: US-1-05
- **正規表現** (参考): `^[a-zA-Z0-9_-]+:[0-9A-HJKMNP-TV-Z]{26}$`

### VR-B-05: `idempotencyKey` のユーザ照合

- **対象**: `Deduct(userID, amount, idempotencyKey)`
- **ルール**: `idempotencyKey` のプレフィックス（`:` の左側）が認証コンテキストの `userID` と一致すること
- **違反時**: `ErrValidation`（他人のキーを横取りできない）
- **根拠**: Q-B4 = A の userID プレフィックス採用
- **対応ストーリー**: US-1-05（セキュリティ補強）

### VR-B-06: `userID` 必須

- **対象**: 全ドメインサービス
- **ルール**: `userID` が空文字列または不正形式の場合は処理せず `ErrUnauthorized`
- **違反時**: `ErrUnauthorized`（HTTP 401）
- **根拠**: Unit A `AuthContextService` の責務だが、Unit B 側でも防御的に検査
- **対応ストーリー**: 全ストーリー（前提）

---

## 3. Decision Rules（判断分岐）

### DR-B-01: 初回 vs 変更の判定

- **対象**: `SetBudget(userID, monthlyBudget)`
- **ルール**: `Wallet` が存在しない（NotFound）場合を **初回**、それ以外を **変更** とする
- **動作**:
  - 初回: `BudgetSettings` を新規作成、`Wallet` を `balance = monthlyBudget` で新規作成
  - 変更: `BudgetSettings` を更新、`Wallet.balance` を差分調整（DR-B-02）
- **根拠**: Q-B3 = B
- **対応ストーリー**: US-0-03

### DR-B-02: SetBudget 変更時の Wallet 差分調整

- **対象**: `SetBudget(userID, monthlyBudget)` の変更パス
- **ルール**:
  ```
  delta = newBudget - oldBudget
  if delta > 0:
      Wallet.balance += delta             # 増額分を当月残高に加算
  elif delta < 0:
      Wallet.balance = min(balance, newBudget)  # 減額時は新予算で打ち切り
  # delta == 0 は no-op
  ```
- **根拠**: Q-B2 = B（即時反映 + 差分調整）
- **対応ストーリー**: US-0-03（変更時挙動）

### DR-B-03: 冪等性ヒット判定

- **対象**: `Deduct(...)` の `IdempotencyRepository.TryAcquire(key, payload)`
- **ルール**:
  ```
  if !acquired:
      # 既存レコードあり
      if hash(prevPayload) == hash(currentPayload):
          return prevResponse with idempotent=true
      else:
          return ErrIdempotencyConflict   # HTTP 409
  else:
      # 新規取得
      proceed to DeductConditional
  ```
- **根拠**: Q-B4 = A, Q-B5 = A
- **対応ストーリー**: US-1-05

### DR-B-04: 残高不足の判定主体

- **対象**: `Deduct(...)` 実行時
- **ルール**: 残高不足の最終判定は **`WalletRepository.DeductConditional` の DynamoDB ConditionExpression** が担う。Unit C 側で事前に `GetBalance` チェックは不要
- **根拠**: Q-B7 = A
- **動作**: `balance < amount` の場合 `ErrInsufficientBalance` を返す（HTTP 402 `INSUFFICIENT_BALANCE`）
- **対応ストーリー**: US-1-04, US-1-06

### DR-B-05: 月初リセットのスキップ判定

- **対象**: `ResetAll(ctx)` の各ユーザ処理
- **ルール**:
  ```
  if BudgetResetLogRepository.Get(resetDate, userID) != nil:
      skip  # 既処理（再実行時の冪等性）
  if BudgetSettings.Get(userID) == nil:
      skip  # 予算未設定ユーザ
  else:
      proceed to reset
  ```
- **根拠**: Q-B8 = A
- **対応ストーリー**: US-3-05

---

## 4. Constraint Rules（不変条件）

### CR-B-01: 残高は非負

- **不変条件**: `∀wallet: wallet.balance >= 0`
- **保証手段**: `DeductConditional` の `ConditionExpression: balance >= :amount`、ConditionFailed で `ErrInsufficientBalance`
- **対応ストーリー**: US-1-06
- **PBT 候補**: P-B-1（任意の `Deduct` シーケンスで `balance < 0` が発生しない）

### CR-B-02: 残高は月間予算以下

- **不変条件**: `∀user: 0 ≤ Wallet.balance ≤ BudgetSettings.monthlyBudget`
- **保証手段**: `SetBudget` 減額時の `min(balance, newBudget)` 打ち切り、`Deduct` は減算のみ、`ResetTo` は `balance = monthlyBudget`
- **対応ストーリー**: US-0-03 変更時、US-1-06
- **PBT 候補**: P-B-2

### CR-B-03: 月間予算の範囲・刻み

- **不変条件**: `1 ≤ BudgetSettings.monthlyBudget ≤ 100,000` かつ `monthlyBudget % 1000 == 0`
- **保証手段**: `SetBudget` の Validation（VR-B-01, VR-B-02）
- **対応ストーリー**: US-0-03

### CR-B-04: 冪等性 — 残高は同一キーで最大 1 回しか減らない

- **不変条件**: `∀key: |{successful Deduct(key)}| ≤ 1`
- **保証手段**: `IdempotencyRepository.TryAcquire` の Put with `ConditionExpression: attribute_not_exists(idempotencyKey)`
- **対応ストーリー**: US-1-05
- **PBT 候補**: P-B-3（同一キーの並列・重複呼び出しで残高減算は 1 回分のみ）

### CR-B-05: 月初リセットは 1 ユーザ 1 月で最大 1 回

- **不変条件**: `∀(date, user): |{ResetTo(user) at date}| ≤ 1`
- **保証手段**: `BudgetResetLog.(resetDate, userID)` 複合キー、Insert with `ConditionExpression: attribute_not_exists(resetDate)`
- **対応ストーリー**: US-3-05

---

## 5. Policy Rules（運用方針）

### PR-B-01: 予算リセットは完全リセット（持ち越しなし）

- **方針**: 月初リセット時、前月残高は失効させ `Wallet.balance` を `BudgetSettings.monthlyBudget` で上書きする
- **根拠**: Q-B10 = A、ダメ化UX（消化促進、節約を肯定しない）
- **対応ストーリー**: US-3-05
- **対応 NFR**: NFR-DEG-04（退化ループの完成）

### PR-B-02: SetBudget は即時反映

- **方針**: 予算変更は即座に反映し、当月残高も差分調整する。「翌月から適用」モデルは採用しない
- **根拠**: Q-B2 = B、ダメ化UX（即時甘やかし、ストレス排除）
- **対応ストーリー**: US-0-03
- **対応 NFR**: NFR-DEG-01（低摩擦オンボーディング）

### PR-B-03: 冪等性レコードの保持期間 24 時間

- **方針**: `IdempotencyRecord` は TTL 24 時間で自動削除
- **根拠**: Q-B4 = A、通信障害リトライ・モバイル復帰のカバー範囲
- **対応ストーリー**: US-1-05

### PR-B-04: 失敗結果も冪等性レコードに保存

- **方針**: `Deduct` が `ErrInsufficientBalance` 等の業務エラーで失敗した場合も、エラー種別を `IdempotencyRecord.response` に保存する。同一キーの再送には保存通り返却
- **根拠**: Q-B5 = A、race condition 回避
- **対応ストーリー**: US-1-05

### PR-B-05: 残高表示の鮮度ポリシー

- **方針**:
  - **バックエンド**: `GetBalance` は DynamoDB ConsistentRead
  - **フロント**: TanStack Query で 30 秒間 fresh、`Deduct` / `SetBudget` 成功時に自動 invalidate
- **根拠**: Q-B6 = A、NFR-DEG-03（残高常時可視化）
- **対応ストーリー**: US-1-02

### PR-B-06: 月初リセットの実行方式

- **方針**: 単一 Lambda の逐次処理。EventBridge Scheduler `cron(0 15 L * ? *)` UTC で起動。失敗時は手動 Lambda 再 invoke で対応（CLI 等の追加実装不要、`(resetDate, userID)` 主キーで自動的に再実行性確保）
- **根拠**: Q-B8 = A
- **対応ストーリー**: US-3-05

---

## 6. ストーリー × ルール トレーサビリティ

| ストーリー | 主要ルール |
|---|---|
| US-0-03 ダメ予算を設定する | VR-B-01, VR-B-02, DR-B-01, DR-B-02, CR-B-03, PR-B-02 |
| US-0-04 初回メイン画面（残高初期表示） | DR-B-01（初回パスで Wallet 作成）, PR-B-05 |
| US-1-02 残高を常に視認できる | PR-B-05 |
| US-1-04 残高不足時にダメになれない | DR-B-04 |
| US-1-05 二重引き落とし防止 | VR-B-03, VR-B-04, VR-B-05, DR-B-03, CR-B-04, PR-B-03, PR-B-04 |
| US-1-06 残高が負にならない | CR-B-01, CR-B-02, DR-B-04 |
| US-3-05 月初リセット | DR-B-05, CR-B-05, PR-B-01, PR-B-06 |

---

## 7. NFR × ルール トレーサビリティ

| NFR | 関連ルール |
|---|---|
| NFR-REL-01（残高不変条件） | CR-B-01, CR-B-02 |
| NFR-REL-02（冪等性） | DR-B-03, CR-B-04, PR-B-03, PR-B-04 |
| NFR-DEG-01（低摩擦オンボーディング） | PR-B-02 |
| NFR-DEG-03（残高常時可視化） | PR-B-05 |
| NFR-DEG-04（退化ループ） | PR-B-01 |

---

## 8. 例外・エラーマッピング（HTTP 応答レベル）

API 応答時のエラーコード対応:

| ルール違反 | Go エラー | HTTP | API Code |
|---|---|---|---|
| VR-B-01 〜 VR-B-04 | `ErrValidation` | 400 | `VALIDATION_FAILED` |
| VR-B-05, VR-B-06 | `ErrUnauthorized` | 401 | `UNAUTHORIZED` |
| DR-B-04 残高不足 | `ErrInsufficientBalance` | 402 | `INSUFFICIENT_BALANCE` |
| DR-B-03 異 payload | `ErrIdempotencyConflict` | 409 | `IDEMPOTENCY_CONFLICT` |
| 想定外エラー | `ErrInternal` | 500 | `INTERNAL_ERROR` |

これらは Application Design 既存の定義 (component-methods.md §7) と整合する。
