"use client";

import { motion, useInView, useReducedMotion } from "motion/react";
import { useRef } from "react";
import { EASE_OUT, STAGGER, inViewOnce } from "./motion";

/*
 * The six stages between a raw number and something a person can act on.
 *
 * The numbers carry the sequence, so there are no connector arrows between
 * the steps — the stagger does that work instead. Fin Orange appears only on
 * the AI stage, which is the accent's documented use; the last stage inverts
 * to charcoal because it is the output the page argues for.
 */

type Step = {
  title: string;
  note: string;
  accent?: "ai" | "output";
};

const STEPS: Step[] = [
  { title: "Raw Metrics", note: "Deploys, latency, errors, revenue" },
  { title: "Feature Extraction", note: "Rates, deltas, rolling baselines" },
  { title: "Rule / Statistical Engine", note: "Known failures by rule, the rest by statistics" },
  { title: "Correlation Detection", note: "What moved together, ranked" },
  { title: "AI Explanation", note: "Ranked evidence becomes a written cause", accent: "ai" },
  { title: "Human-readable Incident", note: "One paragraph anyone can act on", accent: "output" },
];

export function Pipeline() {
  const ref = useRef<HTMLOListElement>(null);
  const inView = useInView(ref, inViewOnce);
  const reduce = useReducedMotion();

  return (
    <ol className="mk-pipeline" ref={ref}>
      {STEPS.map((step, i) => (
        <motion.li
          key={step.title}
          className="mk-step"
          data-accent={step.accent}
          initial={reduce ? { opacity: 0 } : { opacity: 0, transform: "translateY(10px)" }}
          animate={inView ? { opacity: 1, transform: "translateY(0px)" } : undefined}
          transition={{ duration: 0.4, ease: EASE_OUT, delay: i * STAGGER }}
        >
          <span className="mk-step-index">{i + 1}</span>
          <span className="mk-step-body">
            <span className="mk-step-title">{step.title}</span>
            <span className="mk-step-note">{step.note}</span>
          </span>
        </motion.li>
      ))}
    </ol>
  );
}
