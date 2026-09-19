import Link from "next/link";
import { redirect } from "next/navigation";
import { isSignedIn } from "@/lib/api";
import { publicApiBaseUrl } from "@/lib/session";

export const dynamic = "force-dynamic";

export const metadata = { title: "Sign in — OpsPulse" };

export default async function SignInPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  if (await isSignedIn()) redirect("/dashboard");

  const { error } = await searchParams;
  // The browser follows this link itself, so it must be the API's public
  // address rather than an internal container hostname.
  const signInHref = `${publicApiBaseUrl()}/v1/auth/github`;

  return (
    <div className="auth-card">
      <span className="auth-mark">
        <svg width="32" height="32" viewBox="0 0 18 18" aria-hidden="true">
          <rect width="18" height="18" rx="5" fill="#111111" />
          <path
            d="M3.6 9.4h2.6l1.5-3.6 2 7 1.6-3.4h3.1"
            fill="none"
            stroke="#ffffff"
            strokeWidth="1.4"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </span>

      <h1 className="auth-title">Sign in to OpsPulse</h1>
      <p className="auth-sub">
        OpsPulse reads your repositories&apos; activity. Signing in with GitHub
        also lets it install the webhook for you.
      </p>

      {error ? <p className="auth-error">{error}</p> : null}

      <a className="auth-btn" href={signInHref}>
        <svg width="18" height="18" viewBox="0 0 16 16" aria-hidden="true" fill="currentColor">
          <path d="M8 0a8 8 0 0 0-2.5 15.6c.4.1.5-.2.5-.4v-1.4c-2.2.5-2.7-1-2.7-1-.4-.9-.9-1.2-.9-1.2-.7-.5.1-.5.1-.5.8.1 1.2.8 1.2.8.7 1.2 1.9.9 2.4.7 0-.6.3-.9.5-1.1-1.8-.2-3.6-.9-3.6-3.9 0-.9.3-1.6.8-2.1 0-.2-.4-1 .1-2.1 0 0 .7-.2 2.2.8a7.5 7.5 0 0 1 4 0c1.5-1 2.2-.8 2.2-.8.5 1.1.1 1.9.1 2.1.5.5.8 1.2.8 2.1 0 3-1.8 3.7-3.6 3.9.3.3.6.8.6 1.5v2.2c0 .2.1.5.6.4A8 8 0 0 0 8 0Z" />
        </svg>
        Continue with GitHub
      </a>

      <div className="auth-scopes">
        <h2>OpsPulse will ask GitHub for</h2>
        <ul>
          <li>
            <code>read:user</code>
            <span>Your name and avatar, to identify your account.</span>
          </li>
          <li>
            <code>repo</code>
            <span>Read activity from the repositories you choose.</span>
          </li>
          <li>
            <code>admin:repo_hook</code>
            <span>Install the webhook, so you do not have to.</span>
          </li>
        </ul>
      </div>

      <p className="auth-note">
        Your access token is encrypted before it is stored, and is only used
        for the repositories you connect.
      </p>

      <Link href="/" className="auth-back">
        ← Back to opspulse
      </Link>
    </div>
  );
}
