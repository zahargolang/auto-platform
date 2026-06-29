package core_middleware

import (
	core_logger "storage-service/internal/core/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogger_PutsLoggerInContext(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Logger(testLogger()))
	router.GET("/", func(c *gin.Context) {
		// core_logger.FromContext паникует, если логгер не был положен в
		// контекст — сам факт того, что вызов не паникует, и есть проверка.
		log := core_logger.FromContext(c.Request.Context())
		if log == nil {
			t.Fatalf("FromContext() returned nil logger")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
