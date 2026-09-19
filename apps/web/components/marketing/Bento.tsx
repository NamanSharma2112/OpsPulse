"use client";

import { motion, useInView, useReducedMotion } from "motion/react";
import { useRef } from "react";
import { group, inViewOnce, item, itemStill } from "./motion";

/*
 * Two tiles that reveal once on scroll. Each carries a small visual that
 * performs the behaviour it describes rather than an icon that represents it.
 */

export function BentoDetection() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, inViewOnce);
  const reduce = useReducedMotion();

  // The last four buckets are the anomaly the detector flags.
  const bars = [38, 44, 40, 47, 42, 45, 41, 46, 78, 92, 97, 88];

  return (
    <motion.div
      className="mk-bento"
      ref={ref}
      variants={reduce ? itemStill : item}
      initial="hidden"
      animate={inView ? "shown" : "hidden"}
    >
      <h3 className="mk-card-title">Detect what actually moved</h3>
      <p className="mk-body-sm">
        Every metric gets a baseline from its own history, not a threshold
        somebody guessed in 2019.
      </p>

      <div className="mk-bento-visual">
        <div className="mk-bars">
          {bars.map((height, i) => (
            <motion.i
              key={i}
              data-flag={height > 60}
              style={{ height: `${height}%` }}
              initial={reduce ? false : { transform: "scaleY(0)" }}
              animate={inView ? { transform: "scaleY(1)" } : undefined}
              transition={{
                duration: 0.4,
                ease: [0.23, 1, 0.32, 1],
                delay: 0.1 + i * 0.03,
              }}
            />
          ))}
        </div>
      </div>
    </motion.div>
  );
}

export function BentoCorrelation() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, inViewOnce);
  const reduce = useReducedMotion();

  const rows = [
    { name: "Deploy 9f3c1ab", value: 0.97 },
    { name: "DB pool wait", value: 0.94 },
    { name: "Cart completion", value: 0.89 },
    { name: "CDN cache hit", value: 0.11 },
  ];

  return (
    <motion.div
      className="mk-bento"
      ref={ref}
      variants={reduce ? itemStill : item}
      initial="hidden"
      animate={inView ? "shown" : "hidden"}
    >
      <h3 className="mk-card-title">Connect it to a cause</h3>
      <p className="mk-body-sm">
        Engineering and business signals share one timeline, so a latency spike
        and a revenue dip stop being two investigations.
      </p>

      <div className="mk-bento-visual">
        <div className="mk-corr">
          {rows.map((row, i) => (
            <div className="mk-corr-row" key={row.name}>
              <span className="mk-corr-name">{row.name}</span>
              <span className="mk-corr-bar">
                <motion.span
                  initial={reduce ? false : { transform: "scaleX(0)" }}
                  animate={inView ? { transform: "scaleX(1)" } : undefined}
                  transition={{
                    duration: 0.5,
                    ease: [0.23, 1, 0.32, 1],
                    delay: 0.1 + i * 0.06,
                  }}
                  style={{ width: `${Math.round(row.value * 100)}%` }}
                />
              </span>
              <span className="mk-corr-value">{row.value.toFixed(2)}</span>
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  );
}

/** Wraps the pair so they cascade rather than arrive together. */
export function BentoColumn() {
  return (
    <motion.div className="mk-bento-col" variants={group} initial="hidden" animate="shown">
      <BentoDetection />
      <BentoCorrelation />
    </motion.div>
  );
}
