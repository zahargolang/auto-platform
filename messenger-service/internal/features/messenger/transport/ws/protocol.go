package ws

import (
	"time"

	"github.com/google/uuid"

	core_domain "messenger-service/internal/features/messenger/domain"
)

// ClientFrame — то, что клиент шлёт серверу по уже установленному соединению.
type ClientFrame struct {
	Type           string    `json:"type"` // "send_message"
	ConversationID uuid.UUID `json:"conversation_id"`
	Body           string    `json:"body"`
}

// ServerFrame — то, что сервер шлёт клиенту: подтверждение отправки,
// входящее сообщение от собеседника, либо ошибка.
type ServerFrame struct {
	Type    string `json:"type"` // "message_sent" | "message" | "error"
	Payload any    `json:"payload"`
}

// MessagePayload — JSON-форма сообщения в фрейме "message_sent". У
// core_domain.Message нет json-тегов (это domain-тип, не DTO), поэтому
// сериализовать его в ServerFrame напрямую нельзя — поля ушли бы в PascalCase
// и не совпадали бы с snake_case остальных контрактов (REST-история,
// MessageSentEvent фан-аута).
type MessagePayload struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}

func toMessagePayload(m core_domain.Message) MessagePayload {
	return MessagePayload{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Body:           m.Body,
		CreatedAt:      m.CreatedAt,
	}
}
