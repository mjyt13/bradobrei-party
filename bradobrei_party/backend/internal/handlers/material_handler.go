package handlers

import (
	"net/http"
	"strconv"

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

// GET /api/v1/materials
func (h *MaterialHandler) GetAll(c *gin.Context) {
	list, err := h.materialService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GET /api/v1/materials/:id
func (h *MaterialHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	m, err := h.materialService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Материал не найден",
		})
		return
	}
	c.JSON(http.StatusOK, m)
}

// POST /api/v1/materials  (ADMIN)
func (h *MaterialHandler) Create(c *gin.Context) {
	var m models.Material
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}
	if err := h.materialService.Create(&m); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusCreated, m)
}

// PUT /api/v1/materials/:id  (ADMIN)
func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	existing, err := h.materialService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Материал не найден",
		})
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}
	existing.ID = uint(id)
	if err := h.materialService.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DELETE /api/v1/materials/:id  (ADMIN)
func (h *MaterialHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	if err := h.materialService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Материал удалён"})
}

// PUT /api/v1/materials/service/:serviceId  (ADMIN, ADVANCED_MASTER)
// Установить норму расхода для услуги
// body: [{"material_id": 1, "quantity_per_use": 50.5}, ...]
func (h *MaterialHandler) SetServiceMaterials(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("serviceId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	var items []models.ServiceMaterial
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if err := h.materialService.SetServiceMaterials(uint(serviceID), items); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Норма расхода обновлена"})
}
