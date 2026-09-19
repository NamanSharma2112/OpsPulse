import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { ApiUnavailable } from "@/components/Notice";
import { activeProjectId, listIncidents } from "@/lib/api";
import { formatDuration, formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function IncidentsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <ApiUnavailable error={project.error} />;

  const result = await listIncidents(project.data, "", 50);
  if (!result.ok) return <ApiUnavailable error={result.error} />;
  const { incidents } = result.data;
  const open = incidents.filter((i) => i.status !== "resolved");

  return (
    <>
      <h1 className="page-title">Incidents</h1>
      <p className="page-subtitle">
        {open.length} open · raised by failing default-branch workflows and issues labelled{" "}
        <code>incident</code>
      </p>

      <div className="card">
        {incidents.length === 0 ? (
          <Empty message="No incidents recorded. That is the goal." />
        ) : (
          incidents.map((incident) => (
            <div className="row" key={incident.id}>
              <Badge variant={incident.status === "resolved" ? "resolved" : incident.severity} />
              <div className="row-main">
                <div className="row-title">
                  {incident.url ? (
                    <a href={incident.url} target="_blank" rel="noreferrer">
                      {incident.title}
                    </a>
                  ) : (
                    incident.title
                  )}
                </div>
                <div className="row-meta">
                  {incident.source} · opened {formatRelative(incident.opened_at)}
                  {incident.resolved_at
                    ? ` · resolved after ${formatDuration(incident.opened_at, incident.resolved_at)}`
                    : ""}
                </div>
              </div>
              <div className="row-side">{incident.status}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
