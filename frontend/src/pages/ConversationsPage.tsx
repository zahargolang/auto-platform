import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listConversations } from '../api/messenger'
import { getUser } from '../api/user'
import { useAuth } from '../context/useAuth'
import { ApiError } from '../lib/apiClient'
import type { Conversation } from '../types'

interface Row {
  conversation: Conversation
  counterpartName: string
}

export function ConversationsPage() {
  const { user } = useAuth()
  const [rows, setRows] = useState<Row[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!user) return
    let cancelled = false

    listConversations()
      .then(async (conversations) => {
        const withNames = await Promise.all(
          conversations.map(async (conversation) => {
            const counterpartId =
              conversation.seller_id === user.id ? conversation.buyer_id : conversation.seller_id
            try {
              const counterpart = await getUser(counterpartId)
              return { conversation, counterpartName: counterpart.username }
            } catch {
              return { conversation, counterpartName: 'Пользователь' }
            }
          }),
        )
        if (!cancelled) {
          withNames.sort(
            (a, b) =>
              new Date(b.conversation.last_message_at).getTime() -
              new Date(a.conversation.last_message_at).getTime(),
          )
          setRows(withNames)
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Не удалось загрузить переписки')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [user])

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-6 text-2xl font-bold">Сообщения</h1>

      {loading && <p className="text-gray-500">Загрузка…</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && rows.length === 0 && <p className="text-gray-500">У вас пока нет переписок</p>}

      <div className="flex flex-col gap-2">
        {rows.map(({ conversation, counterpartName }) => (
          <Link
            key={conversation.id}
            to={`/messages/${conversation.id}`}
            className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4 hover:border-blue-400"
          >
            <span className="font-medium">{counterpartName}</span>
            <span className="text-xs text-gray-400">
              {new Date(conversation.last_message_at).toLocaleString('ru-RU')}
            </span>
          </Link>
        ))}
      </div>
    </div>
  )
}
