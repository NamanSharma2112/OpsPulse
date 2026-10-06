"use client";

import { animate, motion, useInView, useMotionValue, useReducedMotion, useTransform } from "motion/react";
import { useEffect, useRef } from "react";
import { EASE_QUART, STAGGER } from "./motion";

/*
 * The band of numbers under the marquee. Each value counts up once, when the
 * band is reached.
 *
 * Counting is the point: these are measurements, and a measurement that
 * settles reads as something the system just computed. It is also the one
 * place a number may move — DESIGN.md requires tabular figures precisely so
 * that it can, without the row reflowing underneath it.
 */

type Stat = {
  /** Where the count lands. */
  to: number;
  /** Decimal places to hold while counting, so the width never changes. */
  dp?: number;
  prefix?: string;
  /** Rendered in lime after the number — a unit, not part of the value. */
  suffix?: string;
  label: string;
};

const STATS: Stat[] = [
  { to: 7, label: "GitHub event types ingested" },
  { to: 100, suffix: "%", label: "Events kept, append-only" },
  { to: 4.2, dp: 1, suffix: "h", label: "Median PR cycle time" },
  { to: 1, prefix: "<", suffix: "s", label: "Webhook to dashboard" },
];

/*
 * Fires 200px before the band enters the viewport. That margin is load-bearing
 * rather than cosmetic: each counter is rendered at its settled value so the
 * server HTML is correct, and resetting it to zero to count up has to happen
 * off-screen or the reset itself is visible as a flicker.
 */
const AHEAD = { once: true, margin: "200px" } as const;

export function Stats() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, AHEAD);

  return (
    <div className="ui-container">
      <div className="mk-stats" ref={ref}>
        {STATS.map((stat, i) => (
          <div className="mk-stat" key={stat.label}>
            <div className="mk-stat-value">
              {stat.prefix}
              <Counter stat={stat} run={inView} delay={i * STAGGER} />
              {stat.suffix ? <b>{stat.suffix}</b> : null}
            </div>
            <div className="mk-stat-label">{stat.label}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

function Counter({ stat, run, delay }: { stat: Stat; run: boolean; delay: number }) {
  const reduce = useReducedMotion();
  const dp = stat.dp ?? 0;

  /*
   * The value starts settled, not at zero. useReducedMotion() resolves to
   * null on the server and to the real preference on the client, so anything
   * branching on it at render time hydrates mismatched — and a counter that
   * renders "0" on the server is also the wrong number for anyone whose
   * JavaScript never arrives.
   *
   * It lives in a MotionValue, so the 700ms of interpolation updates the text
   * node directly and never re-renders React.
   */
  const value = useMotionValue(stat.to);
  const text = useTransform(value, (v) => v.toFixed(dp));

  useEffect(() => {
    if (!run || reduce) return;
    value.set(0);
    const controls = animate(value, stat.to, {
      duration: 0.7,
      delay,
      ease: EASE_QUART,
    });
    return () => controls.stop();
  }, [run, reduce, value, stat.to, delay]);

  return <motion.span>{text}</motion.span>;
}
