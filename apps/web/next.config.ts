import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // The shared contracts package ships TypeScript source, so Next compiles it
  // as part of the app rather than expecting a prebuilt bundle.
  transpilePackages: ["@opspulse/types"],
  // Used by the Docker image to emit a minimal standalone server.
  output: process.env.NEXT_OUTPUT === "standalone" ? "standalone" : undefined,
};

export default nextConfig;
