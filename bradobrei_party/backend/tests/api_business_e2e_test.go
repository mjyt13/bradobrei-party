package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"

	"testing"
)

func TestBookingAccessAndFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	_, workerToken := hireEmployeeAndLogin(t, app, adminToken)
	clientToken := registerClientAndLogin(t, app, "client_petr")
	serviceID := seedBookableFixture(t, app)

	createPayload := bookingPayload(app.baseSalonID, serviceID, time.Now().Add(24*time.Hour).UTC())

	status, body := app.request(t, http.MethodPost, "/api/v1/bookings", "", createPayload)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token for booking create, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/bookings", workerToken, createPayload)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on booking create, got %d: %s", status, string(body))
	}

	badPayload := bookingPayload(app.baseSalonID, serviceID, time.Now().Add(25*time.Hour).UTC())
	badPayload["start_time"] = "bad-time"
	status, body = app.request(t, http.MethodPost, "/api/v1/bookings", clientToken, badPayload)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid booking payload, got %d: %s", status, string(body))
	}

	bookingID, createBody := createBookingAsClient(t, app, clientToken, createPayload)
	app.saveArtifact(t, "booking_create", createBody)

	status, body = app.request(t, http.MethodGet, "/api/v1/bookings/my", clientToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 from /bookings/my, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "booking_get_my", body)

	status, body = app.request(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/confirm", bookingID), workerToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for booking confirm, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "booking_confirm", body)

	status, body = app.request(t, http.MethodGet, "/api/v1/bookings/master", workerToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 from /bookings/master, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "booking_master_list", body)

	status, body = app.request(t, http.MethodDelete, fmt.Sprintf("/api/v1/bookings/%d", bookingID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for booking delete, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "booking_delete", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/bookings/%d", bookingID), adminToken, nil)
	if status != http.StatusNotFound {
		t.Fatalf("expected 404 after booking delete, got %d: %s", status, string(body))
	}
}

func TestPaymentAccessAndFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	clientToken := registerClientAndLogin(t, app, "client_olga")
	serviceID := seedBookableFixture(t, app)
	bookingID, _ := createBookingAsClient(t, app, clientToken, bookingPayload(app.baseSalonID, serviceID, time.Now().Add(24*time.Hour).UTC()))

	status, body := app.request(t, http.MethodGet, "/api/v1/payments", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token for payments list, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/payments", clientToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for client on payments list, got %d: %s", status, string(body))
	}

	badPayload := map[string]any{
		"booking_id": 999999,
		"amount":     1200,
		"status":     "SUCCESS",
	}
	status, body = app.request(t, http.MethodPost, "/api/v1/payments", adminToken, badPayload)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payment payload, got %d: %s", status, string(body))
	}

	createPayload := map[string]any{
		"booking_id":              bookingID,
		"amount":                  0,
		"status":                  "SUCCESS",
		"external_transaction_id": "txn_test_success_001",
	}
	status, body = app.request(t, http.MethodPost, "/api/v1/payments", adminToken, createPayload)
	if status != http.StatusCreated {
		t.Fatalf("expected 201 for payment create, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "payment_create", body)

	var createdPayment struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &createdPayment); err != nil {
		t.Fatalf("failed to decode payment create response: %v", err)
	}

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/payments/%d", createdPayment.ID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for payment get by id, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "payment_get", body)
}

func TestMaterialExpensesAndServiceUseFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	_, workerToken := hireEmployeeAndLogin(t, app, adminToken)
	serviceID := seedBookableFixture(t, app)

	service, err := repository.NewServiceRepository(app.db).GetByID(serviceID)
	if err != nil {
		t.Fatalf("failed to load seeded service: %v", err)
	}
	if len(service.Materials) == 0 {
		t.Fatal("expected seeded service to have materials")
	}
	materialID := service.Materials[0].MaterialID

	status, body := app.request(t, http.MethodPost, "/api/v1/material-expenses", "", map[string]any{
		"material_id":    materialID,
		"salon_id":       app.baseSalonID,
		"purchase_price": 350,
		"quantity":       5,
	})
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token for material expense create, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/material-expenses", workerToken, map[string]any{
		"material_id":    materialID,
		"salon_id":       app.baseSalonID,
		"purchase_price": 350,
		"quantity":       5,
	})
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on material expense create, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/material-expenses", adminToken, map[string]any{
		"material_id":    materialID,
		"salon_id":       app.baseSalonID,
		"purchase_price": 350,
		"quantity":       5,
	})
	if status != http.StatusCreated {
		t.Fatalf("expected 201 for material expense create, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "material_expense_create", body)

	inventory, err := repository.NewInventoryRepository(app.db).GetItem(app.baseSalonID, materialID)
	if err != nil {
		t.Fatalf("failed to load inventory after material expense: %v", err)
	}
	if inventory.Quantity != 25 {
		t.Fatalf("expected inventory quantity 25 after purchase, got %.2f", inventory.Quantity)
	}

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/inventories/salon/%d", app.baseSalonID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for inventory list, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "inventory_list", body)

	status, body = app.request(t, http.MethodPut, fmt.Sprintf("/api/v1/inventories/salon/%d/material/%d", app.baseSalonID, materialID), workerToken, map[string]any{
		"quantity": 40,
	})
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on inventory update, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPut, fmt.Sprintf("/api/v1/inventories/salon/%d/material/%d", app.baseSalonID, materialID), adminToken, map[string]any{
		"quantity": 40,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200 for inventory set quantity, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "inventory_update", body)

	inventory, err = repository.NewInventoryRepository(app.db).GetItem(app.baseSalonID, materialID)
	if err != nil {
		t.Fatalf("failed to load inventory after direct update: %v", err)
	}
	if inventory.Quantity != 40 {
		t.Fatalf("expected inventory quantity 40 after direct update, got %.2f", inventory.Quantity)
	}

	status, body = app.request(t, http.MethodPost, fmt.Sprintf("/api/v1/services/%d/use", serviceID), adminToken, map[string]any{
		"salon_id": app.baseSalonID,
		"quantity": 2,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200 for manual service use, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "service_use", body)

	inventory, err = repository.NewInventoryRepository(app.db).GetItem(app.baseSalonID, materialID)
	if err != nil {
		t.Fatalf("failed to load inventory after service use: %v", err)
	}
	if inventory.Quantity != 38 {
		t.Fatalf("expected inventory quantity 38 after service use, got %.2f", inventory.Quantity)
	}

	from := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	to := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/inventory-movement?from=%s&to=%s&salon_id=%d", from, to, app.baseSalonID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for inventory movement report after service use, got %d: %s", status, string(body))
	}

	var reportResp struct {
		Data []struct {
			MaterialName string  `json:"material_name"`
			Purchased    float64 `json:"purchased"`
			WrittenOff   float64 `json:"written_off"`
			Current      float64 `json:"current_balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &reportResp); err != nil {
		t.Fatalf("failed to decode inventory report: %v", err)
	}
	if len(reportResp.Data) == 0 {
		t.Fatal("expected inventory report to contain rows")
	}

	found := false
	for _, row := range reportResp.Data {
		if row.MaterialName != "" {
			if row.Purchased == 5 && row.WrittenOff == 2 && row.Current == 38 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected inventory report row with purchased=5, written_off=2, current_balance=38, got: %s", string(body))
	}

	status, body = app.request(t, http.MethodPost, fmt.Sprintf("/api/v1/services/%d/use", serviceID), adminToken, map[string]any{
		"salon_id": app.baseSalonID,
		"quantity": 0,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid service use payload, got %d: %s", status, string(body))
	}
}

func TestReportsAccessAndFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	_, workerToken := hireEmployeeAndLogin(t, app, adminToken)
	clientToken := registerClientAndLogin(t, app, "client_maria")
	serviceID := seedBookableFixture(t, app)

	reportTime := time.Now().Add(48 * time.Hour).UTC()
	bookingID, _ := createBookingAsClient(t, app, clientToken, bookingPayload(app.baseSalonID, serviceID, reportTime))
	cancelledBookingID, _ := createBookingAsClient(t, app, clientToken, bookingPayload(app.baseSalonID, serviceID, reportTime.Add(2*time.Hour)))

	status, body := app.request(t, http.MethodGet, "/api/v1/reports/employees", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token for reports, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/reports/employees", clientToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for client on reports, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/reports/salon-activity?from=bad-date&to=2026-12-31", adminToken, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid report period, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/confirm", bookingID), workerToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for booking confirm before reports, got %d: %s", status, string(body))
	}

	markBookingCompleted(t, app, bookingID)
	status, body = app.request(t, http.MethodPost, fmt.Sprintf("/api/v1/bookings/%d/cancel", cancelledBookingID), clientToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for booking cancel before reports, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/payments", adminToken, map[string]any{
		"booking_id":              bookingID,
		"amount":                  1800,
		"status":                  "SUCCESS",
		"external_transaction_id": "txn_reports_001",
	})
	if status != http.StatusCreated {
		t.Fatalf("expected 201 for payment create before reports, got %d: %s", status, string(body))
	}

	reviewID := createReviewAsClient(t, app, clientToken)

	from := reportTime.Add(-24 * time.Hour).Format("2006-01-02")
	to := reportTime.Add(3 * 24 * time.Hour).Format("2006-01-02")

	status, body = app.request(t, http.MethodGet, "/api/v1/reports/employees", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for employee report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_employees", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/salon-activity?from=%s&to=%s", from, to), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for salon activity report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_salon_activity", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/service-popularity?from=%s&to=%s", from, to), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for service popularity report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_service_popularity", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/master-activity?from=%s&to=%s", from, to), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for master activity report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_master_activity", body)

	status, body = app.request(t, http.MethodGet, "/api/v1/reports/reviews", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for reviews report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_reviews", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/inventory-movement?from=%s&to=%s&salon_id=%d", from, to, app.baseSalonID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for inventory movement report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_inventory_movement", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/client-loyalty?from=%s&to=%s", from, to), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for client loyalty report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_client_loyalty", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/cancelled-bookings?from=%s&to=%s", from, to), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for cancelled bookings report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_cancelled_bookings", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reports/financial-summary?from=%s&to=%s&salon_id=%d", from, to, app.baseSalonID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for financial summary report, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "report_financial_summary", body)

	htmlReports := []string{
		"/api/v1/reports/employees/html",
		fmt.Sprintf("/api/v1/reports/salon-activity/html?from=%s&to=%s", from, to),
		fmt.Sprintf("/api/v1/reports/service-popularity/html?from=%s&to=%s", from, to),
		fmt.Sprintf("/api/v1/reports/master-activity/html?from=%s&to=%s", from, to),
		"/api/v1/reports/reviews/html",
		fmt.Sprintf("/api/v1/reports/inventory-movement/html?from=%s&to=%s&salon_id=%d", from, to, app.baseSalonID),
		fmt.Sprintf("/api/v1/reports/client-loyalty/html?from=%s&to=%s", from, to),
		fmt.Sprintf("/api/v1/reports/cancelled-bookings/html?from=%s&to=%s", from, to),
		fmt.Sprintf("/api/v1/reports/financial-summary/html?from=%s&to=%s&salon_id=%d", from, to, app.baseSalonID),
	}
	for _, path := range htmlReports {
		status, body = app.request(t, http.MethodGet, path, adminToken, nil)
		if status != http.StatusOK {
			t.Fatalf("expected 200 for html report %s, got %d: %s", path, status, string(body))
		}
		if !looksLikeReportHTML(body) {
			t.Fatalf("expected html report body for %s, got: %s", path, string(body))
		}
	}

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reviews/%d", reviewID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for review get by id, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "review_get", body)
}

func TestReviewAccessAndFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	clientToken := registerClientAndLogin(t, app, "client_review")
	_, workerToken := hireEmployeeAndLogin(t, app, adminToken)

	status, body := app.request(t, http.MethodPost, "/api/v1/reviews", "", map[string]any{
		"text":   "Отзыв без токена",
		"rating": 5,
	})
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token for review create, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/reviews", clientToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for client on reviews list, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/reviews", workerToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on reviews list, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/reviews", clientToken, map[string]any{
		"text":   "",
		"rating": 10,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid review payload, got %d: %s", status, string(body))
	}

	reviewID := createReviewAsClient(t, app, clientToken)

	status, body = app.request(t, http.MethodGet, "/api/v1/reviews", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for admin reviews list, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "review_list", body)

	status, body = app.request(t, http.MethodGet, fmt.Sprintf("/api/v1/reviews/%d", reviewID), adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for admin review get by id, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "review_get_by_id", body)
}

func registerClientAndLogin(t *testing.T, app *testApp, username string) string {
	t.Helper()

	status, body := app.request(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"username":  username,
		"password":  "password123",
		"full_name": "Test Client",
		"phone":     "+79990002233",
		"email":     username + "@example.com",
		"role":      "CLIENT",
	})
	if status != http.StatusCreated {
		t.Fatalf("expected client register 201, got %d: %s", status, string(body))
	}

	return loginAndGetToken(t, app, username, "password123")
}

func seedBookableFixture(t *testing.T, app *testApp) uint {
	t.Helper()

	suffix := time.Now().UnixNano()
	material := models.Material{
		Name: fmt.Sprintf("Material-%d", suffix),
		Unit: "ml",
	}
	if err := app.db.Create(&material).Error; err != nil {
		t.Fatalf("failed to seed material: %v", err)
	}

	service := models.Service{
		Name:            fmt.Sprintf("Service-%d", suffix),
		Description:     "Test service for API e2e",
		Price:           1800,
		DurationMinutes: 75,
	}
	if err := app.db.Create(&service).Error; err != nil {
		t.Fatalf("failed to seed service: %v", err)
	}

	serviceMaterial := models.ServiceMaterial{
		ServiceID:      service.ID,
		MaterialID:     material.ID,
		QuantityPerUse: 1,
	}
	if err := app.db.Create(&serviceMaterial).Error; err != nil {
		t.Fatalf("failed to seed service materials: %v", err)
	}

	inventory := models.Inventory{
		SalonID:     app.baseSalonID,
		MaterialID:  material.ID,
		Quantity:    20,
		LastUpdated: time.Now(),
	}
	if err := app.db.Create(&inventory).Error; err != nil {
		t.Fatalf("failed to seed inventory: %v", err)
	}

	return service.ID
}

func bookingPayload(salonID, serviceID uint, start time.Time) map[string]any {
	return map[string]any{
		"start_time":  start.Format(time.RFC3339),
		"salon_id":    salonID,
		"service_ids": []uint{serviceID},
		"notes":       "E2E booking scenario",
	}
}

func createBookingAsClient(t *testing.T, app *testApp, clientToken string, payload map[string]any) (uint, []byte) {
	t.Helper()

	status, body := app.request(t, http.MethodPost, "/api/v1/bookings", clientToken, payload)
	if status != http.StatusCreated {
		t.Fatalf("expected booking create 201, got %d: %s", status, string(body))
	}

	var resp struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode booking response: %v", err)
	}

	return resp.ID, body
}

func markBookingCompleted(t *testing.T, app *testApp, bookingID uint) {
	t.Helper()

	if err := app.db.Model(&models.Booking{}).
		Where("id = ?", bookingID).
		Update("status", models.BookingCompleted).Error; err != nil {
		t.Fatalf("failed to mark booking as completed: %v", err)
	}
}

func createReviewAsClient(t *testing.T, app *testApp, clientToken string) uint {
	t.Helper()

	status, body := app.request(t, http.MethodPost, "/api/v1/reviews", clientToken, map[string]any{
		"text":   "Отличный тестовый отзыв",
		"rating": 5,
	})
	if status != http.StatusCreated {
		t.Fatalf("expected review create 201, got %d: %s", status, string(body))
	}

	var resp struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode review create response: %v", err)
	}

	app.saveArtifact(t, "review_create", body)
	return resp.ID
}

func looksLikeReportHTML(body []byte) bool {
	html := string(body)
	return strings.Contains(html, "<!doctype html>") ||
		strings.Contains(html, "<html") ||
		strings.Contains(html, "report.css")
}
