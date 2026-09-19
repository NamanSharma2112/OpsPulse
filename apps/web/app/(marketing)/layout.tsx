import { Footer } from "@/components/marketing/Footer";
import { Nav } from "@/components/marketing/Nav";
import "./marketing.css";

/** Marketing shell: cream canvas, sticky nav, dense footer. */
export default function MarketingLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <div className="mk">
      <Nav />
      <main>{children}</main>
      <Footer />
    </div>
  );
}
