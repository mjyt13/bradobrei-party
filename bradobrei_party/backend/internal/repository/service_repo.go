package repository

import (
	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) GetAll() ([]models.Service, error) {
	var services []models.Service
	err := r.db.Preload("Materials.Material").Find(&services).Error
	return services, err
}

func (r *ServiceRepository) GetByID(id uint) (*models.Service, error) {
	var s models.Service
	err := r.db.Preload("Materials.Material").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServiceRepository) Create(s *models.Service) error {
	return r.db.Create(s).Error
}

func (r *ServiceRepository) Update(s *models.Service) error {
	return r.db.Save(s).Error
}

func (r *ServiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}

// GetByMaster — услуги конкретного мастера
func (r *ServiceRepository) GetByMaster(masterID uint) ([]models.Service, error) {
	var services []models.Service
	err := r.db.
		Joins("JOIN employee_services ON employee_services.service_id = services.id").
		Joins("JOIN employee_profiles ON employee_profiles.id = employee_services.employee_profile_id").
		Where("employee_profiles.user_id = ?", masterID).
		Preload("Materials.Material").
		Find(&services).Error
	return services, err
}

// AddToMaster — добавить услугу мастеру
func (r *ServiceRepository) AddToMaster(profileID uint, serviceID uint) error {
	return r.db.Exec(
		"INSERT INTO employee_services (employee_profile_id, service_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		profileID, serviceID,
	).Error
}

// RemoveFromMaster — убрать услугу у мастера
func (r *ServiceRepository) RemoveFromMaster(profileID uint, serviceID uint) error {
	return r.db.Exec(
		"DELETE FROM employee_services WHERE employee_profile_id = ? AND service_id = ?",
		profileID, serviceID,
	).Error
}
