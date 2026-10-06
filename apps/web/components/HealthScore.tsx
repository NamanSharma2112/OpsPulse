"use client";

import { animate, motion, useMotionValue, useReducedMotion, useTransform } from "motion/react";
import { useEffect } from "react";
import type { EngineeringHealth, HealthPillar } from "@opspulse/types";

/*
 * The M3 score: one headline number over three pillars.
 *
 * The ring and the number count up together on first render. This is the one
 * screen in the dashboard with an entrance animation, and it earns it — the
 * number is the page, it is read once per visit, and watching it settle tells
 * you it was computed rather than cached. Everything else in the dashboard is
 * seen dozens of times a day and stays still.
 */

const RADIUS = 52;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

export function HealthScore({ health }: { health: EngineeringHealth }) {
  return (
    <section className="score">
      <div className="score-head">
        <Gauge score={health.score} />
        <div className="score-copy">
          <p className="ui-eyebrow">Engineering health · {health.window}</p>
          <p className="score-summary">{health.summary}</p>
        </div>
      </div>

      <div className="score-pillars">
        {health.pillars.map((pillar) => (
          <Pillar key={pillar.name} pillar={pillar} />
        ))}
      </div>
    </section>
  );
}

function Gauge({ score }: { score: number | null }) {
  const reduce = useReducedMotion();

  // Starts settled so the server HTML carries the real number and hydration
  // matches; the effect resets and counts up only once the client is running.
  const value = useMotionValue(score ?? 0);
  const text = useTransform(value, (v) => String(Math.round(v)));
  const offset = useTransform(value, (v) => CIRCUMFERENCE * (1 - v / 100));

  useEffect(() => {
    if (score === null || reduce) return;
    value.set(0);
    const controls = animate(value, score, {
      duration: 0.9,
      // Modal's out-quart: decelerates hard without overshooting, so the
      // number settles like a measurement rather than springing into place.
      ease: [0.165, 0.84, 0.44, 1],
    });
    return () => controls.stop();
  }, [score, reduce, value]);

  return (
    <div className="score-gauge" data-band={band(score)}>
      <svg viewBox="0 0 120 120" aria-hidden="true">
        <circle className="score-track" cx="60" cy="60" r={RADIUS} />
        <motion.circle
          className="score-arc"
          cx="60"
          cy="60"
          r={RADIUS}
          strokeDasharray={CIRCUMFERENCE}
          style={{ strokeDashoffset: score === null ? CIRCUMFERENCE : offset }}
        />
      </svg>
      <div className="score-number">
        {score === null ? <span className="score-dash">—</span> : <motion.span>{text}</motion.span>}
      </div>
    </div>
  );
}

function Pillar({ pillar }: { pillar: HealthPillar }) {
  return (
    <div className="score-pillar" data-band={band(pillar.score)}>
      <div className="score-pillar-head">
        <span className="score-pillar-name">{pillar.name}</span>
        <span className="score-pillar-score">
          {pillar.score === null ? "—" : pillar.score}
        </span>
      </div>

      <div className="score-signals">
        {pillar.signals.map((signal) => (
          <div className="score-signal" key={signal.name} data-known={signal.known}>
            <span className="score-signal-name">{signal.name}</span>
            <span className="score-signal-value">
              {signal.known ? formatSignal(signal.value, signal.unit) : "no data"}
            </span>
            <span className="score-signal-target">target {signal.target}</span>
            <span className="score-signal-bar">
              <i style={{ width: `${signal.score ?? 0}%` }} data-band={band(signal.score)} />
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function formatSignal(value: number | null, unit: string): string {
  if (value === null) return "—";
  switch (unit) {
    case "%":
      return `${Math.round(value)}%`;
    case "/day":
      return `${value.toFixed(2)}/day`;
    case "h":
      return `${value.toFixed(1)}h`;
    default:
      return String(Math.round(value));
  }
}

/** Colour bands. Unknown is its own band, not a bad one. */
function band(score: number | null): string {
  if (score === null) return "unknown";
  if (score >= 85) return "good";
  if (score >= 70) return "ok";
  if (score >= 50) return "warn";
  return "bad";
}
