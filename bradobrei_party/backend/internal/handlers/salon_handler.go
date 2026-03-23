package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SalonHandler struct {
	salonService *services.SalonService
}

func NewSalonHandler(salonService *services.SalonService) *SalonHandler {
	return &SalonHandler{salonService: salonService}
}

// GET /api/v1/salons
func (h *SalonHandler) GetAll(c *gin.Context) {
	salons, err := h.salonService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, salons)
}

// GET /api/v1/salons/:id
func (h *SalonHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	salon, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Салон не найден",
		})
		return
	}
	c.JSON(http.StatusOK, salon)
}

// GET /api/v1/salons/:id/masters
func (h *SalonHandler) GetMasters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	masters, err := h.salonService.GetMasters(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, masters)
}

// POST /api/v1/salons  (ADMIN, NETWORK_MANAGER)
func (h *SalonHandler) Create(c *gin.Context) {
	var salon models.Salon
	if err := c.ShouldBindJSON(&salon); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Create(&salon); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusCreated, salon)
}

// PUT /api/v1/salons/:id  (ADMIN, NETWORK_MANAGER)
func (h *SalonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	existing, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Салон не найден",
		})
		return
	}

	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DELETE /api/v1/salons/:id  (ADMIN only)
func (h *SalonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.salonService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Салон удалён"})
}
