package services

import (
	"errors"
	"time"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"

	"gorm.io/gorm"
)

type BookingService struct {
	bookingRepo   *repository.BookingRepository
	inventoryRepo *repository.InventoryRepository
	db            *gorm.DB
}

func NewBookingService(
	bookingRepo *repository.BookingRepository,
	inventoryRepo *repository.InventoryRepository,
	db *gorm.DB,
) *BookingService {
	return &BookingService{
		bookingRepo:   bookingRepo,
		inventoryRepo: inventoryRepo,
		db:            db,
	}
}

func (s *BookingService) Create(req dto.CreateBookingRequest, clientID uint) (*models.Booking, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("неверный формат времени, используйте RFC3339")
	}

	var services []models.Service
	if err := s.db.Preload("Materials").Find(&services, req.ServiceIDs).Error; err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("услуги не найдены")
	}

	totalDuration := 0
	totalPrice := 0.0
	for _, svc := range services {
		totalDuration += svc.DurationMinutes
		totalPrice += svc.Price
	}

	if totalDuration < 60 {
		return nil, errors.New("минимальная длительность бронирования - 60 минут")
	}

	if req.MasterID != nil {
		overlap, err := s.bookingRepo.HasOverlap(*req.MasterID, startTime, totalDuration, 0)
		if err != nil {
			return nil, err
		}
		if overlap {
			return nil, errors.New("мастер уже занят в указанное время")
		}
	}

	items := make([]models.BookingItem, 0, len(services))
	for _, svc := range services {
		items = append(items, models.BookingItem{
			ServiceID:      svc.ID,
			Quantity:       1,
			PriceAtBooking: svc.Price,
		})
	}

	booking := &models.Booking{
		ClientID:        clientID,
		MasterID:        req.MasterID,
		SalonID:         req.SalonID,
		StartTime:       startTime,
		DurationMinutes: totalDuration,
		TotalPrice:      totalPrice,
		Status:          models.BookingPending,
		Notes:           req.Notes,
		Items:           items,
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) Confirm(bookingID uint, masterID uint) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, errors.New("бронирование не найдено")
	}

	if booking.Status != models.BookingPending {
		return nil, errors.New("можно подтвердить только бронирование со статусом PENDING")
	}

	for _, item := range booking.Items {
		var materials []models.ServiceMaterial
		if err := s.db.Where("service_id = ?", item.ServiceID).Find(&materials).Error; err != nil {
			return nil, err
		}
		if err := s.inventoryRepo.WriteOff(booking.SalonID, materials, item.Quantity); err != nil {
			return nil, errors.New("недостаточно материалов на складе: " + err.Error())
		}
	}

	booking.Status = models.BookingConfirmed
	booking.MasterID = &masterID
	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) Cancel(bookingID uint, requesterID uint, requesterRole models.UserRole) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return errors.New("бронирование не найдено")
	}

	if requesterRole == models.RoleClient && booking.ClientID != requesterID {
		return errors.New("нет прав для отмены этого бронирования")
	}

	if booking.Status == models.BookingCompleted || booking.Status == models.BookingCancelled {
		return errors.New("нельзя отменить завершённое или уже отменённое бронирование")
	}

	return s.bookingRepo.UpdateStatus(bookingID, models.BookingCancelled)
}

func (s *BookingService) Delete(bookingID uint, requesterID uint, requesterRole models.UserRole) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return errors.New("бронирование не найдено")
	}

	if requesterRole == models.RoleClient && booking.ClientID != requesterID {
		return errors.New("нет прав для удаления этого бронирования")
	}

	if requesterRole != models.RoleClient && requesterRole != models.RoleAdmin {
		return errors.New("нет прав для удаления бронирования")
	}

	return s.bookingRepo.Delete(bookingID)
}

func (s *BookingService) GetByID(id uint) (*models.Booking, error) {
	return s.bookingRepo.GetByID(id)
}

func (s *BookingService) GetAll() ([]models.Booking, error) {
	return s.bookingRepo.GetAll()
}

func (s *BookingService) GetByClient(clientID uint) ([]models.Booking, error) {
	return s.bookingRepo.GetByClientID(clientID)
}

func (s *BookingService) GetByMaster(masterID uint) ([]models.Booking, error) {
	return s.bookingRepo.GetByMasterID(masterID)
}
