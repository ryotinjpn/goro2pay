// Amplify Auth v6 の設定オブジェクト。
// NEXT_PUBLIC_* env から Cognito の値を読み出す。
// Browser に露出してよい (Public な値)。
//
// API_ENDPOINT は server-only env なのでここでは使わない (BFF パターン)。

import type { ResourcesConfig } from "aws-amplify";

export const amplifyConfig: ResourcesConfig = {
  Auth: {
    Cognito: {
      userPoolId: process.env.NEXT_PUBLIC_USER_POOL_ID ?? "",
      userPoolClientId: process.env.NEXT_PUBLIC_USER_POOL_CLIENT_ID ?? "",
      // Region は Amplify が userPoolId からも判定するが明示しておく
      // signUpVerificationMethod は auto-confirm のため不要
    },
  },
};
