import { apiFetch } from '../lib/apiClient'
import type { UserDTO } from '../types'

export function getMyProfile(): Promise<UserDTO> {
  return apiFetch<UserDTO>('/user/me')
}

export function getUser(id: string): Promise<UserDTO> {
  return apiFetch<UserDTO>(`/user/${id}`, { auth: false })
}
