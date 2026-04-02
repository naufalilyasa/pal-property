import type { Metadata } from "next";
import { IBM_Plex_Mono, Manrope } from "next/font/google";

import { Providers } from "@/app/providers";
import { BotChat } from "@/features/chat/components/bot-chat";

import "./globals.css";

const manrope = Manrope({
  variable: "--font-sans",
  subsets: ["latin"],
});

const ibmPlexMono = IBM_Plex_Mono({
  variable: "--font-mono",
  weight: ["400", "500"],
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "PAL Property Seller",
  description: "Seller workspace foundation for managing PAL Property listings.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${manrope.variable} ${ibmPlexMono.variable} antialiased`}>
        <Providers>{children}</Providers>
        <BotChat />
      </body>
    </html>
  );
}
