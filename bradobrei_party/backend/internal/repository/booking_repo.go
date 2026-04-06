package repository

import (
	"time"

	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(b *models.Booking) error {
	return r.db.Create(b).Error
}

func (r *BookingRepository) GetAll() ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Preload("Client").
		Preload("Master").
		Preload("Salon").
		Preload("Items.Service").
		Preload("Payment").
		Order("start_time DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByID(id uint) (*models.Booking, error) {
	var b models.Booking
	err := r.db.
		Preload("Client").
		Preload("Master").
		Preload("Salon").
		Preload("Items.Service").
		Preload("Payment").
		First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepository) GetByClientID(clientID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Where("client_id = ?", clientID).
		Preload("Salon").
		Preload("Items.Service").
		Preload("Payment").
		Order("start_time DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByMasterID(masterID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Where("master_id = ?", masterID).
		Preload("Client").
		Preload("Salon").
		Preload("Items.Service").
		Order("start_time DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) HasOverlap(masterID uint, start time.Time, durationMin int, excludeID uint) (bool, error) {
	end := start.Add(time.Duration(durationMin) * time.Minute)

	var count int64
	q := r.db.Model(&models.Booking{}).
		Where("master_id = ?", masterID).
		Where("status NOT IN ?", []string{string(models.BookingCancelled)}).
		Where(
			"(start_time < ? AND (start_time + (duration_minutes * interval '1 minute')) > ?)",
			end, start,
		)

	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}

	err := q.Count(&count).Error
	return count > 0, err
}

func (r *BookingRepository) UpdateStatus(id uint, status models.BookingStatus) error {
	return r.db.Model(&models.Booking{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BookingRepository) Update(b *models.Booking) error {
	return r.db.Save(b).Error
}

func (r *BookingRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("booking_id = ?", id).Delete(&models.Payment{}).Error; err != nil {
			return err
		}

		if err := tx.Where("booking_id = ?", id).Delete(&models.BookingItem{}).Error; err != nil {
			return err
		}

		return tx.Delete(&models.Booking{}, id).Error
	})
}

func (r *BookingRepository) GetByPeriodAndSalon(salonID uint, from, to time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	q := r.db.
		Where("start_time BETWEEN ? AND ?", from, to).
		Where("status = ?", models.BookingCompleted).
		Preload("Items.Service").
		Preload("Client")

	if salonID > 0 {
		q = q.Where("salon_id = ?", salonID)
	}

	err := q.Find(&bookings).Error
	return bookings, err
}
