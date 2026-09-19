import Link from "next/link";
import { redirect } from "next/navigation";
import { currentUser, isSignedIn } from "@/lib/api";
import { publicApiBaseUrl } from "@/lib/session";
import "./dashboard.css";

const NAV = [
  { href: "/dashboard", label: "Health" },
  { href: "/dashboard/deployments", label: "Deployments" },
  { href: "/dashboard/pull-requests", label: "Pull requests" },
  { href: "/dashboard/incidents", label: "Incidents" },
  { href: "/dashboard/events", label: "Events" },
  { href: "/dashboard/connect", label: "Connect" },
];

/** Shell for the signed-in product. Keeps its own dark theme. */
export default async function DashboardLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  // Everything below this line is private, so the gate lives in the layout
  // rather than being repeated in each page.
  if (!(await isSignedIn())) redirect("/signin");
  const me = await currentUser();

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
            {me.ok ? (
              <span className="nav-user" title={me.data.email}>
                {me.data.github_login ? `@${me.data.github_login}` : me.data.name}
              </span>
            ) : (
              <a className="nav-user" href={`${publicApiBaseUrl()}/v1/auth/github`}>
                Sign in
              </a>
            )}
          </nav>
        </header>
        <main>{children}</main>
      </div>
    </div>
  );
}
