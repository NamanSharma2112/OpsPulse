/** Display helpers shared by the dashboard pages. */

/** Formats an ISO timestamp as a short absolute time. */
export function formatDateTime(iso: string | undefined): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

/** Formats an ISO timestamp as "3h ago". */
export function formatRelative(iso: string | undefined): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";

  const seconds = Math.round((Date.now() - then) / 1000);
  const future = seconds < 0;
  const abs = Math.abs(seconds);

  const units: [number, string][] = [
    [60, "s"],
    [3600, "m"],
    [86400, "h"],
    [2592000, "d"],
  ];
  for (let i = 0; i < units.length; i++) {
    const [limit, suffix] = units[i]!;
    if (abs < limit) {
      const divisor = i === 0 ? 1 : units[i - 1]![0];
      const value = Math.floor(abs / divisor);
      return future ? `in ${value}${suffix}` : `${value}${suffix} ago`;
    }
  }
  return `${Math.floor(abs / 2592000)}mo ago`;
}

/** Formats a 0-1 ratio as a whole percentage. */
export function formatPercent(ratio: number): string {
  return `${Math.round(ratio * 100)}%`;
}

/** Duration between two timestamps, as "4m" or "1h 12m". */
export function formatDuration(from: string, to: string | undefined): string {
  if (!to) return "—";
  const ms = new Date(to).getTime() - new Date(from).getTime();
  if (Number.isNaN(ms) || ms < 0) return "—";

  const minutes = Math.round(ms / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

/** Shortens a commit SHA for display. */
export function shortSha(sha: string): string {
  return sha ? sha.slice(0, 7) : "—";
}
