// app/layout.tsx
import { ReactNode } from 'react'
import { AuthProvider } from '../components/AuthContext'
import { SocketProvider } from './context/socketContext'  // ✅ Import this
import NavBar from '../components/NavBar'

export const metadata = {
  title: 'My App',
  description: 'A simple auth app',
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-green-50 min-h-screen flex flex-col items-center">
        <SocketProvider> {/* ✅ Wrap everything in the SocketProvider */}
          <AuthProvider>
            <NavBar />
            <hr className="w-full border-green-300" />
            <main className="flex-grow w-full max-w-3xl p-6">
              {children}
            </main>
          </AuthProvider>
        </SocketProvider>
      </body>
    </html>
  )
}
