package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService *services.BookingService
}

func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// POST /api/v1/bookings
func (h *BookingHandler) Create(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	booking, err := h.bookingService.Create(req, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "booking_failed", Code: 400, Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GET /api/v1/bookings/:id
func (h *BookingHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	booking, err := h.bookingService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Бронирование не найдено",
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// GET /api/v1/bookings/my
func (h *BookingHandler) GetMy(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	bookings, err := h.bookingService.GetByClient(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// GET /api/v1/bookings/master
func (h *BookingHandler) GetByMaster(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	bookings, err := h.bookingService.GetByMaster(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// POST /api/v1/bookings/:id/confirm
func (h *BookingHandler) Confirm(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	booking, err := h.bookingService.Confirm(uint(id), claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "confirm_failed", Code: 400, Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// POST /api/v1/bookings/:id/cancel
func (h *BookingHandler) Cancel(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.bookingService.Cancel(uint(id), claims.UserID, claims.Role); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "cancel_failed", Code: 400, Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Бронирование отменено"})
}
