import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { deleteListing, getListings } from '../api/listings'
import { useAuth } from '../context/useAuth'
import { formatPrice } from '../lib/constants'
import { ApiError } from '../lib/apiClient'
import type { Listing } from '../types'

// У listing-service нет отдельного "GET /api/listings/mine" — только
// POST/PATCH/DELETE (см. helm/.../ingress.yaml, путь защищён по path).
// Поэтому тянем общий публичный список с большим limit и фильтруем по
// user_id на клиенте — нормально для реалистичного количества объявлений
// одного пользователя, но не масштабируется на тысячи объявлений.
export function MyListingsPage() {
  const { user } = useAuth()
  const [listings, setListings] = useState<Listing[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!user) return
    let cancelled = false
    getListings({ limit: 100 })
      .then((res) => {
        if (!cancelled) setListings(res.items.filter((l) => l.user_id === user.id))
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Не удалось загрузить объявления')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [user])

  async function handleDelete(id: string) {
    if (!window.confirm('Удалить объявление?')) return
    try {
      await deleteListing(id)
      setListings((prev) => prev.filter((l) => l.id !== id))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось удалить объявление')
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Мои объявления</h1>
        <Link
          to="/mine/new"
          className="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
        >
          + Создать объявление
        </Link>
      </div>

      {loading && <p className="text-gray-500">Загрузка…</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && listings.length === 0 && (
        <p className="text-gray-500">У вас пока нет объявлений</p>
      )}

      <div className="flex flex-col gap-3">
        {listings.map((listing) => (
          <div
            key={listing.id}
            className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4"
          >
            <Link to={`/listings/${listing.id}`} className="hover:text-blue-600">
              <div className="font-semibold">
                {listing.make} {listing.model}, {listing.year}
              </div>
              <div className="text-sm text-gray-500">
                {formatPrice(listing.price)} · {listing.status}
              </div>
            </Link>
            <div className="flex gap-2">
              <Link
                to={`/mine/${listing.id}/edit`}
                className="rounded border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50"
              >
                Редактировать
              </Link>
              <button
                onClick={() => handleDelete(listing.id)}
                className="rounded border border-red-300 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
              >
                Удалить
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
