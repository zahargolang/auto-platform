package auth_transport_http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
)

func TestRegister_Success(t *testing.T) {
	wantUser := core_domain.AuthUser{
		ID:          uuid.New(),
		Username:    "john_doe",
		PhoneNumber: "+77001234567",
	}
	svc := &fakeService{
		registerFunc: func(ctx context.Context, username, phoneNumber, password string) (core_domain.AuthUser, error) {
			return wantUser, nil
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"username":"john_doe","phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Register(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("Register() status = %d, want %d (body: %s)", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp UserDTOResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Register() invalid JSON body: %v (body: %s)", err, w.Body.String())
	}
	if resp.ID != wantUser.ID || resp.Username != wantUser.Username || resp.PhoneNumber != wantUser.PhoneNumber {
		t.Fatalf("Register() response = %+v, want fields matching %+v", resp, wantUser)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	svc := &fakeService{}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`not-json`)
	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Register() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestRegister_MissingRequiredFields(t *testing.T) {
	svc := &fakeService{}
	h := NewAuthHTTPHandler(svc)

	// "username" отсутствует — должно быть отбито биндингом (`binding:"required"`),
	// не дойдя до сервисного слоя вовсе.
	c, w := newTestContext(`{"phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Register() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestRegister_ServiceError проверяет, что хендлер пишет ровно один
// JSON-ответ при ошибке сервиса (раньше после c.JSON(500, ...) не было
// `return`, и код дополнительно писал второй, "успешный" ответ поверх
// первого — здесь это и проверяем: декодер не должен находить ничего
// после первого объекта).
//
// Дальнейшее уточнение статусов под конкретные типы ошибок
// (ErrPhoneNumberUse/ErrInvalidArgument — 400, а не 500) в хендлере пока
// не реализовано — он всегда отвечает 500 на любую ошибку сервиса.
func TestRegister_ServiceError(t *testing.T) {
	svc := &fakeService{
		registerFunc: func(ctx context.Context, username, phoneNumber, password string) (core_domain.AuthUser, error) {
			return core_domain.AuthUser{}, errors.New("invalid argument")
		},
	}
	h := NewAuthHTTPHandler(svc)

	c, w := newTestContext(`{"username":"john_doe","phone_number":"+77001234567","password":"Sup3rSecret!"}`)
	h.Register(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Register() status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	dec := json.NewDecoder(w.Body)
	var first map[string]any
	if err := dec.Decode(&first); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}
	if dec.More() {
		t.Fatalf("Register() wrote more than one JSON object to the response body — missing `return` after the error branch")
	}
}
