import Link from "next/link";

const COLUMNS = [
  {
    title: "Product",
    links: [
      { href: "#how-it-works", label: "How it works" },
      { href: "#pipeline", label: "Pipeline" },
      { href: "#integrations", label: "Integrations" },
      { href: "/dashboard", label: "Live dashboard" },
    ],
  },
  {
    title: "Developers",
    links: [
      { href: "/dashboard", label: "API reference" },
      { href: "/dashboard", label: "Webhook setup" },
      { href: "/dashboard", label: "Self-hosting" },
      { href: "/dashboard", label: "Changelog" },
    ],
  },
  {
    title: "Company",
    links: [
      { href: "/dashboard", label: "About" },
      { href: "/dashboard", label: "Security" },
      { href: "/dashboard", label: "Privacy" },
      { href: "/dashboard", label: "Contact" },
    ],
  },
];

export function Footer() {
  return (
    <footer className="mk-footer">
      <div className="mk-container">
        <div className="mk-footer-grid">
          <div className="mk-footer-col">
            <span className="mk-wordmark" style={{ marginBottom: 12 }}>
              OpsPulse
            </span>
            <p className="mk-body-sm" style={{ maxWidth: "34ch" }}>
              Engineering and business operations intelligence. Raw signals in,
              written incidents out.
            </p>
          </div>
          {COLUMNS.map((column) => (
            <div className="mk-footer-col" key={column.title}>
              <h4>{column.title}</h4>
              <ul>
                {column.links.map((link) => (
                  <li key={link.label}>
                    <Link href={link.href}>{link.label}</Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div className="mk-footer-bottom">
          <p className="mk-caption">© {new Date().getFullYear()} OpsPulse</p>
          <p className="mk-caption">Built on an append-only event log.</p>
        </div>
      </div>
    </footer>
  );
}
