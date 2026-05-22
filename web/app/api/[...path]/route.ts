// BFF catch-all proxy Route Handler (NFR Design LC-AUTH-18)
//
// ブラウザ → /api/* (同一オリジン) → 本 Route Handler → 上流 API Gateway
// `Authorization: Bearer <accessToken>` ヘッダはそのまま透過する (Server で変換しない)。
// `API_ENDPOINT` は server-only env (`NEXT_PUBLIC_` プレフィックスなし)。

import { NextRequest } from "next/server";

export async function GET(request: NextRequest, ctx: { params: { path: string[] } }) {
  return proxy(request, ctx.params.path);
}
export async function POST(request: NextRequest, ctx: { params: { path: string[] } }) {
  return proxy(request, ctx.params.path);
}
export async function PUT(request: NextRequest, ctx: { params: { path: string[] } }) {
  return proxy(request, ctx.params.path);
}
export async function DELETE(request: NextRequest, ctx: { params: { path: string[] } }) {
  return proxy(request, ctx.params.path);
}
export async function PATCH(request: NextRequest, ctx: { params: { path: string[] } }) {
  return proxy(request, ctx.params.path);
}

async function proxy(request: NextRequest, pathSegments: string[]): Promise<Response> {
  const apiEndpoint = process.env.API_ENDPOINT;
  if (!apiEndpoint) {
    return new Response("API_ENDPOINT not configured", { status: 500 });
  }

  const authHeader = request.headers.get("Authorization");
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    return new Response("Missing Authorization header", { status: 401 });
  }

  const upstreamUrl = `${apiEndpoint}/api/${pathSegments.join("/")}${request.nextUrl.search}`;
  const upstream = await fetch(upstreamUrl, {
    method: request.method,
    headers: {
      Authorization: authHeader,
      "Content-Type": request.headers.get("Content-Type") ?? "application/json",
    },
    body: ["GET", "HEAD"].includes(request.method) ? undefined : await request.text(),
  });

  // 上流レスポンスをそのまま透過 (401 を含む全 status)
  return new Response(upstream.body, {
    status: upstream.status,
    headers: upstream.headers,
  });
}
