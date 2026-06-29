import { useEffect, useState, type FormEvent } from 'react'
import { getListings } from '../api/listings'
import { ListingCard } from '../components/ListingCard'
import { BODY_TYPES, FUEL_TYPES, TRANSMISSION_TYPES } from '../lib/constants'
import type { Listing, ListingFilters } from '../types'
import { ApiError } from '../lib/apiClient'

const EMPTY_FILTERS: ListingFilters = {}

export function HomePage() {
  const [filters, setFilters] = useState<ListingFilters>(EMPTY_FILTERS)
  const [appliedFilters, setAppliedFilters] = useState<ListingFilters>(EMPTY_FILTERS)
  const [listings, setListings] = useState<Listing[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    getListings(appliedFilters)
      .then((res) => {
        if (!cancelled) setListings(res.items)
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
  }, [appliedFilters])

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setAppliedFilters(filters)
  }

  function handleReset() {
    setFilters(EMPTY_FILTERS)
    setAppliedFilters(EMPTY_FILTERS)
  }

  function setField<K extends keyof ListingFilters>(key: K, value: ListingFilters[K]) {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <div>
      <h1 className="mb-4 text-2xl font-bold">Объявления</h1>

      <form
        onSubmit={handleSubmit}
        className="mb-6 grid grid-cols-2 gap-3 rounded-lg border border-gray-200 bg-white p-4 sm:grid-cols-4"
      >
        <input
          placeholder="Марка"
          value={filters.make ?? ''}
          onChange={(e) => setField('make', e.target.value)}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        />
        <input
          placeholder="Модель"
          value={filters.model ?? ''}
          onChange={(e) => setField('model', e.target.value)}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        />
        <input
          placeholder="Город"
          value={filters.city ?? ''}
          onChange={(e) => setField('city', e.target.value)}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        />
        <select
          value={filters.body_type ?? ''}
          onChange={(e) => setField('body_type', e.target.value as ListingFilters['body_type'])}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        >
          <option value="">Тип кузова</option>
          {BODY_TYPES.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <select
          value={filters.fuel_type ?? ''}
          onChange={(e) => setField('fuel_type', e.target.value as ListingFilters['fuel_type'])}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        >
          <option value="">Тип топлива</option>
          {FUEL_TYPES.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <select
          value={filters.transmission ?? ''}
          onChange={(e) =>
            setField('transmission', e.target.value as ListingFilters['transmission'])
          }
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        >
          <option value="">Трансмиссия</option>
          {TRANSMISSION_TYPES.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <input
          type="number"
          placeholder="Цена от"
          value={filters.price_from ?? ''}
          onChange={(e) => setField('price_from', e.target.valueAsNumber || undefined)}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        />
        <input
          type="number"
          placeholder="Цена до"
          value={filters.price_to ?? ''}
          onChange={(e) => setField('price_to', e.target.valueAsNumber || undefined)}
          className="rounded border border-gray-300 px-2 py-1.5 text-sm"
        />
        <div className="col-span-2 flex gap-2 sm:col-span-4">
          <button
            type="submit"
            className="rounded bg-blue-600 px-4 py-1.5 text-sm text-white hover:bg-blue-700"
          >
            Применить
          </button>
          <button
            type="button"
            onClick={handleReset}
            className="rounded border border-gray-300 px-4 py-1.5 text-sm hover:bg-gray-50"
          >
            Сбросить
          </button>
        </div>
      </form>

      {loading && <p className="text-gray-500">Загрузка…</p>}
      {error && <p className="text-red-600">{error}</p>}
      {!loading && !error && listings.length === 0 && (
        <p className="text-gray-500">Ничего не найдено</p>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {listings.map((listing) => (
          <ListingCard key={listing.id} listing={listing} />
        ))}
      </div>
    </div>
  )
}
