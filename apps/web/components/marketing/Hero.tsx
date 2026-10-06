"use client";

import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";
import { LiveDemo } from "./LiveDemo";
import { Terminal } from "./Terminal";
import { group, item, itemStill } from "./motion";

/**
 * The hero enters as one cascade rather than six separate arrivals, and the
 * panels come last so the eye lands on the sentence before the interface.
 *
 * This is the one place on the site with an on-load animation: it is seen
 * once per visit, and its purpose is explanation — the panels demonstrate the
 * product rather than describing it.
 */
export function Hero() {
  const reduce = useReducedMotion();
  const child = reduce ? itemStill : item;

  return (
    <section className="mk-hero">
      {/* Decorative engineering grid behind the fold. */}
      <div className="ui-grid-bg" aria-hidden="true" />

      <motion.div
        className="ui-container mk-hero-grid"
        variants={group}
        initial="hidden"
        animate="shown"
      >
        <div className="mk-hero-copy">
          <motion.span className="ui-chip" variants={child}>
            <span className="ui-chip-dot" />
            correlation engine with written explanations
          </motion.span>

          <motion.h1 className="ui-display-xl" variants={child}>
            The platform for engineering operations intelligence
          </motion.h1>

          <motion.p className="ui-body-lg" variants={child}>
            OpsPulse ingests every event your repositories emit, works out
            which of your signals moved together, and writes the incident in a
            paragraph anyone can act on.
          </motion.p>

          <motion.div className="mk-hero-actions" variants={child}>
            <Link href="/dashboard" className="ui-btn ui-btn-primary">
              Get started
            </Link>
            <Link href="#how-it-works" className="ui-btn ui-btn-secondary">
              See how it works
            </Link>
          </motion.div>
        </div>

        <div className="mk-hero-panels">
          <motion.div variants={child}>
            <Terminal />
          </motion.div>
          <motion.div variants={child}>
            <LiveDemo />
          </motion.div>
        </div>
      </motion.div>
    </section>
  );
}
