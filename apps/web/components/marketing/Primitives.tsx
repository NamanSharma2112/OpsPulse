"use client";

import { motion, useInView, useReducedMotion } from "motion/react";
import { useRef } from "react";
import { group, inViewOnce, item, itemStill } from "./motion";

/*
 * Three numbered blocks, the way DESIGN.md numbers a section: zero-padded
 * index in lime, then the claim.
 *
 * These are the three things the platform is, as opposed to the pipeline
 * below, which is what it does to one signal.
 */

const PRIMITIVES = [
  {
    index: "01",
    title: "Append-only event log",
    body: "Every webhook is stored raw and never mutated. Metrics are projections over that log, so a metric you invent next month can be backfilled from the history you are collecting today.",
  },
  {
    index: "02",
    title: "One timeline for engineering and business",
    body: "Deploys, latency, errors and revenue land on the same clock. A latency spike and a revenue dip stop being two separate investigations with two separate owners.",
  },
  {
    index: "03",
    title: "Explanations, not dashboards",
    body: "Ranked correlations become a written incident: what moved, when it moved, what shipped just before, and what it cost. One paragraph, not nine charts.",
  },
];

export function Primitives() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, inViewOnce);
  const reduce = useReducedMotion();

  return (
    <motion.div
      className="mk-primitives"
      ref={ref}
      variants={group}
      initial="hidden"
      animate={inView ? "shown" : "hidden"}
    >
      {PRIMITIVES.map((primitive) => (
        <motion.div
          className="mk-primitive"
          key={primitive.index}
          variants={reduce ? itemStill : item}
        >
          <span className="mk-primitive-index">{primitive.index}</span>
          <h3>{primitive.title}</h3>
          <p>{primitive.body}</p>
        </motion.div>
      ))}
    </motion.div>
  );
}
