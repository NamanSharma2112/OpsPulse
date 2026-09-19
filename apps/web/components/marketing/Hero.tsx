"use client";

import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";
import { LiveDemo } from "./LiveDemo";
import { group, item, itemStill } from "./motion";

/**
 * The hero enters as one cascade rather than five separate arrivals. The demo
 * panel comes last, so the eye lands on the sentence before the interface.
 */
export function Hero() {
  const reduce = useReducedMotion();
  const child = reduce ? itemStill : item;

  return (
    <section className="mk-hero">
      <motion.div
        className="mk-container mk-hero-grid"
        variants={group}
        initial="hidden"
        animate="shown"
      >
        <div className="mk-hero-copy">
          <motion.span className="mk-chip" variants={child}>
            <span className="mk-chip-dot" />
            Correlation engine with written explanations
          </motion.span>

          <motion.h1 className="mk-display-xl" variants={child}>
            Engineering &amp; business operations intelligence
          </motion.h1>

          <motion.p className="mk-body-lg" variants={child}>
            OpsPulse works out which of your signals moved together, then writes
            the incident in a paragraph anyone can act on.
          </motion.p>

          <motion.div className="mk-hero-actions" variants={child}>
            <Link href="/dashboard" className="mk-btn mk-btn-primary">
              Get started
            </Link>
            <Link href="#how-it-works" className="mk-btn mk-btn-secondary">
              See how it works
            </Link>
          </motion.div>
        </div>

        <motion.div variants={child}>
          <LiveDemo />
        </motion.div>
      </motion.div>
    </section>
  );
}
