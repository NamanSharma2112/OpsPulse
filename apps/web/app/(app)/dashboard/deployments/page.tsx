import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { ApiUnavailable } from "@/components/Notice";
import { activeProjectId, listDeployments } from "@/lib/api";
import { formatDuration, formatRelative, shortSha } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function DeploymentsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <ApiUnavailable error={project.error} />;

  const result = await listDeployments(project.data, 50);
  if (!result.ok) return <ApiUnavailable error={result.error} />;
  const { deployments } = result.data;

  return (
    <>
      <h1 className="page-title">Deployments</h1>
      <p className="page-subtitle">
        Projected from GitHub <code>deployment</code> and <code>deployment_status</code> events
      </p>

      <div className="card">
        {deployments.length === 0 ? (
          <Empty message="No deployments recorded yet." />
        ) : (
          deployments.map((d) => (
            <div className="row" key={d.id}>
              <Badge variant={d.status} />
              <div className="row-main">
                <div className="row-title">
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
                <div className="row-meta">
                  {d.ref}
                  {d.actor ? ` · by ${d.actor}` : ""}
                  {d.finished_at ? ` · took ${formatDuration(d.started_at, d.finished_at)}` : ""}
                </div>
              </div>
              <div className="row-side">{formatRelative(d.started_at)}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
