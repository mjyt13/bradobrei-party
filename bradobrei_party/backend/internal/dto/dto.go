package dto

import "bradobrei/backend/internal/models"

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

// GeocodeAddressRequest — серверный геокодинг адреса (секретный ключ провайдера только на backend).
type GeocodeAddressRequest struct {
	Address string `json:"address" binding:"required" example:"Екатеринбург, ул. Малышева, 12"`
}

// GeocodeAddressResponse — нормализованные координаты и строка адреса для карты в SPA.
type GeocodeAddressResponse struct {
	Latitude         float64 `json:"latitude" example:"56.838011"`
	Longitude        float64 `json:"longitude" example:"60.597474"`
	FormattedAddress string  `json:"formatted_address"`
	Provider         string  `json:"provider" example:"yandex"`
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

type CreateServiceRequest struct {
	Name            string  `json:"name" binding:"required" example:"Мужская стрижка"`
	Description     string  `json:"description" example:"Стрижка с оформлением висков и затылка"`
	Price           float64 `json:"price" binding:"required,gt=0" example:"1800"`
	DurationMinutes int     `json:"duration_minutes" binding:"required,min=1" example:"75"`
}

type UpdateServiceRequest struct {
	Name            string  `json:"name" binding:"required" example:"Мужская стрижка"`
	Description     string  `json:"description" example:"Стрижка с оформлением висков и затылка"`
	Price           float64 `json:"price" binding:"required,gt=0" example:"1900"`
	DurationMinutes int     `json:"duration_minutes" binding:"required,min=1" example:"90"`
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

type UpdateEmployeeRequest struct {
	Username string          `json:"username" binding:"required,min=3,max=50" example:"master_ivan"`
	FullName string          `json:"full_name" binding:"required" example:"Иван Барбер"`
	Phone    string          `json:"phone" example:"+79990001122"`
	Email    string          `json:"email" binding:"omitempty,email" example:"ivan.barber@example.com"`
	Role     models.UserRole `json:"role" binding:"required" example:"ADVANCED_MASTER"`

	Specialization string  `json:"specialization" example:"Fade, beard styling"`
	ExpectedSalary float64 `json:"expected_salary" example:"85000"`
	WorkSchedule   string  `json:"work_schedule" example:"{\"mon\":\"10:00-19:00\",\"wed\":\"10:00-19:00\"}"`
	SalonIDs       []uint  `json:"salon_ids" example:"1,2"`
}

type PatchEmployeeRequest struct {
	Username       *string          `json:"username,omitempty" example:"master_ivan"`
	FullName       *string          `json:"full_name,omitempty" example:"Иван Барбер"`
	Phone          *string          `json:"phone,omitempty" example:"+79990001122"`
	Email          *string          `json:"email,omitempty" example:"ivan.barber@example.com"`
	Role           *models.UserRole `json:"role,omitempty" example:"ADVANCED_MASTER"`
	Specialization *string          `json:"specialization,omitempty" example:"Fade, beard styling"`
	ExpectedSalary *float64         `json:"expected_salary,omitempty" example:"85000"`
	WorkSchedule   *string          `json:"work_schedule,omitempty" example:"{\"mon\":\"10:00-19:00\",\"wed\":\"10:00-19:00\"}"`
	SalonIDs       *[]uint          `json:"salon_ids,omitempty" example:"1,2"`
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

type CreateMaterialExpenseRequest struct {
	MaterialID    uint    `json:"material_id" binding:"required" example:"1"`
	SalonID       uint    `json:"salon_id" binding:"required" example:"1"`
	PurchasePrice float64 `json:"purchase_price" binding:"required,gt=0" example:"350"`
	Quantity      float64 `json:"quantity" binding:"required,gt=0" example:"10"`
}

type UpdateMaterialExpenseRequest struct {
	MaterialID    uint    `json:"material_id" binding:"required" example:"1"`
	SalonID       uint    `json:"salon_id" binding:"required" example:"1"`
	PurchasePrice float64 `json:"purchase_price" binding:"required,gt=0" example:"350"`
	Quantity      float64 `json:"quantity" binding:"required,gt=0" example:"10"`
}

type UseServiceRequest struct {
	SalonID  uint `json:"salon_id" binding:"required" example:"1"`
	Quantity int  `json:"quantity" binding:"required,min=1" example:"1"`
}

type SetInventoryQuantityRequest struct {
	Quantity float64 `json:"quantity" binding:"required,gte=0" example:"25"`
}
