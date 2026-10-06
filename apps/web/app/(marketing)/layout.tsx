import { Footer } from "@/components/marketing/Footer";
import { Nav } from "@/components/marketing/Nav";
import "../ui.css";
import "./marketing.css";

/** Marketing shell: graphite ground, sticky nav, dense footer. */
export default function MarketingLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <div className="ui">
      <Nav />
      <main>{children}</main>
      <Footer />
    </div>
  );
}
