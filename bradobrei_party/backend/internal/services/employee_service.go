package services

import (
	"errors"
	"fmt"
	"strings"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"

	"gorm.io/gorm"
)

type EmployeeService struct {
	employeeRepo *repository.EmployeeRepository
	userRepo     *repository.UserRepository
	db           *gorm.DB
}

func NewEmployeeService(
	employeeRepo *repository.EmployeeRepository,
	userRepo *repository.UserRepository,
	db *gorm.DB,
) *EmployeeService {
	return &EmployeeService{employeeRepo: employeeRepo, userRepo: userRepo, db: db}
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

func (s *EmployeeService) HireEmployee(req dto.HireEmployeeRequest) (*models.EmployeeProfile, error) {
	if req.Role == models.RoleClient {
		return nil, errors.New("нельзя нанять пользователя с ролью CLIENT")
	}
	if req.ExpectedSalary <= 0 {
		return nil, errors.New("expected_salary должен быть больше 0")
	}

	if req.WorkSchedule != "" {
		normalized, err := normalizeScheduleJSON(req.WorkSchedule, "work_schedule")
		if err != nil {
			return nil, err
		}
		if normalized != nil {
			req.WorkSchedule = *normalized
		}
	}

	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("пользователь с таким логином уже существует")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.Email != "" {
		if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
			return nil, errors.New("пользователь с таким email уже зарегистрирован")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

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

	var schedulePtr *string
	if req.WorkSchedule != "" {
		schedulePtr = &req.WorkSchedule
	}

	profile := &models.EmployeeProfile{
		Specialization: req.Specialization,
		ExpectedSalary: req.ExpectedSalary,
		WorkSchedule:   schedulePtr,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		ur := repository.NewUserRepository(tx)
		er := repository.NewEmployeeRepository(tx)

		if err := ur.Create(user); err != nil {
			return hireUserCreateError(err)
		}

		profile.UserID = user.ID
		if err := er.Create(profile); err != nil {
			return fmt.Errorf("не удалось создать профиль: %w", err)
		}

		if req.SalonID > 0 {
			if err := er.AssignToSalon(profile.ID, req.SalonID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.employeeRepo.GetByID(profile.ID)
}

func hireUserCreateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "email") {
			return errors.New("пользователь с таким email уже зарегистрирован")
		}
		return errors.New("пользователь с таким логином уже существует")
	}
	s := err.Error()
	if strings.Contains(s, "23505") {
		low := strings.ToLower(s)
		if strings.Contains(low, "email") || strings.Contains(s, "uni_users_email") {
			return errors.New("пользователь с таким email уже зарегистрирован")
		}
		return errors.New("пользователь с таким логином уже существует")
	}
	return fmt.Errorf("не удалось создать пользователя: %w", err)
}

func (s *EmployeeService) UpdateSchedule(userID uint, schedule string) error {
	profile, err := s.employeeRepo.GetByUserID(userID)
	if err != nil {
		return errors.New("профиль сотрудника не найден")
	}

	normalized, err := normalizeScheduleJSON(schedule, "schedule")
	if err != nil {
		return err
	}
	if normalized == nil {
		return errors.New("schedule не должен быть пустым")
	}

	return s.employeeRepo.UpdateSchedule(profile.ID, *normalized)
}

func (s *EmployeeService) UpdateProfile(profile *models.EmployeeProfile) error {
	return s.employeeRepo.Update(profile)
}

func (s *EmployeeService) UpdateEmployee(profileID uint, req dto.UpdateEmployeeRequest) (*models.EmployeeProfile, error) {
	if req.Role == models.RoleClient {
		return nil, errors.New("сотруднику нельзя назначить роль CLIENT")
	}
	if req.ExpectedSalary <= 0 {
		return nil, errors.New("expected_salary должен быть больше 0")
	}

	profile, err := s.employeeRepo.GetByID(profileID)
	if err != nil {
		return nil, errors.New("профиль сотрудника не найден")
	}

	user, err := s.userRepo.GetByID(profile.UserID)
	if err != nil {
		return nil, errors.New("учётная запись сотрудника не найдена")
	}

	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}

	var schedulePtr *string
	if req.WorkSchedule != "" {
		normalized, err := normalizeScheduleJSON(req.WorkSchedule, "work_schedule")
		if err != nil {
			return nil, err
		}
		schedulePtr = normalized
	}

	user.Username = req.Username
	user.FullName = req.FullName
	user.Phone = req.Phone
	user.Email = emailPtr
	user.Role = req.Role

	profile.Specialization = req.Specialization
	profile.ExpectedSalary = req.ExpectedSalary
	profile.WorkSchedule = schedulePtr

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("не удалось обновить пользователя: " + err.Error())
	}
	if err := s.employeeRepo.Update(profile); err != nil {
		return nil, errors.New("не удалось обновить профиль: " + err.Error())
	}
	if err := s.employeeRepo.ReplaceSalonAssignments(profile.ID, uniqueSalonIDs(req.SalonIDs)); err != nil {
		return nil, errors.New("не удалось обновить закрепление за салонами: " + err.Error())
	}

	return s.employeeRepo.GetByID(profile.ID)
}

func (s *EmployeeService) PatchEmployee(profileID uint, req dto.PatchEmployeeRequest) (*models.EmployeeProfile, error) {
	profile, err := s.employeeRepo.GetByID(profileID)
	if err != nil {
		return nil, errors.New("профиль сотрудника не найден")
	}

	user, err := s.userRepo.GetByID(profile.UserID)
	if err != nil {
		return nil, errors.New("учётная запись сотрудника не найдена")
	}

	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Email != nil {
		if *req.Email == "" {
			user.Email = nil
		} else {
			user.Email = req.Email
		}
	}
	if req.Role != nil {
		if *req.Role == models.RoleClient {
			return nil, errors.New("сотруднику нельзя назначить роль CLIENT")
		}
		user.Role = *req.Role
	}
	if req.Specialization != nil {
		profile.Specialization = *req.Specialization
	}
	if req.ExpectedSalary != nil {
		if *req.ExpectedSalary <= 0 {
			return nil, errors.New("expected_salary должен быть больше 0")
		}
		profile.ExpectedSalary = *req.ExpectedSalary
	}
	if req.WorkSchedule != nil {
		if *req.WorkSchedule == "" {
			profile.WorkSchedule = nil
		} else {
			normalized, err := normalizeScheduleJSON(*req.WorkSchedule, "work_schedule")
			if err != nil {
				return nil, err
			}
			profile.WorkSchedule = normalized
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("не удалось обновить пользователя: " + err.Error())
	}
	if err := s.employeeRepo.Update(profile); err != nil {
		return nil, errors.New("не удалось обновить профиль: " + err.Error())
	}
	if req.SalonIDs != nil {
		if err := s.employeeRepo.ReplaceSalonAssignments(profile.ID, uniqueSalonIDs(*req.SalonIDs)); err != nil {
			return nil, errors.New("не удалось обновить закрепление за салонами: " + err.Error())
		}
	}

	return s.employeeRepo.GetByID(profile.ID)
}

func (s *EmployeeService) FireEmployee(profileID uint) error {
	profile, err := s.employeeRepo.GetByID(profileID)
	if err != nil {
		return errors.New("профиль сотрудника не найден")
	}

	return s.employeeRepo.Fire(profile.ID, profile.UserID)
}

func (s *EmployeeService) AssignToSalon(profileID, salonID uint) error {
	return s.employeeRepo.AssignToSalon(profileID, salonID)
}

func (s *EmployeeService) RemoveFromSalon(profileID, salonID uint) error {
	return s.employeeRepo.RemoveFromSalon(profileID, salonID)
}

func uniqueSalonIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))

	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}
