"use client";

import { useReveal } from "./Reveal";

/*
 * The six stages between a raw number and something a person can act on.
 * Steps stagger in on scroll; the AI stage is the one place Fin Orange
 * appears, and the final stage inverts to charcoal because it is the output
 * the whole page argues for.
 */

type Step = {
  title: string;
  note: string;
  accent?: "ai" | "output";
};

const STEPS: Step[] = [
  {
    title: "Raw Metrics",
    note: "Deploys, latency, errors, paging, revenue — pulled from the systems you already run.",
  },
  {
    title: "Feature Extraction",
    note: "Each series is normalised into rates, deltas and rolling baselines.",
  },
  {
    title: "Rule / Statistical Engine",
    note: "Deterministic rules catch the known failures; statistics catch the rest.",
  },
  {
    title: "Correlation Detection",
    note: "Signals that moved together in the same window are ranked against the anomaly.",
  },
  {
    title: "AI Explanation",
    note: "The ranked evidence — never raw logs — is turned into a written cause.",
    accent: "ai",
  },
  {
    title: "Human-readable Incident",
    note: "One paragraph an on-call engineer and a finance lead can both act on.",
    accent: "output",
  },
];

export function Pipeline() {
  const { ref, visible } = useReveal<HTMLDivElement>(0.15);

  return (
    <div className="mk-pipeline" ref={ref} data-visible={visible}>
      {STEPS.map((step, i) => (
        <div key={step.title}>
          <div
            className="mk-step"
            data-accent={step.accent}
            style={{ transitionDelay: `${i * 90}ms` }}
          >
            <span className="mk-step-index">{i + 1}</span>
            <span>
              <span className="mk-step-title">{step.title}</span>
              <span className="mk-step-note">{step.note}</span>
            </span>
          </div>
          {i < STEPS.length - 1 ? (
            <span className="mk-step-link" aria-hidden="true">
              <svg width="12" height="22" viewBox="0 0 12 22">
                <path
                  d="M6 2v14m0 0 3.2-3.4M6 16l-3.2-3.4"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.3"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </span>
          ) : null}
        </div>
      ))}
    </div>
  );
}
