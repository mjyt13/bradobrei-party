package services

import (
	"errors"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"

	"gorm.io/gorm"
)

type ServiceService struct {
	serviceRepo  *repository.ServiceRepository
	employeeRepo *repository.EmployeeRepository
	db           *gorm.DB
}

func NewServiceService(
	serviceRepo *repository.ServiceRepository,
	employeeRepo *repository.EmployeeRepository,
	_ *repository.InventoryRepository,
	db *gorm.DB,
) *ServiceService {
	return &ServiceService{
		serviceRepo:  serviceRepo,
		employeeRepo: employeeRepo,
		db:           db,
	}
}

func (s *ServiceService) GetAll() ([]models.Service, error) {
	return s.serviceRepo.GetAll()
}

func (s *ServiceService) GetByID(id uint) (*models.Service, error) {
	return s.serviceRepo.GetByID(id)
}

func (s *ServiceService) Create(svc *models.Service) error {
	// ТЗ 2.3.7: минимальная длительность — 60 минут
	if svc.DurationMinutes < 60 {
		return errors.New("длительность услуги должна быть не менее 60 минут")
	}
	if svc.Price <= 0 {
		return errors.New("стоимость услуги должна быть больше нуля")
	}
	return s.serviceRepo.Create(svc)
}

func (s *ServiceService) Update(svc *models.Service) error {
	if svc.DurationMinutes < 60 {
		return errors.New("длительность услуги должна быть не менее 60 минут")
	}
	return s.serviceRepo.Update(svc)
}

func (s *ServiceService) Delete(id uint) error {
	return s.serviceRepo.Delete(id)
}

func (s *ServiceService) GetByMaster(masterID uint) ([]models.Service, error) {
	return s.serviceRepo.GetByMaster(masterID)
}

// AddToMaster — ADVANCED_MASTER может добавлять услуги себе и другим (ТЗ 2.3.3)
func (s *ServiceService) AddToMaster(requesterID uint, requesterRole models.UserRole, targetUserID uint, serviceID uint) error {
	// BASIC_MASTER может только запрашивать изменение у себя — здесь просто добавляем
	// ADVANCED_MASTER — может менять у любого
	if requesterRole == models.RoleBasicMaster && requesterID != targetUserID {
		return errors.New("недостаточно прав: вы можете изменять только свой список услуг")
	}

	profile, err := s.employeeRepo.GetByUserID(targetUserID)
	if err != nil {
		return errors.New("профиль сотрудника не найден")
	}

	return s.serviceRepo.AddToMaster(profile.ID, serviceID)
}

// RemoveFromMaster — только ADVANCED_MASTER или Admin (ТЗ 2.3.3)
func (s *ServiceService) RemoveFromMaster(profileID uint, serviceID uint) error {
	return s.serviceRepo.RemoveFromMaster(profileID, serviceID)
}

// UseService — ручное списание материалов по услуге в конкретном салоне.
func (s *ServiceService) UseService(serviceID, salonID uint, quantity int, userID uint) error {
	if quantity <= 0 {
		return errors.New("количество использований услуги должно быть больше нуля")
	}

	service, err := s.serviceRepo.GetByID(serviceID)
	if err != nil {
		return errors.New("услуга не найдена")
	}

	if len(service.Materials) == 0 {
		return errors.New("для услуги не задана норма расхода материалов")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		inventoryRepo := repository.NewInventoryRepository(tx)
		usageRepo := repository.NewServiceUsageRepository(tx)

		if err := inventoryRepo.WriteOff(salonID, service.Materials, quantity); err != nil {
			return errors.New("не удалось списать материалы со склада: " + err.Error())
		}

		return usageRepo.Create(&models.ServiceUsage{
			ServiceID: serviceID,
			SalonID:   salonID,
			UserID:    userID,
			Quantity:  quantity,
		})
	})
}
