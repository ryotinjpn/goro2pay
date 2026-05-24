import type { Metadata } from "next";
import { AppProviders } from "./providers";
// Unit C: 全画面共通のトースト Host (P-FE-TOAST-02)
import { ToastHost } from "@/components/order/ToastHost";

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
    <html lang="ja">
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
