package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		if rawQuery != "" {
			path += "?" + rawQuery
		}

		log.Printf(
			"[HTTP] %s %s -> %d (%s) ip=%s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start).Round(time.Millisecond),
			c.ClientIP(),
		)
	}
}

func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			log.Printf("[ERROR] %s %s -> %v", c.Request.Method, c.Request.URL.Path, err)
		}
	}
}

func RecoveryWithLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v\n%s", r, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error:   "internal_server_error",
					Code:    500,
					Message: "Внутренняя ошибка сервера. Изменения откатены.",
				})
			}
		}()
		c.Next()
	}
}

func ErrorJournal(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = writer

		c.Next()

		statusCode := c.Writer.Status()
		if statusCode < 400 {
			return
		}

		entry := models.ErrorLog{
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      c.Request.URL.RawQuery,
			StatusCode: statusCode,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}

		if claims, ok := GetCurrentClaims(c); ok {
			userID := claims.UserID
			userRole := claims.Role
			entry.UserID = &userID
			entry.UserRole = &userRole
		}

		var payload dto.ErrorResponse
		if err := json.Unmarshal(writer.body.Bytes(), &payload); err == nil {
			entry.ErrorCode = payload.Error
			entry.Message = payload.Message
		}

		if entry.ErrorCode == "" {
			entry.ErrorCode = http.StatusText(statusCode)
		}

		if entry.Message == "" {
			if len(c.Errors) > 0 {
				entry.Message = c.Errors.String()
			} else {
				entry.Message = strings.TrimSpace(writer.body.String())
			}
		}

		if err := db.Session(&gorm.Session{}).Create(&entry).Error; err != nil {
			log.Printf("[ERROR_LOG] failed to persist error journal entry: %v", err)
		}
	}
}
