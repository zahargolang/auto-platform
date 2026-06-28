package core_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) {
		if c.GetHeader(RequestIDHeader) == "" {
			t.Fatalf("request_id was not set on the incoming request")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Header().Get(RequestIDHeader) == "" {
		t.Fatalf("response is missing %s header", RequestIDHeader)
	}
}

func TestRequestID_PreservesClientProvided(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "client-supplied-id")
	router.ServeHTTP(w, req)

	if got := w.Header().Get(RequestIDHeader); got != "client-supplied-id" {
		t.Fatalf("RequestID() overwrote client-supplied id, got %q", got)
	}
}
