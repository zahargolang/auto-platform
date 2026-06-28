package core_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTrace_PassesThroughStatusCode(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Logger(testLogger()), Trace())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusTeapot)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	// Trace ничего не должен подменять в ответе — только читать
	// c.Writer.Status() после того, как хендлер отработал.
	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}
}
