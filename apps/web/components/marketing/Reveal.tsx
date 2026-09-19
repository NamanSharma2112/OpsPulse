"use client";

import { motion, useInView, useReducedMotion } from "motion/react";
import { useRef } from "react";
import { EASE_OUT, inViewOnce } from "./motion";

/**
 * Reveals its children once, the first time they scroll into view. Used for
 * section headers and the closing panel — marketing surfaces only, never
 * functional UI somebody visits daily.
 */
export function Reveal({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, inViewOnce);
  const reduce = useReducedMotion();

  return (
    <motion.div
      ref={ref}
      className={className}
      initial={reduce ? { opacity: 0 } : { opacity: 0, transform: "translateY(12px)" }}
      animate={inView ? { opacity: 1, transform: "translateY(0px)" } : undefined}
      transition={{ duration: 0.45, ease: EASE_OUT }}
    >
      {children}
    </motion.div>
  );
}
