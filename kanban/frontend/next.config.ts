import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emits a self-contained server bundle in .next/standalone, which is what the
  // runtime stage of the Dockerfile copies. Without this the image has no
  // server.js to run.
  output: "standalone",

  // NEXT_PUBLIC_* values are inlined at build time, so the API base URL must be
  // known when `next build` runs, not just at container start.
  env: {
    NEXT_PUBLIC_API_BASE_URL:
      process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1",
  },
};

export default nextConfig;
