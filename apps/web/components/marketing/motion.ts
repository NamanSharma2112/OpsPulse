/**
 * Shared motion values. Everything animated on the marketing page draws its
 * curve and duration from here, so the page moves like one system rather than
 * a dozen independent decisions.
 *
 * These mirror the CSS custom properties in app/globals.css — CSS handles
 * hover and press, Motion handles entrances, exits and layout.
 */

/** Strong ease-out. Entrances and exits. */
export const EASE_OUT = [0.23, 1, 0.32, 1] as const;

/** Strong ease-in-out. Movement between two on-screen positions. */
export const EASE_IN_OUT = [0.77, 0, 0.175, 1] as const;

/** Entrance for a single element. */
export const enter = {
  duration: 0.42,
  ease: EASE_OUT,
};

/** The sliding tab indicator: a real position change, so ease-in-out. */
export const indicator = {
  duration: 0.24,
  ease: EASE_IN_OUT,
};

/** Crossfading one panel of content for another. */
export const crossfade = {
  duration: 0.2,
  ease: EASE_OUT,
};

/**
 * A group entrance. 50ms lands inside the 30–80ms band where a stagger reads
 * as one cascade rather than a queue.
 */
export const STAGGER = 0.05;

/**
 * Variants for a staggered group. The parent orchestrates; children only
 * describe their own two states.
 *
 * `transform` is written as a full string because Motion's `x`/`y`/`scale`
 * shorthands are not hardware accelerated and drop frames under load.
 */
export const group = {
  hidden: {},
  shown: {
    transition: { staggerChildren: STAGGER, delayChildren: 0.04 },
  },
};

export const item = {
  hidden: { opacity: 0, transform: "translateY(8px)" },
  shown: { opacity: 1, transform: "translateY(0px)", transition: enter },
};

/**
 * The reduced-motion counterpart: the same states, minus the movement. The
 * element still arrives, it just doesn't travel.
 */
export const itemStill = {
  hidden: { opacity: 0, transform: "translateY(0px)" },
  shown: { opacity: 1, transform: "translateY(0px)", transition: { duration: 0.2 } },
};

/** Shared options for scroll reveals: fire once, slightly before fully in view. */
export const inViewOnce = { once: true, margin: "-80px" } as const;
