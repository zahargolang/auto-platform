import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { getListing } from '../api/listings'
import { createOrGetConversation } from '../api/messenger'
import { useAuth } from '../context/useAuth'
import { formatPrice, labelFor, BODY_TYPES, FUEL_TYPES, TRANSMISSION_TYPES } from '../lib/constants'
import { ApiError } from '../lib/apiClient'
import type { Listing } from '../types'

export function ListingDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const [listing, setListing] = useState<Listing | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [contacting, setContacting] = useState(false)

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    getListing(id)
      .then((res) => {
        if (!cancelled) setListing(res)
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError && err.status === 404 ? 'Объявление не найдено' : 'Не удалось загрузить объявление')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id])

  async function handleContactSeller() {
    if (!listing) return
    setContacting(true)
    try {
      const conv = await createOrGetConversation(listing.id, listing.user_id)
      navigate(`/messages/${conv.id}`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось начать переписку')
    } finally {
      setContacting(false)
    }
  }

  if (loading) return <p className="text-gray-500">Загрузка…</p>
  if (error) return <p className="text-red-600">{error}</p>
  if (!listing) return null

  const isOwner = user?.id === listing.user_id

  return (
    <div className="mx-auto max-w-2xl">
      <h1 className="text-2xl font-bold">
        {listing.make} {listing.model}, {listing.year}
      </h1>
      <p className="mt-2 text-3xl font-bold text-blue-600">{formatPrice(listing.price)}</p>

      <dl className="mt-6 grid grid-cols-2 gap-3 rounded-lg border border-gray-200 bg-white p-4 text-sm">
        <Field label="Пробег" value={`${listing.mileage.toLocaleString('ru-RU')} км`} />
        <Field label="Цвет" value={listing.color} />
        <Field label="Кузов" value={labelFor(BODY_TYPES, listing.body_type)} />
        <Field label="Топливо" value={labelFor(FUEL_TYPES, listing.fuel_type)} />
        <Field label="Трансмиссия" value={labelFor(TRANSMISSION_TYPES, listing.transmission)} />
        <Field label="Объём двигателя" value={`${listing.engine_volume} л`} />
        <Field label="Город" value={listing.city} />
        <Field label="Регион" value={listing.region} />
      </dl>

      <div className="mt-4">
        <h2 className="mb-1 font-semibold">Описание</h2>
        <p className="whitespace-pre-wrap text-gray-700">{listing.description}</p>
      </div>

      <div className="mt-6">
        {isOwner ? (
          <Link
            to="/mine"
            className="inline-block rounded bg-gray-100 px-4 py-2 text-sm hover:bg-gray-200"
          >
            Это ваше объявление — управлять в «Мои объявления»
          </Link>
        ) : isAuthenticated ? (
          <button
            onClick={handleContactSeller}
            disabled={contacting}
            className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {contacting ? 'Открываем чат…' : 'Написать продавцу'}
          </button>
        ) : (
          <Link
            to="/login"
            className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
          >
            Войдите, чтобы написать продавцу
          </Link>
        )}
      </div>
    </div>
  )
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-gray-500">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  )
}
