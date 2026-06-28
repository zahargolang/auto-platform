package auth_transport_http

import (
	"context"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	core_domain "github.com/zosinkin/social_network/internal/core/domain"
)

// fakeService — ручная fake-реализация Service (см. transport.go): каждый
// тест задаёт только то поле-функцию, которое ему нужно.
type fakeService struct {
	registerFunc func(ctx context.Context, username, phoneNumber, password string) (core_domain.AuthUser, error)
	loginFunc    func(ctx context.Context, phoneNumber, password string) (string, error)
	loginWithRefreshFunc func(
		ctx context.Context,
		phoneNumber, password string,
		refreshTokenTTL time.Duration,
	) (string, string, error)
	refreshAccessTokenFunc func(ctx context.Context, refreshToken string) (string, error)
}

func (f *fakeService) Register(ctx context.Context, username, phoneNumber, password string) (core_domain.AuthUser, error) {
	return f.registerFunc(ctx, username, phoneNumber, password)
}

func (f *fakeService) Login(ctx context.Context, phoneNumber, password string) (string, error) {
	return f.loginFunc(ctx, phoneNumber, password)
}

func (f *fakeService) LoginWithRefresh(
	ctx context.Context,
	phoneNumber, password string,
	refreshTokenTTL time.Duration,
) (string, string, error) {
	return f.loginWithRefreshFunc(ctx, phoneNumber, password, refreshTokenTTL)
}

func (f *fakeService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	return f.refreshAccessTokenFunc(ctx, refreshToken)
}

// newTestContext строит *gin.Context с заданным телом запроса — без
// поднятия реального HTTP-сервера, как это принято для юнит-тестов
// gin-хендлеров.
func newTestContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}
