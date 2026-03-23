package handlers

import (
	"net/http"
	"time"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// parsePeriod — читает ?from=&to= из query params
func parsePeriod(c *gin.Context) (time.Time, time.Time, error) {
	var from, to time.Time
	var err error

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return from, to, err
		}
	} else {
		from = time.Now().AddDate(0, -1, 0) // по умолчанию — прошлый месяц
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return from, to, err
		}
	} else {
		to = time.Now()
	}

	return from, to, nil
}

// GET /api/v1/reports/employees
// Доступ: ADMIN, HR, NETWORK_MANAGER
func (h *ReportHandler) Employees(c *gin.Context) {
	data, err := h.reportService.GetEmployeeList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": "employee_list", "data": data})
}

// GET /api/v1/reports/salon-activity?from=2026-01-01&to=2026-01-31
// Доступ: ADMIN, ACCOUNTANT, NETWORK_MANAGER
func (h *ReportHandler) SalonActivity(c *gin.Context) {
	from, to, err := parsePeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: "Формат даты: YYYY-MM-DD",
		})
		return
	}

	data, err := h.reportService.GetSalonActivity(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": "salon_activity", "period": gin.H{"from": from, "to": to}, "data": data})
}

// GET /api/v1/reports/service-popularity?from=&to=
// Доступ: ADMIN, ACCOUNTANT, NETWORK_MANAGER
func (h *ReportHandler) ServicePopularity(c *gin.Context) {
	from, to, err := parsePeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: "Формат даты: YYYY-MM-DD",
		})
		return
	}

	data, err := h.reportService.GetServicePopularity(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": "service_popularity", "period": gin.H{"from": from, "to": to}, "data": data})
}

// GET /api/v1/reports/master-activity?from=&to=
// Доступ: ADMIN, ACCOUNTANT, NETWORK_MANAGER
func (h *ReportHandler) MasterActivity(c *gin.Context) {
	from, to, err := parsePeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: "Формат даты: YYYY-MM-DD",
		})
		return
	}

	data, err := h.reportService.GetMasterActivity(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": "master_activity", "period": gin.H{"from": from, "to": to}, "data": data})
}

// GET /api/v1/reports/reviews?from=&to=
// Доступ: ADMIN
func (h *ReportHandler) Reviews(c *gin.Context) {
	from, to, _ := parsePeriod(c) // нулевые значения допустимы — берёт все

	data, err := h.reportService.GetReviews(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": "reviews", "data": data})
}
