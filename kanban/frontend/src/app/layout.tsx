import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import type { ReactNode } from "react";

import { ThemeProvider, themeInitScript } from "@/providers/theme-provider";
import { ToastProvider } from "@/providers/toast-provider";

import "./globals.css";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Kanban",
  description: "A focused board for moving deals forward.",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  // The board is a fixed-height app shell; pinch-zoom stays available but
  // double-tap zoom does not fight the drag gestures.
  maximumScale: 5,
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    // `suppressHydrationWarning` is required: the inline script below mutates
    // the class list before React hydrates, so the server HTML is stale by one
    // attribute.
    <html
      lang="en"
      suppressHydrationWarning
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <head>
        {/* Applies the stored/OS theme before first paint — without this a
            dark-mode reload flashes white. */}
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className="h-full">
        <ThemeProvider>
          <ToastProvider>{children}</ToastProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
