package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type EmployeeHandler struct {
	employeeService *services.EmployeeService
}

func NewEmployeeHandler(employeeService *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: employeeService}
}

// GET /api/v1/employees  (ADMIN, HR, NETWORK_MANAGER)
func (h *EmployeeHandler) GetAll(c *gin.Context) {
	list, err := h.employeeService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GET /api/v1/employees/:id  (ADMIN, HR, NETWORK_MANAGER)
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}
	profile, err := h.employeeService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Профиль сотрудника не найден",
		})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// GET /api/v1/employees/me  — свой профиль (любой сотрудник)
func (h *EmployeeHandler) GetMe(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)
	profile, err := h.employeeService.GetMyProfile(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "not_found", Code: 404, Message: "Профиль не найден",
		})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// POST /api/v1/employees  — нанять сотрудника (HR, ADMIN)
// ТЗ 2.3.4: HR создаёт нового сотрудника в системе
func (h *EmployeeHandler) Hire(c *gin.Context) {
	var req models.HireEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	// Хешируем пароль здесь, передаём уже готовый хеш в сервис
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	req.PasswordHash = string(hash)

	profile, err := h.employeeService.HireEmployee(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "hire_failed", Code: 400, Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, profile)
}

// PATCH /api/v1/employees/me/schedule  — мастер меняет своё расписание (ТЗ 2.3.2)
// body: {"schedule": "{\"mon\":\"9-18\",\"tue\":\"9-18\"}"}
func (h *EmployeeHandler) UpdateMySchedule(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	var body struct {
		Schedule string `json:"schedule" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if err := h.employeeService.UpdateSchedule(claims.UserID, body.Schedule); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "update_failed", Code: 400, Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Расписание обновлено"})
}

// POST /api/v1/employees/:id/assign-salon  (ADMIN, NETWORK_MANAGER)
// body: {"salon_id": 2}
func (h *EmployeeHandler) AssignToSalon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	var body struct {
		SalonID uint `json:"salon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	if err := h.employeeService.AssignToSalon(uint(id), body.SalonID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Сотрудник прикреплён к салону"})
}

// DELETE /api/v1/employees/:id/assign-salon/:salonId  (ADMIN, NETWORK_MANAGER)
func (h *EmployeeHandler) RemoveFromSalon(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	salonID, _ := strconv.ParseUint(c.Param("salonId"), 10, 64)

	if err := h.employeeService.RemoveFromSalon(uint(id), uint(salonID)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Сотрудник откреплён от салона"})
}
