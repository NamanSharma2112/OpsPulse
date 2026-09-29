import { type NextRequest, NextResponse } from "next/server";
import { apiBaseUrl } from "@/lib/session";

/**
 * Serves the API under the dashboard's own origin at /v1/*.
 *
 * This is what makes the session cookie work on a stock deployment. The
 * cookie the OAuth callback sets is host-only, so if the browser talked to
 * the API on a second hostname the dashboard would never receive it — a
 * failure that only shows up in production, because locally both share
 * `localhost` and cookies ignore ports. Proxying keeps one origin, so no
 * shared parent domain or cookie-domain setting is needed.
 *
 * A route handler rather than a next.config rewrite: rewrite destinations are
 * resolved when the app is built, so an env-var destination bakes in whatever
 * was set at build time. This reads the address per request instead.
 */

export const dynamic = "force-dynamic";

// Headers that describe one hop and must not be forwarded to the next.
const HOP_BY_HOP = new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
  // Re-encoded by fetch, so the original values would be wrong.
  "content-encoding",
  "content-length",
  "host",
]);

async function proxy(request: NextRequest): Promise<Response> {
  const incoming = new URL(request.url);
  const target = `${apiBaseUrl()}${incoming.pathname}${incoming.search}`;

  const headers = new Headers();
  request.headers.forEach((value, key) => {
    if (!HOP_BY_HOP.has(key.toLowerCase())) headers.set(key, value);
  });
  // The API sees the origin the browser actually used.
  headers.set("X-Forwarded-Host", incoming.host);
  headers.set("X-Forwarded-Proto", incoming.protocol.replace(":", ""));

  const hasBody = request.method !== "GET" && request.method !== "HEAD";

  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: request.method,
      headers,
      body: hasBody ? await request.arrayBuffer() : undefined,
      // The OAuth flow is a chain of redirects the browser must follow
      // itself, so they are passed through rather than resolved here.
      redirect: "manual",
      cache: "no-store",
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown error";
    return NextResponse.json(
      { error: { message: `could not reach the API: ${message}` } },
      { status: 502 },
    );
  }

  const responseHeaders = new Headers();
  upstream.headers.forEach((value, key) => {
    if (!HOP_BY_HOP.has(key.toLowerCase())) responseHeaders.append(key, value);
  });

  // 304 and 204 carry no body, and giving them one is a protocol error.
  const body = upstream.status === 204 || upstream.status === 304 ? null : upstream.body;

  return new NextResponse(body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: responseHeaders,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
export const HEAD = proxy;
export const OPTIONS = proxy;
