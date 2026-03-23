package repository

import (
	"time"

	"bradobrei/backend/internal/models"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// --- DTO для отчётов ---

type EmployeeReportRow struct {
	FullName       string  `json:"full_name"`
	Specialization string  `json:"specialization"`
	Salons         string  `json:"salons"` // JSON-массив адресов
	ExpectedSalary float64 `json:"expected_salary"`
}

type SalonActivityRow struct {
	SalonID      uint    `json:"salon_id"`
	Address      string  `json:"address"`
	ClientCount  int64   `json:"client_count"`
	ServiceCount int64   `json:"service_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type ServicePopularityRow struct {
	ServiceID    uint    `json:"service_id"`
	ServiceName  string  `json:"service_name"`
	UsageCount   int64   `json:"usage_count"`
	RelativeFreq float64 `json:"relative_freq"` // доля от всех оказанных
}

type MasterActivityRow struct {
	MasterID     uint    `json:"master_id"`
	FullName     string  `json:"full_name"`
	ServiceCount int64   `json:"service_count"`
	Revenue      float64 `json:"revenue"`
	MaterialCost float64 `json:"material_cost"` // TODO: считать через ServiceMaterial
}

// Report 2.2.1 — Список действующих сотрудников
func (r *ReportRepository) GetEmployeeList() ([]models.User, error) {
	var users []models.User
	err := r.db.
		Where("role != ? AND deleted_at IS NULL", models.RoleClient).
		Preload("EmployeeProfile.Salons").
		Find(&users).Error
	return users, err
}

// Report 2.2.2 — Месячная активность сети
func (r *ReportRepository) GetSalonActivity(from, to time.Time) ([]SalonActivityRow, error) {
	var rows []SalonActivityRow
	err := r.db.Raw(`
		SELECT
			s.id        AS salon_id,
			s.address   AS address,
			COUNT(DISTINCT b.client_id)     AS client_count,
			COUNT(bi.id)                    AS service_count,
			COALESCE(SUM(b.total_price), 0) AS total_revenue
		FROM salons s
		LEFT JOIN bookings b ON b.salon_id = s.id
			AND b.start_time BETWEEN ? AND ?
			AND b.status = 'COMPLETED'
		LEFT JOIN booking_items bi ON bi.booking_id = b.id
		GROUP BY s.id, s.address
		ORDER BY total_revenue DESC
	`, from, to).Scan(&rows).Error
	return rows, err
}

// Report 2.2.3 — Популярность услуг
func (r *ReportRepository) GetServicePopularity(from, to time.Time) ([]ServicePopularityRow, error) {
	var rows []ServicePopularityRow
	err := r.db.Raw(`
		WITH totals AS (
			SELECT COUNT(*) AS total
			FROM booking_items bi
			JOIN bookings b ON b.id = bi.booking_id
			WHERE b.start_time BETWEEN ? AND ? AND b.status = 'COMPLETED'
		)
		SELECT
			svc.id          AS service_id,
			svc.name        AS service_name,
			COUNT(bi.id)    AS usage_count,
			CASE WHEN (SELECT total FROM totals) > 0
				THEN ROUND(COUNT(bi.id)::numeric / (SELECT total FROM totals), 4)
				ELSE 0
			END             AS relative_freq
		FROM services svc
		LEFT JOIN booking_items bi ON bi.service_id = svc.id
		LEFT JOIN bookings b ON b.id = bi.booking_id
			AND b.start_time BETWEEN ? AND ? AND b.status = 'COMPLETED'
		GROUP BY svc.id, svc.name
		ORDER BY usage_count DESC
	`, from, to, from, to).Scan(&rows).Error
	return rows, err
}

// Report 2.2.4 — Активность мастеров
func (r *ReportRepository) GetMasterActivity(from, to time.Time) ([]MasterActivityRow, error) {
	var rows []MasterActivityRow
	err := r.db.Raw(`
		SELECT
			u.id            AS master_id,
			u.full_name     AS full_name,
			COUNT(bi.id)    AS service_count,
			COALESCE(SUM(b.total_price), 0) AS revenue,
			0               AS material_cost
		FROM users u
		JOIN bookings b ON b.master_id = u.id
			AND b.start_time BETWEEN ? AND ?
			AND b.status = 'COMPLETED'
		LEFT JOIN booking_items bi ON bi.booking_id = b.id
		WHERE u.role IN ('BASIC_MASTER', 'ADVANCED_MASTER')
		GROUP BY u.id, u.full_name
		ORDER BY revenue DESC
	`, from, to).Scan(&rows).Error
	return rows, err
}

// Report 2.2.5 — Отзывы об ИС
func (r *ReportRepository) GetReviews(from, to time.Time) ([]models.Review, error) {
	var reviews []models.Review
	q := r.db.Preload("User")
	if !from.IsZero() {
		q = q.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("created_at <= ?", to)
	}
	err := q.Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}
