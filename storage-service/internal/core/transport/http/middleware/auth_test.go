package core_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireUserID_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(RequireUserID())
	router.GET("/", func(c *gin.Context) {
		t.Fatal("handler should not be reached without X-User-Id")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireUserID_InvalidHeader(t *testing.T) {
	router := gin.New()
	router.Use(RequireUserID())
	router.GET("/", func(c *gin.Context) {
		t.Fatal("handler should not be reached with invalid X-User-Id")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", "not-a-uuid")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireUserID_SetsUserID(t *testing.T) {
	id := uuid.New()

	router := gin.New()
	router.Use(RequireUserID())
	router.GET("/", func(c *gin.Context) {
		got, ok := c.MustGet("user_id").(uuid.UUID)
		if !ok || got != id {
			t.Fatalf("user_id in context = %v, want %v", got, id)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", id.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
