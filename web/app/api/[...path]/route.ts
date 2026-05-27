// BFF catch-all proxy Route Handler (NFR Design LC-AUTH-18)
//
// ブラウザ → /api/* (同一オリジン) → 本 Route Handler → 上流 API Gateway
// `Authorization: Bearer <accessToken>` ヘッダはそのまま透過する (Server で変換しない)。
// `API_ENDPOINT` は server-only env (`NEXT_PUBLIC_` プレフィックスなし)。

import { NextRequest } from "next/server";

function mockResponse(pathSegments: string[], method: string): Response | null {
  const route = pathSegments.join("/");
  const now = new Date().toISOString();

  if (route === "wallet" && method === "GET") {
    return Response.json({ balance: 25000, monthlyBudget: 30000, updatedAt: now });
  }
  if (route === "wallet/budget" && method === "POST") {
    return Response.json({ monthlyBudget: 30000, appliedFrom: "2026-06-01" });
  }
  if (route === "metrics" && method === "GET") {
    return Response.json({ damageCount: 3, consumptionRate: 17, monthlyBudget: 30000, remainingBalance: 25000, thresholdExceeded: false, summaryText: "今月はまだいける。" });
  }
  if (route === "orders" && method === "GET") {
    return Response.json({
      items: [
        { orderId: "mock-01", storeName: "松屋", menuName: "牛めし", amount: 480, createdAt: now },
        { orderId: "mock-02", storeName: "吉野家", menuName: "牛丼", amount: 426, createdAt: now },
      ],
    });
  }
  if (route === "orders" && method === "POST") {
    return Response.json({ orderId: "mock-new", storeName: "すき家", menuName: "牛丼", amount: 430, remainingBalance: 24570, idempotent: false });
  }
  if (route === "suggest" && method === "GET") {
    return Response.json({ hasSuggestion: true, suggestionId: "mock-s1", title: "今日のおすすめ", plan: { storeName: "松屋", menuName: "牛めし", amount: 480, category: "food" } });
  }
  if (route === "budget/raise/recommendation" && method === "GET") {
    return Response.json({ currentMonthlyBudget: 30000, recommendedMonthlyBudget: 35000 });
  }
  if (route === "budget/raise" && method === "POST") {
    return Response.json({ newMonthlyBudget: 35000, appliedFrom: "2026-06-01" });
  }
  if (route === "auth/logout" && method === "POST") {
    return new Response(null, { status: 204 });
  }
  return null;
}

// Next.js 15 の breaking change により、Route Handler の `context.params` は
// Promise<{...}> 型になった。同期アクセスは型エラー & ランタイム警告になる。
// 参考: https://nextjs.org/docs/app/api-reference/file-conventions/route#context-optional
type RouteCtx = { params: Promise<{ path: string[] }> };

export async function GET(request: NextRequest, ctx: RouteCtx) {
  const { path } = await ctx.params;
  return proxy(request, path);
}
export async function POST(request: NextRequest, ctx: RouteCtx) {
  const { path } = await ctx.params;
  return proxy(request, path);
}
export async function PUT(request: NextRequest, ctx: RouteCtx) {
  const { path } = await ctx.params;
  return proxy(request, path);
}
export async function DELETE(request: NextRequest, ctx: RouteCtx) {
  const { path } = await ctx.params;
  return proxy(request, path);
}
export async function PATCH(request: NextRequest, ctx: RouteCtx) {
  const { path } = await ctx.params;
  return proxy(request, path);
}

async function proxy(request: NextRequest, pathSegments: string[]): Promise<Response> {
  const apiEndpoint = process.env.API_ENDPOINT;
  if (!apiEndpoint) {
    return new Response("API_ENDPOINT not configured", { status: 500 });
  }
  if (apiEndpoint === "mock") {
    const mock = mockResponse(pathSegments, request.method);
    if (mock) return mock;
    return new Response("Mock not found", { status: 404 });
  }

  // CSRF: 本 BFF は Authorization: Bearer 必須なので、cookie 自動送信に
  // 依存する古典的 CSRF 攻撃には脆弱でない (Bearer は cookie ではなく
  // header 必須で、ブラウザが他 origin から自動付与しない)。
  // 将来 cookie ベース認証へ切り替える場合は、ここで Origin / Referer の
  // ホワイトリスト検証を追加する (例: Amplify ドメインのみ許容)。
  const authHeader = request.headers.get("Authorization");
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    return new Response("Missing Authorization header", { status: 401 });
  }

  // 上流に渡すヘッダは Authorization 必須、Content-Type は body がある場合のみ。
  // body 無しに application/json をデフォで付けると fetch 仕様 (RFC 7231) と
  // 整合せず一部の上流が拒否する場合があるため省略する。
  const upstreamHeaders: HeadersInit = { Authorization: authHeader };
  const contentType = request.headers.get("Content-Type");
  if (contentType) {
    upstreamHeaders["Content-Type"] = contentType;
  }

  const upstreamUrl = `${apiEndpoint}/api/${pathSegments.join("/")}${request.nextUrl.search}`;
  const upstream = await fetch(upstreamUrl, {
    method: request.method,
    headers: upstreamHeaders,
    body: ["GET", "HEAD"].includes(request.method) ? undefined : await request.text(),
  });

  // ホップバイホップ系のヘッダを除去してからクライアントへ転送する。
  // - Content-Encoding: undici fetch は gzip/br を自動 decode するため、
  //   元の encoding ヘッダを残すとブラウザが二重 decode しようとして
  //   ERR_CONTENT_DECODING_FAILED で失敗する。
  // - Content-Length / Transfer-Encoding: stream を再パイプするため
  //   再計算が必要。Next.js / Node が再付与する。
  // - Connection: HTTP/1.1 hop-by-hop ヘッダ。proxy 越しに残してはいけない。
  const responseHeaders = new Headers(upstream.headers);
  for (const h of ["content-encoding", "content-length", "transfer-encoding", "connection"]) {
    responseHeaders.delete(h);
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: responseHeaders,
  });
}
