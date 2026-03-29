package dto

import "bradobrei/backend/internal/models"

// ErrorResponse — единый формат ошибок HTTP API.
type ErrorResponse struct {
	Error   string `json:"error" example:"bad_request"`
	Code    int    `json:"code" example:"400"`
	Message string `json:"message,omitempty" example:"invalid request body"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.dev-token"`
}

type UserResponse struct {
	User models.User `json:"user"`
}

type RegisterRequest struct {
	Username string          `json:"username" binding:"required,min=3,max=50" example:"admin"`
	Password string          `json:"password" binding:"required,min=6" example:"password"`
	FullName string          `json:"full_name" binding:"required" example:"Иван Петров"`
	Phone    string          `json:"phone" example:"+79991234567"`
	Email    string          `json:"email" binding:"omitempty,email" example:"admin@example.com"`
	Role     models.UserRole `json:"role" example:"ADMINISTRATOR"`
}

type CreateBookingRequest struct {
	StartTime  string `json:"start_time" binding:"required" example:"2026-03-29T14:00:00Z"`
	SalonID    uint   `json:"salon_id" binding:"required" example:"1"`
	MasterID   *uint  `json:"master_id" example:"2"`
	ServiceIDs []uint `json:"service_ids" binding:"required,min=1" example:"1,2"`
	Notes      string `json:"notes" example:"Клиент просит оформить бороду и стрижку без ожидания."`
}

type CreateReviewRequest struct {
	Text   string `json:"text" binding:"required" example:"Отличный сервис и удобная запись."`
	Rating int    `json:"rating" binding:"required,min=1,max=5" example:"5"`
}

type CreatePaymentRequest struct {
	BookingID             uint                 `json:"booking_id" binding:"required" example:"1"`
	Amount                float64              `json:"amount" example:"2500"`
	Status                models.PaymentStatus `json:"status" example:"PENDING"`
	ExternalTransactionID string               `json:"external_transaction_id" example:"txn_local_12345"`
}

type HireEmployeeRequest struct {
	Username string          `json:"username" binding:"required,min=3,max=50" example:"master_ivan"`
	Password string          `json:"password" binding:"required,min=6" example:"password"`
	FullName string          `json:"full_name" binding:"required" example:"Иван Барбер"`
	Phone    string          `json:"phone" example:"+79990001122"`
	Email    string          `json:"email" binding:"omitempty,email" example:"ivan.barber@example.com"`
	Role     models.UserRole `json:"role" binding:"required" example:"ADVANCED_MASTER"`

	Specialization string  `json:"specialization" example:"Fade, beard styling"`
	ExpectedSalary float64 `json:"expected_salary" example:"85000"`
	WorkSchedule   string  `json:"work_schedule" example:"{\"mon\":\"10:00-19:00\",\"wed\":\"10:00-19:00\"}"`
	SalonID        uint    `json:"salon_id" example:"1"`

	PasswordHash string `json:"-"`
}

type UpdateScheduleRequest struct {
	Schedule string `json:"schedule" binding:"required" example:"{\"mon\":\"10:00-19:00\",\"tue\":\"12:00-20:00\"}"`
}

type AssignSalonRequest struct {
	SalonID uint `json:"salon_id" binding:"required" example:"1"`
}

type AssignServiceToMasterRequest struct {
	TargetUserID uint `json:"target_user_id" binding:"required" example:"12"`
}
