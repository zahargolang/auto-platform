import { apiFetch } from '../lib/apiClient'
import type { Conversation, Message } from '../types'

export function listConversations(): Promise<Conversation[]> {
  return apiFetch<Conversation[]>('/messenger/mine/conversations')
}

export function createOrGetConversation(
  listingId: string,
  recipientId: string,
): Promise<Conversation> {
  return apiFetch<Conversation>('/messenger/mine/conversation', {
    method: 'POST',
    body: { listing_id: listingId, recipient_id: recipientId },
  })
}

export function listMessages(
  conversationId: string,
  page = 1,
  limit = 50,
): Promise<Message[]> {
  return apiFetch<Message[]>(
    `/messenger/mine/conversations/${conversationId}/messages?page=${page}&limit=${limit}`,
  )
}
