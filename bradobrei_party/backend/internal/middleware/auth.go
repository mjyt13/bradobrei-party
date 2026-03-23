package middleware

import (
	"net/http"
	"strings"

	"bradobrei/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserContextKey = "currentUser"

// Claims — JWT payload
type Claims struct {
	UserID uint            `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthRequired — проверяет JWT токен, кладёт user в контекст
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized", Code: 401, Message: "Отсутствует токен авторизации",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized", Code: 401, Message: "Недействительный токен",
			})
			return
		}

		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// RequireRoles — RBAC: пропускает только указанные роли
func RequireRoles(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		claims, ok := c.MustGet(UserContextKey).(*Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized", Code: 401,
			})
			return
		}

		if !allowed[claims.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error: "forbidden", Code: 403, Message: "Недостаточно прав для данного действия",
			})
			return
		}

		c.Next()
	}
}

// GetCurrentClaims — хелпер для получения claims из контекста
func GetCurrentClaims(c *gin.Context) (*Claims, bool) {
	v, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	claims, ok := v.(*Claims)
	return claims, ok
}
