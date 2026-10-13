import type { NextConfig } from "next";

/**
 * Content-Security-Policy.
 *
 * The bearer token is kept in localStorage (see src/lib/token.ts), which makes
 * "any script on this page can read it" the thing worth defending against. This
 * policy is that defence.
 *
 * `'unsafe-inline'` is present for scripts because the theme is applied by a
 * small blocking <script> in the document head (src/app/layout.tsx) — without
 * it the page flashes white before hydration. A nonce would be stricter, but
 * Next only issues nonces from middleware, and this app deliberately has no
 * middleware: the session token lives in localStorage, which the Edge runtime
 * cannot read. `script-src-elem` is therefore the part that would change first
 * if that decision is ever revisited.
 *
 * 'unsafe-eval' is required by the React development build only; production
 * does not need it, so it is dropped for production builds.
 */
const isDev = process.env.NODE_ENV !== "production";

const scriptSrc = ["'self'", "'unsafe-inline'", ...(isDev ? ["'unsafe-eval'"] : [])].join(" ");

const csp = [
  "default-src 'self'",
  `script-src ${scriptSrc}`,
  // Google Fonts is loaded for Geist/Geist Mono.
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
  "font-src 'self' https://fonts.gstatic.com data:",
  "img-src 'self' data: blob:",
  // The API is called directly from the browser, hence the explicit origin.
  "connect-src 'self' http://localhost:8080 http://localhost:8081 http://localhost:3000",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self'",
  "object-src 'none'",
].join("; ");

const nextConfig: NextConfig = {
  // Emits a self-contained server bundle in .next/standalone, which is what the
  // runtime stage of the Dockerfile copies. Without this the image has no
  // server.js to run.
  output: "standalone",

  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "Content-Security-Policy", value: csp },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          { key: "X-Frame-Options", value: "DENY" },
          { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
        ],
      },
    ];
  },

  // NEXT_PUBLIC_* values are inlined at build time, so the API base URL must be
  // known when `next build` runs, not just at container start.
  env: {
    NEXT_PUBLIC_API_BASE_URL:
      process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1",
  },
};

export default nextConfig;