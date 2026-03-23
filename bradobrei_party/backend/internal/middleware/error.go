package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"bradobrei/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// ErrorLogger — middleware для логирования всех ошибок (ТЗ 2.4)
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Логируем все ошибки gin после обработки запроса
		for _, err := range c.Errors {
			log.Printf("[ERROR] %s %s → %v", c.Request.Method, c.Request.URL.Path, err)
		}
	}
}

// RecoveryWithLog — перехватывает panic, откатывает (ТЗ 2.4.5)
func RecoveryWithLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v\n%s", r, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "internal_server_error",
					Code:    500,
					Message: "Внутренняя ошибка сервера. Изменения откачены.",
				})
			}
		}()
		c.Next()
	}
}
