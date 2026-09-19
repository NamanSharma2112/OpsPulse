import type {
  DeploymentList,
  EventList,
  IncidentList,
  Organization,
  OrganizationList,
  Project,
  ProjectHealth,
  ProjectList,
  PullRequestList,
  RepositoryList,
  User,
} from "@opspulse/types";
import { apiBaseUrl, sessionToken } from "./session";

/**
 * Result of an API call. The dashboard renders partial data rather than
 * failing outright, so these return an error instead of throwing.
 */
export type ApiResult<T> = { ok: true; data: T } | { ok: false; error: string };

/** A GitHub repository the signed-in user can administer. */
export interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  private: boolean;
  default_branch: string;
  html_url: string;
}

async function request<T>(
  path: string,
  init?: { method?: string; body?: unknown },
): Promise<ApiResult<T>> {
  const token = await sessionToken();
  if (!token) return { ok: false, error: "not signed in" };

  try {
    const res = await fetch(`${apiBaseUrl()}${path}`, {
      method: init?.method ?? "GET",
      headers: {
        Authorization: `Bearer ${token}`,
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
      },
      body: init?.body ? JSON.stringify(init.body) : undefined,
      // Operational data goes stale fast; always read through to the API.
      cache: "no-store",
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      const message =
        body && typeof body === "object" && "error" in body
          ? (body as { error: { message: string } }).error.message
          : `request failed with status ${res.status}`;
      return { ok: false, error: message };
    }
    if (res.status === 204) return { ok: true, data: undefined as T };
    return { ok: true, data: (await res.json()) as T };
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown error";
    return { ok: false, error: `could not reach the API: ${message}` };
  }
}

/** True when the request carries a session. */
export async function isSignedIn(): Promise<boolean> {
  return (await sessionToken()) !== "";
}

export function currentUser(): Promise<ApiResult<User>> {
  return request<User>("/v1/auth/me");
}

// Reads ----------------------------------------------------------------------
export function listOrganizations(): Promise<ApiResult<OrganizationList>> {
  return request<OrganizationList>("/v1/organizations");
}

export function listProjects(): Promise<ApiResult<ProjectList>> {
  return request<ProjectList>("/v1/projects");
}

export function listRepositories(projectID: string): Promise<ApiResult<RepositoryList>> {
  return request<RepositoryList>(`/v1/projects/${projectID}/repositories`);
}

/** Repositories on GitHub the signed-in user can administer. */
export function listGitHubRepositories(): Promise<ApiResult<{ repositories: GitHubRepo[] }>> {
  return request<{ repositories: GitHubRepo[] }>("/v1/github/repositories");
}

/** The raw event feed — the source of truth every projection is built from. */
export function listEvents(projectID: string, limit = 50): Promise<ApiResult<EventList>> {
  return request<EventList>(`/v1/projects/${projectID}/events?limit=${limit}`);
}

export function getHealth(projectID: string): Promise<ApiResult<ProjectHealth>> {
  return request<ProjectHealth>(`/v1/projects/${projectID}/health`);
}

export function listDeployments(projectID: string, limit = 25): Promise<ApiResult<DeploymentList>> {
  return request<DeploymentList>(`/v1/projects/${projectID}/deployments?limit=${limit}`);
}

export function listPullRequests(
  projectID: string,
  state = "",
  limit = 25,
): Promise<ApiResult<PullRequestList>> {
  const query = new URLSearchParams({ limit: String(limit) });
  if (state) query.set("state", state);
  return request<PullRequestList>(`/v1/projects/${projectID}/pull-requests?${query}`);
}

export function listIncidents(
  projectID: string,
  status = "",
  limit = 25,
): Promise<ApiResult<IncidentList>> {
  const query = new URLSearchParams({ limit: String(limit) });
  if (status) query.set("status", status);
  return request<IncidentList>(`/v1/projects/${projectID}/incidents?${query}`);
}

// Writes ---------------------------------------------------------------------
export function createOrganization(name: string): Promise<ApiResult<Organization>> {
  return request<Organization>("/v1/organizations", { method: "POST", body: { name } });
}

export function createProject(
  organizationID: string,
  name: string,
): Promise<ApiResult<Project>> {
  return request<Project>("/v1/projects", {
    method: "POST",
    body: { organization_id: organizationID, name },
  });
}

/** What the API reports after connecting a repository. */
export interface ConnectResult {
  repository: { id: string; external_id: string; name: string };
  webhook_secret: string;
  webhook_url: string;
  webhook_installed: boolean;
  manual_reason?: string;
}

export function connectRepository(
  projectID: string,
  repo: string,
  defaultBranch: string,
): Promise<ApiResult<ConnectResult>> {
  return request<ConnectResult>(`/v1/projects/${projectID}/repositories`, {
    method: "POST",
    body: { repo, default_branch: defaultBranch },
  });
}

/**
 * Resolves the project the dashboard should display: the one pinned by
 * OPSPULSE_PROJECT_ID, otherwise the first the session can see.
 */
export async function activeProjectId(): Promise<ApiResult<string>> {
  const pinned = process.env.OPSPULSE_PROJECT_ID;
  if (pinned) return { ok: true, data: pinned };

  const projects = await listProjects();
  if (!projects.ok) return projects;
  const first = projects.data.projects[0];
  if (!first) {
    return { ok: false, error: "no projects yet" };
  }
  return { ok: true, data: first.id };
}
