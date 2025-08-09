import type { Metadata } from "next";
import "./style.css";

export const metadata: Metadata = {
  title: "Social Dilemma",
  description: "A social network built with Next.js and Go.",
};

import { AuthProvider } from "../context/AuthContext";
import Navbar from "./components/Navbar";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <Navbar />
          <main className="main-content">{children}</main>
        </AuthProvider>
      </body>
    </html>
  );
}
