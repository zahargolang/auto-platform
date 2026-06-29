const ACCESS_KEY = 'auto-platform.access_token'
const REFRESH_KEY = 'auto-platform.refresh_token'

export interface Tokens {
  accessToken: string | null
  refreshToken: string | null
}

export function getTokens(): Tokens {
  return {
    accessToken: localStorage.getItem(ACCESS_KEY),
    refreshToken: localStorage.getItem(REFRESH_KEY),
  }
}

export function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_KEY, accessToken)
  localStorage.setItem(REFRESH_KEY, refreshToken)
}

export function setAccessToken(accessToken: string): void {
  localStorage.setItem(ACCESS_KEY, accessToken)
}

export function clearTokens(): void {
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

// Декодирует payload JWT без проверки подписи — только чтобы вытащить
// claims (sub/username/phoneNumber) для отображения в UI. Сама проверка
// подписи и срока действия делается на бэкенде при каждом запросе.
export function decodeJwt<T = unknown>(token: string): T | null {
  try {
    const payload = token.split('.')[1]
    const json = decodeURIComponent(
      atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
        .split('')
        .map((c) => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join(''),
    )
    return JSON.parse(json) as T
  } catch {
    return null
  }
}
