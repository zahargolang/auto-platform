import { useCallback, useEffect, useRef, useState } from 'react'
import { getTokens } from './tokenStorage'
import type { ClientFrame, ServerFrame } from '../types'

const MAX_RECONNECT_DELAY_MS = 10_000

// Один WebSocket на время жизни компонента, который его монтирует (страница
// чата). Токен передаём через ?token= в query — нативный WebSocket API не
// умеет ставить кастомные заголовки на хендшейк, поэтому Authorization тут
// не сработает (gateway достаёт токен сам из X-Original-URL, см.
// auth-service/.../validate.go: extractToken).
//
// При разрыве (закрытие, ошибка, протухший токен) переподключаемся с
// экспоненциальной задержкой (1с, 2с, 4с... до 10с) — без этого "connected"
// один раз падал бы в false навсегда, и UI вечно показывал бы
// "переподключение…", реально ничего не предпринимая.
export function useMessengerSocket(onFrame: (frame: ServerFrame) => void) {
  const wsRef = useRef<WebSocket | null>(null)
  const [connected, setConnected] = useState(false)
  const onFrameRef = useRef(onFrame)
  onFrameRef.current = onFrame

  useEffect(() => {
    let cancelled = false
    let reconnectAttempt = 0
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined

    function connect() {
      const { accessToken } = getTokens()
      if (!accessToken || cancelled) return

      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const url = `${protocol}://${window.location.host}/api/messenger/ws?token=${encodeURIComponent(accessToken)}`
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        reconnectAttempt = 0
        setConnected(true)
      }
      ws.onclose = () => {
        setConnected(false)
        if (cancelled) return
        const delay = Math.min(1000 * 2 ** reconnectAttempt, MAX_RECONNECT_DELAY_MS)
        reconnectAttempt += 1
        reconnectTimer = setTimeout(connect, delay)
      }
      ws.onerror = () => {
        ws.close()
      }
      ws.onmessage = (event: MessageEvent<string>) => {
        try {
          const frame = JSON.parse(event.data) as ServerFrame
          onFrameRef.current(frame)
        } catch {
          // не JSON — игнорируем
        }
      }
    }

    connect()

    return () => {
      cancelled = true
      clearTimeout(reconnectTimer)
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [])

  const sendMessage = useCallback((conversationId: string, body: string) => {
    const frame: ClientFrame = { type: 'send_message', conversation_id: conversationId, body }
    wsRef.current?.send(JSON.stringify(frame))
  }, [])

  return { connected, sendMessage }
}
