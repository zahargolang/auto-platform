package auth_transport_http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	h := NewAuthHTTPHandler(&fakeService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	h.Health(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Health() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthorized(t *testing.T) {
	h := NewAuthHTTPHandler(&fakeService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/authorized", nil)

	h.Authorized(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Authorized() status = %d, want %d", w.Code, http.StatusOK)
	}
}
