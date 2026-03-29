package middleware

import (
	"net/http"
	"strings"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserContextKey = "currentUser"

// Claims stores the JWT payload used across protected endpoints.
type Claims struct {
	UserID uint            `json:"user_id"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func normalizeRole(role models.UserRole) models.UserRole {
	switch role {
	case "ADMIN":
		return models.RoleAdmin
	default:
		return role
	}
}

// ExtractToken supports both "Bearer <token>" and a raw JWT for local dev convenience.
func ExtractToken(authHeader string) (string, bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", false
	}

	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[len("Bearer "):])
		return token, token != ""
	}

	return authHeader, true
}

// AuthRequired validates JWT and stores parsed claims in the Gin context.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := ExtractToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Code:    401,
				Message: "Отсутствует токен авторизации",
			})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Code:    401,
				Message: "Недействительный токен",
			})
			return
		}

		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// RequireRoles enforces RBAC for protected endpoints.
func RequireRoles(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		claims, ok := c.MustGet(UserContextKey).(*Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: "unauthorized",
				Code:  401,
			})
			return
		}

		if !allowed[normalizeRole(claims.Role)] {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Code:    403,
				Message: "Недостаточно прав для данного действия",
			})
			return
		}

		c.Next()
	}
}

// GetCurrentClaims returns parsed JWT claims from the Gin context.
func GetCurrentClaims(c *gin.Context) (*Claims, bool) {
	v, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}

	claims, ok := v.(*Claims)
	return claims, ok
}
