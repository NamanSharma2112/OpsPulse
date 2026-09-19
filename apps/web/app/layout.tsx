import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";

/*
 * Saans and SaansMono are proprietary. DESIGN.md names Inter at weight 500
 * as the closest free substitute, with JetBrains Mono for the mono role.
 */
const sans = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "OpsPulse — Engineering & Business Operations Intelligence",
  description:
    "OpsPulse turns raw engineering and business signals into human-readable incidents: correlation detection, statistical rules, and an explanation you can act on.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={`${sans.variable} ${mono.variable}`}>
      <body>{children}</body>
    </html>
  );
}
