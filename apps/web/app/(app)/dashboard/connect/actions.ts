"use server";

import { revalidatePath } from "next/cache";
import {
  connectRepository,
  createOrganization,
  createProject,
  listOrganizations,
  listProjects,
} from "@/lib/api";

export type ActionState = { error?: string; message?: string };

/**
 * Connects a GitHub repository, creating the organization and project it
 * needs along the way.
 *
 * The three steps are one action because a first-time user has none of them
 * and should not have to fill three forms to see a single event.
 */
export async function connectRepositoryAction(
  _prev: ActionState,
  form: FormData,
): Promise<ActionState> {
  const repo = String(form.get("repo") ?? "").trim();
  const defaultBranch = String(form.get("default_branch") ?? "main").trim();
  if (!repo.includes("/")) {
    return { error: 'Pick a repository, or type one as "owner/name".' };
  }

  // Reuse an existing organization rather than making a second one.
  const organizations = await listOrganizations();
  if (!organizations.ok) return { error: organizations.error };

  let organizationID = organizations.data.organizations[0]?.id;
  if (!organizationID) {
    const name = String(form.get("organization_name") ?? "").trim() || "My organization";
    const created = await createOrganization(name);
    if (!created.ok) return { error: created.error };
    organizationID = created.data.id;
  }

  const projects = await listProjects();
  if (!projects.ok) return { error: projects.error };

  let projectID = projects.data.projects[0]?.id;
  if (!projectID) {
    // Name the project after the repository, which is what it watches.
    const name = String(form.get("project_name") ?? "").trim() || repo.split("/")[1] || "Project";
    const created = await createProject(organizationID, name);
    if (!created.ok) return { error: created.error };
    projectID = created.data.id;
  }

  const result = await connectRepository(projectID, repo, defaultBranch || "main");
  if (!result.ok) return { error: result.error };

  revalidatePath("/dashboard", "layout");

  if (result.data.webhook_installed) {
    return { message: `Connected ${repo}. The webhook is installed and active.` };
  }
  return {
    message:
      `Connected ${repo}, but the webhook was not installed: ` +
      `${result.data.manual_reason ?? "unknown reason"}. Add it by hand at ` +
      `${result.data.webhook_url} with secret ${result.data.webhook_secret}`,
  };
}
