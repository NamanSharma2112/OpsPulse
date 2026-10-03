"use client";

import { useActionState, useState } from "react";
import { passwordSignInAction, type SignInState } from "./actions";

/**
 * Email and password sign-in. Offered alongside GitHub, and the only option
 * on a server with no OAuth app configured — which is the usual case when
 * running locally.
 */
export function PasswordForm({ defaultOpen }: { defaultOpen: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  const [mode, setMode] = useState<"login" | "register">("login");
  const [state, action, pending] = useActionState<SignInState, FormData>(
    passwordSignInAction,
    {},
  );

  if (!open) {
    return (
      <button type="button" className="auth-toggle" onClick={() => setOpen(true)}>
        Use an email address instead
      </button>
    );
  }

  return (
    <form action={action} className="auth-form">
      <input type="hidden" name="mode" value={mode} />

      {mode === "register" ? (
        <label className="ui-field">
          <span>Name</span>
          <input className="ui-input" name="name" autoComplete="name" disabled={pending} />
        </label>
      ) : null}

      <label className="ui-field">
        <span>Email</span>
        <input
          className="ui-input"
          name="email"
          type="email"
          autoComplete="email"
          required
          disabled={pending}
        />
      </label>

      <label className="ui-field">
        <span>Password</span>
        <input
          className="ui-input"
          name="password"
          type="password"
          autoComplete={mode === "register" ? "new-password" : "current-password"}
          required
          disabled={pending}
        />
      </label>

      <button type="submit" className="ui-btn ui-btn-primary auth-btn" disabled={pending}>
        {pending ? "Signing in…" : mode === "register" ? "Create account" : "Sign in"}
      </button>

      {state.error ? <p className="ui-notice ui-notice-error">{state.error}</p> : null}

      <button
        type="button"
        className="auth-toggle"
        onClick={() => setMode(mode === "login" ? "register" : "login")}
        disabled={pending}
      >
        {mode === "login" ? "Create an account instead" : "I already have an account"}
      </button>
    </form>
  );
}
