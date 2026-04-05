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
	MaterialCost float64 `json:"material_cost"`
}

type InventoryMovementRow struct {
	SalonID        uint    `json:"salon_id"`
	SalonAddress   string  `json:"salon_address"`
	MaterialName   string  `json:"material_name"`
	Unit           string  `json:"unit"`
	OpeningBalance float64 `json:"opening_balance"`
	Purchased      float64 `json:"purchased"`
	WrittenOff     float64 `json:"written_off"`
	CurrentBalance float64 `json:"current_balance"`
}

type ClientLoyaltyRow struct {
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

type CancelledBookingRow struct {
	BookingID           uint      `json:"booking_id"`
	PlannedVisit        time.Time `json:"planned_visit"`
	ClientFullName      string    `json:"client_full_name"`
	MasterFullName      string    `json:"master_full_name"`
	CancellationReason  string    `json:"cancellation_reason"`
	CancellationRatePct float64   `json:"cancellation_rate_pct"`
	Status              string    `json:"status"`
}

type FinancialSummaryRow struct {
	SalonID         uint      `json:"salon_id"`
	SalonAddress    string    `json:"salon_address"`
	ExpenseItem     string    `json:"expense_item"`
	Amount          float64   `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
	TotalBalance    float64   `json:"total_balance"`
}

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
			s.id AS salon_id,
			s.address AS address,
			COUNT(DISTINCT b.client_id) AS client_count,
			COUNT(bi.id) AS service_count,
			COALESCE(SUM(b.total_price), 0) AS total_revenue
		FROM salons s
		LEFT JOIN bookings b ON b.salon_id = s.id
			AND b.start_time BETWEEN ? AND ?
			AND b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
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
			WHERE b.start_time BETWEEN ? AND ?
			  AND b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
		)
		SELECT
			svc.id AS service_id,
			svc.name AS service_name,
			COUNT(bi.id) AS usage_count,
			CASE
				WHEN (SELECT total FROM totals) > 0
					THEN ROUND(COUNT(bi.id)::numeric / (SELECT total FROM totals), 4)
				ELSE 0
			END AS relative_freq
		FROM services svc
		LEFT JOIN booking_items bi ON bi.service_id = svc.id
		LEFT JOIN bookings b ON b.id = bi.booking_id
			AND b.start_time BETWEEN ? AND ?
			AND b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
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
			u.id AS master_id,
			u.full_name AS full_name,
			COUNT(bi.id) AS service_count,
			COALESCE(SUM(b.total_price), 0) AS revenue,
			COALESCE(SUM(sm.quantity_per_use * bi.quantity), 0) AS material_cost
		FROM users u
		JOIN bookings b ON b.master_id = u.id
			AND b.start_time BETWEEN ? AND ?
			AND b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
		LEFT JOIN booking_items bi ON bi.booking_id = b.id
		LEFT JOIN service_materials sm ON sm.service_id = bi.service_id
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

func (r *ReportRepository) GetInventoryMovement(from, to time.Time, salonID uint) ([]InventoryMovementRow, error) {
	var rows []InventoryMovementRow
	err := r.db.Raw(`
		WITH usage AS (
			SELECT
				b.salon_id,
				sm.material_id,
				COALESCE(SUM(sm.quantity_per_use * bi.quantity), 0) AS written_off
			FROM bookings b
			JOIN booking_items bi ON bi.booking_id = b.id
			JOIN service_materials sm ON sm.service_id = bi.service_id
			WHERE b.start_time BETWEEN ? AND ?
			  AND b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
			GROUP BY b.salon_id, sm.material_id
		),
		manual_usage AS (
			SELECT
				su.salon_id,
				sm.material_id,
				COALESCE(SUM(sm.quantity_per_use * su.quantity), 0) AS written_off
			FROM service_usages su
			JOIN service_materials sm ON sm.service_id = su.service_id
			WHERE su.created_at BETWEEN ? AND ?
			GROUP BY su.salon_id, sm.material_id
		),
		total_usage AS (
			SELECT
				salon_id,
				material_id,
				COALESCE(SUM(written_off), 0) AS written_off
			FROM (
				SELECT salon_id, material_id, written_off FROM usage
				UNION ALL
				SELECT salon_id, material_id, written_off FROM manual_usage
			) src
			GROUP BY salon_id, material_id
		),
		purchases AS (
			SELECT
				me.salon_id,
				me.material_id,
				COALESCE(SUM(me.quantity), 0) AS purchased
			FROM material_expenses me
			WHERE me.created_at BETWEEN ? AND ?
			GROUP BY me.salon_id, me.material_id
		),
		row_source AS (
			SELECT salon_id, material_id FROM total_usage
			UNION
			SELECT salon_id, material_id FROM purchases
			UNION
			SELECT salon_id, material_id FROM inventories
		)
		SELECT
			rs.salon_id AS salon_id,
			s.address AS salon_address,
			m.name AS material_name,
			m.unit AS unit,
			COALESCE(inv.quantity, 0) + COALESCE(u.written_off, 0) - COALESCE(p.purchased, 0) AS opening_balance,
			COALESCE(p.purchased, 0) AS purchased,
			COALESCE(u.written_off, 0) AS written_off,
			COALESCE(inv.quantity, 0) AS current_balance
		FROM row_source rs
		JOIN salons s ON s.id = rs.salon_id
		JOIN materials m ON m.id = rs.material_id
		LEFT JOIN inventories inv ON inv.salon_id = rs.salon_id AND inv.material_id = rs.material_id
		LEFT JOIN total_usage u ON u.salon_id = rs.salon_id AND u.material_id = rs.material_id
		LEFT JOIN purchases p ON p.salon_id = rs.salon_id AND p.material_id = rs.material_id
		WHERE (? = 0 OR rs.salon_id = ?)
		ORDER BY s.address, m.name
	`, from, to, from, to, from, to, salonID, salonID).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) GetClientLoyalty(from, to time.Time) ([]ClientLoyaltyRow, error) {
	var rows []ClientLoyaltyRow
	err := r.db.Raw(`
		WITH visit_stats AS (
			SELECT
				b.client_id,
				MIN(b.start_time) FILTER (WHERE b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')) AS first_visit,
				MAX(b.start_time) FILTER (WHERE b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')) AS last_visit,
				COUNT(*) FILTER (
					WHERE b.status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
					  AND b.start_time BETWEEN ? AND ?
				) AS visit_count
			FROM bookings b
			GROUP BY b.client_id
		),
		payment_stats AS (
			SELECT
				b.client_id,
				COALESCE(SUM(p.amount) FILTER (
					WHERE p.status = 'SUCCESS'
					  AND COALESCE(p.completed_at, p.created_at) BETWEEN ? AND ?
				), 0) AS paid_total,
				COUNT(*) FILTER (
					WHERE p.status = 'SUCCESS'
					  AND COALESCE(p.completed_at, p.created_at) BETWEEN ? AND ?
				) AS successful_payments
			FROM payments p
			JOIN bookings b ON b.id = p.booking_id
			GROUP BY b.client_id
		)
		SELECT
			u.id AS client_id,
			u.full_name AS full_name,
			u.phone AS phone,
			COALESCE(u.email, '') AS email,
			vs.first_visit AS first_visit,
			vs.last_visit AS last_visit,
			COALESCE(vs.visit_count, 0) AS visit_count,
			COALESCE(ps.paid_total, 0) AS paid_total,
			CASE
				WHEN COALESCE(ps.successful_payments, 0) >= 3 OR COALESCE(ps.paid_total, 0) >= 10000 THEN 'ACTIVE'
				ELSE 'BASIC'
			END AS bonus_status
		FROM users u
		LEFT JOIN visit_stats vs ON vs.client_id = u.id
		LEFT JOIN payment_stats ps ON ps.client_id = u.id
		WHERE u.role = 'CLIENT'
		  AND u.deleted_at IS NULL
		ORDER BY paid_total DESC, last_visit DESC NULLS LAST, u.full_name
	`, from, to, from, to, from, to).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) GetCancelledBookings(from, to time.Time) ([]CancelledBookingRow, error) {
	var rows []CancelledBookingRow
	err := r.db.Raw(`
		WITH totals AS (
			SELECT
				COUNT(*) AS total_requests,
				COUNT(*) FILTER (
					WHERE status = 'CANCELLED'
					   OR (status = 'PENDING' AND start_time < NOW())
				) AS cancelled_count
			FROM bookings
			WHERE start_time BETWEEN ? AND ?
		)
		SELECT
			b.id AS booking_id,
			b.start_time AS planned_visit,
			client.full_name AS client_full_name,
			COALESCE(master.full_name, 'Не назначен') AS master_full_name,
			CASE
				WHEN b.status = 'CANCELLED' AND NULLIF(b.notes, '') IS NOT NULL THEN b.notes
				WHEN b.status = 'CANCELLED' THEN 'Отмена без указания причины'
				ELSE 'Визит не был подтверждён или не состоялся'
			END AS cancellation_reason,
			CASE
				WHEN totals.total_requests > 0 THEN ROUND((totals.cancelled_count::numeric / totals.total_requests) * 100, 2)
				ELSE 0
			END AS cancellation_rate_pct,
			b.status AS status
		FROM bookings b
		JOIN users client ON client.id = b.client_id
		LEFT JOIN users master ON master.id = b.master_id
		CROSS JOIN totals
		WHERE b.start_time BETWEEN ? AND ?
		  AND (
				b.status = 'CANCELLED'
				OR (b.status = 'PENDING' AND b.start_time < NOW())
		  )
		ORDER BY b.start_time DESC
	`, from, to, from, to).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) GetFinancialSummary(from, to time.Time, salonID uint) ([]FinancialSummaryRow, error) {
	var rows []FinancialSummaryRow
	err := r.db.Raw(`
		WITH tx AS (
			SELECT
				b.salon_id AS salon_id,
				s.address AS salon_address,
				CASE p.status
					WHEN 'SUCCESS' THEN 'Выручка от оплаченных услуг'
					WHEN 'REFUNDED' THEN 'Возврат клиенту'
					WHEN 'FAILED' THEN 'Неуспешная транзакция'
					ELSE 'Ожидаемое поступление'
				END AS expense_item,
				p.amount AS amount,
				COALESCE(p.completed_at, p.created_at, b.start_time) AS transaction_date,
				CASE
					WHEN p.status = 'SUCCESS' THEN p.amount
					WHEN p.status = 'REFUNDED' THEN -p.amount
					ELSE 0
				END AS signed_amount
			FROM payments p
			JOIN bookings b ON b.id = p.booking_id
			JOIN salons s ON s.id = b.salon_id
			WHERE COALESCE(p.completed_at, p.created_at, b.start_time) BETWEEN ? AND ?
			  AND (? = 0 OR b.salon_id = ?)
			UNION ALL
			SELECT
				me.salon_id AS salon_id,
				s.address AS salon_address,
				'Закупка материала: ' || m.name AS expense_item,
				(me.purchase_price * me.quantity) AS amount,
				me.created_at AS transaction_date,
				-(me.purchase_price * me.quantity) AS signed_amount
			FROM material_expenses me
			JOIN salons s ON s.id = me.salon_id
			JOIN materials m ON m.id = me.material_id
			WHERE me.created_at BETWEEN ? AND ?
			  AND (? = 0 OR me.salon_id = ?)
		)
		SELECT
			salon_id,
			salon_address,
			expense_item,
			amount,
			transaction_date,
			COALESCE(SUM(signed_amount) OVER (PARTITION BY salon_id), 0) AS total_balance
		FROM tx
		ORDER BY salon_address, transaction_date, expense_item
	`, from, to, salonID, salonID, from, to, salonID, salonID).Scan(&rows).Error
	return rows, err
}
