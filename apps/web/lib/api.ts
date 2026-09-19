import type {
  DeploymentList,
  IncidentList,
  ProjectHealth,
  ProjectList,
  PullRequestList,
} from "@opspulse/types";

const API_URL = process.env.OPSPULSE_API_URL ?? "http://localhost:8080";
const API_TOKEN = process.env.OPSPULSE_API_TOKEN ?? "";

/**
 * Result of an API read. The dashboard renders partial data rather than
 * failing outright, so every fetch returns an error instead of throwing.
 */
export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string };

/** True when the dashboard has been given a token to read the API with. */
export function isConfigured(): boolean {
  return API_TOKEN !== "";
}

export function apiUrl(): string {
  return API_URL;
}

async function get<T>(path: string): Promise<ApiResult<T>> {
  if (!API_TOKEN) {
    return { ok: false, error: "OPSPULSE_API_TOKEN is not set" };
  }
  try {
    const res = await fetch(`${API_URL}${path}`, {
      headers: { Authorization: `Bearer ${API_TOKEN}` },
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
    return { ok: true, data: (await res.json()) as T };
  } catch (err) {
    // Most often the API container is not up yet.
    const message = err instanceof Error ? err.message : "unknown error";
    return { ok: false, error: `could not reach the API at ${API_URL}: ${message}` };
  }
}

export function listProjects(): Promise<ApiResult<ProjectList>> {
  return get<ProjectList>("/v1/projects");
}

/**
 * Resolves the project the dashboard should display: the one pinned by
 * OPSPULSE_PROJECT_ID, otherwise the first the token can see.
 */
export async function activeProjectId(): Promise<ApiResult<string>> {
  const pinned = process.env.OPSPULSE_PROJECT_ID;
  if (pinned) return { ok: true, data: pinned };

  const projects = await listProjects();
  if (!projects.ok) return projects;
  const first = projects.data.projects[0];
  if (!first) {
    return { ok: false, error: "no projects yet — create one to start collecting events" };
  }
  return { ok: true, data: first.id };
}

export function getHealth(projectID: string): Promise<ApiResult<ProjectHealth>> {
  return get<ProjectHealth>(`/v1/projects/${projectID}/health`);
}

export function listDeployments(projectID: string, limit = 25): Promise<ApiResult<DeploymentList>> {
  return get<DeploymentList>(`/v1/projects/${projectID}/deployments?limit=${limit}`);
}

export function listPullRequests(
  projectID: string,
  state = "",
  limit = 25,
): Promise<ApiResult<PullRequestList>> {
  const query = new URLSearchParams({ limit: String(limit) });
  if (state) query.set("state", state);
  return get<PullRequestList>(`/v1/projects/${projectID}/pull-requests?${query}`);
}

export function listIncidents(
  projectID: string,
  status = "",
  limit = 25,
): Promise<ApiResult<IncidentList>> {
  const query = new URLSearchParams({ limit: String(limit) });
  if (status) query.set("status", status);
  return get<IncidentList>(`/v1/projects/${projectID}/incidents?${query}`);
}
