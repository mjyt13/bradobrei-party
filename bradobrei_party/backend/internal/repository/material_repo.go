package repository

import (
	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type MaterialRepository struct {
	db *gorm.DB
}

func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

func (r *MaterialRepository) GetAll() ([]models.Material, error) {
	var materials []models.Material
	err := r.db.Find(&materials).Error
	return materials, err
}

func (r *MaterialRepository) GetByID(id uint) (*models.Material, error) {
	var m models.Material
	err := r.db.First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MaterialRepository) Create(m *models.Material) error {
	return r.db.Create(m).Error
}

func (r *MaterialRepository) Update(m *models.Material) error {
	return r.db.Save(m).Error
}

func (r *MaterialRepository) Delete(id uint) error {
	return r.db.Delete(&models.Material{}, id).Error
}

// SetServiceMaterials — обновить норму расхода для услуги (транзакция)
func (r *MaterialRepository) SetServiceMaterials(serviceID uint, items []models.ServiceMaterial) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удалить старые нормы
		if err := tx.Where("service_id = ?", serviceID).Delete(&models.ServiceMaterial{}).Error; err != nil {
			return err
		}
		// Вставить новые
		if len(items) > 0 {
			for i := range items {
				items[i].ServiceID = serviceID
			}
			return tx.Create(&items).Error
		}
		return nil
	})
}
