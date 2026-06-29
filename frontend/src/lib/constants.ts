import type { BodyType, FuelType, TransmissionType } from '../types'

export const BODY_TYPES: { value: BodyType; label: string }[] = [
  { value: 'sedan', label: 'Седан' },
  { value: 'suv', label: 'Внедорожник' },
  { value: 'hatchback', label: 'Хэтчбек' },
  { value: 'coupe', label: 'Купе' },
  { value: 'wagon', label: 'Универсал' },
  { value: 'minivan', label: 'Минивэн' },
  { value: 'pickup', label: 'Пикап' },
]

export const FUEL_TYPES: { value: FuelType; label: string }[] = [
  { value: 'gasoline', label: 'Бензин' },
  { value: 'diesel', label: 'Дизель' },
  { value: 'electric', label: 'Электро' },
  { value: 'hybrid', label: 'Гибрид' },
  { value: 'lpg', label: 'Газ' },
]

export const TRANSMISSION_TYPES: { value: TransmissionType; label: string }[] = [
  { value: 'automatic', label: 'Автомат' },
  { value: 'manual', label: 'Механика' },
  { value: 'robot', label: 'Робот' },
  { value: 'variator', label: 'Вариатор' },
]

export function formatPrice(price: number): string {
  return new Intl.NumberFormat('ru-RU').format(price) + ' ₽'
}

export function labelFor(options: { value: string; label: string }[], value: string): string {
  return options.find((o) => o.value === value)?.label ?? value
}
