package models

import (
	"time"

	"gorm.io/gorm"
)

// ====================== ENUMS ======================

type UserRole string
type BookingStatus string
type PaymentStatus string

const (
	RoleClient         UserRole = "CLIENT"
	RoleBasicMaster    UserRole = "BASIC_MASTER"
	RoleAdvancedMaster UserRole = "ADVANCED_MASTER"
	RoleHR             UserRole = "HR_SPECIALIST"
	RoleAccountant     UserRole = "ACCOUNTANT"
	RoleNetworkManager UserRole = "NETWORK_MANAGER"
	RoleAdmin          UserRole = "ADMINISTRATOR"
)

const (
	BookingPending    BookingStatus = "PENDING"
	BookingConfirmed  BookingStatus = "CONFIRMED"
	BookingInProgress BookingStatus = "IN_PROGRESS"
	BookingCompleted  BookingStatus = "COMPLETED"
	BookingCancelled  BookingStatus = "CANCELLED"
)

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentSuccess  PaymentStatus = "SUCCESS"
	PaymentFailed   PaymentStatus = "FAILED"
	PaymentRefunded PaymentStatus = "REFUNDED"
)

// ====================== СУЩНОСТИ ======================

// User — единая таблица пользователей (7 ролей по ТЗ 2.3)
type User struct {
	ID           uint           `gorm:"primaryKey"                      json:"id"`
	Username     string         `gorm:"unique;not null;size:50"          json:"username"`
	PasswordHash string         `gorm:"not null"                         json:"-"`
	FullName     string         `gorm:"not null;size:100"                json:"full_name"`
	Phone        string         `gorm:"size:20"                          json:"phone"`
	Email        *string        `gorm:"unique;size:100"                  json:"email,omitempty"`
	Role         UserRole       `gorm:"type:varchar(30);default:'CLIENT'" json:"role"`
	CreatedAt    time.Time      `                                        json:"created_at"`
	UpdatedAt    time.Time      `                                        json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                            json:"-"`

	EmployeeProfile  *EmployeeProfile `gorm:"foreignKey:UserID"    json:"employee_profile,omitempty"`
	BookingsAsClient []Booking        `gorm:"foreignKey:ClientID"  json:"-"`
	BookingsAsMaster []Booking        `gorm:"foreignKey:MasterID"  json:"-"`
	Reviews          []Review         `gorm:"foreignKey:UserID"    json:"-"`
}

// EmployeeProfile — профиль сотрудника (отчёт 2.2.1)
type EmployeeProfile struct {
	ID             uint      `gorm:"primaryKey"           json:"id"`
	UserID         uint      `gorm:"unique;not null"      json:"user_id"`
	Specialization string    `gorm:"size:100"             json:"specialization"`
	ExpectedSalary float64   `gorm:"type:decimal(10,2)"   json:"expected_salary"`
	WorkSchedule   *string   `gorm:"type:jsonb"           json:"work_schedule,omitempty"` // {"mon":"9-18","tue":"9-18",...}
	CreatedAt      time.Time `                            json:"created_at"`
	UpdatedAt      time.Time `                            json:"updated_at"`

	User     User      `gorm:"foreignKey:UserID"               json:"user,omitempty"`
	Salons   []Salon   `gorm:"many2many:employee_salons;"      json:"salons,omitempty"`
	Services []Service `gorm:"many2many:employee_services;"    json:"services,omitempty"`
}

// Salon — салон (PostGIS, отчёты 2.2.2)
type Salon struct {
	ID             uint      `gorm:"primaryKey"                      json:"id"`
	Name           string    `gorm:"not null;size:100"               json:"name"`
	Address        string    `gorm:"not null"                        json:"address"`
	Location       *string   `gorm:"type:geography(POINT,4326)"      json:"location,omitempty"` // PostGIS, опционально
	WorkingHours   *string   `gorm:"type:jsonb"                      json:"working_hours,omitempty"`
	Status         string    `gorm:"default:'OPEN'"                  json:"status"` // OPEN/CLOSED
	MaxStaff       int       `                                       json:"max_staff"`
	BaseHourlyRate float64   `gorm:"type:decimal(10,2)"              json:"base_hourly_rate"`
	CreatedAt      time.Time `                                       json:"created_at"`
	UpdatedAt      time.Time `                                       json:"updated_at"`

	Inventory []Inventory       `gorm:"foreignKey:SalonID"         json:"-"`
	Bookings  []Booking         `gorm:"foreignKey:SalonID"         json:"-"`
	Employees []EmployeeProfile `gorm:"many2many:employee_salons;" json:"employees,omitempty"`
}

// Service — услуга (длительность ≥ 60 мин по ТЗ)
type Service struct {
	ID              uint      `gorm:"primaryKey"              json:"id"`
	Name            string    `gorm:"not null;unique;size:100" json:"name"`
	Description     string    `                               json:"description"`
	Price           float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	DurationMinutes int       `gorm:"not null"                json:"duration_minutes"`
	CreatedAt       time.Time `                               json:"created_at"`
	UpdatedAt       time.Time `                               json:"updated_at"`

	Materials []ServiceMaterial `gorm:"foreignKey:ServiceID"        json:"materials,omitempty"`
	Bookings  []BookingItem     `gorm:"foreignKey:ServiceID"        json:"-"`
	Employees []EmployeeProfile `gorm:"many2many:employee_services;" json:"employees,omitempty"`
}

// Material — расходный материал
type Material struct {
	ID   uint   `gorm:"primaryKey"       json:"id"`
	Name string `gorm:"not null;unique"  json:"name"`
	Unit string `gorm:"size:20"          json:"unit"` // мл, шт, гр

	Services    []ServiceMaterial `gorm:"foreignKey:MaterialID" json:"-"`
	Inventories []Inventory       `gorm:"foreignKey:MaterialID" json:"-"`
}

// ServiceMaterial — норма расхода материала на услугу
type ServiceMaterial struct {
	ServiceID      uint    `gorm:"primaryKey" json:"service_id"`
	MaterialID     uint    `gorm:"primaryKey" json:"material_id"`
	QuantityPerUse float64 `gorm:"not null"   json:"quantity_per_use"`

	Service  Service  `json:"service,omitempty"`
	Material Material `json:"material,omitempty"`
}

// Inventory — складской запас в салоне
type Inventory struct {
	ID          uint      `gorm:"primaryKey"  json:"id"`
	SalonID     uint      `gorm:"not null"    json:"salon_id"`
	MaterialID  uint      `gorm:"not null"    json:"material_id"`
	Quantity    float64   `gorm:"not null"    json:"quantity"`
	LastUpdated time.Time `gorm:"autoUpdateTime" json:"last_updated"`

	Salon    Salon    `json:"salon,omitempty"`
	Material Material `json:"material,omitempty"`
}

// Booking — бронирование (ключевая сущность для отчётов 2.2.2–2.2.4)
type Booking struct {
	ID              uint          `gorm:"primaryKey"                              json:"id"`
	StartTime       time.Time     `gorm:"not null;index:idx_booking_time"         json:"start_time"`
	DurationMinutes int           `gorm:"not null"                                json:"duration_minutes"`
	Status          BookingStatus `gorm:"default:'PENDING';index:idx_booking_status" json:"status"`
	TotalPrice      float64       `gorm:"type:decimal(10,2)"                      json:"total_price"`
	Notes           string        `                                               json:"notes"`
	ClientID        uint          `gorm:"not null"                                json:"client_id"`
	MasterID        *uint         `                                               json:"master_id,omitempty"`
	SalonID         uint          `gorm:"not null"                                json:"salon_id"`
	CreatedAt       time.Time     `                                               json:"created_at"`
	UpdatedAt       time.Time     `                                               json:"updated_at"`

	Client  User          `gorm:"foreignKey:ClientID"  json:"client,omitempty"`
	Master  *User         `gorm:"foreignKey:MasterID"  json:"master,omitempty"`
	Salon   Salon         `                            json:"salon,omitempty"`
	Items   []BookingItem `gorm:"foreignKey:BookingID" json:"items,omitempty"`
	Payment *Payment      `gorm:"foreignKey:BookingID" json:"payment,omitempty"`
}

// BookingItem — позиция бронирования (одна услуга)
type BookingItem struct {
	ID             uint    `gorm:"primaryKey"                      json:"id"`
	BookingID      uint    `gorm:"not null"                        json:"booking_id"`
	ServiceID      uint    `gorm:"not null"                        json:"service_id"`
	Quantity       int     `gorm:"default:1"                       json:"quantity"`
	PriceAtBooking float64 `gorm:"type:decimal(10,2);not null"     json:"price_at_booking"`

	Booking Booking `json:"-"`
	Service Service `json:"service,omitempty"`
}

// Payment — оплата бронирования
type Payment struct {
	ID                    uint          `gorm:"primaryKey"                  json:"id"`
	BookingID             uint          `gorm:"unique"                      json:"booking_id"`
	Amount                float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status                PaymentStatus `gorm:"default:'PENDING'"           json:"status"`
	ExternalTransactionID string        `                                   json:"external_transaction_id,omitempty"`
	CreatedAt             time.Time     `                                   json:"created_at"`
	CompletedAt           *time.Time    `                                   json:"completed_at,omitempty"`

	Booking Booking `json:"-"`
}

// Review — отзыв об информационной системе (ТЗ 2.2.5)
type Review struct {
	ID        uint      `gorm:"primaryKey"                    json:"id"`
	UserID    uint      `gorm:"not null"                      json:"user_id"`
	Text      string    `gorm:"not null"                      json:"text"`
	Rating    int       `gorm:"check:rating BETWEEN 1 AND 5"  json:"rating"` // 1-5
	CreatedAt time.Time `                                     json:"created_at"`

	User User `json:"user,omitempty"`
}

// ====================== DTO (запросы/ответы) ======================

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email" binding:"omitempty,email"`
	Role     UserRole `json:"role"`
}

type CreateBookingRequest struct {
	StartTime  string `json:"start_time" binding:"required"` // RFC3339
	SalonID    uint   `json:"salon_id"   binding:"required"`
	MasterID   *uint  `json:"master_id"`
	ServiceIDs []uint `json:"service_ids" binding:"required,min=1"`
	Notes      string `json:"notes"`
}

type CreateReviewRequest struct {
	Text   string `json:"text"   binding:"required"`
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// HireEmployeeRequest — HR нанимает нового сотрудника (ТЗ 2.3.4)
type HireEmployeeRequest struct {
	// Данные учётной записи
	Username string   `json:"username"  binding:"required,min=3,max=50"`
	Password string   `json:"password"  binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required"`
	Phone    string   `json:"phone"`
	Email    string   `json:"email"     binding:"omitempty,email"`
	Role     UserRole `json:"role"      binding:"required"`

	// Данные профиля сотрудника
	Specialization string  `json:"specialization"`
	ExpectedSalary float64 `json:"expected_salary"`
	WorkSchedule   string  `json:"work_schedule"` // пустая строка → NULL в jsonb
	SalonID        uint    `json:"salon_id"`      // опционально — сразу прикрепить к салону

	// Заполняется в хэндлере перед передачей в сервис (не из JSON)
	PasswordHash string `json:"-"`
}

// UpdateScheduleRequest — запрос на изменение расписания мастера (ТЗ 2.3.2)
type UpdateScheduleRequest struct {
	Schedule string `json:"schedule" binding:"required"`
}
