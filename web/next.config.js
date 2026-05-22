/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // BFF パターン: API_ENDPOINT は server-only env として扱う (NEXT_PUBLIC_ なし)
  // 公開する env は NEXT_PUBLIC_ プレフィックス付きのみ (USER_POOL_ID 等)
};

module.exports = nextConfig;
