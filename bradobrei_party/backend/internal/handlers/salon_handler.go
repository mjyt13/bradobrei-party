package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
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

// GetAll godoc
// @Summary Список салонов
// @Description Базовый справочный эндпоинт для просмотра доступных филиалов и аналитики 2.2.2.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Salon
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons [get]
func (h *SalonHandler) GetAll(c *gin.Context) {
	salons, err := h.salonService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, salons)
}

// GetByID godoc
// @Summary Салон по ID
// @Description Возвращает один салон с базовой информацией.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID салона"
// @Success 200 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /salons/{id} [get]
func (h *SalonHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	salon, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Салон не найден",
		})
		return
	}
	c.JSON(http.StatusOK, salon)
}

// GetMasters godoc
// @Summary Мастера салона
// @Description Возвращает мастеров, закреплённых за выбранным салоном.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID салона"
// @Success 200 {array} models.EmployeeProfile
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id}/masters [get]
func (h *SalonHandler) GetMasters(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	masters, err := h.salonService.GetMasters(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, masters)
}

// Create godoc
// @Summary Создать салон
// @Description Создаёт новый салон.
// @Tags salons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Salon true "Данные салона"
// @Success 201 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons [post]
func (h *SalonHandler) Create(c *gin.Context) {
	var salon models.Salon
	if err := c.ShouldBindJSON(&salon); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Create(&salon); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusCreated, salon)
}

// Update godoc
// @Summary Обновить салон
// @Description Обновляет данные существующего салона.
// @Tags salons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID салона"
// @Param request body models.Salon true "Обновлённые данные салона"
// @Success 200 {object} models.Salon
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id} [put]
func (h *SalonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	existing, err := h.salonService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Салон не найден",
		})
		return
	}

	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.salonService.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// Delete godoc
// @Summary Удалить салон
// @Description Удаляет салон из системы.
// @Tags salons
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID салона"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /salons/{id} [delete]
func (h *SalonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	if err := h.salonService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Салон удалён"})
}
