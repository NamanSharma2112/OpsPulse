import Link from "next/link";
import { BentoColumn } from "@/components/marketing/Bento";
import { Hero } from "@/components/marketing/Hero";
import { Integrations } from "@/components/marketing/Integrations";
import { Pipeline } from "@/components/marketing/Pipeline";
import { Reveal } from "@/components/marketing/Reveal";

export default function LandingPage() {
  return (
    <>
      <Hero />
      <Integrations />

      <section className="mk-section" id="how-it-works">
        <div className="mk-container">
          <Reveal className="mk-section-head">
            <p className="mk-eyebrow">How it works</p>
            <h2 className="mk-display-lg">
              From a number nobody noticed to an incident somebody can fix
            </h2>
          </Reveal>

          <div className="mk-bento-grid">
            <BentoColumn />
            <div id="pipeline">
              <Pipeline />
            </div>
          </div>
        </div>
      </section>

      <section className="mk-section-tight">
        <div className="mk-container">
          <Reveal className="mk-cta">
            <div className="mk-cta-copy">
              <h2 className="mk-display-md">Point it at one repository</h2>
              <p className="mk-body">
                OpsPulse keeps every raw event, so the metrics you care about
                next month can be backfilled from the history you collect today.
              </p>
            </div>
            <div className="mk-cta-actions">
              <Link href="/dashboard" className="mk-btn mk-btn-primary">
                Open the dashboard
              </Link>
            </div>
          </Reveal>
        </div>
      </section>
    </>
  );
}
