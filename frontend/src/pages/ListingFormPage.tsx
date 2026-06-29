import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { createListing, getListing, updateListing } from '../api/listings'
import { ApiError } from '../lib/apiClient'
import { BODY_TYPES, FUEL_TYPES, TRANSMISSION_TYPES } from '../lib/constants'
import type { BodyType, FuelType, ListingStatus, TransmissionType } from '../types'

// PATCH /api/listings/mine/:id принимает только подмножество полей
// (UpdateListingRequest на бэкенде) — марка/модель/год/тип кузова/топлива/
// трансмиссии/объём двигателя неизменяемы после создания. Поэтому форма
// редактирования показывает меньше полей, чем форма создания.
const STATUSES: ListingStatus[] = ['active', 'inactive', 'sold']

interface FormState {
  title: string
  description: string
  price: string
  make: string
  model: string
  year: string
  mileage: string
  color: string
  bodyType: BodyType
  fuelType: FuelType
  transmission: TransmissionType
  engineVolume: string
  city: string
  region: string
  status: ListingStatus
}

const EMPTY: FormState = {
  title: '',
  description: '',
  price: '',
  make: '',
  model: '',
  year: '',
  mileage: '0',
  color: '',
  bodyType: 'sedan',
  fuelType: 'gasoline',
  transmission: 'automatic',
  engineVolume: '',
  city: '',
  region: '',
  status: 'active',
}

export function ListingFormPage() {
  const { id } = useParams<{ id: string }>()
  const isEdit = Boolean(id)
  const navigate = useNavigate()

  const [form, setForm] = useState<FormState>(EMPTY)
  const [loading, setLoading] = useState(isEdit)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    getListing(id)
      .then((l) =>
        setForm({
          title: l.title,
          description: l.description,
          price: String(l.price),
          make: l.make,
          model: l.model,
          year: String(l.year),
          mileage: String(l.mileage),
          color: l.color,
          bodyType: l.body_type,
          fuelType: l.fuel_type,
          transmission: l.transmission,
          engineVolume: String(l.engine_volume),
          city: l.city,
          region: l.region,
          status: l.status,
        }),
      )
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Не удалось загрузить объявление'))
      .finally(() => setLoading(false))
  }, [id])

  function setField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSaving(true)
    try {
      if (isEdit && id) {
        await updateListing(id, {
          title: form.title,
          description: form.description,
          price: Number(form.price),
          status: form.status,
          mileage: Number(form.mileage),
          color: form.color,
          city: form.city,
          region: form.region,
        })
      } else {
        await createListing({
          title: form.title,
          description: form.description,
          price: Number(form.price),
          make: form.make,
          model: form.model,
          year: Number(form.year),
          mileage: Number(form.mileage),
          color: form.color,
          body_type: form.bodyType,
          fuel_type: form.fuelType,
          transmission: form.transmission,
          engine_volume: Number(form.engineVolume),
          city: form.city,
          region: form.region,
        })
      }
      navigate('/mine')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось сохранить объявление')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <p className="text-gray-500">Загрузка…</p>

  const inputClass = 'rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none'

  return (
    <div className="mx-auto max-w-xl">
      <h1 className="mb-6 text-2xl font-bold">
        {isEdit ? 'Редактировать объявление' : 'Новое объявление'}
      </h1>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <label className="flex flex-col gap-1">
          <span className="text-sm text-gray-600">Заголовок</span>
          <input
            required
            minLength={3}
            maxLength={200}
            value={form.title}
            onChange={(e) => setField('title', e.target.value)}
            className={inputClass}
          />
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-sm text-gray-600">Описание</span>
          <textarea
            required
            rows={4}
            value={form.description}
            onChange={(e) => setField('description', e.target.value)}
            className={inputClass}
          />
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-sm text-gray-600">Цена, ₽</span>
          <input
            type="number"
            required
            min={1}
            value={form.price}
            onChange={(e) => setField('price', e.target.value)}
            className={inputClass}
          />
        </label>

        {!isEdit && (
          <div className="grid grid-cols-2 gap-4">
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Марка</span>
              <input
                required
                value={form.make}
                onChange={(e) => setField('make', e.target.value)}
                className={inputClass}
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Модель</span>
              <input
                required
                value={form.model}
                onChange={(e) => setField('model', e.target.value)}
                className={inputClass}
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Год выпуска</span>
              <input
                type="number"
                required
                min={1901}
                value={form.year}
                onChange={(e) => setField('year', e.target.value)}
                className={inputClass}
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Объём двигателя, л</span>
              <input
                type="number"
                step="0.1"
                min={0}
                value={form.engineVolume}
                onChange={(e) => setField('engineVolume', e.target.value)}
                className={inputClass}
              />
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Тип кузова</span>
              <select
                value={form.bodyType}
                onChange={(e) => setField('bodyType', e.target.value as BodyType)}
                className={inputClass}
              >
                {BODY_TYPES.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Тип топлива</span>
              <select
                value={form.fuelType}
                onChange={(e) => setField('fuelType', e.target.value as FuelType)}
                className={inputClass}
              >
                {FUEL_TYPES.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1">
              <span className="text-sm text-gray-600">Трансмиссия</span>
              <select
                value={form.transmission}
                onChange={(e) => setField('transmission', e.target.value as TransmissionType)}
                className={inputClass}
              >
                {TRANSMISSION_TYPES.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <label className="flex flex-col gap-1">
            <span className="text-sm text-gray-600">Пробег, км</span>
            <input
              type="number"
              min={0}
              value={form.mileage}
              onChange={(e) => setField('mileage', e.target.value)}
              className={inputClass}
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-sm text-gray-600">Цвет</span>
            <input
              required
              value={form.color}
              onChange={(e) => setField('color', e.target.value)}
              className={inputClass}
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-sm text-gray-600">Город</span>
            <input
              required
              value={form.city}
              onChange={(e) => setField('city', e.target.value)}
              className={inputClass}
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-sm text-gray-600">Регион</span>
            <input
              required
              value={form.region}
              onChange={(e) => setField('region', e.target.value)}
              className={inputClass}
            />
          </label>
        </div>

        {isEdit && (
          <label className="flex flex-col gap-1">
            <span className="text-sm text-gray-600">Статус</span>
            <select
              value={form.status}
              onChange={(e) => setField('status', e.target.value as ListingStatus)}
              className={inputClass}
            >
              {STATUSES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>
        )}

        {error && <p className="text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={saving}
          className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {saving ? 'Сохраняем…' : isEdit ? 'Сохранить' : 'Создать'}
        </button>
      </form>
    </div>
  )
}
