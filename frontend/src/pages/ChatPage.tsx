import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react'
import { useParams } from 'react-router-dom'
import { listConversations, listMessages } from '../api/messenger'
import { getUser } from '../api/user'
import { useAuth } from '../context/useAuth'
import { useMessengerSocket } from '../lib/useMessengerSocket'
import { ApiError } from '../lib/apiClient'
import type { Message, ServerFrame } from '../types'

export function ChatPage() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()

  const [messages, setMessages] = useState<Message[]>([])
  const [counterpartName, setCounterpartName] = useState('Чат')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  const handleFrame = useCallback(
    (frame: ServerFrame) => {
      if (frame.type === 'message_sent') {
        const p = frame.payload
        if (p.conversation_id !== id) return
        setMessages((prev) => [
          ...prev,
          { id: p.id, conversation_id: p.conversation_id, sender_id: p.sender_id, body: p.body, created_at: p.created_at },
        ])
      } else if (frame.type === 'message') {
        const p = frame.payload
        if (p.conversation_id !== id) return
        setMessages((prev) => [
          ...prev,
          { id: p.message_id, conversation_id: p.conversation_id, sender_id: p.sender_id, body: p.body, created_at: p.created_at },
        ])
      } else if (frame.type === 'error') {
        setError(typeof frame.payload === 'string' ? frame.payload : 'Ошибка сокета')
      }
    },
    [id],
  )

  const { connected, sendMessage } = useMessengerSocket(handleFrame)

  useEffect(() => {
    if (!id || !user) return
    let cancelled = false

    Promise.all([listMessages(id), listConversations()])
      .then(([msgs, conversations]) => {
        if (cancelled) return
        setMessages(msgs)
        const conv = conversations.find((c) => c.id === id)
        if (!conv) return
        const counterpartId = conv.seller_id === user.id ? conv.buyer_id : conv.seller_id
        getUser(counterpartId)
          .then((u) => {
            if (!cancelled) setCounterpartName(u.username)
          })
          .catch(() => undefined)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(
            err instanceof ApiError && err.status === 403
              ? 'У вас нет доступа к этой переписке'
              : 'Не удалось загрузить сообщения',
          )
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [id, user])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!id || !draft.trim()) return
    sendMessage(id, draft.trim())
    setDraft('')
  }

  if (loading) return <p className="text-gray-500">Загрузка…</p>
  if (error) return <p className="text-red-600">{error}</p>

  return (
    <div className="mx-auto flex h-[70vh] max-w-xl flex-col">
      <div className="mb-3 flex items-center justify-between border-b border-gray-200 pb-2">
        <h1 className="text-xl font-bold">{counterpartName}</h1>
        <span className={`text-xs ${connected ? 'text-green-600' : 'text-gray-400'}`}>
          {connected ? 'online' : 'переподключение…'}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto rounded-lg border border-gray-200 bg-white p-4">
        {messages.length === 0 && <p className="text-sm text-gray-400">Сообщений пока нет</p>}
        <div className="flex flex-col gap-2">
          {messages.map((m) => {
            const mine = m.sender_id === user?.id
            return (
              <div key={m.id} className={`flex ${mine ? 'justify-end' : 'justify-start'}`}>
                <div
                  className={`max-w-xs rounded-lg px-3 py-2 text-sm ${
                    mine ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-900'
                  }`}
                >
                  <p>{m.body}</p>
                  <p className={`mt-1 text-xs ${mine ? 'text-blue-100' : 'text-gray-400'}`}>
                    {new Date(m.created_at).toLocaleTimeString('ru-RU')}
                  </p>
                </div>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </div>

      <form onSubmit={handleSubmit} className="mt-3 flex gap-2">
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Сообщение…"
          className="flex-1 rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
        />
        <button
          type="submit"
          disabled={!connected || !draft.trim()}
          className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          Отправить
        </button>
      </form>
    </div>
  )
}
