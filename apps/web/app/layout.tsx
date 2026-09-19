import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "OpsPulse",
  description: "Delivery and reliability signals for your GitHub repositories",
};

const NAV = [
  { href: "/", label: "Health" },
  { href: "/deployments", label: "Deployments" },
  { href: "/pull-requests", label: "Pull requests" },
  { href: "/incidents", label: "Incidents" },
];

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>
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
      </body>
    </html>
  );
}
