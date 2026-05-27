/**
 * Pre Sign-up Lambda Trigger (Q-D7=A Node.js)
 *
 * 役割: Cognito User Pool に対して auto-confirm + auto-verify-email を有効化する。
 *       メール検証コードフロー (Q-A1=B) を回避し、登録後即座に CONFIRMED 状態にする。
 *
 * 失敗時の挙動 (A-NFR-REL-01):
 *   - Lambda runtime に例外を伝播 → Cognito が SignUp 全体を失敗扱いにする
 *   - DLQ や追加ログは設置しない (本 MVP 範囲)
 */
exports.handler = async (event) => {
  event.response.autoConfirmUser = true;
  event.response.autoVerifyEmail = true;
  return event;
};
