/**
 * Continuous marquee of the sources OpsPulse ingests. The row is rendered
 * twice so the -100% translation loops seamlessly; the duplicate is hidden
 * from assistive technology.
 */

type Integration = { name: string; icon: React.ReactNode };

const INTEGRATIONS: Integration[] = [
  { name: "GitHub", icon: <Circle d="M9 1.5a7.5 7.5 0 0 0-2.4 14.6c.4.1.5-.2.5-.4v-1.3c-2 .5-2.5-.9-2.5-.9-.3-.9-.8-1.1-.8-1.1-.7-.4 0-.4 0-.4.8 0 1.2.8 1.2.8.7 1.2 1.8.8 2.3.6 0-.5.3-.9.5-1.1-1.6-.2-3.3-.8-3.3-3.6 0-.8.3-1.5.8-2 0-.2-.4-.9.1-2 0 0 .6-.2 2 .8a7 7 0 0 1 3.7 0c1.4-1 2-.8 2-.8.5 1.1.1 1.8.1 2 .5.5.8 1.2.8 2 0 2.8-1.7 3.4-3.3 3.6.3.2.5.7.5 1.4v2.1c0 .2.1.5.5.4A7.5 7.5 0 0 0 9 1.5Z" /> },
  { name: "Datadog", icon: <Circle d="M15 3 9.8 6.3l-.3 2.5-1.4.9-.2 1.7-1 .6-2-2 1.2-.8.3 1 .6-.4.4-2.4 1.6-1-.3-1.2L4 8.4l-1-1.6 6.4-3.6L15 3Zm-5.4 9.6.8-.5-.2 2.2-2.4 1.2-.5-1 1.6-.8.7-1.1Z" /> },
  { name: "PagerDuty", icon: <Circle d="M5.5 2h4a4 4 0 0 1 0 8h-2v2.5h-2V2Zm2 2v4h2a2 2 0 0 0 0-4h-2ZM5.5 14h2v2h-2v-2Z" /> },
  { name: "Sentry", icon: <Circle d="M9 2.4a1.3 1.3 0 0 0-1.1.6L6 6.3l1.1.7 1.9-3.3 5.2 9H12a6.6 6.6 0 0 0-2.4-4.3l-1.1.8a5.3 5.3 0 0 1 2 3.5h2.8a.7.7 0 0 0 .6-1L10.1 3a1.3 1.3 0 0 0-1.1-.6ZM5.4 8.2 3.2 12a.7.7 0 0 0 .6 1h3v-1.3H4.5l1.6-2.8-.7-.7Z" /> },
  { name: "Stripe", icon: <Circle d="M8.6 6.9c0-.5.4-.7 1-.7.9 0 2 .3 2.9.8V4.3A7.6 7.6 0 0 0 9.6 3.8C7.4 3.8 6 5 6 7c0 3.1 4.2 2.6 4.2 3.9 0 .5-.5.7-1.1.7-1 0-2.2-.4-3.2-1v2.8c1 .4 2.1.6 3.2.6 2.3 0 3.8-1.1 3.8-3.2 0-3.3-4.3-2.7-4.3-3.9Z" /> },
  { name: "Slack", icon: <Circle d="M4.6 10.8a1.5 1.5 0 1 1-1.5-1.5h1.5v1.5Zm.8 0a1.5 1.5 0 0 1 3 0v3.7a1.5 1.5 0 0 1-3 0v-3.7Zm1.5-6a1.5 1.5 0 1 1 1.5-1.5v1.5H6.9Zm0 .8a1.5 1.5 0 0 1 0 3H3.2a1.5 1.5 0 0 1 0-3h3.7Zm6 1.5a1.5 1.5 0 1 1 1.5 1.5h-1.5V7.1Zm-.8 0a1.5 1.5 0 0 1-3 0V3.4a1.5 1.5 0 0 1 3 0v3.7Zm-1.5 6a1.5 1.5 0 1 1-1.5 1.5v-1.5h1.5Zm0-.8a1.5 1.5 0 0 1 0-3h3.7a1.5 1.5 0 0 1 0 3h-3.7Z" /> },
  { name: "Linear", icon: <Circle d="M2.3 10.3 7.7 15.7a6.9 6.9 0 0 1-5.4-5.4Zm-.2-1.7L9.4 15.9c.6-.1 1.1-.3 1.6-.5L2.6 7c-.2.5-.4 1-.5 1.6Zm1 -2.6 7.9 7.9c.4-.3.8-.6 1.1-.9L3.9 5.5c-.3.3-.6.7-.8 1.1Zm1.5-1.8 7.2 7.2a6.9 6.9 0 0 0-7.2-7.2Z" /> },
  { name: "Grafana", icon: <Circle d="M9 2 3 5v4c0 3.4 2.6 6.3 6 7 3.4-.7 6-3.6 6-7V5L9 2Zm0 2.2 4 2v2.8c0 2.4-1.7 4.5-4 5.1-2.3-.6-4-2.7-4-5.1V6.2l4-2Z" /> },
  { name: "Snowflake", icon: <Circle d="M8.3 1.8h1.4v3.4l2.5-1.5.7 1.2-2.6 1.5 2.6 1.5-.7 1.2-2.5-1.4v3.4H8.3v-3.4l-2.5 1.4-.7-1.2 2.6-1.5-2.6-1.5.7-1.2 2.5 1.5V1.8Zm-4 11.1 3.5-2 .7 1.2-3.5 2-.7-1.2Zm9.4 0-.7 1.2-3.5-2 .7-1.2 3.5 2Z" /> },
  { name: "AWS CloudWatch", icon: <Circle d="M13.4 7.1A4.6 4.6 0 0 0 4.6 6 3.5 3.5 0 0 0 5 13h8.1a3 3 0 0 0 .3-5.9ZM6.4 9.3h1.5l1-2.2 1.4 4 .9-1.8h2v1.2h-1.2l-1.5 3-1.4-4-1 2h-2.3l.6-2.2Z" /> },
  { name: "Vercel", icon: <Circle d="M9 3 16 15H2L9 3Z" /> },
  { name: "PostgreSQL", icon: <Circle d="M9 2c3.1 0 5.6 1 5.6 3.3 0 1.8-.5 4.5-1.3 6.4-.6 1.5-1.4 2.4-2.4 2.4-.6 0-1-.3-1.4-.6-.2-.2-.4-.3-.5-.3s-.3.1-.5.3c-.4.3-.8.6-1.4.6-1 0-1.8-.9-2.4-2.4-.8-1.9-1.3-4.6-1.3-6.4C3.4 3 5.9 2 9 2Zm0 1.4c-2.5 0-4.2.7-4.2 1.9 0 1.6.5 4.1 1.2 5.9.4 1 .9 1.5 1.1 1.5.1 0 .2 0 .5-.2.4-.4.9-.7 1.4-.7s1 .3 1.4.7c.3.2.4.2.5.2.2 0 .7-.5 1.1-1.5.7-1.8 1.2-4.3 1.2-5.9 0-1.2-1.7-1.9-4.2-1.9Z" /> },
];

export function Integrations() {
  return (
    <section className="mk-marquee" id="integrations">
      <div className="mk-container mk-marquee-label">
        <p className="mk-body-sm">
          OpsPulse reads the systems you already run — code, deploys, paging,
          errors, revenue.
        </p>
      </div>
      <div className="mk-marquee-track">
        <Row />
        <Row aria-hidden />
      </div>
    </section>
  );
}

function Row({ "aria-hidden": hidden }: { "aria-hidden"?: boolean }) {
  return (
    <div className="mk-marquee-row" aria-hidden={hidden}>
      {INTEGRATIONS.map((integration) => (
        <span className="mk-logo" key={integration.name}>
          {integration.icon}
          {integration.name}
        </span>
      ))}
    </div>
  );
}

/** Wraps a path in a consistently sized 18px glyph. */
function Circle({ d }: { d: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
      <path d={d} fill="currentColor" />
    </svg>
  );
}
