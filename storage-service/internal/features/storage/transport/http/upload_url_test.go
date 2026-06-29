package transport_http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	core_logger "storage-service/internal/core/logger"
	storage_service "storage-service/internal/features/storage/service"
)

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func newTestContext(userID uuid.UUID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/mine/upload-url", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := core_logger.ToContext(req.Context(), testLogger())
	c.Request = req.WithContext(ctx)
	c.Set("user_id", userID)
	return c, w
}

func TestNewUploadURL_OK(t *testing.T) {
	ownerID := uuid.New()
	h := NewHandler(&fakeService{
		newUploadURLFunc: func(ctx context.Context, gotOwnerID uuid.UUID, filename string) (storage_service.UploadURL, error) {
			if gotOwnerID != ownerID {
				t.Fatalf("ownerID = %v, want %v", gotOwnerID, ownerID)
			}
			if filename != "photo.jpg" {
				t.Fatalf("filename = %q, want %q", filename, "photo.jpg")
			}
			return storage_service.UploadURL{UploadURL: "https://s3/up", PublicURL: "https://s3/pub"}, nil
		},
	})

	c, w := newTestContext(ownerID, `{"filename":"photo.jpg"}`)
	h.NewUploadURL(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "https://s3/pub") {
		t.Fatalf("body = %s, want public_url present", w.Body.String())
	}
}

func TestNewUploadURL_MissingFilename(t *testing.T) {
	h := NewHandler(&fakeService{})

	c, w := newTestContext(uuid.New(), `{}`)
	h.NewUploadURL(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestNewUploadURL_ServiceError(t *testing.T) {
	h := NewHandler(&fakeService{
		newUploadURLFunc: func(ctx context.Context, ownerID uuid.UUID, filename string) (storage_service.UploadURL, error) {
			return storage_service.UploadURL{}, errors.New("boom")
		},
	})

	c, w := newTestContext(uuid.New(), `{"filename":"photo.jpg"}`)
	h.NewUploadURL(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
