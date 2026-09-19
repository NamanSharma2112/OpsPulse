import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { ApiUnavailable } from "@/components/Notice";
import { activeProjectId, listEvents, listRepositories } from "@/lib/api";
import { formatDateTime, formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function EventsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <ApiUnavailable error={project.error} />;

  const [events, repositories] = await Promise.all([
    listEvents(project.data, 100),
    listRepositories(project.data),
  ]);
  if (!events.ok) return <ApiUnavailable error={events.error} />;

  // Events store a repository id; show the name the reader recognises.
  const repoNames = new Map(
    repositories.ok ? repositories.data.repositories.map((r) => [r.id, r.external_id]) : [],
  );

  return (
    <>
      <h1 className="page-title">Events</h1>
      <p className="page-subtitle">
        Every webhook delivery, stored verbatim. Deployments, pull requests and
        incidents are all derived from this feed.
      </p>

      <div className="card">
        {events.data.events.length === 0 ? (
          <Empty message="No events yet. Connect a repository and its first webhook will land here." />
        ) : (
          events.data.events.map((event) => (
            <div className="row" key={event.id}>
              <Badge variant="neutral" label={event.source} />
              <div className="row-main">
                <div className="row-title">
                  {event.type}
                  {event.action ? (
                    <span style={{ opacity: 0.6 }}>{` · ${event.action}`}</span>
                  ) : null}
                </div>
                <div className="row-meta">
                  {repoNames.get(event.repository_id ?? "") ?? "—"}
                  {event.actor ? ` · by ${event.actor}` : ""}
                  {` · ${formatDateTime(event.occurred_at)}`}
                </div>
              </div>
              <div className="row-side">{formatRelative(event.received_at)}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
