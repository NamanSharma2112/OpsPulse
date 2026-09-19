import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { ApiUnavailable } from "@/components/Notice";
import { activeProjectId, getHealth, listDeployments, listIncidents } from "@/lib/api";
import { formatPercent, formatRelative, shortSha } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function HealthPage() {
  const project = await activeProjectId();
  if (!project.ok) return <ApiUnavailable error={project.error} />;

  const [health, deployments, incidents] = await Promise.all([
    getHealth(project.data),
    listDeployments(project.data, 5),
    listIncidents(project.data, "open", 5),
  ]);

  if (!health.ok) return <ApiUnavailable error={health.error} />;
  const h = health.data;

  return (
    <>
      <h1 className="page-title">Health</h1>
      <p className="page-subtitle">
        Rolling {h.window} window · {h.deployments.succeeded + h.deployments.failed} deployments
      </p>

      <div className="tiles">
        <div className="tile">
          <div className="tile-label">Status</div>
          <div className="tile-value">
            <Badge variant={h.status} />
          </div>
          <div className="tile-note">
            {h.status === "healthy"
              ? "No open incidents or failed deploys"
              : `${h.open_incidents} open incident${h.open_incidents === 1 ? "" : "s"}`}
          </div>
        </div>

        <div className="tile">
          <div className="tile-label">Deploy success</div>
          <div className="tile-value">{formatPercent(h.deployments.success_rate)}</div>
          <div className="tile-note">
            {h.deployments.succeeded} succeeded · {h.deployments.failed} failed
          </div>
        </div>

        <div className="tile">
          <div className="tile-label">Open pull requests</div>
          <div className="tile-value">{h.open_pull_requests}</div>
          <div className="tile-note">Awaiting review or merge</div>
        </div>

        <div className="tile">
          <div className="tile-label">Last deploy</div>
          <div className="tile-value" style={{ fontSize: 18 }}>
            {h.last_deployment ? formatRelative(h.last_deployment.started_at) : "—"}
          </div>
          <div className="tile-note">
            {h.last_deployment
              ? `${h.last_deployment.environment} · ${shortSha(h.last_deployment.sha)}`
              : "No deployments recorded"}
          </div>
        </div>
      </div>

      <h2 className="page-title" style={{ fontSize: 17 }}>
        Open incidents
      </h2>
      <div className="card">
        {!incidents.ok ? (
          <Empty message={incidents.error} />
        ) : incidents.data.incidents.length === 0 ? (
          <Empty message="Nothing on fire." />
        ) : (
          incidents.data.incidents.map((incident) => (
            <div className="row" key={incident.id}>
              <Badge variant={incident.severity} />
              <div className="row-main">
                <div className="row-title">{incident.title}</div>
                <div className="row-meta">
                  {incident.source} · opened {formatRelative(incident.opened_at)}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      <h2 className="page-title" style={{ fontSize: 17 }}>
        Recent deployments
      </h2>
      <div className="card">
        {!deployments.ok ? (
          <Empty message={deployments.error} />
        ) : deployments.data.deployments.length === 0 ? (
          <Empty message="No deployments yet." />
        ) : (
          deployments.data.deployments.map((d) => (
            <div className="row" key={d.id}>
              <Badge variant={d.status} />
              <div className="row-main">
                <div className="row-title">
                  {d.environment} · {shortSha(d.sha)}
                </div>
                <div className="row-meta">
                  {d.ref}
                  {d.actor ? ` · by ${d.actor}` : ""}
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
