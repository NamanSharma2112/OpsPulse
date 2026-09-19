/**
 * Shown when the dashboard cannot read the API: missing token, unreachable
 * service, or no projects registered yet. It explains the fix rather than
 * rendering an empty screen.
 */
export function Notice({
  title,
  body,
  hint,
}: {
  title: string;
  body: string;
  hint?: string;
}) {
  return (
    <div className="notice">
      <div className="notice-title">{title}</div>
      <div className="notice-body">{body}</div>
      {hint ? <pre>{hint}</pre> : null}
    </div>
  );
}

/** Notice variant for a fetch that failed. */
export function ApiUnavailable({ error }: { error: string }) {
  return (
    <Notice
      title="No data to show"
      body={error}
      hint={`# start the stack, then create a project
docker compose up -d

curl -X POST $OPSPULSE_API_URL/v1/auth/register \\
  -H 'Content-Type: application/json' \\
  -d '{"email":"you@example.com","name":"You","password":"a-long-password"}'`}
    />
  );
}
