'use client'

import { createContext, useContext } from 'react'
import useSharedSocket from '../hooks/useSharedSocket'

// Create the context
const SocketContext = createContext<{ send: (data: any) => void } | null>(null)

// Provide it to the app
export function SocketProvider({ children }: { children: React.ReactNode }) {
  const { send } = useSharedSocket((data) => {
    console.log("WebSocket message received:", data)
  })

  return (
    <SocketContext.Provider value={{ send }}>
      {children}
    </SocketContext.Provider>
  )
}

// Make it usable in components
export function useSocket() {
  const ctx = useContext(SocketContext)
  if (!ctx) throw new Error("useSocket must be used inside SocketProvider")
  return ctx
}
