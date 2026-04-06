package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService *services.BookingService
}

func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// Create godoc
// @Summary Создать бронирование
// @Description Основа пользовательского сценария записи клиента. Используется в отчётах 2.2.2, 2.2.4, 2.2.7 и 2.2.8.
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBookingRequest true "Данные бронирования"
// @Success 201 {object} models.Booking
// @Failure 400 {object} dto.ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) Create(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	booking, err := h.bookingService.Create(req, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "booking_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetAll godoc
// @Summary Список бронирований
// @Description Возвращает все бронирования для административного и аналитического просмотра.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Booking
// @Failure 500 {object} dto.ErrorResponse
// @Router /bookings [get]
func (h *BookingHandler) GetAll(c *gin.Context) {
	bookings, err := h.bookingService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// GetByID godoc
// @Summary Бронирование по ID
// @Description Возвращает одно бронирование со связанными услугами и платёжом.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} models.Booking
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	booking, err := h.bookingService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Бронирование не найдено",
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// GetMy godoc
// @Summary Мои бронирования
// @Description Возвращает бронирования текущего клиента.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Booking
// @Failure 500 {object} dto.ErrorResponse
// @Router /bookings/my [get]
func (h *BookingHandler) GetMy(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	bookings, err := h.bookingService.GetByClient(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// GetByMaster godoc
// @Summary Бронирования мастера
// @Description Рабочий эндпоинт мастера для просмотра назначенных записей.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Booking
// @Failure 500 {object} dto.ErrorResponse
// @Router /bookings/master [get]
func (h *BookingHandler) GetByMaster(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	bookings, err := h.bookingService.GetByMaster(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// Confirm godoc
// @Summary Подтвердить бронирование
// @Description Подтверждает бронирование и списывает материалы.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} models.Booking
// @Failure 400 {object} dto.ErrorResponse
// @Router /bookings/{id}/confirm [post]
func (h *BookingHandler) Confirm(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	booking, err := h.bookingService.Confirm(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "confirm_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// Cancel godoc
// @Summary Отменить бронирование
// @Description Отменяет бронирование по правилам роли и владельца записи.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) Cancel(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.bookingService.Cancel(uint(id), claims.UserID, claims.Role); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "cancel_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Бронирование отменено"})
}

// Delete godoc
// @Summary Удалить бронирование
// @Description Полностью удаляет бронирование и связанные позиции/платёж. Клиент может удалить только свою запись, администратор - любую.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Router /bookings/{id} [delete]
func (h *BookingHandler) Delete(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.bookingService.Delete(uint(id), claims.UserID, claims.Role); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "delete_failed",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Бронирование удалено"})
}
