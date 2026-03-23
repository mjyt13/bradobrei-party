package handlers

import (
	"net/http"

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

// POST /api/v1/reviews
func (h *ReviewHandler) Create(c *gin.Context) {
	claims, _ := middleware.GetCurrentClaims(c)

	var req models.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "bad_request", Code: 400, Message: err.Error(),
		})
		return
	}

	review := models.Review{
		UserID: claims.UserID,
		Text:   req.Text,
		Rating: req.Rating,
	}

	if err := h.db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GET /api/v1/reviews
func (h *ReviewHandler) GetAll(c *gin.Context) {
	var reviews []models.Review
	if err := h.db.Preload("User").Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal", Code: 500})
		return
	}
	c.JSON(http.StatusOK, reviews)
}
