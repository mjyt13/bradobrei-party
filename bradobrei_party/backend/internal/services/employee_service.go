package services

import (
	"errors"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type EmployeeService struct {
	employeeRepo *repository.EmployeeRepository
	userRepo     *repository.UserRepository
}

func NewEmployeeService(
	employeeRepo *repository.EmployeeRepository,
	userRepo *repository.UserRepository,
) *EmployeeService {
	return &EmployeeService{employeeRepo: employeeRepo, userRepo: userRepo}
}

func (s *EmployeeService) GetAll() ([]models.EmployeeProfile, error) {
	return s.employeeRepo.GetAll()
}

func (s *EmployeeService) GetByID(id uint) (*models.EmployeeProfile, error) {
	return s.employeeRepo.GetByID(id)
}

func (s *EmployeeService) GetMyProfile(userID uint) (*models.EmployeeProfile, error) {
	return s.employeeRepo.GetByUserID(userID)
}

// HireEmployee — HR создаёт нового сотрудника (ТЗ 2.3.4)
// Создаёт User + EmployeeProfile за одну операцию
func (s *EmployeeService) HireEmployee(req models.HireEmployeeRequest) (*models.EmployeeProfile, error) {
	// Проверить что роль — сотрудник, не клиент
	if req.Role == models.RoleClient {
		return nil, errors.New("нельзя нанять пользователя с ролью CLIENT")
	}

	// Создать пользователя (пароль хешируется в хэндлере перед передачей)
	// Email — указатель, чтобы пустое значение сохранялось как NULL (не нарушает unique constraint)
	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}

	user := &models.User{
		Username:     req.Username,
		PasswordHash: req.PasswordHash,
		FullName:     req.FullName,
		Phone:        req.Phone,
		Email:        emailPtr,
		Role:         req.Role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("не удалось создать пользователя: " + err.Error())
	}

	// WorkSchedule — указатель, чтобы пустое значение не шло в jsonb
	var schedulePtr *string
	if req.WorkSchedule != "" {
		schedulePtr = &req.WorkSchedule
	}

	// Создать профиль сотрудника
	profile := &models.EmployeeProfile{
		UserID:         user.ID,
		Specialization: req.Specialization,
		ExpectedSalary: req.ExpectedSalary,
		WorkSchedule:   schedulePtr,
	}

	if err := s.employeeRepo.Create(profile); err != nil {
		return nil, errors.New("не удалось создать профиль: " + err.Error())
	}

	// Прикрепить к салону если указан
	if req.SalonID > 0 {
		if err := s.employeeRepo.AssignToSalon(profile.ID, req.SalonID); err != nil {
			return nil, err
		}
	}

	// Вернуть с полной информацией
	return s.employeeRepo.GetByID(profile.ID)
}

// UpdateSchedule — мастер запрашивает изменение расписания (ТЗ 2.3.2)
func (s *EmployeeService) UpdateSchedule(userID uint, schedule string) error {
	profile, err := s.employeeRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("профиль сотрудника не найден")
	}
	return s.employeeRepo.UpdateSchedule(profile.ID, schedule)
}

// UpdateProfile — Admin/ADVANCED_MASTER редактирует профиль
func (s *EmployeeService) UpdateProfile(profile *models.EmployeeProfile) error {
	return s.employeeRepo.Update(profile)
}

func (s *EmployeeService) AssignToSalon(profileID, salonID uint) error {
	return s.employeeRepo.AssignToSalon(profileID, salonID)
}

func (s *EmployeeService) RemoveFromSalon(profileID, salonID uint) error {
	return s.employeeRepo.RemoveFromSalon(profileID, salonID)
}
