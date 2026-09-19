"use client";

import { useReveal } from "./Reveal";

/*
 * Two bento tiles that animate in on scroll. Each carries a small live
 * visual rather than an icon, so the tile shows the behaviour it describes.
 */

export function BentoDetection() {
  const { ref, visible } = useReveal<HTMLDivElement>();

  // The last four buckets are the anomaly the detector flags.
  const bars = [38, 44, 40, 47, 42, 45, 41, 46, 78, 92, 97, 88];

  return (
    <div className="mk-bento" ref={ref} data-visible={visible}>
      <div className="mk-bento-head">
        <span className="mk-bento-icon">
          <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true">
            <path
              d="M1.5 11.5h2.2l1.6-6 2.2 9 1.8-7 1.4 4h3.8"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.4"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
        <h3 className="mk-card-title">Detect what actually moved</h3>
      </div>
      <p className="mk-body-sm">
        Every metric gets a baseline from its own history, not a threshold
        somebody guessed in 2019. OpsPulse flags the deviation and leaves the
        ordinary noise alone.
      </p>
      <div className="mk-bento-visual">
        <div className="mk-bars">
          {bars.map((height, i) => (
            <i
              key={i}
              data-flag={height > 60}
              style={{
                height: `${height}%`,
                transitionDelay: `${i * 45}ms`,
              }}
            />
          ))}
        </div>
        <p className="mk-caption" style={{ marginTop: 10 }}>
          4 of 12 buckets outside the expected band
        </p>
      </div>
    </div>
  );
}

export function BentoCorrelation() {
  const { ref, visible } = useReveal<HTMLDivElement>();

  const rows = [
    { name: "Deploy 9f3c1ab", value: 0.97 },
    { name: "DB pool wait", value: 0.94 },
    { name: "Cart completion", value: 0.89 },
    { name: "CDN cache hit", value: 0.11 },
  ];

  return (
    <div className="mk-bento" ref={ref} data-visible={visible}>
      <div className="mk-bento-head">
        <span className="mk-bento-icon">
          <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true">
            <path
              d="M6.5 9.5a3 3 0 0 0 4.2 0l2-2a3 3 0 0 0-4.2-4.2l-.8.8M9.5 6.5a3 3 0 0 0-4.2 0l-2 2a3 3 0 0 0 4.2 4.2l.8-.8"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.4"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
        <h3 className="mk-card-title">Connect it to a cause</h3>
      </div>
      <p className="mk-body-sm">
        Engineering signals and business signals share one timeline, so a
        latency spike and a revenue dip stop being two separate investigations.
      </p>
      <div className="mk-bento-visual">
        <div className="mk-corr">
          {rows.map((row, i) => (
            <div className="mk-corr-row" key={row.name}>
              <span className="mk-corr-name">{row.name}</span>
              <span className="mk-corr-bar">
                <span
                  style={{
                    ["--w" as string]: `${Math.round(row.value * 100)}%`,
                    transitionDelay: `${i * 90}ms`,
                  }}
                />
              </span>
              <span className="mk-corr-value">{row.value.toFixed(2)}</span>
            </div>
          ))}
        </div>
        <p className="mk-caption" style={{ marginTop: 10 }}>
          Ranked by correlation against the anomaly window
        </p>
      </div>
    </div>
  );
}
