import { useEffect, useState } from 'react'
import { getMyProfile } from '../api/user'
import { ApiError } from '../lib/apiClient'
import type { UserDTO } from '../types'

export function ProfilePage() {
  const [profile, setProfile] = useState<UserDTO | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getMyProfile()
      .then(setProfile)
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Не удалось загрузить профиль'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p className="text-gray-500">Загрузка…</p>
  if (error) return <p className="text-red-600">{error}</p>
  if (!profile) return null

  return (
    <div className="mx-auto max-w-sm">
      <h1 className="mb-6 text-2xl font-bold">Профиль</h1>
      <dl className="flex flex-col gap-3 rounded-lg border border-gray-200 bg-white p-4">
        <div>
          <dt className="text-sm text-gray-500">Имя пользователя</dt>
          <dd className="font-medium">{profile.username}</dd>
        </div>
        <div>
          <dt className="text-sm text-gray-500">Телефон</dt>
          <dd className="font-medium">{profile.phone_number}</dd>
        </div>
        <div>
          <dt className="text-sm text-gray-500">ID</dt>
          <dd className="font-mono text-xs text-gray-600">{profile.id}</dd>
        </div>
      </dl>
    </div>
  )
}
