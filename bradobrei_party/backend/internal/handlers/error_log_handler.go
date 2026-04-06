package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ErrorLogHandler struct {
	service *services.ErrorLogService
}

func NewErrorLogHandler(service *services.ErrorLogService) *ErrorLogHandler {
	return &ErrorLogHandler{service: service}
}

// GetAll godoc
// @Summary Журнал ошибок и исключительных ситуаций
// @Tags error-logs
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Ограничение по количеству записей"
// @Success 200 {array} models.ErrorLog
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /error-logs [get]
func (h *ErrorLogHandler) GetAll(c *gin.Context) {
	limit := 0
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "bad_request",
				Code:    400,
				Message: "limit должен быть неотрицательным целым числом",
			})
			return
		}
		limit = parsed
	}

	logs, err := h.service.GetAll(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal",
			Code:    500,
			Message: "Не удалось загрузить журнал ошибок",
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}
