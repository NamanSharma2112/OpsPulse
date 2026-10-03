"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { passwordSignIn } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

export type SignInState = { error?: string };

/**
 * Signs in with email and password.
 *
 * The API returns a token; the cookie is set here rather than forwarded from
 * the API's response, because a server action's own fetch does not pass
 * Set-Cookie on to the browser. Same cookie name and attributes as the OAuth
 * callback sets, so both paths produce an identical session.
 */
export async function passwordSignInAction(
  _prev: SignInState,
  form: FormData,
): Promise<SignInState> {
  const mode = form.get("mode") === "register" ? "register" : "login";
  const email = String(form.get("email") ?? "").trim();
  const password = String(form.get("password") ?? "");
  const name = String(form.get("name") ?? "").trim();

  if (!email || !password) {
    return { error: "Enter an email address and a password." };
  }
  if (mode === "register" && password.length < 8) {
    return { error: "Use a password of at least 8 characters." };
  }

  const body: Record<string, string> =
    mode === "register" ? { email, password, name } : { email, password };

  const result = await passwordSignIn(mode, body);
  if (!result.ok) {
    return {
      error:
        result.error === "invalid credentials"
          ? "That email and password did not match an account."
          : result.error,
    };
  }

  const jar = await cookies();
  jar.set(SESSION_COOKIE, result.data.token, {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.SECURE_COOKIES === "true",
    path: "/",
    expires: new Date(result.data.expires_at),
  });

  // Outside the try/catch above: redirect works by throwing.
  redirect("/dashboard");
}
