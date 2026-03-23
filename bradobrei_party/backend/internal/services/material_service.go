package services

import (
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type MaterialService struct {
	materialRepo *repository.MaterialRepository
}

func NewMaterialService(materialRepo *repository.MaterialRepository) *MaterialService {
	return &MaterialService{materialRepo: materialRepo}
}

func (s *MaterialService) GetAll() ([]models.Material, error) {
	return s.materialRepo.GetAll()
}

func (s *MaterialService) GetByID(id uint) (*models.Material, error) {
	return s.materialRepo.GetByID(id)
}

func (s *MaterialService) Create(m *models.Material) error {
	return s.materialRepo.Create(m)
}

func (s *MaterialService) Update(m *models.Material) error {
	return s.materialRepo.Update(m)
}

func (s *MaterialService) Delete(id uint) error {
	return s.materialRepo.Delete(id)
}

// SetServiceMaterials — задать норму расхода материалов для услуги
func (s *MaterialService) SetServiceMaterials(serviceID uint, items []models.ServiceMaterial) error {
	return s.materialRepo.SetServiceMaterials(serviceID, items)
}
