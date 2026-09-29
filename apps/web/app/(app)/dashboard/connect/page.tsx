import { Badge } from "@/components/Badge";
import { listGitHubRepositories, listProjects, listRepositories } from "@/lib/api";
import { ConnectForm } from "./ConnectForm";

export const dynamic = "force-dynamic";

export default async function ConnectPage() {
  const github = await listGitHubRepositories();

  // Repositories already connected, so the picker does not offer them twice.
  const projects = await listProjects();
  const connected: string[] = [];
  if (projects.ok) {
    for (const project of projects.data.projects) {
      const repos = await listRepositories(project.id);
      if (repos.ok) {
        connected.push(...repos.data.repositories.map((r) => r.external_id.toLowerCase()));
      }
    }
  }

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Connect a repository</h1>
        <p className="ui-body">
          OpsPulse installs the webhook for you and starts recording events as
          soon as GitHub delivers the first one.
        </p>
      </header>

      <div className="ui-card ui-card-pad">
        {github.ok ? (
          <ConnectForm repositories={github.data.repositories} connected={connected} />
        ) : (
          <>
            <p className="ui-body-sm" style={{ marginBottom: 16 }}>
              {github.error}. Enter a repository by hand — you will need to add
              the webhook yourself.
            </p>
            <ConnectForm repositories={[]} connected={connected} />
          </>
        )}
      </div>

      {connected.length > 0 ? (
        <>
          <h2 className="dash-section-title">Already connected</h2>
          <div className="ui-card">
            {connected.map((name) => (
              <div className="ui-row" key={name}>
                <Badge variant="success" label="webhook active" />
                <div className="ui-row-main">
                  <div className="ui-row-title">{name}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : null}
    </>
  );
}
