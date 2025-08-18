import type { Metadata } from "next";
import "./style.css";
import { AuthProvider } from "../context/AuthContext";
import { PopupProvider } from "../context/PopupContext";
import ClientLayout from "./components/ClientLayout";

export const metadata: Metadata = {
  title: "Social Dilemma",
  description: "A social network built with Next.js and Go.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <PopupProvider>
            <ClientLayout>{children}</ClientLayout>
          </PopupProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
