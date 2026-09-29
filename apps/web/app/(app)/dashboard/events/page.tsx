import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { Notice } from "@/components/Notice";
import { activeProjectId, listEvents, listRepositories } from "@/lib/api";
import { formatDateTime, formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function EventsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <Notice title="No data to show" body={project.error} />;

  const [events, repositories] = await Promise.all([
    listEvents(project.data, 100),
    listRepositories(project.data),
  ]);
  if (!events.ok) return <Notice title="No data to show" body={events.error} />;

  // Events store a repository id; show the name the reader recognises.
  const repoNames = new Map(
    repositories.ok ? repositories.data.repositories.map((r) => [r.id, r.external_id]) : [],
  );

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Events</h1>
      <p className="ui-body">
        Every webhook delivery, stored verbatim. Deployments, pull requests and
        incidents are all derived from this feed.
      </p>
      </header>

      <div className="ui-card">
        {events.data.events.length === 0 ? (
          <Empty message="No events yet. Connect a repository and its first webhook will land here." />
        ) : (
          events.data.events.map((event) => (
            <div className="ui-row" key={event.id}>
              <Badge variant="neutral" label={event.source} />
              <div className="ui-row-main">
                <div className="ui-row-title">
                  {event.type}
                  {event.action ? (
                    <span style={{ opacity: 0.6 }}>{` · ${event.action}`}</span>
                  ) : null}
                </div>
                <div className="ui-row-meta">
                  {repoNames.get(event.repository_id ?? "") ?? "—"}
                  {event.actor ? ` · by ${event.actor}` : ""}
                  {` · ${formatDateTime(event.occurred_at)}`}
                </div>
              </div>
              <div className="ui-row-side">{formatRelative(event.received_at)}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
