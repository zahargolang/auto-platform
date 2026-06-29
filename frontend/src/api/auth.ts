import { apiFetch } from '../lib/apiClient'
import type { LoginResponse, UserDTO } from '../types'

export function login(phoneNumber: string, password: string): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/auth/login', {
    method: 'POST',
    auth: false,
    body: { phone_number: phoneNumber, password },
  })
}

export function register(
  username: string,
  phoneNumber: string,
  password: string,
): Promise<UserDTO> {
  return apiFetch<UserDTO>('/auth/register', {
    method: 'POST',
    auth: false,
    body: { username, phone_number: phoneNumber, password },
  })
}
