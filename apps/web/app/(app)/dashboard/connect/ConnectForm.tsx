"use client";

import { useActionState } from "react";
import type { GitHubRepo } from "@/lib/api";
import { connectRepositoryAction, type ActionState } from "./actions";

/**
 * Picks a repository and connects it. When the session can list repositories
 * from GitHub the picker is a select; otherwise it falls back to typing
 * "owner/name", which is what a password-only account has to do.
 */
export function ConnectForm({
  repositories,
  connected,
}: {
  repositories: GitHubRepo[];
  connected: string[];
}) {
  const [state, action, pending] = useActionState<ActionState, FormData>(
    connectRepositoryAction,
    {},
  );

  const available = repositories.filter((r) => !connected.includes(r.full_name.toLowerCase()));

  return (
    <form action={action} className="connect-form">
      <label className="ui-field">
        <span>Repository</span>
        {available.length > 0 ? (
          <select
            className="ui-select"
            name="repo"
            defaultValue={available[0]?.full_name}
            disabled={pending}
          >
            {available.map((repo) => (
              <option key={repo.id} value={repo.full_name}>
                {repo.full_name}
                {repo.private ? " (private)" : ""}
              </option>
            ))}
          </select>
        ) : (
          <input
            className="ui-input"
            name="repo"
            placeholder="owner/name"
            autoComplete="off"
            spellCheck={false}
            disabled={pending}
            required
          />
        )}
      </label>

      <label className="ui-field">
        <span>Default branch</span>
        <input className="ui-input" name="default_branch" defaultValue="main" disabled={pending} />
      </label>

      <div className="connect-actions">
        <button type="submit" className="ui-btn ui-btn-primary" disabled={pending}>
          {pending ? "Connecting…" : "Connect repository"}
        </button>
        {pending ? <span className="ui-caption">Installing the webhook…</span> : null}
      </div>

      {state.error ? <p className="ui-notice ui-notice-error">{state.error}</p> : null}
      {state.message ? (
        <p className="ui-notice ui-notice-ok">
          {state.message}
          {state.secret ? <code className="connect-secret">{state.secret}</code> : null}
        </p>
      ) : null}
    </form>
  );
}
