package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole хранит прикладную роль пользователя.
// В физической модели роль вынесена в справочник roles, а в коде оставлена
// строковым enum, чтобы упростить JWT, RBAC и фильтры в запросах.
type UserRole string

// BookingStatus хранит жизненный цикл бронирования.
// В SQL-модели это можно было бы вынести в booking_statuses, но текущая
// бизнес-логика работает со строковыми статусами напрямую.
type BookingStatus string

// PaymentStatus хранит статус платёжной операции.
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

// User — базовая учётная запись системы.
// По смыслу соответствует users из физической модели, но для API ФИО хранится
// одной строкой, а роль денормализована в строковый enum.
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

// EmployeeProfile — расширение учётной записи сотрудника.
// Покрывает идею workers/profiles из физической модели.
// Участвует в выполнении требования 2.2.1 "Реестр персонала":
// здесь хранятся специализация, закрепления за салонами и расчётный оклад.
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

// Salon — филиал сети.
// Address хранит человекочитаемый адрес, а Location — пространственную точку PostGIS.
// Геополе переведено на geometry(Point,4326), чтобы модель совпадала с индексацией
// и геооперациями в PostGIS.
// Участвует в выполнении требования 2.2.2 "Аналитический отчёт об операционной активности филиалов":
// здесь лежат идентификатор филиала, адрес и параметры режима работы.
type Salon struct {
	ID             uint      `gorm:"primaryKey"                      json:"id"`
	Name           string    `gorm:"not null;size:100"               json:"name"`
	Address        string    `gorm:"not null"                        json:"address"`
	Location       *string   `gorm:"type:geometry(POINT,4326)"      json:"location,omitempty"` // PostGIS, опционально
	// Latitude/Longitude только для JSON (карта в SPA), не колонки БД.
	Latitude  *float64 `json:"latitude,omitempty"  gorm:"-"`
	Longitude *float64 `json:"longitude,omitempty" gorm:"-"`
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

// Service — услуга салона.
// Используется и как прайс-лист, и как источник данных для отчёта 2.2.3
// "Статистика востребованности услуг".
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

// Material — расходный материал со справочной единицей измерения.
type Material struct {
	ID   uint   `gorm:"primaryKey"       json:"id"`
	Name string `gorm:"not null;unique"  json:"name"`
	Unit string `gorm:"size:20"          json:"unit"` // мл, шт, гр

	Services    []ServiceMaterial `gorm:"foreignKey:MaterialID" json:"-"`
	Inventories []Inventory       `gorm:"foreignKey:MaterialID" json:"-"`
}

// ServiceMaterial — норма расхода материала на одну услугу.
// По смыслу соответствует service_consumption из физической модели.
type ServiceMaterial struct {
	ServiceID      uint    `gorm:"primaryKey" json:"service_id"`
	MaterialID     uint    `gorm:"primaryKey" json:"material_id"`
	QuantityPerUse float64 `gorm:"not null"   json:"quantity_per_use"`

	Service  Service  `json:"service,omitempty"`
	Material Material `json:"material,omitempty"`
}

// Inventory — остаток материала в конкретном салоне.
// По смыслу соответствует supplies из физической модели.
// Участвует в требовании 2.2.6 "Ведомость движения ТМЦ":
// хранит фактический остаток материала в филиале.
type Inventory struct {
	ID          uint      `gorm:"primaryKey"     json:"id"`
	SalonID     uint      `gorm:"not null"       json:"salon_id"`
	MaterialID  uint      `gorm:"not null"       json:"material_id"`
	Quantity    float64   `gorm:"not null"       json:"quantity"`
	LastUpdated time.Time `gorm:"autoUpdateTime" json:"last_updated"`

	Salon    Salon    `json:"salon,omitempty"`
	Material Material `json:"material,omitempty"`
}

// MaterialExpense фиксирует закупку или расходную операцию по материалу для салона.
// На текущем этапе хранит цену закупки и количество, а сервисный слой синхронно
// отражает операцию в фактическом остатке inventory.
type MaterialExpense struct {
	ID            uint      `gorm:"primaryKey"                  json:"id"`
	MaterialID    uint      `gorm:"not null;index"              json:"material_id"`
	SalonID       uint      `gorm:"not null;index"              json:"salon_id"`
	PurchasePrice float64   `gorm:"type:decimal(10,2);not null" json:"purchase_price"`
	Quantity      float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	CreatedAt     time.Time `                                   json:"created_at"`
	UpdatedAt     time.Time `                                   json:"updated_at"`

	Material Material `gorm:"foreignKey:MaterialID" json:"material,omitempty"`
	Salon    Salon    `gorm:"foreignKey:SalonID"    json:"salon,omitempty"`
}

// ServiceUsage фиксирует ручное использование услуги вне сценария подтверждения бронирования.
// Нужен как журнал складских операций, чтобы отчёт 2.2.6 видел списания,
// выполненные отдельной командой /services/{id}/use.
type ServiceUsage struct {
	ID        uint      `gorm:"primaryKey"     json:"id"`
	ServiceID uint      `gorm:"not null;index" json:"service_id"`
	SalonID   uint      `gorm:"not null;index" json:"salon_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Quantity  int       `gorm:"not null"       json:"quantity"`
	CreatedAt time.Time `                      json:"created_at"`

	Service Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Salon   Salon   `gorm:"foreignKey:SalonID"   json:"salon,omitempty"`
	User    User    `gorm:"foreignKey:UserID"    json:"user,omitempty"`
}

// Booking — запись клиента на набор услуг.
// Нужна для отчётов по загрузке салонов, мастеров и лояльности клиентов.
// Уже сейчас участвует минимум в требованиях 2.2.2, 2.2.4, 2.2.7 и 2.2.8,
// потому что связывает клиента, мастера, салон, время визита и итоговую стоимость.
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

// BookingItem — конкретная услуга внутри бронирования.
// По смыслу соответствует booking_service, но хранит снимок цены на момент заказа.
type BookingItem struct {
	ID             uint    `gorm:"primaryKey"                      json:"id"`
	BookingID      uint    `gorm:"not null"                        json:"booking_id"`
	ServiceID      uint    `gorm:"not null"                        json:"service_id"`
	Quantity       int     `gorm:"default:1"                       json:"quantity"`
	PriceAtBooking float64 `gorm:"type:decimal(10,2);not null"     json:"price_at_booking"`

	Booking Booking `json:"-"`
	Service Service `json:"service,omitempty"`
}

// Payment — платёж по бронированию.
// В физической модели статус платежа мог бы жить в transaction_statuses,
// но текущей прикладной логике удобнее строковый enum.
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

// Review — отзыв пользователя о работе системы или качестве обслуживания.
// В физической модели feedback/evaluations разбивались на две таблицы; здесь для API
// удобнее хранить текст и числовую оценку в одной сущности.
// Используется для требования 2.2.5 "Журнал мониторинга качества обслуживания и обратной связи".
type Review struct {
	ID        uint      `gorm:"primaryKey"                    json:"id"`
	UserID    uint      `gorm:"not null"                      json:"user_id"`
	Text      string    `gorm:"not null"                      json:"text"`
	Rating    int       `gorm:"check:rating BETWEEN 1 AND 5"  json:"rating"` // 1-5
	CreatedAt time.Time `                                     json:"created_at"`

	User User `json:"user,omitempty"`
}

// ErrorLog хранит журнал исключительных ситуаций для требований 2.4.
// Заполняется middleware после ответа сервера и позволяет просматривать историю
// 400/401/403/404/500 без отдельного доступа к stdout контейнера.
type ErrorLog struct {
	ID         uint       `gorm:"primaryKey"                        json:"id"`
	Method     string     `gorm:"not null;size:10;index"           json:"method"`
	Path       string     `gorm:"not null;index"                   json:"path"`
	Query      string     `gorm:"type:text"                        json:"query,omitempty"`
	StatusCode int        `gorm:"not null;index"                   json:"status_code"`
	ErrorCode  string     `gorm:"size:100;index"                   json:"error_code,omitempty"`
	Message    string     `gorm:"type:text"                        json:"message,omitempty"`
	IP         string     `gorm:"size:64"                          json:"ip,omitempty"`
	UserAgent  string     `gorm:"type:text"                        json:"user_agent,omitempty"`
	UserID     *uint      `gorm:"index"                            json:"user_id,omitempty"`
	UserRole   *UserRole  `gorm:"type:varchar(30);index"           json:"user_role,omitempty"`
	CreatedAt  time.Time  `                                        json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ReportMeta описывает служебные метаданные экспортируемого отчёта.
// Эти поля не относятся к ORM-таблицам: они нужны для HTML/PDF-представления,
// когда один и тот же набор данных оформляется как печатный документ.
type ReportMeta struct {
	ReportCode   string     `json:"report_code"`
	Title        string     `json:"title"`
	Subtitle     string     `json:"subtitle,omitempty"`
	GeneratedAt  time.Time  `json:"generated_at"`
	PeriodFrom   *time.Time `json:"period_from,omitempty"`
	PeriodTo     *time.Time `json:"period_to,omitempty"`
	GeneratedBy  string     `json:"generated_by,omitempty"`
	Organization string     `json:"organization,omitempty"`
}

// EmployeeRegistryReportDocument — представление отчёта 2.2.1 для HTML/PDF.
type EmployeeRegistryReportDocument struct {
	Meta ReportMeta                  `json:"meta"`
	Rows []EmployeeRegistryReportRow `json:"rows"`
}

type EmployeeRegistryReportRow struct {
	FullName       string   `json:"full_name"`
	Role           UserRole `json:"role"`
	Phone          string   `json:"phone"`
	Email          string   `json:"email,omitempty"`
	Specialization string   `json:"specialization,omitempty"`
	Salons         []string `json:"salons,omitempty"`
	ExpectedSalary float64  `json:"expected_salary"`
}

// SalonActivityReportDocument — представление отчёта 2.2.2 для HTML/PDF.
type SalonActivityReportDocument struct {
	Meta ReportMeta               `json:"meta"`
	Rows []SalonActivityReportRow `json:"rows"`
}

type SalonActivityReportRow struct {
	SalonID      uint    `json:"salon_id"`
	SalonName    string  `json:"salon_name,omitempty"`
	Address      string  `json:"address"`
	ClientCount  int64   `json:"client_count"`
	ServiceCount int64   `json:"service_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

// MasterActivityReportDocument — представление отчёта 2.2.4 для HTML/PDF.
type MasterActivityReportDocument struct {
	Meta ReportMeta                `json:"meta"`
	Rows []MasterActivityReportRow `json:"rows"`
}

type MasterActivityReportRow struct {
	MasterID     uint    `json:"master_id"`
	FullName     string  `json:"full_name"`
	ServiceCount int64   `json:"service_count"`
	Revenue      float64 `json:"revenue"`
	MaterialCost float64 `json:"material_cost"`
}

// ReviewsReportDocument — представление отчёта 2.2.5 для HTML/PDF.
type ReviewsReportDocument struct {
	Meta ReportMeta         `json:"meta"`
	Rows []ReviewsReportRow `json:"rows"`
}

type ReviewsReportRow struct {
	Author    string    `json:"author"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// ServicePopularityReportDocument — представление отчёта 2.2.3 для HTML/PDF.
type ServicePopularityReportDocument struct {
	Meta ReportMeta                   `json:"meta"`
	Rows []ServicePopularityReportRow `json:"rows"`
}

type ServicePopularityReportRow struct {
	ServiceID    uint    `json:"service_id"`
	ServiceName  string  `json:"service_name"`
	UsageCount   int64   `json:"usage_count"`
	RelativeFreq float64 `json:"relative_freq"`
}

// InventoryMovementReportDocument — представление отчёта 2.2.6 для HTML/PDF.
type InventoryMovementReportDocument struct {
	Meta ReportMeta                   `json:"meta"`
	Rows []InventoryMovementReportRow `json:"rows"`
}

type InventoryMovementReportRow struct {
	SalonAddress   string  `json:"salon_address"`
	MaterialName   string  `json:"material_name"`
	Unit           string  `json:"unit"`
	OpeningBalance float64 `json:"opening_balance"`
	Purchased      float64 `json:"purchased"`
	WrittenOff     float64 `json:"written_off"`
	CurrentBalance float64 `json:"current_balance"`
}

// ClientLoyaltyReportDocument — представление отчёта 2.2.7 для HTML/PDF.
type ClientLoyaltyReportDocument struct {
	Meta ReportMeta               `json:"meta"`
	Rows []ClientLoyaltyReportRow `json:"rows"`
}

type ClientLoyaltyReportRow struct {
	ClientID    uint       `json:"client_id"`
	FullName    string     `json:"full_name"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email,omitempty"`
	FirstVisit  *time.Time `json:"first_visit,omitempty"`
	LastVisit   *time.Time `json:"last_visit,omitempty"`
	VisitCount  int64      `json:"visit_count"`
	PaidTotal   float64    `json:"paid_total"`
	BonusStatus string     `json:"bonus_status"`
}

// CancelledBookingsReportDocument — представление отчёта 2.2.8 для HTML/PDF.
type CancelledBookingsReportDocument struct {
	Meta ReportMeta                   `json:"meta"`
	Rows []CancelledBookingsReportRow `json:"rows"`
}

type CancelledBookingsReportRow struct {
	BookingID           uint      `json:"booking_id"`
	PlannedVisit        time.Time `json:"planned_visit"`
	ClientFullName      string    `json:"client_full_name"`
	MasterFullName      string    `json:"master_full_name"`
	CancellationReason  string    `json:"cancellation_reason"`
	CancellationRatePct float64   `json:"cancellation_rate_pct"`
	Status              string    `json:"status"`
}

// FinancialSummaryReportDocument — представление отчёта 2.2.9 для HTML/PDF.
type FinancialSummaryReportDocument struct {
	Meta ReportMeta                  `json:"meta"`
	Rows []FinancialSummaryReportRow `json:"rows"`
}

type FinancialSummaryReportRow struct {
	SalonAddress    string    `json:"salon_address"`
	ExpenseItem     string    `json:"expense_item"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	TotalBalance    float64   `json:"total_balance"`
}
