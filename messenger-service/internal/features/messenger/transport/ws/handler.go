package ws

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	core_logger "messenger-service/internal/core/logger"
	core_domain "messenger-service/internal/features/messenger/domain"
)

type Service interface {
	SendMessage(ctx context.Context, conversationID uuid.UUID, senderID uuid.UUID, body string) (core_domain.Message, error)
}

type Handler struct {
	hub      *Hub
	service  Service
	upgrader websocket.Upgrader
}

func NewHandler(hub *Hub, service Service) *Handler {
	allowed := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	return &Handler{
		hub:     hub,
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // не браузерный клиент (curl, тест) — не блокируем
				}
				for _, a := range allowed {
					if a == "*" || a == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// ServeWS — это "взять трубку": user_id уже проверен gateway'ем (см.
// RequireUserID на группе, см. transport.go) до того, как запрос попал
// сюда — на обычном HTTP-запросе ещё можно вернуть 401 кодом статуса,
// после апгрейда это уже не получится. Дальше переводим соединение в
// режим WebSocket и регистрируем пользователя в Hub.
func (h *Handler) ServeWS(c *gin.Context) {
	log := core_logger.FromContext(c.Request.Context())
	userID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	//Для того, чтобы сделать запрос на подключение к WS, используется upgrader
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("websocket upgrade error: ", zap.Error(err))
		return
	}

	h.hub.Register(userID, conn)
	h.readLoop(c.Request.Context(), userID, conn)
}

// readLoop — это "слушать трубку": пока соединение открыто, читаем всё,
// что присылает клиент, и реагируем. Выполняется в своей goroutine на
// каждое соединение (запускается прямо из ServeWS, который сам уже
// исполняется в горутине Gin на каждый запрос).
func (h *Handler) readLoop(ctx context.Context, userID uuid.UUID, conn *websocket.Conn) {
	defer func() {
		h.hub.Unregister(userID, conn)
		_ = conn.Close()
	}()

	for {
		var frame ClientFrame
		if err := conn.ReadJSON(&frame); err != nil {
			// клиент закрыл соединение (закрыл вкладку) или сетевая ошибка —
			// для нас это просто "разговор окончен", не повод логировать как ошибку
			return
		}

		switch frame.Type {
		case "send_message":
			msg, err := h.service.SendMessage(ctx, frame.ConversationID, userID, frame.Body)
			if err != nil {
				_ = conn.WriteJSON(ServerFrame{Type: "error", Payload: err.Error()})
				continue
			}
			_ = conn.WriteJSON(ServerFrame{Type: "message_sent", Payload: toMessagePayload(msg)})
		default:
			_ = conn.WriteJSON(ServerFrame{Type: "error", Payload: "unknown frame type"})
		}
	}
}
