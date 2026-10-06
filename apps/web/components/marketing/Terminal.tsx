"use client";

import { useEffect, useRef, useState } from "react";
import { useInView, useReducedMotion } from "motion/react";

/*
 * The hero's left panel: the three commands that actually connect a
 * repository, typed out.
 *
 * It types rather than fades because the point of the panel is that setup is
 * three lines long — you can't read that from a block that appears whole.
 *
 * Driven by one interval on a character index rather than by Motion: there is
 * nothing to interpolate, and a single timer is cheaper than a per-character
 * animation. It runs once, when the panel scrolls into view, and never loops.
 */

type Line =
  | { kind: "cmd"; text: string }
  | { kind: "out"; text: string }
  | { kind: "ok"; text: string };

const SCRIPT: Line[] = [
  { kind: "cmd", text: "opspulse connect --repo acme/storefront" },
  { kind: "out", text: "webhook installed · 7 event types" },
  { kind: "cmd", text: "opspulse watch" },
  { kind: "out", text: "push        9f3c1ab  main" },
  { kind: "out", text: "deployment  production  success" },
  { kind: "ok", text: "pull_request #482 merged in 4h 12m" },
  { kind: "cmd", text: "opspulse explain --last" },
  { kind: "ok", text: "checkout p95 tripled after 14:02 deploy" },
];

/** Commands type at human speed; output arrives whole, the way it would. */
const CHAR_MS = 22;
const LINE_PAUSE_MS = 260;

export function Terminal() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true, margin: "-60px" });
  const reduce = useReducedMotion();

  // How many lines are complete, and how far into the line after them.
  const [done, setDone] = useState(0);
  const [chars, setChars] = useState(0);

  const skip = reduce || !inView;

  useEffect(() => {
    if (skip) return;
    if (done >= SCRIPT.length) return;

    const line = SCRIPT[done]!;

    // Output lines are not typed — they are printed after a beat.
    if (line.kind !== "cmd") {
      const timer = window.setTimeout(() => setDone((d) => d + 1), LINE_PAUSE_MS);
      return () => window.clearTimeout(timer);
    }

    if (chars < line.text.length) {
      const timer = window.setTimeout(() => setChars((c) => c + 1), CHAR_MS);
      return () => window.clearTimeout(timer);
    }

    const timer = window.setTimeout(() => {
      setDone((d) => d + 1);
      setChars(0);
    }, LINE_PAUSE_MS);
    return () => window.clearTimeout(timer);
  }, [skip, done, chars]);

  // With reduced motion, or before the panel is reached, the whole transcript
  // is simply present. Nothing is hidden from anyone.
  const visible = skip ? SCRIPT.length : done;
  const typing = !skip && done < SCRIPT.length ? SCRIPT[done] : undefined;

  return (
    <div className="ui-mockup">
      <div className="ui-mockup-bar">
        <div className="ui-mockup-dots">
          <i />
          <i />
          <i />
        </div>
        <span className="ui-mockup-title">opspulse — setup</span>
      </div>

      <div className="ui-mockup-body">
        {/* The transcript is a log, so it is announced as one rather than
            read out character by character as it types. */}
        <div className="mk-term" ref={ref} role="log" aria-live="off">
          {SCRIPT.slice(0, visible).map((line, i) => (
            <Row key={i} line={line} text={line.text} />
          ))}

          {typing ? (
            <Row
              line={typing}
              text={typing.kind === "cmd" ? typing.text.slice(0, chars) : typing.text}
              caret
            />
          ) : null}
        </div>
      </div>
    </div>
  );
}

function Row({ line, text, caret }: { line: Line; text: string; caret?: boolean }) {
  if (line.kind === "cmd") {
    return (
      <div className="mk-term-line">
        <span className="mk-term-prompt" aria-hidden="true">
          $
        </span>
        <span className="mk-term-cmd">
          {text}
          {caret ? <span className="mk-term-caret" aria-hidden="true" /> : null}
        </span>
      </div>
    );
  }

  return (
    <div className="mk-term-line">
      <span className={line.kind === "ok" ? "mk-term-out mk-term-ok" : "mk-term-out"}>
        {text}
      </span>
    </div>
  );
}
