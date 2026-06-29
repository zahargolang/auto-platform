import { createContext } from 'react'
import type { AuthUser, UserDTO } from '../types'

export interface AuthContextValue {
  user: AuthUser | null
  isAuthenticated: boolean
  login: (phoneNumber: string, password: string) => Promise<void>
  register: (username: string, phoneNumber: string, password: string) => Promise<UserDTO>
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)
