package services

import (
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type SalonService struct {
	salonRepo *repository.SalonRepository
}

func NewSalonService(salonRepo *repository.SalonRepository) *SalonService {
	return &SalonService{salonRepo: salonRepo}
}

func (s *SalonService) GetAll() ([]models.Salon, error) {
	return s.salonRepo.GetAll()
}

func (s *SalonService) GetByID(id uint) (*models.Salon, error) {
	return s.salonRepo.GetByID(id)
}

func (s *SalonService) Create(salon *models.Salon) error {
	return s.salonRepo.Create(salon)
}

func (s *SalonService) Update(salon *models.Salon) error {
	return s.salonRepo.Update(salon)
}

func (s *SalonService) Delete(id uint) error {
	return s.salonRepo.Delete(id)
}

func (s *SalonService) GetMasters(salonID uint) ([]models.EmployeeProfile, error) {
	return s.salonRepo.GetMasters(salonID)
}
