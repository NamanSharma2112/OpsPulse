"use client";

import { useEffect, useMemo, useState } from "react";

/*
 * The hero's product mockup: a working miniature of what OpsPulse does.
 * Pick a scenario and the panel walks the same path the real pipeline takes —
 * raw signals, a correlation, then a written incident.
 *
 * All series are fixed arrays rather than random, so the server and client
 * render identically.
 */

type Signal = {
  name: string;
  unit: string;
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
    tab: "Checkout latency",
    window: "14:02 – 14:26 UTC",
    severity: "Major",
    headline: "Checkout p95 tripled after the 14:02 deploy",
    explanation:
      "Latency, connection-pool saturation and cart abandonment all turned at 14:04, two minutes after deploy 9f3c1ab reached production. The pool ceiling is the only upstream change in the window.",
    impact: "≈ 1,240 carts affected · $18.4k at risk",
    signals: [
      {
        name: "Checkout p95",
        unit: "ms",
        series: [180, 176, 184, 179, 188, 182, 190, 420, 610, 720, 760, 740],
        delta: "+311%",
        correlation: 0.97,
      },
      {
        name: "DB pool wait",
        unit: "ms",
        series: [4, 5, 4, 6, 5, 5, 7, 96, 180, 240, 268, 255],
        delta: "+4,900%",
        correlation: 0.94,
      },
      {
        name: "Cart completion",
        unit: "%",
        series: [72, 71, 73, 72, 74, 72, 71, 58, 44, 36, 33, 34],
        delta: "−53%",
        correlation: 0.89,
      },
    ],
  },
  {
    id: "payments",
    tab: "Payment failures",
    window: "09:40 – 10:05 UTC",
    severity: "Critical",
    headline: "Card declines spiked on one processor region",
    explanation:
      "Declines rose only for EU-routed traffic while US volume held flat. The processor's regional status feed flipped to degraded at 09:41, which matches the first failed batch exactly.",
    impact: "≈ 3,180 payments declined · $52.1k blocked",
    signals: [
      {
        name: "Decline rate",
        unit: "%",
        series: [2, 2, 3, 2, 2, 3, 2, 18, 34, 41, 39, 37],
        delta: "+1,750%",
        correlation: 0.99,
      },
      {
        name: "Processor RTT",
        unit: "ms",
        series: [120, 118, 124, 119, 122, 121, 130, 340, 520, 580, 566, 540],
        delta: "+372%",
        correlation: 0.91,
      },
      {
        name: "Retry queue",
        unit: "k",
        series: [1, 1, 1, 2, 1, 1, 2, 9, 22, 31, 36, 34],
        delta: "+3,300%",
        correlation: 0.87,
      },
    ],
  },
  {
    id: "regression",
    tab: "Deploy regression",
    window: "22:15 – 22:48 UTC",
    severity: "Minor",
    headline: "Background job throughput halved after a worker release",
    explanation:
      "Queue depth climbed steadily while worker CPU stayed flat — the workers are idle, not saturated. A serialisation change in the 22:15 release is the only correlated deploy.",
    impact: "≈ 42 min of delayed exports",
    signals: [
      {
        name: "Jobs / min",
        unit: "",
        series: [860, 872, 851, 866, 870, 858, 861, 520, 430, 402, 398, 410],
        delta: "−53%",
        correlation: 0.95,
      },
      {
        name: "Queue depth",
        unit: "k",
        series: [2, 2, 3, 2, 3, 2, 3, 12, 24, 33, 40, 44],
        delta: "+1,366%",
        correlation: 0.92,
      },
      {
        name: "Worker CPU",
        unit: "%",
        series: [58, 61, 59, 60, 57, 62, 60, 59, 58, 57, 59, 58],
        delta: "±0%",
        correlation: 0.12,
      },
    ],
  },
];

const STAGES = ["Ingesting signals", "Detecting correlation", "Explaining"] as const;

export function LiveDemo() {
  const [active, setActive] = useState(0);
  const [stage, setStage] = useState<number>(STAGES.length);
  const scenario = SCENARIOS[active]!;

  // Replay the analysis whenever the scenario changes.
  useEffect(() => {
    setStage(0);
    const timers = STAGES.map((_, i) =>
      window.setTimeout(() => setStage(i + 1), 420 * (i + 1)),
    );
    return () => timers.forEach(window.clearTimeout);
  }, [active]);

  const done = stage >= STAGES.length;

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
              {s.tab}
            </button>
          ))}
        </div>

        <div className="demo-signals">
          {scenario.signals.map((signal, i) => (
            <SignalRow
              key={`${scenario.id}-${signal.name}`}
              signal={signal}
              revealed={stage >= 1}
              correlated={stage >= 2}
              delay={i * 70}
            />
          ))}
        </div>

        <div className="demo-stages" aria-live="polite">
          {STAGES.map((label, i) => (
            <span key={label} className="demo-stage" data-state={stageState(stage, i)}>
              <StageIcon done={stage > i} />
              {label}
            </span>
          ))}
        </div>

        <div className="demo-output" data-visible={done}>
          <div className="demo-output-head">
            <span className="demo-sev" data-sev={scenario.severity.toLowerCase()}>
              {scenario.severity}
            </span>
            <span className="demo-window">{scenario.window}</span>
          </div>
          <p className="demo-headline">{scenario.headline}</p>
          <p className="demo-explanation">{scenario.explanation}</p>
          <p className="demo-impact">{scenario.impact}</p>
        </div>
      </div>
    </div>
  );
}

function stageState(stage: number, index: number): "done" | "active" | "idle" {
  if (stage > index) return "done";
  if (stage === index) return "active";
  return "idle";
}

function SignalRow({
  signal,
  revealed,
  correlated,
  delay,
}: {
  signal: Signal;
  revealed: boolean;
  correlated: boolean;
  delay: number;
}) {
  const path = useMemo(() => sparkline(signal.series), [signal.series]);
  // Below 0.5 the signal is noise, and the panel says so rather than
  // quietly dropping it.
  const isCorrelated = signal.correlation >= 0.5;

  return (
    <div
      className="demo-signal"
      data-revealed={revealed}
      data-correlated={correlated && isCorrelated}
      style={{ transitionDelay: `${delay}ms` }}
    >
      <div className="demo-signal-meta">
        <span className="demo-signal-name">{signal.name}</span>
        <span className="demo-signal-delta">{signal.delta}</span>
      </div>
      <svg className="demo-spark" viewBox="0 0 120 32" preserveAspectRatio="none" aria-hidden="true">
        <path d={path} fill="none" stroke="currentColor" strokeWidth="1.6" vectorEffect="non-scaling-stroke" />
      </svg>
      <span className="demo-signal-corr">
        {correlated ? `r ${signal.correlation.toFixed(2)}` : "—"}
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

function StageIcon({ done }: { done: boolean }) {
  return (
    <svg width="12" height="12" viewBox="0 0 12 12" aria-hidden="true">
      {done ? (
        <path
          d="M2.5 6.4 4.8 8.7 9.5 3.6"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.6"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      ) : (
        <circle cx="6" cy="6" r="3" fill="none" stroke="currentColor" strokeWidth="1.5" />
      )}
    </svg>
  );
}
