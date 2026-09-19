"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

const LINKS = [
  { href: "#how-it-works", label: "How it works" },
  { href: "#integrations", label: "Integrations" },
  { href: "#pipeline", label: "Pipeline" },
  { href: "/dashboard", label: "Live dashboard" },
];

/**
 * Sticky top bar. It sits transparent over the cream canvas and solidifies
 * with a hairline once the page scrolls, per DESIGN.md's top-nav.
 */
export function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header className="mk-nav" data-scrolled={scrolled} data-open={open}>
      <div className="mk-container">
        <div className="mk-nav-inner">
          <Link href="/" className="mk-wordmark" onClick={() => setOpen(false)}>
            <Mark />
            OpsPulse
          </Link>

          <nav className="mk-nav-links">
            {LINKS.map((link) => (
              <Link key={link.href} href={link.href}>
                {link.label}
              </Link>
            ))}
          </nav>

          <div className="mk-nav-actions">
            <Link href="/dashboard" className="mk-btn mk-btn-tertiary">
              Sign in
            </Link>
            <Link href="/dashboard" className="mk-btn mk-btn-primary">
              Get started
            </Link>
            <button
              type="button"
              className="mk-nav-toggle"
              aria-label={open ? "Close menu" : "Open menu"}
              aria-expanded={open}
              onClick={() => setOpen((v) => !v)}
            >
              <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true">
                {open ? (
                  <path
                    d="M3 3l10 10M13 3L3 13"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                ) : (
                  <path
                    d="M2 4h12M2 8h12M2 12h12"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                )}
              </svg>
            </button>
          </div>
        </div>

        <div className="mk-nav-mobile">
          {LINKS.map((link) => (
            <Link key={link.href} href={link.href} onClick={() => setOpen(false)}>
              {link.label}
            </Link>
          ))}
        </div>
      </div>
    </header>
  );
}

function Mark() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
      <rect width="18" height="18" rx="5" fill="#111111" />
      <path
        d="M3.6 9.4h2.6l1.5-3.6 2 7 1.6-3.4h3.1"
        fill="none"
        stroke="#ffffff"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
