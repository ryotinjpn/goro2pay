/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // BFF パターン: API_ENDPOINT は server-only env として扱う (NEXT_PUBLIC_ なし)
  // 公開する env は NEXT_PUBLIC_ プレフィックス付きのみ (USER_POOL_ID 等)
  //
  // NOTE: cold start を縮めたい場合は output: 'standalone' を検討する。
  // Amplify Hosting (WEB_COMPUTE) は公式テンプレートで動くため本 MVP では
  // 不要だが、Lambda コンテナ起動時間が問題になる場合のオプション。
};

module.exports = nextConfig;
