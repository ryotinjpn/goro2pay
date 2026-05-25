# Tech Stack Decisions — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. 継承テックスタック（Unit A/B/C/D と共通）

Unit E は新規テックスタックを導入しない。以下はすべて既存の決定を継承する。

| 領域 | 技術 | 決定出典 |
|---|---|---|
| Backend 言語 | Go 1.24 | Unit A |
| HTTP フレームワーク | Gin | Unit A |
| DynamoDB SDK | aws-sdk-go-v2 | Unit A/B/C |
| 並列処理 | `golang.org/x/sync/errgroup` | Unit C P-DI-01 |
| PBT | `gopkg.in/check.v1` + `github.com/leanovate/gopter` | Unit C P-PBT-01 |
| Frontend フレームワーク | Next.js 14 (App Router) | Unit A |
| 状態管理 | Jotai | Unit A/C |
| データフェッチ | TanStack Query v5 | Unit A/C |
| スタイリング | Tailwind CSS | Unit A |
| Frontend テスト | Vitest + React Testing Library | Unit A/C |
| IaC | Terraform | Unit A |
| 新規 DynamoDB テーブル | **なし** | unit-of-work.md |

---

## 2. Unit E 固有の技術選定

### 2.1 DynamoDB 並列読み取り: errgroup (Q-N1=C)

**選定**: `golang.org/x/sync/errgroup`

**理由**:
- Unit C で採用済み（P-DI-01）のパターンを踏襲
- `errgroup.WithContext` でキャンセル伝播が自動
- `g.Wait()` でいずれかのエラーを即座に捕捉（Q-N2=A の全失敗即 500 と整合）

```go
import "golang.org/x/sync/errgroup"
```

### 2.2 PBT ライブラリ: gopter (Q-N3=A)

**選定**: `github.com/leanovate/gopter`

**理由**: Unit C で採用済み。`ConsumptionRate` と `ComputeRecommendedBudget` の不変条件テスト（P-E-PBT-01 / P-E-PBT-02）に適用。

### 2.3 アニメーション: Tailwind CSS `animate-pulse` (Q-N4=B)

**選定**: Tailwind CSS 組み込みユーティリティ

**理由**: 追加ライブラリ不要。`animate-pulse` は CSS `@keyframes pulse` で実装済み。Unit A/C と同じ Tailwind バージョンで利用可能。

---

## 3. 依存パッケージ（追加なし）

Unit E が新たに追加する go.mod / package.json エントリはなし。すべて Unit A/B/C/D で導入済みのパッケージを使用する。

```
golang.org/x/sync  ← errgroup (既存)
github.com/leanovate/gopter ← PBT (Unit C で追加済み)
```
