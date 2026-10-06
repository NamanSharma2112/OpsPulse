import Link from "next/link";
import { BentoColumn } from "@/components/marketing/Bento";
import { Hero } from "@/components/marketing/Hero";
import { Integrations } from "@/components/marketing/Integrations";
import { Pipeline } from "@/components/marketing/Pipeline";
import { Primitives } from "@/components/marketing/Primitives";
import { Reveal } from "@/components/marketing/Reveal";
import { Stats } from "@/components/marketing/Stats";

export default function LandingPage() {
  return (
    <>
      <Hero />
      <Integrations />
      <Stats />

      <section className="mk-section" id="platform">
        <div className="ui-container">
          <Reveal className="mk-section-head">
            <p className="ui-eyebrow" data-index="01">
              The platform
            </p>
            <h2 className="ui-display-lg">
              Infrastructure built for signals, not for screenshots
            </h2>
          </Reveal>

          <Primitives />
        </div>
      </section>

      <section className="mk-section" id="how-it-works">
        <div className="ui-container">
          <Reveal className="mk-section-head">
            <p className="ui-eyebrow" data-index="02">
              How it works
            </p>
            <h2 className="ui-display-lg">
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
        <div className="ui-container">
          <Reveal className="mk-cta">
            <div className="mk-cta-copy">
              <p className="ui-eyebrow" data-index="03">
                Get started
              </p>
              <h2 className="ui-display-md">Point it at one repository</h2>
              <p className="ui-body">
                OpsPulse keeps every raw event, so the metrics you care about
                next month can be backfilled from the history you collect today.
              </p>
            </div>
            <div className="mk-cta-actions">
              <Link href="/dashboard" className="ui-btn ui-btn-primary">
                Open the dashboard
              </Link>
            </div>
          </Reveal>
        </div>
      </section>
    </>
  );
}
