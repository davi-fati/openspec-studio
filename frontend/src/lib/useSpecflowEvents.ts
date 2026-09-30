import { useEffect, useRef } from 'react'

const RECONNECT_DELAY_MS = 2000

export interface SpecflowEvent {
  flowId: number
  itemId: number
  type: 'flow_status' | 'item_status' | 'item_log'
  data: string
}

/**
 * Subscribes to the backend's single SSE stream, listening specifically for
 * "specflow" events (flow/item status transitions and near-live log
 * chunks) - a separate event name from the generic "update" other pages
 * use, since Specflow needs structured per-item data, not just "refetch".
 */
export function useSpecflowEvents(onEvent: (ev: SpecflowEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    let source: EventSource | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let cancelled = false

    const connect = () => {
      if (cancelled) return
      source = new EventSource('/api/events')

      source.addEventListener('specflow', (e) => {
        try {
          onEventRef.current(JSON.parse((e as MessageEvent).data))
        } catch {
          // ignore malformed event
        }
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
