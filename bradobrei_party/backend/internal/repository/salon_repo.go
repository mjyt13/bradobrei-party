package repository

import (
	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type SalonRepository struct {
	db *gorm.DB
}

func NewSalonRepository(db *gorm.DB) *SalonRepository {
	return &SalonRepository{db: db}
}

func (r *SalonRepository) Create(s *models.Salon) error {
	return r.db.Create(s).Error
}

func (r *SalonRepository) GetByID(id uint) (*models.Salon, error) {
	var s models.Salon
	err := r.db.Preload("Employees.User").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SalonRepository) GetAll() ([]models.Salon, error) {
	var salons []models.Salon
	err := r.db.Where("status = ?", "OPEN").Find(&salons).Error
	return salons, err
}

func (r *SalonRepository) Update(s *models.Salon) error {
	return r.db.Save(s).Error
}

func (r *SalonRepository) Delete(id uint) error {
	return r.db.Delete(&models.Salon{}, id).Error
}

// GetMasters — мастера конкретного салона (для клиента ТЗ 2.3.7)
func (r *SalonRepository) GetMasters(salonID uint) ([]models.EmployeeProfile, error) {
	var profiles []models.EmployeeProfile
	err := r.db.
		Joins("JOIN employee_salons ON employee_salons.employee_profile_id = employee_profiles.id").
		Where("employee_salons.salon_id = ?", salonID).
		Preload("User").
		Preload("Services").
		Find(&profiles).Error
	return profiles, err
}
