'use client'

import { useEffect, useRef } from 'react'

export default function useSharedSocket(onMessage: (data: any) => void) {
  const workerRef = useRef<SharedWorker | null>(null)
  const portRef = useRef<MessagePort | null>(null)

  useEffect(() => {
    const worker = new SharedWorker('/shared-worker.js') // must match `public` location
    workerRef.current = worker
    const port = worker.port
    portRef.current = port
    port.start()

    port.onmessage = (e) => {
      onMessage(e.data)
    }

    port.postMessage({ type: 'login' })

    return () => {
      port.postMessage({ type: 'logout' })
      port.close()
    }
  }, [onMessage])

  const send = (data: any) => {
    portRef.current?.postMessage(data)
  }

  return { send }
}
