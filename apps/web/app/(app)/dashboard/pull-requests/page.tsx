import { Badge } from "@/components/Badge";
import { Empty } from "@/components/Empty";
import { Notice } from "@/components/Notice";
import { activeProjectId, listPullRequests } from "@/lib/api";
import { formatDuration, formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function PullRequestsPage() {
  const project = await activeProjectId();
  if (!project.ok) return <Notice title="No data to show" body={project.error} />;

  const result = await listPullRequests(project.data, "", 50);
  if (!result.ok) return <Notice title="No data to show" body={result.error} />;
  const prs = result.data.pull_requests;

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Pull requests</h1>
      <p className="ui-body">
        {prs.filter((pr) => pr.state === "open").length} open ·{" "}
        {prs.filter((pr) => pr.state === "merged").length} merged
      </p>
      </header>

      <div className="ui-card">
        {prs.length === 0 ? (
          <Empty message="No pull requests recorded yet." />
        ) : (
          prs.map((pr) => (
            <div className="ui-row" key={pr.id}>
              <Badge variant={pr.draft ? "neutral" : pr.state} label={pr.draft ? "draft" : pr.state} />
              <div className="ui-row-main">
                <div className="ui-row-title">
                  {pr.url ? (
                    <a href={pr.url} target="_blank" rel="noreferrer">
                      #{pr.number} {pr.title}
                    </a>
                  ) : (
                    <>
                      #{pr.number} {pr.title}
                    </>
                  )}
                </div>
                <div className="ui-row-meta">
                  by {pr.author} · opened {formatRelative(pr.opened_at)}
                  {pr.merged_at ? ` · merged in ${formatDuration(pr.opened_at, pr.merged_at)}` : ""}
                </div>
              </div>
              <div className="ui-row-side">{formatRelative(pr.updated_at)}</div>
            </div>
          ))
        )}
      </div>
    </>
  );
}
