package handlers

import (
	"net/http"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя в системе.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные регистрации"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "registration_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// Login godoc
// @Summary Вход в систему
// @Description Проверяет логин и пароль, возвращает только JWT для дальнейшей авторизации.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Me godoc
// @Summary Текущий пользователь
// @Description Возвращает пользователя, идентифицированного по JWT из заголовка Authorization.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := middleware.GetCurrentClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Code:    401,
			Message: "Не удалось определить пользователя из токена",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{User: *user})
}

// DocsRedirect godoc
// @Summary Перейти к документации
// @Description Перенаправляет на Swagger UI.
// @Tags docs
// @Router / [get]
func (h *AuthHandler) DocsRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "/swagger/index.html")
}
