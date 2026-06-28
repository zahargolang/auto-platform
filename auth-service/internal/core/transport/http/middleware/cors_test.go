package core_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newCORSRouter(allowedOrigins []string) *gin.Engine {
	router := gin.New()
	router.Use(CORS(allowedOrigins))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestCORS_AllowedOrigin(t *testing.T) {
	router := newCORSRouter([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "https://example.com")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	router := newCORSRouter([]string{"https://example.com"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty for a disallowed origin", got)
	}
}

func TestCORS_WildcardAllowsAnyOrigin(t *testing.T) {
	router := newCORSRouter([]string{"*"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://anything.example")
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://anything.example" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q with wildcard config", got, "https://anything.example")
	}
}

func TestCORS_PreflightShortCircuits(t *testing.T) {
	handlerCalled := false
	router := gin.New()
	router.Use(CORS([]string{"https://example.com"}))
	router.OPTIONS("/", func(c *gin.Context) {
		handlerCalled = true
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if handlerCalled {
		t.Fatalf("OPTIONS preflight reached the handler, want it short-circuited by CORS")
	}
}
