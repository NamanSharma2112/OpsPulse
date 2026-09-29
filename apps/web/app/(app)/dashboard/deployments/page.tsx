import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { Notice } from "@/components/Notice";
import { activeProjectId, listDeployments } from "@/lib/api";
import { formatDuration, formatRelative, shortSha } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function DeploymentsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <Notice title="No data to show" body={project.error} />;

  const result = await listDeployments(project.data, 50);
  if (!result.ok) return <Notice title="No data to show" body={result.error} />;
  const { deployments } = result.data;

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Deployments</h1>
      <p className="ui-body">
        Projected from GitHub <code>deployment</code> and <code>deployment_status</code> events
      </p>
      </header>

      <div className="ui-card">
        {deployments.length === 0 ? (
          <Empty message="No deployments recorded yet." />
        ) : (
          deployments.map((d) => (
            <div className="ui-row" key={d.id}>
              <Badge variant={d.status} />
              <div className="ui-row-main">
                <div className="ui-row-title">
                  {d.url ? (
                    <a href={d.url} target="_blank" rel="noreferrer">
                      {d.environment} · {shortSha(d.commit_sha)}
                    </a>
                  ) : (
                    <>
                      {d.environment} · {shortSha(d.commit_sha)}
                    </>
                  )}
                </div>
                <div className="ui-row-meta">
                  {d.ref}
                  {d.actor ? ` · by ${d.actor}` : ""}
                  {d.finished_at ? ` · took ${formatDuration(d.started_at, d.finished_at)}` : ""}
                </div>
              </div>
              <div className="ui-row-side">{formatRelative(d.started_at)}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
