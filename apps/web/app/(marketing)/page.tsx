import Link from "next/link";
import { BentoCorrelation, BentoDetection } from "@/components/marketing/Bento";
import { Integrations } from "@/components/marketing/Integrations";
import { LiveDemo } from "@/components/marketing/LiveDemo";
import { Pipeline } from "@/components/marketing/Pipeline";

export default function LandingPage() {
  return (
    <>
      {/* Hero: copy left, the working product mockup right. */}
      <section className="mk-hero">
        <div className="mk-container mk-hero-grid">
          <div className="mk-hero-copy">
            <span className="mk-chip">
              <span className="mk-chip-dot" />
              Correlation engine, now with written explanations
            </span>
            <h1 className="mk-display-xl" style={{ marginTop: 20 }}>
              Engineering &amp; business operations intelligence
            </h1>
            <p className="mk-body-lg">
              OpsPulse watches the systems you already run, works out which
              signals moved together, and writes the incident in a paragraph
              your on-call engineer and your finance lead can both act on.
            </p>
            <div className="mk-hero-actions">
              <Link href="/dashboard" className="mk-btn mk-btn-primary">
                Get started
              </Link>
              <Link href="#how-it-works" className="mk-btn mk-btn-secondary">
                See how it works
              </Link>
            </div>
            <div className="mk-hero-note">
              <span>Connect a repo in two minutes</span>
              <span>·</span>
              <span>Self-hostable</span>
              <span>·</span>
              <span>No agent to install</span>
            </div>
          </div>

          <div className="mk-hero-demo">
            <LiveDemo />
          </div>
        </div>
      </section>

      {/* The integrations strip. */}
      <Integrations />

      {/* Bento tiles beside the pipeline, per the wireframe. */}
      <section className="mk-section" id="how-it-works">
        <div className="mk-container">
          <div className="mk-section-head">
            <p className="mk-eyebrow">How it works</p>
            <h2 className="mk-display-lg">
              From a number nobody noticed to an incident somebody can fix
            </h2>
            <p className="mk-body-lg">
              Most monitoring tells you a metric crossed a line. OpsPulse tells
              you what changed, what it correlates with, and what it cost —
              in that order.
            </p>
          </div>

          <div className="mk-bento-grid">
            <div className="mk-bento-col">
              <BentoDetection />
              <BentoCorrelation />
            </div>

            <div id="pipeline">
              <Pipeline />
            </div>
          </div>
        </div>
      </section>

      {/* Closing CTA. */}
      <section className="mk-section-tight">
        <div className="mk-container">
          <div className="mk-cta">
            <div className="mk-cta-copy">
              <h2 className="mk-display-md">Point it at one repository</h2>
              <p className="mk-body">
                Start with a single service. OpsPulse keeps every raw event, so
                the metrics you decide to care about next month can be
                backfilled from the history you are already collecting today.
              </p>
            </div>
            <div className="mk-cta-actions">
              <Link href="/dashboard" className="mk-btn mk-btn-primary">
                Open the dashboard
              </Link>
              <Link href="#how-it-works" className="mk-btn mk-btn-secondary">
                Read the pipeline
              </Link>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
