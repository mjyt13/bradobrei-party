package services

import (
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type ErrorLogService struct {
	repo *repository.ErrorLogRepository
}

func NewErrorLogService(repo *repository.ErrorLogRepository) *ErrorLogService {
	return &ErrorLogService{repo: repo}
}

func (s *ErrorLogService) GetAll(limit int) ([]models.ErrorLog, error) {
	return s.repo.GetAll(limit)
}
