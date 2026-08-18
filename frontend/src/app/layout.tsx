import type { Metadata } from "next";
import { Inter, Lora, Lexend } from "next/font/google";
import { Providers } from "@/components/providers";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

const lora = Lora({
  variable: "--font-lora",
  subsets: ["latin"],
  display: "swap",
});

const lexend = Lexend({
  variable: "--font-lexend",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "GCET Faculty Portal",
  description: "Official portal for GCET Faculty members to manage attendance, timetable, and more.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full">
      <body
        className={`${inter.variable} ${lora.variable} ${lexend.variable} antialiased h-full`}
      >
        <Providers>
          <div className="min-h-full bg-background flex flex-col">
            {children}
          </div>
        </Providers>
      </body>
    </html>
  );
}
