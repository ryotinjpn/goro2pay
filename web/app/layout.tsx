import type { Metadata } from "next";
import {
  Cormorant_Garamond,
  Zen_Old_Mincho,
  DotGothic16,
  Inter,
} from "next/font/google";

import { AppProviders } from "./providers";
// Unit C: 全画面共通のトースト Host (P-FE-TOAST-02)
import { ToastHost } from "@/components/order/ToastHost";

import "./globals.css";

const fontDisplaySerif = Cormorant_Garamond({
  subsets: ["latin"],
  weight: ["500", "700"],
  style: ["normal", "italic"],
  display: "swap",
  variable: "--font-display-serif",
});

const fontBodyMincho = Zen_Old_Mincho({
  subsets: ["latin"],
  weight: ["500", "700", "900"],
  display: "swap",
  variable: "--font-body-mincho",
});

const fontMonoPixel = DotGothic16({
  subsets: ["latin"],
  weight: ["400"],
  display: "swap",
  variable: "--font-mono-pixel",
});

const fontBodySans = Inter({
  subsets: ["latin"],
  weight: ["600", "800"],
  display: "swap",
  variable: "--font-body-sans",
});

export const metadata: Metadata = {
  title: "ゴロゴロPay",
  description: "面倒は、こちらで引き受ける。",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="ja"
      className={[
        fontDisplaySerif.variable,
        fontBodyMincho.variable,
        fontMonoPixel.variable,
        fontBodySans.variable,
      ].join(" ")}
    >
      <body>
        <AppProviders>
          {children}
          {/* Jotai store 配下に配置する必要があるため AppProviders 内側 */}
          <ToastHost />
        </AppProviders>
      </body>
    </html>
  );
}
