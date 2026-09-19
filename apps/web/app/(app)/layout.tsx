import Link from "next/link";
import "./dashboard.css";

const NAV = [
  { href: "/dashboard", label: "Health" },
  { href: "/dashboard/deployments", label: "Deployments" },
  { href: "/dashboard/pull-requests", label: "Pull requests" },
  { href: "/dashboard/incidents", label: "Incidents" },
];

/** Shell for the signed-in product. Keeps its own dark theme. */
export default function DashboardLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <div className="dash">
      <div className="shell">
        <header className="masthead">
          <Link href="/" className="brand">
            <span className="brand-mark">◆</span>
            <span>OpsPulse</span>
            <span className="brand-sub">delivery &amp; reliability</span>
          </Link>
          <nav className="nav">
            {NAV.map((item) => (
              <Link key={item.href} href={item.href}>
                {item.label}
              </Link>
            ))}
          </nav>
        </header>
        <main>{children}</main>
      </div>
    </div>
  );
}
