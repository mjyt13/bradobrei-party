package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// GetAll godoc
// @Summary Список платежей
// @Description Возвращает список платежей по бронированиям.
// @Tags payments
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Payment
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments [get]
func (h *PaymentHandler) GetAll(c *gin.Context) {
	payments, err := h.paymentService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// GetByID godoc
// @Summary Платёж по ID
// @Description Возвращает один платёж по идентификатору.
// @Tags payments
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID платежа"
// @Success 200 {object} models.Payment
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	payment, err := h.paymentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Платёж не найден",
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// Create godoc
// @Summary Создать платёж
// @Description Создаёт платёж для существующего бронирования.
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePaymentRequest true "Данные платежа"
// @Success 201 {object} models.Payment
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments [post]
func (h *PaymentHandler) Create(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	payment, err := h.paymentService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "payment_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, payment)
}
