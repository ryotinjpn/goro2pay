import type { Metadata } from "next";
import { AppProviders } from "./providers";

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
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
