package handlers

import (
	"net/http"
	"strconv"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewHandler struct {
	db *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{db: db}
}

// Create godoc
// @Summary Создать отзыв
// @Description Сохраняет отзыв текущего пользователя о качестве обслуживания.
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReviewRequest true "Данные отзыва"
// @Success 201 {object} models.Review
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /reviews [post]
func (h *ReviewHandler) Create(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	review := models.Review{
		UserID: claims.UserID,
		Text:   req.Text,
		Rating: req.Rating,
	}

	if err := h.db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetAll godoc
// @Summary Список отзывов
// @Description Возвращает все отзывы с данными пользователя.
// @Tags reviews
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Review
// @Failure 500 {object} dto.ErrorResponse
// @Router /reviews [get]
func (h *ReviewHandler) GetAll(c *gin.Context) {
	var reviews []models.Review
	if err := h.db.Preload("User").Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

// GetByID godoc
// @Summary Отзыв по ID
// @Description Возвращает один отзыв с данными автора.
// @Tags reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID отзыва"
// @Success 200 {object} models.Review
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /reviews/{id} [get]
func (h *ReviewHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Code: 400})
		return
	}

	var review models.Review
	if err := h.db.Preload("User").First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Code:    404,
			Message: "Отзыв не найден",
		})
		return
	}

	c.JSON(http.StatusOK, review)
}
