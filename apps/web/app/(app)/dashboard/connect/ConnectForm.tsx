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
      {available.length > 0 ? (
        <label className="connect-field">
          <span>Repository</span>
          <select name="repo" defaultValue={available[0]?.full_name} disabled={pending}>
            {available.map((repo) => (
              <option key={repo.id} value={repo.full_name}>
                {repo.full_name}
                {repo.private ? " (private)" : ""}
              </option>
            ))}
          </select>
        </label>
      ) : (
        <label className="connect-field">
          <span>Repository</span>
          <input
            name="repo"
            placeholder="owner/name"
            autoComplete="off"
            spellCheck={false}
            disabled={pending}
            required
          />
        </label>
      )}

      <label className="connect-field">
        <span>Default branch</span>
        <input name="default_branch" defaultValue="main" disabled={pending} />
      </label>

      <button type="submit" className="connect-submit" disabled={pending}>
        {pending ? "Connecting…" : "Connect repository"}
      </button>

      {state.error ? <p className="connect-error">{state.error}</p> : null}
      {state.message ? <p className="connect-message">{state.message}</p> : null}
    </form>
  );
}
