import Link from "next/link";
import { Empty } from "@/components/Empty";
import { HealthScore } from "@/components/HealthScore";
import { Notice } from "@/components/Notice";
import { activeProjectId, getAnalytics, getScore, listAnomalies } from "@/lib/api";
import { formatRelative } from "@/lib/format";

export const dynamic = "force-dynamic";

/** Windows the API accepts. Anything else is rejected rather than coerced. */
const WINDOWS = ["7d", "14d", "30d", "90d"] as const;

export default async function AnalyticsPage({
  searchParams,
}: {
  searchParams: Promise<{ window?: string }>;
}) {
  const project = await activeProjectId();
  if (!project.ok) {
    return (
      <Notice
        title="Nothing connected yet"
        body="Analytics are derived from the events your repositories send, so there is nothing to measure until one is connected."
        action={
          <Link href="/dashboard/connect" className="ui-btn ui-btn-primary">
            Connect a repository
          </Link>
        }
      />
    );
  }

  const requested = (await searchParams).window;
  const window = WINDOWS.includes(requested as (typeof WINDOWS)[number]) ? requested! : "30d";

  const [analytics, score, anomalies] = await Promise.all([
    getAnalytics(project.data, window),
    getScore(project.data, window),
    listAnomalies(project.data, 10),
  ]);

  return (
    <>
      <header className="dash-head">
        <h1 className="ui-title">Analytics</h1>
        <p className="ui-body">
          Delivery metrics and the health score derived from them, over a
          rolling window.
        </p>
      </header>

      <nav className="window-tabs" aria-label="Window">
        {WINDOWS.map((w) => (
          <Link
            key={w}
            href={`/dashboard/analytics?window=${w}`}
            className="window-tab"
            data-active={w === window}
          >
            {w}
          </Link>
        ))}
      </nav>

      {score.ok ? <HealthScore health={score.data} /> : <Empty message={score.error} />}

      <h2 className="dash-section-title">Delivery metrics</h2>
      {!analytics.ok ? (
        <Empty message={analytics.error} />
      ) : (
        <div className="dash-tiles">
          <Tile
            label="PR cycle time"
            value={fmtHours(analytics.data.pr_cycle_time_hours)}
            note={
              analytics.data.pr_cycle_time_hours === null
                ? "No pull requests merged in this window"
                : `median of ${analytics.data.prs_merged} · p90 ${fmtHours(analytics.data.pr_cycle_time_p90_hours)}`
            }
          />
          <Tile
            label="Deploy frequency"
            value={`${analytics.data.deployments_per_day.toFixed(2)}/day`}
            note={`${analytics.data.deployments_succeeded} to production in ${window}`}
          />
          <Tile
            label="Open issues"
            value={String(analytics.data.open_issues)}
            note="Open right now, not just this window"
          />
          <Tile
            label="Commit frequency"
            value={`${analytics.data.commits_per_day.toFixed(1)}/day`}
            note={`${analytics.data.commits} commits in ${window}`}
          />
          <Tile
            label="PR merge rate"
            value={fmtPercent(analytics.data.pr_merge_rate)}
            note={
              analytics.data.prs_decided === 0
                ? "No pull requests decided yet"
                : `${analytics.data.prs_decided} decided · open ones excluded`
            }
          />
          <Tile
            label="Deploy success"
            value={fmtPercent(analytics.data.deployment_success_rate)}
            note={`${analytics.data.deployments_finished} finished deployments`}
          />
        </div>
      )}

      <h2 className="dash-section-title">Anomalies</h2>
      <div className="ui-card">
        {!anomalies.ok ? (
          <Empty message={anomalies.error} />
        ) : anomalies.data.anomalies.length === 0 ? (
          <Empty message="Nothing has moved outside its baseline. Detection needs about a week of history before it will judge anything." />
        ) : (
          anomalies.data.anomalies.map((a) => (
            <div className="ui-row anomaly-row" key={a.id}>
              <span className="anomaly-arrow" data-direction={a.direction}>
                {a.direction === "above" ? "↑" : "↓"}
              </span>
              <div className="ui-row-main">
                <div className="ui-row-title">{a.explanation}</div>
                <div className="ui-row-meta">
                  {a.metric} · detected {formatRelative(a.detected_at)}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </>
  );
}

function Tile({ label, value, note }: { label: string; value: string; note: string }) {
  return (
    <div className="dash-tile">
      <div className="dash-tile-label">{label}</div>
      <div className="dash-tile-value">{value}</div>
      <div className="dash-tile-note">{note}</div>
    </div>
  );
}

/** A null metric is unknown, and says so rather than showing a zero. */
function fmtHours(v: number | null): string {
  if (v === null) return "—";
  if (v < 1) return `${Math.round(v * 60)}m`;
  if (v < 48) return `${v.toFixed(1)}h`;
  return `${(v / 24).toFixed(1)}d`;
}

function fmtPercent(v: number | null): string {
  return v === null ? "—" : `${Math.round(v * 100)}%`;
}
