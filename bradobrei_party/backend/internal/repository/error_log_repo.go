package repository

import (
	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type ErrorLogRepository struct {
	db *gorm.DB
}

func NewErrorLogRepository(db *gorm.DB) *ErrorLogRepository {
	return &ErrorLogRepository{db: db}
}

func (r *ErrorLogRepository) Create(entry *models.ErrorLog) error {
	return r.db.Create(entry).Error
}

func (r *ErrorLogRepository) GetAll(limit int) ([]models.ErrorLog, error) {
	var logs []models.ErrorLog

	query := r.db.Preload("User").Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}
