package auth_transport_http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	core_errors "github.com/zosinkin/social_network/internal/core/errors"
	core_logger "github.com/zosinkin/social_network/internal/core/logger"
)

type CreateUserResponse UserDTOResponse

// Register регистрирует нового пользователя.
//
//	@Summary		Регистрация
//	@Description	Создаёт нового пользователя по номеру телефона и паролю
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserRequest	true	"Данные регистрации"
//	@Success		201		{object}	UserDTOResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/register [post]
func (h *AuthHTTPHandler) Register(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"failed to decode and validate HTTP request": err.Error(),
		})
		return
	}

	user, err := h.authService.Register(
		c.Request.Context(),
		req.Username,
		req.PhoneNumber,
		req.Password,
	)
	if err != nil {
		log := core_logger.FromContext(c.Request.Context())
		switch {
		case errors.Is(err, core_errors.ErrPhoneNumberUse):
			log.Debug("register: phone number already in use", zap.Error(err))
			c.JSON(http.StatusConflict, gin.H{
				"error": "phone number is already in use",
			})
		case errors.Is(err, core_errors.ErrInvalidArgument):
			log.Debug("register: invalid argument", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:
			log.Error("register: unexpected error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"failed to create user": err.Error(),
			})
		}
		return
	}
	resp := UserDTOResponse{
		ID:          user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
	}

	c.JSON(http.StatusCreated, resp)
}

// Login авторизует пользователя и выдаёт access/refresh токены.
//
//	@Summary		Вход
//	@Description	Авторизует пользователя по номеру телефона и паролю
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Данные входа"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/login [post]
func (h *AuthHTTPHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "Invalid request payload",
		})
		c.Abort()
		return
	}

	refreshTokenTTL := 7 * 24 * time.Hour

	accessToken, refreshToken, err := h.authService.LoginWithRefresh(
		c.Request.Context(),
		req.PhoneNumber,
		req.Password,
		refreshTokenTTL,
	)
	if err != nil {
		log := core_logger.FromContext(c.Request.Context())
		if errors.Is(err, core_errors.ErrInvalidCredentials) {
			log.Debug("login: invalid credentials", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			c.Abort()
			return
		} else {
			log.Error("login: unexpected error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			c.Abort()
			return
		}
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}
