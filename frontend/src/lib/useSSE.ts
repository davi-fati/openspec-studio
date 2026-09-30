import { useEffect, useRef } from 'react'

const RECONNECT_DELAY_MS = 2000

/**
 * Subscribes to the backend's single SSE stream at /api/events and calls
 * onUpdate whenever an "update" event arrives. Reconnects automatically on
 * drop, so callers never see a dead connection.
 */
export function useSSE(onUpdate: () => void) {
  const onUpdateRef = useRef(onUpdate)
  onUpdateRef.current = onUpdate

  useEffect(() => {
    let source: EventSource | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let cancelled = false

    const connect = () => {
      if (cancelled) return
      source = new EventSource('/api/events')

      source.addEventListener('update', () => {
        onUpdateRef.current()
      })

      source.onerror = () => {
        source?.close()
        if (!cancelled) {
          reconnectTimer = setTimeout(connect, RECONNECT_DELAY_MS)
        }
      }
    }

    connect()

    return () => {
      cancelled = true
      source?.close()
      if (reconnectTimer) clearTimeout(reconnectTimer)
    }
  }, [])
}
