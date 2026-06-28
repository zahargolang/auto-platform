package core_middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPanic_RecoversAndReturns500(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Logger(testLogger()), Panic())
	router.GET("/", func(c *gin.Context) {
		panic("something went terribly wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Сам факт того, что ServeHTTP не паникует наружу (не роняет тестовый
	// процесс), уже подтверждает recover — но проверим и итоговый ответ.
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body: %v (body: %s)", err, w.Body.String())
	}
	if body["error"] == "" {
		t.Fatalf("response body missing \"error\" field: %s", w.Body.String())
	}
}

func TestPanic_PassesThroughWhenNoPanic(t *testing.T) {
	router := gin.New()
	router.Use(RequestID(), Logger(testLogger()), Panic())
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
