import { cookies } from "next/headers";

/** Name of the httpOnly cookie the API sets on a successful OAuth callback. */
export const SESSION_COOKIE = "opspulse_session";

/**
 * Reads the session token from the request.
 *
 * The cookie is httpOnly, so it is only ever read here on the server and
 * forwarded to the API as a bearer token — browser script never sees it.
 *
 * OPSPULSE_API_TOKEN remains as a fallback so a server without a GitHub
 * OAuth app configured can still be driven from the environment.
 */
export async function sessionToken(): Promise<string> {
  const jar = await cookies();
  const fromCookie = jar.get(SESSION_COOKIE)?.value;
  if (fromCookie) return fromCookie;
  return process.env.OPSPULSE_API_TOKEN ?? "";
}

export function apiBaseUrl(): string {
  return process.env.OPSPULSE_API_URL ?? "http://localhost:8080";
}

/**
 * The API's public address, used for links the browser follows directly —
 * the GitHub sign-in redirect. In Docker the server-side base is an internal
 * hostname the browser cannot resolve, so this is configured separately.
 */
export function publicApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL ?? apiBaseUrl();
}
