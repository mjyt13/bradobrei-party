package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type MaterialHandler struct {
	materialService *services.MaterialService
}

func NewMaterialHandler(materialService *services.MaterialService) *MaterialHandler {
	return &MaterialHandler{materialService: materialService}
}

// GetAll godoc
// @Summary Список материалов
// @Description Возвращает все материалы.
// @Tags materials
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Material
// @Failure 500 {object} dto.ErrorResponse
// @Router /materials [get]
func (h *MaterialHandler) GetAll(c *gin.Context) {
	list, err := h.materialService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetByID godoc
// @Summary Материал по ID
// @Description Возвращает один материал по идентификатору.
// @Tags materials
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID материала"
// @Success 200 {object} models.Material
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /materials/{id} [get]
func (h *MaterialHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	m, err := h.materialService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Материал не найден",
		})
		return
	}
	c.JSON(http.StatusOK, m)
}

// Create godoc
// @Summary Создать материал
// @Description Создаёт новый материал.
// @Tags materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Material true "Данные материала"
// @Success 201 {object} models.Material
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /materials [post]
func (h *MaterialHandler) Create(c *gin.Context) {
	var m models.Material
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if err := h.materialService.Create(&m); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusCreated, m)
}

// Update godoc
// @Summary Обновить материал
// @Description Обновляет существующий материал.
// @Tags materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID материала"
// @Param request body models.Material true "Обновлённые данные материала"
// @Success 200 {object} models.Material
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /materials/{id} [put]
func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	existing, err := h.materialService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Материал не найден",
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
	existing.ID = uint(id)
	if err := h.materialService.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// Delete godoc
// @Summary Удалить материал
// @Description Удаляет материал.
// @Tags materials
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID материала"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /materials/{id} [delete]
func (h *MaterialHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	if err := h.materialService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Материал удалён"})
}

// SetServiceMaterials godoc
// @Summary Установить норму расхода для услуги
// @Description Назначает список материалов и норм расхода для выбранной услуги.
// @Tags materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param serviceId path int true "ID услуги"
// @Param request body []models.ServiceMaterial true "Материалы услуги"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /materials/service/{serviceId} [put]
func (h *MaterialHandler) SetServiceMaterials(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	var items []models.ServiceMaterial
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.materialService.SetServiceMaterials(uint(serviceID), items); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Норма расхода обновлена"})
}
