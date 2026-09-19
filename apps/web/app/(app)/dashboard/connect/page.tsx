import { Empty } from "@/components/Empty";
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
      <h1 className="page-title">Connect a repository</h1>
      <p className="page-subtitle">
        OpsPulse installs the webhook for you and starts recording events as
        soon as GitHub delivers the first one.
      </p>

      <div className="card" style={{ padding: 20 }}>
        {github.ok ? (
          <ConnectForm repositories={github.data.repositories} connected={connected} />
        ) : (
          <>
            <p className="row-meta" style={{ marginBottom: 16 }}>
              {github.error}. Enter a repository by hand — you will need to add
              the webhook yourself.
            </p>
            <ConnectForm repositories={[]} connected={connected} />
          </>
        )}
      </div>

      {connected.length > 0 ? (
        <>
          <h2 className="page-title" style={{ fontSize: 17 }}>
            Already connected
          </h2>
          <div className="card">
            {connected.map((name) => (
              <div className="row" key={name}>
                <div className="row-main">
                  <div className="row-title">{name}</div>
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <Empty message="" />
      )}
    </>
  );
}
