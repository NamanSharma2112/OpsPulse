import Link from "next/link";
import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { Notice } from "@/components/Notice";
import { activeProjectId, getHealth, listDeployments, listIncidents } from "@/lib/api";
import { formatPercent, formatRelative, shortSha } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function HealthPage() {
  const project = await activeProjectId();
  if (!project.ok) {
    return (
      <Notice
        title="Nothing connected yet"
        body="Connect a repository and OpsPulse installs its webhook, then starts recording events as soon as GitHub delivers the first one."
        action={
          <Link href="/dashboard/connect" className="ui-btn ui-btn-primary">
            Connect a repository
          </Link>
        }
      />
    );
  }

  const [health, deployments, incidents] = await Promise.all([
    getHealth(project.data),
    listDeployments(project.data, 5),
    listIncidents(project.data, "open", 5),
  ]);

  if (!health.ok) {
    return <Notice title="No data to show" body={health.error} />;
  }
  const h = health.data;

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Health</h1>
        <p className="ui-body">
          Rolling {h.window} window ·{" "}
          {h.deployments.succeeded + h.deployments.failed} deployments
        </p>
      </header>

      <div className="dash-tiles">
        <div className="dash-tile">
          <div className="dash-tile-label">Status</div>
          <div className="dash-tile-value">
            <Badge variant={h.status} />
          </div>
          <div className="dash-tile-note">
            {h.status === "healthy"
              ? "No open incidents or failed deploys"
              : `${h.open_incidents} open incident${h.open_incidents === 1 ? "" : "s"}`}
          </div>
        </div>

        <div className="dash-tile">
          <div className="dash-tile-label">Deploy success</div>
          <div className="dash-tile-value">{formatPercent(h.deployments.success_rate)}</div>
          <div className="dash-tile-note">
            {h.deployments.succeeded} succeeded · {h.deployments.failed} failed
          </div>
        </div>

        <div className="dash-tile">
          <div className="dash-tile-label">Open pull requests</div>
          <div className="dash-tile-value">{h.open_pull_requests}</div>
          <div className="dash-tile-note">Awaiting review or merge</div>
        </div>

        <div className="dash-tile">
          <div className="dash-tile-label">Last deploy</div>
          <div className="dash-tile-value" style={{ fontSize: 20 }}>
            {h.last_deployment ? formatRelative(h.last_deployment.started_at) : "—"}
          </div>
          <div className="dash-tile-note">
            {h.last_deployment
              ? `${h.last_deployment.environment} · ${shortSha(h.last_deployment.commit_sha)}`
              : "No deployments recorded"}
          </div>
        </div>
      </div>

      <h2 className="dash-section-title">Open incidents</h2>
      <div className="ui-card">
        {!incidents.ok ? (
          <Empty message={incidents.error} />
        ) : incidents.data.incidents.length === 0 ? (
          <Empty message="Nothing on fire." />
        ) : (
          incidents.data.incidents.map((incident) => (
            <div className="ui-row" key={incident.id}>
              <Badge variant={incident.severity} />
              <div className="ui-row-main">
                <div className="ui-row-title">{incident.title}</div>
                <div className="ui-row-meta">
                  {incident.source} · opened {formatRelative(incident.opened_at)}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      <h2 className="dash-section-title">Recent deployments</h2>
      <div className="ui-card">
        {!deployments.ok ? (
          <Empty message={deployments.error} />
        ) : deployments.data.deployments.length === 0 ? (
          <Empty message="No deployments yet." />
        ) : (
          deployments.data.deployments.map((d) => (
            <div className="ui-row" key={d.id}>
              <Badge variant={d.status} />
              <div className="ui-row-main">
                <div className="ui-row-title">
                  {d.environment} · {shortSha(d.commit_sha)}
                </div>
                <div className="ui-row-meta">
                  {d.ref}
                  {d.actor ? ` · by ${d.actor}` : ""}
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
