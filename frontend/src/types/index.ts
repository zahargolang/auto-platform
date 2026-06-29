// Контракты — зеркало DTO из Go-хендлеров соответствующих сервисов.
// auth-service: internal/features/auth/transport/http/dto.go
// listing-service: internal/features/listings/transport/http/dto.go
// user-service: internal/feauture/users/transport/http/dto.go
// messenger-service: internal/features/messenger/transport/http/dto.go, transport/ws/protocol.go

export interface AuthUser {
  id: string
  username: string
  phoneNumber: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
}

export interface UserDTO {
  id: string
  version: number
  username: string
  phone_number: string
}

export type ListingStatus = 'active' | 'inactive' | 'sold'
export type BodyType = 'sedan' | 'suv' | 'hatchback' | 'coupe' | 'wagon' | 'minivan' | 'pickup'
export type FuelType = 'gasoline' | 'diesel' | 'electric' | 'hybrid' | 'lpg'
export type TransmissionType = 'automatic' | 'manual' | 'robot' | 'variator'

export interface Listing {
  id: string
  user_id: string
  title: string
  description: string
  price: number
  status: ListingStatus
  make: string
  model: string
  year: number
  mileage: number
  color: string
  body_type: BodyType
  fuel_type: FuelType
  transmission: TransmissionType
  engine_volume: number
  city: string
  region: string
  created_at: string
  updated_at: string
}

export interface ListingsResponse {
  items: Listing[]
  total: number
}

export interface ListingFilters {
  make?: string
  model?: string
  city?: string
  region?: string
  fuel_type?: FuelType | ''
  transmission?: TransmissionType | ''
  body_type?: BodyType | ''
  year_from?: number
  year_to?: number
  price_from?: number
  price_to?: number
  mileage_from?: number
  mileage_to?: number
  page?: number
  limit?: number
}

export interface ListingFormValues {
  title: string
  description: string
  price: number
  make: string
  model: string
  year: number
  mileage: number
  color: string
  body_type: BodyType
  fuel_type: FuelType
  transmission: TransmissionType
  engine_volume: number
  city: string
  region: string
}

export interface Conversation {
  id: string
  listing_id: string
  seller_id: string
  buyer_id: string
  created_at: string
  last_message_at: string
}

export interface Message {
  id: string
  conversation_id: string
  sender_id: string
  body: string
  created_at: string
}

export interface ClientFrame {
  type: 'send_message'
  conversation_id: string
  body: string
}

// payload "message_sent" — подтверждение собственного отправленного
// сообщения (ws/protocol.go: MessagePayload).
export interface MessageSentPayload {
  id: string
  conversation_id: string
  sender_id: string
  body: string
  created_at: string
}

// payload "message" — входящее сообщение от собеседника через Kafka-фанаут
// (transport/kafka/event_dto.go: MessageSentEvent) — обратите внимание,
// id сообщения тут называется message_id, а не id.
export interface IncomingMessagePayload {
  message_id: string
  conversation_id: string
  sender_id: string
  recipient_id: string
  body: string
  created_at: string
}

export type ServerFrame =
  | { type: 'message_sent'; payload: MessageSentPayload }
  | { type: 'message'; payload: IncomingMessagePayload }
  | { type: 'error'; payload: string }
