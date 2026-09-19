"use client";

import { AnimatePresence, motion, useReducedMotion } from "motion/react";
import { useEffect, useMemo, useState } from "react";
import { crossfade, indicator } from "./motion";

/*
 * The hero's product mockup: a working miniature of what OpsPulse does.
 * Pick a scenario and the panel shows which signals correlate, then writes
 * the incident.
 *
 * All series are fixed arrays rather than random, so the server and client
 * render identically.
 */

type Signal = {
  name: string;
  series: number[];
  delta: string;
  correlation: number;
};

type Scenario = {
  id: string;
  tab: string;
  signals: Signal[];
  window: string;
  severity: string;
  headline: string;
  explanation: string;
  impact: string;
};

const SCENARIOS: Scenario[] = [
  {
    id: "checkout",
    tab: "Checkout",
    window: "14:02 – 14:26 UTC",
    severity: "Major",
    headline: "Checkout p95 tripled after the 14:02 deploy",
    explanation:
      "Latency, pool saturation and cart abandonment all turned at 14:04, two minutes after deploy 9f3c1ab reached production.",
    impact: "≈ 1,240 carts · $18.4k at risk",
    signals: [
      {
        name: "Checkout p95",
        series: [180, 176, 184, 179, 188, 182, 190, 420, 610, 720, 760, 740],
        delta: "+311%",
        correlation: 0.97,
      },
      {
        name: "DB pool wait",
        series: [4, 5, 4, 6, 5, 5, 7, 96, 180, 240, 268, 255],
        delta: "+4,900%",
        correlation: 0.94,
      },
      {
        name: "Cart completion",
        series: [72, 71, 73, 72, 74, 72, 71, 58, 44, 36, 33, 34],
        delta: "−53%",
        correlation: 0.89,
      },
    ],
  },
  {
    id: "payments",
    tab: "Payments",
    window: "09:40 – 10:05 UTC",
    severity: "Critical",
    headline: "Card declines spiked on one processor region",
    explanation:
      "Declines rose only for EU-routed traffic while US volume held flat. The processor's regional feed went degraded at 09:41.",
    impact: "≈ 3,180 payments · $52.1k blocked",
    signals: [
      {
        name: "Decline rate",
        series: [2, 2, 3, 2, 2, 3, 2, 18, 34, 41, 39, 37],
        delta: "+1,750%",
        correlation: 0.99,
      },
      {
        name: "Processor RTT",
        series: [120, 118, 124, 119, 122, 121, 130, 340, 520, 580, 566, 540],
        delta: "+372%",
        correlation: 0.91,
      },
      {
        name: "Retry queue",
        series: [1, 1, 1, 2, 1, 1, 2, 9, 22, 31, 36, 34],
        delta: "+3,300%",
        correlation: 0.87,
      },
    ],
  },
  {
    id: "regression",
    tab: "Deploys",
    window: "22:15 – 22:48 UTC",
    severity: "Minor",
    headline: "Job throughput halved after a worker release",
    explanation:
      "Queue depth climbed while worker CPU stayed flat — the workers are idle, not saturated. One correlated deploy at 22:15.",
    impact: "≈ 42 min of delayed exports",
    signals: [
      {
        name: "Jobs / min",
        series: [860, 872, 851, 866, 870, 858, 861, 520, 430, 402, 398, 410],
        delta: "−53%",
        correlation: 0.95,
      },
      {
        name: "Queue depth",
        series: [2, 2, 3, 2, 3, 2, 3, 12, 24, 33, 40, 44],
        delta: "+1,366%",
        correlation: 0.92,
      },
      {
        name: "Worker CPU",
        series: [58, 61, 59, 60, 57, 62, 60, 59, 58, 57, 59, 58],
        delta: "±0%",
        correlation: 0.12,
      },
    ],
  },
];

/** Below this a signal is noise, and the panel says so rather than hiding it. */
const CORRELATION_FLOOR = 0.5;

export function LiveDemo() {
  const [active, setActive] = useState(0);
  const [settled, setSettled] = useState(true);
  const scenario = SCENARIOS[active]!;

  // Correlation resolves a beat after the scenario changes, so the rows are
  // seen as raw signals before they are ranked.
  useEffect(() => {
    setSettled(false);
    const timer = window.setTimeout(() => setSettled(true), 260);
    return () => window.clearTimeout(timer);
  }, [active]);

  return (
    <div className="mk-mockup demo">
      <div className="mk-mockup-bar">
        <div className="mk-mockup-dots">
          <i />
          <i />
          <i />
        </div>
        <span className="mk-mockup-title">OpsPulse — incident analysis</span>
        <span className="demo-live">
          <i />
          live
        </span>
      </div>

      <div className="mk-mockup-body">
        <div className="demo-tabs" role="tablist" aria-label="Example incidents">
          {SCENARIOS.map((s, i) => (
            <button
              key={s.id}
              type="button"
              role="tab"
              aria-selected={i === active}
              className="demo-tab"
              data-active={i === active}
              onClick={() => setActive(i)}
            >
              {/* One element slides between tabs rather than two cross-fading. */}
              {i === active ? (
                <motion.span
                  layoutId="demo-tab-indicator"
                  className="demo-tab-indicator"
                  transition={indicator}
                />
              ) : null}
              <span className="demo-tab-label">{s.tab}</span>
            </button>
          ))}
        </div>

        <div className="demo-signals">
          {scenario.signals.map((signal) => (
            <SignalRow
              key={`${scenario.id}-${signal.name}`}
              signal={signal}
              correlated={settled && signal.correlation >= CORRELATION_FLOOR}
              settled={settled}
            />
          ))}
        </div>

        <div className="demo-output">
          <AnimatePresence mode="wait" initial={false}>
            <motion.div
              key={scenario.id}
              initial={{ opacity: 0, transform: "translateY(4px)" }}
              animate={{ opacity: 1, transform: "translateY(0px)" }}
              exit={{ opacity: 0, transform: "translateY(-4px)" }}
              transition={crossfade}
            >
              <div className="demo-output-head">
                <span className="demo-sev" data-sev={scenario.severity.toLowerCase()}>
                  {scenario.severity}
                </span>
                <span className="demo-window">{scenario.window}</span>
              </div>
              <p className="demo-headline">{scenario.headline}</p>
              <p className="demo-explanation">{scenario.explanation}</p>
              <p className="demo-impact">{scenario.impact}</p>
            </motion.div>
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}

function SignalRow({
  signal,
  correlated,
  settled,
}: {
  signal: Signal;
  correlated: boolean;
  settled: boolean;
}) {
  const reduce = useReducedMotion();
  const path = useMemo(() => sparkline(signal.series), [signal.series]);

  return (
    <div className="demo-signal" data-correlated={correlated}>
      <div className="demo-signal-meta">
        <span className="demo-signal-name">{signal.name}</span>
        <span className="demo-signal-delta">{signal.delta}</span>
      </div>

      <svg className="demo-spark" viewBox="0 0 120 32" preserveAspectRatio="none" aria-hidden="true">
        {/* The trace draws itself in, so the row reads as a signal arriving. */}
        <motion.path
          d={path}
          fill="none"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          vectorEffect="non-scaling-stroke"
          initial={reduce ? false : { pathLength: 0 }}
          animate={{ pathLength: 1 }}
          transition={{ duration: 0.5, ease: [0.23, 1, 0.32, 1] }}
        />
      </svg>

      <span className="demo-signal-corr">
        {settled ? `r ${signal.correlation.toFixed(2)}` : "—"}
      </span>
    </div>
  );
}

/** Maps a series onto a 120×32 viewBox as an SVG path. */
function sparkline(series: number[]): string {
  const min = Math.min(...series);
  const max = Math.max(...series);
  const span = max - min || 1;
  const stepX = 120 / (series.length - 1);

  return series
    .map((value, i) => {
      const x = (i * stepX).toFixed(1);
      // 3px padding top and bottom keeps the stroke inside the box.
      const y = (29 - ((value - min) / span) * 26).toFixed(1);
      return `${i === 0 ? "M" : "L"}${x} ${y}`;
    })
    .join(" ");
}
