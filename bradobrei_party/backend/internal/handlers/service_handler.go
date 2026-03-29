package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceService *services.ServiceService
}

func NewServiceHandler(serviceService *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

// GetAll godoc
// @Summary Список услуг
// @Description Возвращает все услуги сети.
// @Tags services
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Service
// @Failure 500 {object} dto.ErrorResponse
// @Router /services [get]
func (h *ServiceHandler) GetAll(c *gin.Context) {
	list, err := h.serviceService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetByID godoc
// @Summary Услуга по ID
// @Description Возвращает одну услугу по идентификатору.
// @Tags services
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID услуги"
// @Success 200 {object} models.Service
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /services/{id} [get]
func (h *ServiceHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	svc, err := h.serviceService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Услуга не найдена",
		})
		return
	}
	c.JSON(http.StatusOK, svc)
}

// GetMy godoc
// @Summary Мои услуги
// @Description Возвращает услуги, привязанные к текущему мастеру.
// @Tags services
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Service
// @Failure 500 {object} dto.ErrorResponse
// @Router /services/my [get]
func (h *ServiceHandler) GetMy(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	list, err := h.serviceService.GetByMaster(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Create godoc
// @Summary Создать услугу
// @Description Создаёт новую услугу.
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Service true "Данные услуги"
// @Success 201 {object} models.Service
// @Failure 400 {object} dto.ErrorResponse
// @Router /services [post]
func (h *ServiceHandler) Create(c *gin.Context) {
	var svc models.Service
	if err := c.ShouldBindJSON(&svc); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	if err := h.serviceService.Create(&svc); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, svc)
}

// Update godoc
// @Summary Обновить услугу
// @Description Обновляет существующую услугу.
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID услуги"
// @Param request body models.Service true "Обновлённые данные услуги"
// @Success 200 {object} models.Service
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /services/{id} [put]
func (h *ServiceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	existing, err := h.serviceService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Услуга не найдена",
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
	if err := h.serviceService.Update(existing); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// Delete godoc
// @Summary Удалить услугу
// @Description Удаляет услугу.
// @Tags services
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID услуги"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /services/{id} [delete]
func (h *ServiceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	if err := h.serviceService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Услуга удалена"})
}

// AssignToMaster godoc
// @Summary Назначить услугу мастеру
// @Description Привязывает услугу к мастеру.
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID услуги"
// @Param request body dto.AssignServiceToMasterRequest true "Пользователь-мастер"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /services/{id}/assign-master [post]
func (h *ServiceHandler) AssignToMaster(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	var req dto.AssignServiceToMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	if err := h.serviceService.AddToMaster(
		claims.UserID, claims.Role, req.TargetUserID, uint(serviceID),
	); err != nil {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Code:    403,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Услуга добавлена мастеру"})
}

// RemoveFromMaster godoc
// @Summary Убрать услугу у мастера
// @Description Отвязывает услугу от профиля мастера.
// @Tags services
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID услуги"
// @Param profileId path int true "ID профиля мастера"
// @Success 200 {object} map[string]string
// @Failure 500 {object} dto.ErrorResponse
// @Router /services/{id}/assign-master/{profileId} [delete]
func (h *ServiceHandler) RemoveFromMaster(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	profileID, _ := strconv.ParseUint(c.Param("profileId"), 10, 64)

	if err := h.serviceService.RemoveFromMaster(uint(profileID), uint(serviceID)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Услуга убрана у мастера"})
}
