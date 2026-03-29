package services

import (
	"errors"
	"time"

	"bradobrei/backend/internal/dto"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	bookingRepo *repository.BookingRepository
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	bookingRepo *repository.BookingRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *PaymentService) GetAll() ([]models.Payment, error) {
	return s.paymentRepo.GetAll()
}

func (s *PaymentService) GetByID(id uint) (*models.Payment, error) {
	return s.paymentRepo.GetByID(id)
}

func (s *PaymentService) Create(req dto.CreatePaymentRequest) (*models.Payment, error) {
	booking, err := s.bookingRepo.GetByID(req.BookingID)
	if err != nil {
		return nil, errors.New("бронирование не найдено")
	}

	status := req.Status
	if status == "" {
		status = models.PaymentPending
	}

	switch status {
	case models.PaymentPending, models.PaymentSuccess, models.PaymentFailed, models.PaymentRefunded:
	default:
		return nil, errors.New("некорректный статус платежа")
	}

	amount := req.Amount
	if amount <= 0 {
		amount = booking.TotalPrice
	}
	if amount <= 0 {
		return nil, errors.New("сумма платежа должна быть больше нуля")
	}

	var completedAt *time.Time
	if status == models.PaymentSuccess {
		now := time.Now()
		completedAt = &now
	}

	payment := &models.Payment{
		BookingID:             req.BookingID,
		Amount:                amount,
		Status:                status,
		ExternalTransactionID: req.ExternalTransactionID,
		CompletedAt:           completedAt,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}
