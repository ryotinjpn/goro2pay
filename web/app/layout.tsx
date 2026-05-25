import type { Metadata } from "next";
import { Cormorant_Garamond, Zen_Old_Mincho, DotGothic16 } from "next/font/google";
import "./globals.css";
import { AppProviders } from "./providers";
// Unit C: 全画面共通のトースト Host (P-FE-TOAST-02)
import { ToastHost } from "@/components/order/ToastHost";

const cormorant = Cormorant_Garamond({
  subsets: ["latin"],
  weight: ["700"],
  style: ["italic"],
  display: "swap",
  variable: "--font-cormorant",
});

const zenMincho = Zen_Old_Mincho({
  subsets: ["latin"],
  weight: ["700", "900"],
  display: "swap",
  variable: "--font-zen-mincho",
});

const dotGothic = DotGothic16({
  subsets: ["latin"],
  weight: ["400"],
  display: "swap",
  variable: "--font-dotgothic",
});

export const metadata: Metadata = {
  title: "ゴロゴロPay",
  description: "めんどくさいを丸投げ",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="ja"
      className={`${cormorant.variable} ${zenMincho.variable} ${dotGothic.variable}`}
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
