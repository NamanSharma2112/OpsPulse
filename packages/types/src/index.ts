/**
 * Shared contracts for the OpsPulse API.
 *
 * These mirror the JSON the Go service emits (services/api/internal/domain).
 * Changing a field here without changing it there will not fail the build, so
 * keep the two in step.
 */

export interface User {
  id: string;
  email: string;
  name: string;
  github_id?: number;
  github_login?: string;
  avatar_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Session {
  token: string;
  expires_at: string;
  user: User;
}

export type OrgRole = "owner" | "admin" | "member";

export interface Org {
  id: string;
  name: string;
  slug: string;
  github_login?: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  org_id: string;
  name: string;
  slug: string;
  repo_owner: string;
  repo_name: string;
  default_branch: string;
  created_at: string;
  updated_at: string;
}

export type DeploymentStatus =
  | "pending"
  | "running"
  | "success"
  | "failure"
  | "inactive";

export interface Deployment {
  id: string;
  project_id: string;
  external_id: string;
  environment: string;
  ref: string;
  sha: string;
  status: DeploymentStatus;
  actor?: string;
  url?: string;
  started_at: string;
  finished_at?: string;
}

export type PullRequestState = "open" | "closed" | "merged";

export interface PullRequest {
  id: string;
  project_id: string;
  number: number;
  title: string;
  author: string;
  state: PullRequestState;
  draft: boolean;
  url?: string;
  opened_at: string;
  merged_at?: string;
  closed_at?: string;
  updated_at: string;
}

export type IncidentSeverity = "critical" | "major" | "minor";
export type IncidentStatus = "open" | "acknowledged" | "resolved";

export interface Incident {
  id: string;
  project_id: string;
  external_id: string;
  title: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  source: string;
  url?: string;
  opened_at: string;
  resolved_at?: string;
}

export interface OpsEvent {
  id: string;
  project_id: string;
  delivery_id: string;
  type: string;
  action?: string;
  actor?: string;
  occurred_at: string;
  received_at: string;
}

export type HealthStatus = "healthy" | "degraded" | "critical";

export interface ProjectHealth {
  project_id: string;
  window: string;
  status: HealthStatus;
  open_incidents: number;
  open_pull_requests: number;
  deployments: {
    succeeded: number;
    failed: number;
    success_rate: number;
  };
  last_deployment?: Deployment;
}

/** Every endpoint returns this shape on error. */
export interface ApiError {
  error: { message: string };
}

// Collection envelopes returned by the list endpoints.
export interface ProjectList { projects: Project[] }
export interface OrgList { orgs: Org[] }
export interface DeploymentList { deployments: Deployment[] }
export interface PullRequestList { pull_requests: PullRequest[] }
export interface IncidentList { incidents: Incident[] }
export interface EventList { events: OpsEvent[] }
