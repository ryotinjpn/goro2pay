// Package wallet は Unit B (`budget` / ダメ予算) のドメイン層を提供する。
//
// 凍結 IF (unit-interfaces.md §3.1):
//   - WalletService interface (GetBalance / SetBudget / Deduct / ResetAll)
//   - 公開 DTO: WalletSnapshot / DeductResult / ResetResult
//   - Sentinel errors: ErrInsufficientBalance / ErrIdempotencyConflict / ErrBudgetOutOfRange
//
// 設計成果物:
//   - aidlc-docs/construction/budget/functional-design/ (UC-B-01〜06, VR/DR/CR/PR)
//   - aidlc-docs/construction/budget/nfr-design/ (P-REL-01/02/03, P-OBS-01/02, P-TEST-01)
//   - aidlc-docs/construction/budget/infrastructure-design/ (DynamoDB 4 テーブル + Scheduler)
package wallet
