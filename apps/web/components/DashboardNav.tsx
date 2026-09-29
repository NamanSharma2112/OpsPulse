"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion, useReducedMotion } from "motion/react";
import { useEffect, useState } from "react";
import { indicator } from "./marketing/motion";
import { Wordmark } from "./Wordmark";

const NAV = [
  { href: "/dashboard", label: "Health" },
  { href: "/dashboard/deployments", label: "Deployments" },
  { href: "/dashboard/pull-requests", label: "Pull requests" },
  { href: "/dashboard/incidents", label: "Incidents" },
  { href: "/dashboard/events", label: "Events" },
  { href: "/dashboard/connect", label: "Connect" },
];

type NavUser = {
  name: string;
  email: string;
  login?: string;
  avatar?: string;
};

/**
 * The product's navigation. Same bar as the marketing site — sticky,
 * hairline on scroll, one pill travelling between links — so moving from the
 * site into the app does not feel like moving between two products.
 */
export function DashboardNav({ user }: { user: NavUser | null }) {
  const pathname = usePathname();
  const [scrolled, setScrolled] = useState(false);
  const [hovered, setHovered] = useState<string | null>(null);
  const reduce = useReducedMotion();

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  // The pill rests on the current page and follows the pointer on hover.
  const marked = hovered ?? pathname;

  return (
    <header className="ui-nav" data-scrolled={scrolled}>
      <div className="ui-container">
        <div className="ui-nav-inner">
          <Link href="/" className="ui-wordmark">
            <Wordmark />
            OpsPulse
          </Link>

          <nav className="ui-nav-links" onMouseLeave={() => setHovered(null)}>
            {NAV.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="ui-nav-link"
                data-current={pathname === item.href}
                onMouseEnter={() => setHovered(item.href)}
              >
                {marked === item.href && !reduce ? (
                  <motion.span
                    layoutId="dash-nav-indicator"
                    className="ui-nav-link-bg"
                    transition={indicator}
                  />
                ) : null}
                <span className="ui-nav-link-label">{item.label}</span>
              </Link>
            ))}
          </nav>

          <div className="ui-nav-actions">
            {user ? (
              <span className="dash-user" title={user.email}>
                {user.avatar ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img className="dash-user-avatar" src={user.avatar} alt="" width={22} height={22} />
                ) : null}
                {user.login ? `@${user.login}` : user.name}
              </span>
            ) : null}
          </div>
        </div>
      </div>
    </header>
  );
}
