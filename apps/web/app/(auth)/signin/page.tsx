import Link from "next/link";
import { redirect } from "next/navigation";
import { authProviders, isSignedIn } from "@/lib/api";
import { SIGN_IN_PATH } from "@/lib/session";
import { PasswordForm } from "./PasswordForm";
import { Wordmark } from "@/components/Wordmark";

export const dynamic = "force-dynamic";

export const metadata = { title: "Sign in — OpsPulse" };

export default async function SignInPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  if (await isSignedIn()) redirect("/dashboard");

  const { error } = await searchParams;
  const providers = await authProviders();

  return (
    <div className="auth-card">
      <span className="auth-mark">
        {/* The shared mark, not a copy of it — a second definition is how the
            sign-in page quietly ends up a version behind. */}
        <Wordmark size={32} />
      </span>

      <h1 className="auth-title">Sign in to OpsPulse</h1>
      <p className="auth-sub">
        {providers.github
          ? "OpsPulse reads your repositories' activity. Signing in with GitHub also lets it install the webhook for you."
          : "Sign in to connect a repository and start collecting events."}
      </p>

      {error ? <p className="ui-notice ui-notice-error auth-alert">{error}</p> : null}

      {providers.github ? (
        <>
          <a className="ui-btn ui-btn-primary auth-btn" href={SIGN_IN_PATH}>
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
        </>
      ) : (
        <p className="auth-hint">
          Sign-in with GitHub is not configured on this server. Set{" "}
          <code>GITHUB_CLIENT_ID</code> and <code>GITHUB_CLIENT_SECRET</code> to
          enable it — until then, connecting a repository needs the webhook
          added by hand.
        </p>
      )}

      {/* The form is the only option when GitHub is unconfigured, and a
          secondary one when it is. */}
      <PasswordForm defaultOpen={!providers.github} />

      <Link href="/" className="auth-back">
        ← Back to opspulse
      </Link>
    </div>
  );
}
