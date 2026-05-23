# Auth Unit — Database Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation Step 12 (スキップ)

## 結論: Unit A は DynamoDB テーブルを所有しない

unit-of-work.md §3.1 / unit-interfaces.md §2 の通り、Unit A (auth) は **Amazon Cognito User Pool** のみをデータストアとして使用する。アプリ側 DynamoDB テーブルは持たない。

## 永続化対象

| データ | 保存先 |
|---|---|
| ユーザの credential | Cognito User Pool (内部に管理、アプリは参照しない) |
| ユーザの email | Cognito User Pool (アプリ DB には保存しない、A-NFR-SEC-06) |
| ユーザの sub (UUID) | Cognito User Pool (各 Unit が外部キーとして参照) |
| Session トークン | Browser localStorage (Amplify Auth 標準、A-NFR-SEC-03) |

## マイグレーションスクリプト

**該当なし**。Cognito User Pool は Terraform で管理 (Infrastructure Design `cognito.tf`)、初回 apply で User Pool が作成される。
