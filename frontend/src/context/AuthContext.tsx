import { useState, type ReactNode } from 'react'
import * as authApi from '../api/auth'
import { clearTokens, decodeJwt, getTokens, setTokens } from '../lib/tokenStorage'
import type { AuthUser } from '../types'
import { AuthContext } from './authContext'

interface AccessTokenClaims {
  sub: string
  username: string
  phoneNumber: string
}

function userFromToken(token: string): AuthUser | null {
  const claims = decodeJwt<AccessTokenClaims>(token)
  if (!claims) return null
  return { id: claims.sub, username: claims.username, phoneNumber: claims.phoneNumber }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(() => {
    const { accessToken } = getTokens()
    return accessToken ? userFromToken(accessToken) : null
  })

  async function login(phoneNumber: string, password: string) {
    const res = await authApi.login(phoneNumber, password)
    setTokens(res.access_token, res.refresh_token)
    setUser(userFromToken(res.access_token))
  }

  async function register(username: string, phoneNumber: string, password: string) {
    return authApi.register(username, phoneNumber, password)
  }

  function logout() {
    clearTokens()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: user !== null, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
