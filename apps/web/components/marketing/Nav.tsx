"use client";

import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";
import { useEffect, useState } from "react";
import { indicator } from "./motion";

const LINKS = [
  { href: "#how-it-works", label: "How it works" },
  { href: "#integrations", label: "Integrations" },
  { href: "/dashboard", label: "Dashboard" },
];

/**
 * Sticky top bar. It sits transparent over the cream canvas and solidifies
 * with a hairline once the page scrolls.
 *
 * The hovered link is marked by one pill that travels between them, rather
 * than each link lighting its own background — one object moving reads as a
 * single response to the pointer.
 */
export function Nav() {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);
  const [hovered, setHovered] = useState<string | null>(null);
  const reduce = useReducedMotion();

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

          <nav className="mk-nav-links" onMouseLeave={() => setHovered(null)}>
            {LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="mk-nav-link"
                onMouseEnter={() => setHovered(link.href)}
              >
                {hovered === link.href && !reduce ? (
                  <motion.span
                    layoutId="mk-nav-hover"
                    className="mk-nav-link-bg"
                    transition={indicator}
                  />
                ) : null}
                <span className="mk-nav-link-label">{link.label}</span>
              </Link>
            ))}
          </nav>

          <div className="mk-nav-actions">
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
                  <path d="M3 3l10 10M13 3L3 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                ) : (
                  <path d="M2 4h12M2 8h12M2 12h12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
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
