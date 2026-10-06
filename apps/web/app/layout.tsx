import type { Metadata } from "next";
import { DM_Mono, Inter } from "next/font/google";
import "./globals.css";

/*
 * ABC Diatype is licensed, so DESIGN.md names Inter as the substitute — with
 * `cv11` enabled in globals.css, which swaps in the single-storey alternates
 * that make Inter read as a neutral grotesque rather than as Inter.
 *
 * DM Mono is the real thing: it is Modal's mono and it is freely available,
 * so it ships as specified rather than as a stand-in.
 */
const sans = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const mono = DM_Mono({
  subsets: ["latin"],
  weight: ["300", "400", "500"],
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
