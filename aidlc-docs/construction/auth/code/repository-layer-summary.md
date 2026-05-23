# Auth Unit — Repository Layer Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation Step 8 (スキップ)

## 結論: Unit A は Repository を所有しない

unit-of-work.md §3.1 / unit-interfaces.md §2 の通り、Unit A (auth) は:

- DynamoDB テーブルを所有しない
- ユーザ情報は **Amazon Cognito User Pool** に集約 (NFR-COMP-02 / A-NFR-COMP-01)
- アプリ DB には **userId (Cognito sub) のみ** を外部キーとして保持

そのため、Unit A の Code Generation で `internal/repo/auth/` 等の Repository パッケージは**作成しない**。

## userId の永続化責務

| Unit | テーブル | userId の格納 |
|---|---|---|
| Unit A (auth) | (なし) | Cognito User Pool 内のみ |
| Unit B (budget) | GoroPay_Wallet / GoroPay_BudgetSettings 等 | PK として保持 |
| Unit C (order) | GoroPay_OrderHistory | PK として保持 |
| Unit D (suggest) | GoroPay_Suggestion | userId を属性として保持 |
| Unit E (metrics) | (Unit B/C 所有テーブルを読取参照) | — |

## トレーサビリティ

- NFR-COMP-02: 個人情報はメールアドレスのみ (Cognito 内のみ)
- A-NFR-SEC-06: アプリ側 DynamoDB には sub のみ、メールアドレスは保持しない
