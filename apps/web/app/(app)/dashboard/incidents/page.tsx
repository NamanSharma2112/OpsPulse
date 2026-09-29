import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { Notice } from "@/components/Notice";
import { activeProjectId, listIncidents } from "@/lib/api";
import { formatDuration, formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function IncidentsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <Notice title="No data to show" body={project.error} />;

  const result = await listIncidents(project.data, "", 50);
  if (!result.ok) return <Notice title="No data to show" body={result.error} />;
  const { incidents } = result.data;
  const open = incidents.filter((i) => i.status !== "resolved");

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Incidents</h1>
      <p className="ui-body">
        {open.length} open · raised by failing default-branch workflows and issues labelled{" "}
        <code>incident</code>
      </p>
      </header>

      <div className="ui-card">
        {incidents.length === 0 ? (
          <Empty message="No incidents recorded. That is the goal." />
        ) : (
          incidents.map((incident) => (
            <div className="ui-row" key={incident.id}>
              <Badge variant={incident.status === "resolved" ? "resolved" : incident.severity} />
              <div className="ui-row-main">
                <div className="ui-row-title">
                  {incident.url ? (
                    <a href={incident.url} target="_blank" rel="noreferrer">
                      {incident.title}
                    </a>
                  ) : (
                    incident.title
                  )}
                </div>
                <div className="ui-row-meta">
                  {incident.source} · opened {formatRelative(incident.opened_at)}
                  {incident.resolved_at
                    ? ` · resolved after ${formatDuration(incident.opened_at, incident.resolved_at)}`
                    : ""}
                </div>
              </div>
              <div className="ui-row-side">{incident.status}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
