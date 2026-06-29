import { getTokens, setAccessToken, clearTokens } from './tokenStorage'

const BASE = '/api'

export class ApiError extends Error {
  status: number
  body: unknown

  constructor(status: number, message: string, body?: unknown) {
    super(message)
    this.status = status
    this.body = body
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  /** Прикреплять Authorization: Bearer <access_token>. По умолчанию true. */
  auth?: boolean
}

let refreshPromise: Promise<string | null> | null = null

// Обновляет access-токен через refresh-токен (POST /api/auth/refresh).
// Совмещает параллельные запросы в один — если несколько запросов
// получили 401 одновременно, рефреш токена выполнится только раз.
async function refreshAccessToken(): Promise<string | null> {
  const { refreshToken } = getTokens()
  if (!refreshToken) return null

  try {
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (!res.ok) {
      clearTokens()
      return null
    }
    const data = (await res.json()) as { access_token: string }
    setAccessToken(data.access_token)
    return data.access_token
  } catch {
    return null
  }
}

async function doFetch(path: string, options: RequestOptions): Promise<Response> {
  const { auth = true, body, headers, ...rest } = options
  const finalHeaders = new Headers(headers)

  let finalBody: BodyInit | undefined
  if (body !== undefined) {
    finalHeaders.set('Content-Type', 'application/json')
    finalBody = JSON.stringify(body)
  }

  if (auth) {
    const { accessToken } = getTokens()
    if (accessToken) finalHeaders.set('Authorization', `Bearer ${accessToken}`)
  }

  return fetch(`${BASE}${path}`, { ...rest, headers: finalHeaders, body: finalBody })
}

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  let res = await doFetch(path, options)

  if (res.status === 401 && options.auth !== false) {
    refreshPromise ??= refreshAccessToken().finally(() => {
      refreshPromise = null
    })
    const newToken = await refreshPromise
    if (newToken) {
      res = await doFetch(path, options)
    }
  }

  if (!res.ok) {
    let body: unknown
    try {
      body = await res.json()
    } catch {
      // тело не json или пустое — оставляем body как undefined
    }
    const message =
      (body as { error?: string } | undefined)?.error ?? `${res.status} ${res.statusText}`
    throw new ApiError(res.status, message, body)
  }

  if (res.status === 204) return undefined as T
  const text = await res.text()
  return (text ? JSON.parse(text) : undefined) as T
}
