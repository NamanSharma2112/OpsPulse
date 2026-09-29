/**
 * Shown when the dashboard cannot read the API: unreachable service, or no
 * projects connected yet. It explains the next step rather than rendering an
 * empty screen.
 */
export function Notice({
  title,
  body,
  action,
}: {
  title: string;
  body: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="ui-card ui-card-pad">
      <h2 className="ui-card-title">{title}</h2>
      <p className="ui-body" style={{ marginTop: 8 }}>
        {body}
      </p>
      {action ? <div style={{ marginTop: 18 }}>{action}</div> : null}
    </div>
  );
}
