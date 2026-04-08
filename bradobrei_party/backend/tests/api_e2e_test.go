package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"bradobrei/backend/internal/handlers"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	reportspkg "bradobrei/backend/internal/reports"
	"bradobrei/backend/internal/repository"
	"bradobrei/backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testApp struct {
	router       *gin.Engine
	rootDir      string
	artifactPath string
	db           *gorm.DB
	adminDB      *gorm.DB
	schema       string
	baseSalonID  uint
}

func TestAuthFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	assertLoginFailure(t, app, "admin", "wrong-password")

	adminToken := loginAndGetToken(t, app, "admin", "password")
	status, body := app.request(t, http.MethodGet, "/api/v1/me", adminToken, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200 from /me, got %d: %s", status, string(body))
	}
}

func TestEmployeeAccessAndCrudFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	assertLoginFailure(t, app, "admin", "wrong-password")
	adminToken := loginAndGetToken(t, app, "admin", "password")

	employeeID, _ := hireEmployeeAndLogin(t, app, adminToken)
	workerToken := loginAndGetToken(t, app, "worker_ivan", "password123")

	status, body := app.request(t, http.MethodGet, "/api/v1/employees", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodGet, "/api/v1/employees", workerToken, nil)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on /employees, got %d: %s", status, string(body))
	}

	badUpdate := map[string]any{
		"username":        "worker_ivan",
		"full_name":       "Иван Барбер",
		"phone":           "+79990001122",
		"email":           "not-an-email",
		"role":            "ADVANCED_MASTER",
		"specialization":  "Fade",
		"expected_salary": 90000,
		"work_schedule":   `{"mon":"10:00-19:00"}`,
		"salon_ids":       []uint{app.baseSalonID},
	}
	status, body = app.request(t, http.MethodPut, fmt.Sprintf("/api/v1/employees/%d", employeeID), adminToken, badUpdate)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid employee update, got %d: %s", status, string(body))
	}

	badSalaryPatch := map[string]any{
		"expected_salary": 0,
	}
	status, body = app.request(t, http.MethodPatch, fmt.Sprintf("/api/v1/employees/%d", employeeID), adminToken, badSalaryPatch)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero employee salary, got %d: %s", status, string(body))
	}

	badSchedulePatch := map[string]any{
		"work_schedule": `{"mon":"letters-instead-of-time"}`,
	}
	status, body = app.request(t, http.MethodPatch, fmt.Sprintf("/api/v1/employees/%d", employeeID), adminToken, badSchedulePatch)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid employee schedule, got %d: %s", status, string(body))
	}

	patchPayload := map[string]any{
		"full_name":       "Иван Барбер Обновлённый",
		"expected_salary": 95000,
		"salon_ids":       []uint{app.baseSalonID},
	}
	status, body = app.request(t, http.MethodPatch, fmt.Sprintf("/api/v1/employees/%d", employeeID), adminToken, patchPayload)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for employee patch, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "employee_update", body)
}

func TestSalonAccessAndCrudFlow(t *testing.T) {
	app := newTestApp(t)
	defer app.close(t)

	adminToken := loginAndGetToken(t, app, "admin", "password")
	_, _ = hireEmployeeAndLogin(t, app, adminToken)
	workerToken := loginAndGetToken(t, app, "worker_ivan", "password123")

	noTokenPayload := map[string]any{
		"name":             "Тестовый салон",
		"address":          "Екатеринбург",
		"location":         "56.8389, 60.6057",
		"working_hours":    `{"mon":"10:00-20:00"}`,
		"status":           "OPEN",
		"max_staff":        8,
		"base_hourly_rate": 1500,
	}
	status, body := app.request(t, http.MethodPost, "/api/v1/salons", "", noTokenPayload)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", status, string(body))
	}

	status, body = app.request(t, http.MethodPost, "/api/v1/salons", workerToken, noTokenPayload)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403 for worker on salon create, got %d: %s", status, string(body))
	}

	badSalonPayload := map[string]any{
		"name":             "Плохой салон",
		"address":          "Пермь",
		"location":         "POINT(bad)",
		"working_hours":    `{"mon":"10:00-20:00"}`,
		"status":           "OPEN",
		"max_staff":        8,
		"base_hourly_rate": 1400,
	}
	status, body = app.request(t, http.MethodPost, "/api/v1/salons", adminToken, badSalonPayload)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid salon payload, got %d: %s", status, string(body))
	}

	badScheduleSalonPayload := map[string]any{
		"name":             "Плохой график",
		"address":          "Пермь",
		"location":         "58.0141, 56.2230",
		"working_hours":    `{"mon":"aa:bb-cc:dd"}`,
		"status":           "OPEN",
		"max_staff":        8,
		"base_hourly_rate": 1400,
	}
	status, body = app.request(t, http.MethodPost, "/api/v1/salons", adminToken, badScheduleSalonPayload)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid salon schedule, got %d: %s", status, string(body))
	}

	createSalonPayload := map[string]any{
		"name":             "Золотые руки",
		"address":          "Пермь",
		"location":         "58.0141, 56.2230",
		"working_hours":    `{"mon":"10:00-20:00","tue":"10:00-20:00"}`,
		"status":           "OPEN",
		"max_staff":        8,
		"base_hourly_rate": 1400,
	}
	status, body = app.request(t, http.MethodPost, "/api/v1/salons", adminToken, createSalonPayload)
	if status != http.StatusCreated {
		t.Fatalf("expected 201 for salon create, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "salon_create", body)

	var createdSalon struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &createdSalon); err != nil {
		t.Fatalf("failed to decode created salon: %v", err)
	}

	updateSalonPayload := map[string]any{
		"name":             "Золотые руки Плюс",
		"address":          "Пермь, центр",
		"location":         "58.0141, 56.2230",
		"working_hours":    `{"mon":"09:00-21:00","tue":"09:00-21:00"}`,
		"status":           "OPEN",
		"max_staff":        10,
		"base_hourly_rate": 1700,
	}
	status, body = app.request(t, http.MethodPut, fmt.Sprintf("/api/v1/salons/%d", createdSalon.ID), adminToken, updateSalonPayload)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for salon update, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "salon_update", body)
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rootDir := backendRoot(t)
	_ = godotenv.Load(filepath.Join(rootDir, ".env"))
	os.Setenv("JWT_SECRET", "test-secret")

	required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			t.Skipf("skipping integration tests: %s is not configured", key)
		}
	}

	publicDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=public",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"),
	)
	adminDB, err := gorm.Open(postgres.Open(publicDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect admin db: %v", err)
	}

	schema := fmt.Sprintf("test_api_%d", time.Now().UnixNano())
	if err := adminDB.Exec("CREATE SCHEMA IF NOT EXISTS " + schema).Error; err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	testDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=%s,public",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"), schema,
	)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{}, &models.EmployeeProfile{}, &models.Salon{},
		&models.Service{}, &models.Material{}, &models.ServiceMaterial{},
		&models.Inventory{}, &models.MaterialExpense{}, &models.ServiceUsage{}, &models.Booking{}, &models.BookingItem{},
		&models.Payment{}, &models.Review{}, &models.ErrorLog{},
	); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	baseSalon := models.Salon{
		Name:           "Базовый салон",
		Address:        "Екатеринбург",
		Status:         "OPEN",
		MaxStaff:       5,
		BaseHourlyRate: 1200,
	}
	if err := db.Create(&baseSalon).Error; err != nil {
		t.Fatalf("failed to seed salon: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash admin password: %v", err)
	}
	adminUser := models.User{
		Username:     "admin",
		PasswordHash: string(hash),
		FullName:     "Администратор Системы",
		Phone:        "+79990000000",
		Role:         models.RoleAdmin,
	}
	if err := db.Create(&adminUser).Error; err != nil {
		t.Fatalf("failed to seed admin: %v", err)
	}

	router := buildTestRouter(db)
	return &testApp{
		router:       router,
		rootDir:      rootDir,
		artifactPath: filepath.Join(rootDir, "test_artifacts", "api_outputs.json"),
		db:           db,
		adminDB:      adminDB,
		schema:       schema,
		baseSalonID:  baseSalon.ID,
	}
}

func buildTestRouter(db *gorm.DB) *gin.Engine {
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	salonRepo := repository.NewSalonRepository(db)
	invRepo := repository.NewInventoryRepository(db)
	reportRepo := repository.NewReportRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	authSvc := services.NewAuthService(userRepo)
	bookingSvc := services.NewBookingService(bookingRepo, invRepo, db)
	salonSvc := services.NewSalonService(salonRepo, nil)
	reportSvc := services.NewReportService(reportRepo)
	employeeSvc := services.NewEmployeeService(employeeRepo, userRepo, db)
	paymentSvc := services.NewPaymentService(paymentRepo, bookingRepo)
	serviceSvc := services.NewServiceService(repository.NewServiceRepository(db), employeeRepo, invRepo, db)
	materialExpenseSvc := services.NewMaterialExpenseService(db)
	inventorySvc := services.NewInventoryService(invRepo)

	authH := handlers.NewAuthHandler(authSvc)
	bookingH := handlers.NewBookingHandler(bookingSvc)
	salonH := handlers.NewSalonHandler(salonSvc)
	reportRenderer, err := reportspkg.NewRenderer(nil)
	if err != nil {
		panic(err)
	}
	reportH := handlers.NewReportHandler(reportSvc)
	reportFileH := handlers.NewReportFileHandler(reportSvc, reportRenderer)
	employeeH := handlers.NewEmployeeHandler(employeeSvc)
	reviewH := handlers.NewReviewHandler(db)
	paymentH := handlers.NewPaymentHandler(paymentSvc)
	serviceH := handlers.NewServiceHandler(serviceSvc)
	materialExpenseH := handlers.NewMaterialExpenseHandler(materialExpenseSvc)
	inventoryH := handlers.NewInventoryHandler(inventorySvc)

	r := gin.New()
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/register", authH.Register)
	}

	api := v1.Group("/")
	api.Use(middleware.AuthRequired(os.Getenv("JWT_SECRET")))
	{
		api.GET("/me", authH.Me)

		salons := api.Group("/salons")
		{
			salons.GET("", salonH.GetAll)
			salons.GET("/:id", salonH.GetByID)
			salons.GET("/:id/masters", salonH.GetMasters)
			salons.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Create)
			salons.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Update)
		}

		emps := api.Group("/employees")
		{
			emps.GET("", middleware.RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleNetworkManager), employeeH.GetAll)
			emps.GET("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleNetworkManager), employeeH.GetByID)
			emps.GET("/me", middleware.RequireRoles(
				models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleHR,
				models.RoleAccountant, models.RoleNetworkManager, models.RoleAdmin,
			), employeeH.GetMe)
			emps.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleHR), employeeH.Hire)
			emps.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleHR), employeeH.Update)
			emps.PATCH("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleHR), employeeH.Patch)
			emps.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleHR), employeeH.Fire)
		}

		bookings := api.Group("/bookings")
		{
			bookings.GET("", middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager), bookingH.GetAll)
			bookings.POST("", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.Create)
			bookings.GET("/my", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.GetMy)
			bookings.GET("/master", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleAdmin), bookingH.GetByMaster)
			bookings.GET("/:id", bookingH.GetByID)
			bookings.DELETE("/:id", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.Delete)
			bookings.POST("/:id/confirm", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleAdmin), bookingH.Confirm)
			bookings.POST("/:id/cancel", bookingH.Cancel)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("", reviewH.Create)
			reviews.GET("", middleware.RequireRoles(models.RoleAdmin), reviewH.GetAll)
			reviews.GET("/:id", middleware.RequireRoles(models.RoleAdmin), reviewH.GetByID)
		}

		payments := api.Group("/payments")
		{
			payments.GET("", middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager), paymentH.GetAll)
			payments.GET("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager), paymentH.GetByID)
			payments.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant), paymentH.Create)
		}

		svcs := api.Group("/services")
		{
			svcs.POST("/:id/use", middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster, models.RoleBasicMaster), serviceH.Use)
		}

		materialExpenses := api.Group("/material-expenses")
		materialExpenses.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant))
		{
			materialExpenses.GET("", materialExpenseH.GetAll)
			materialExpenses.GET("/:id", materialExpenseH.GetByID)
			materialExpenses.POST("", materialExpenseH.Create)
			materialExpenses.PUT("/:id", materialExpenseH.Update)
			materialExpenses.DELETE("/:id", materialExpenseH.Delete)
		}

		inventories := api.Group("/inventories")
		{
			inventories.GET("/salon/:salonId",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager, models.RoleAdvancedMaster),
				inventoryH.GetBySalon)
			inventories.PUT("/salon/:salonId/material/:materialId",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant),
				inventoryH.SetQuantity)
		}

		reports := api.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager, models.RoleHR))
		{
			reports.GET("/employees", reportH.Employees)
			reports.GET("/employees/html", reportFileH.EmployeesHTML)
			reports.GET("/employees/pdf", reportFileH.EmployeesPDF)
			reports.GET("/salon-activity", reportH.SalonActivity)
			reports.GET("/salon-activity/html", reportFileH.SalonActivityHTML)
			reports.GET("/salon-activity/pdf", reportFileH.SalonActivityPDF)
			reports.GET("/service-popularity", reportH.ServicePopularity)
			reports.GET("/service-popularity/html", reportFileH.ServicePopularityHTML)
			reports.GET("/service-popularity/pdf", reportFileH.ServicePopularityPDF)
			reports.GET("/master-activity", reportH.MasterActivity)
			reports.GET("/master-activity/html", reportFileH.MasterActivityHTML)
			reports.GET("/master-activity/pdf", reportFileH.MasterActivityPDF)
			reports.GET("/reviews", reportH.Reviews)
			reports.GET("/reviews/html", reportFileH.ReviewsHTML)
			reports.GET("/reviews/pdf", reportFileH.ReviewsPDF)
			reports.GET("/inventory-movement", reportH.InventoryMovement)
			reports.GET("/inventory-movement/html", reportFileH.InventoryMovementHTML)
			reports.GET("/inventory-movement/pdf", reportFileH.InventoryMovementPDF)
			reports.GET("/client-loyalty", reportH.ClientLoyalty)
			reports.GET("/client-loyalty/html", reportFileH.ClientLoyaltyHTML)
			reports.GET("/client-loyalty/pdf", reportFileH.ClientLoyaltyPDF)
			reports.GET("/cancelled-bookings", reportH.CancelledBookings)
			reports.GET("/cancelled-bookings/html", reportFileH.CancelledBookingsHTML)
			reports.GET("/cancelled-bookings/pdf", reportFileH.CancelledBookingsPDF)
			reports.GET("/financial-summary", reportH.FinancialSummary)
			reports.GET("/financial-summary/html", reportFileH.FinancialSummaryHTML)
			reports.GET("/financial-summary/pdf", reportFileH.FinancialSummaryPDF)
		}
	}

	return r
}

func (a *testApp) close(t *testing.T) {
	t.Helper()
	if a.adminDB != nil {
		if err := a.adminDB.Exec("DROP SCHEMA IF EXISTS " + a.schema + " CASCADE").Error; err != nil {
			t.Fatalf("failed to drop schema: %v", err)
		}
	}
	if sqlDB, err := a.db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	if sqlDB, err := a.adminDB.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func (a *testApp) request(t *testing.T, method, path, token string, body any) (int, []byte) {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	a.router.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

func assertLoginFailure(t *testing.T, app *testApp, username, password string) {
	t.Helper()
	status, body := app.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	if status != http.StatusUnauthorized {
		t.Fatalf("expected failed login 401, got %d: %s", status, string(body))
	}
}

func loginAndGetToken(t *testing.T, app *testApp, username, password string) string {
	t.Helper()
	status, body := app.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	if status != http.StatusOK {
		t.Fatalf("expected successful login 200, got %d: %s", status, string(body))
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	return resp.Token
}

func hireEmployeeAndLogin(t *testing.T, app *testApp, adminToken string) (uint, string) {
	t.Helper()
	payload := map[string]any{
		"username":        "worker_ivan",
		"password":        "password123",
		"full_name":       "Иван Барбер",
		"phone":           "+79990001122",
		"email":           "worker.ivan@example.com",
		"role":            "ADVANCED_MASTER",
		"specialization":  "Fade, beard styling",
		"expected_salary": 85000,
		"work_schedule":   `{"mon":"10:00-19:00","wed":"10:00-19:00"}`,
		"salon_id":        app.baseSalonID,
	}

	status, body := app.request(t, http.MethodPost, "/api/v1/employees", adminToken, payload)
	if status != http.StatusCreated {
		t.Fatalf("expected employee create 201, got %d: %s", status, string(body))
	}
	app.saveArtifact(t, "employee_create", body)

	var resp struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to decode employee create response: %v", err)
	}

	token := loginAndGetToken(t, app, "worker_ivan", "password123")
	return resp.ID, token
}

func (a *testApp) saveArtifact(t *testing.T, key string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(a.artifactPath), 0o755); err != nil {
		t.Fatalf("failed to create artifact dir: %v", err)
	}

	store := map[string]json.RawMessage{}
	if existing, err := os.ReadFile(a.artifactPath); err == nil && len(existing) > 0 {
		_ = json.Unmarshal(existing, &store)
	}

	store[key] = json.RawMessage(body)
	pretty, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		t.Fatalf("failed to encode artifact file: %v", err)
	}
	if err := os.WriteFile(a.artifactPath, pretty, 0o644); err != nil {
		t.Fatalf("failed to write artifact file: %v", err)
	}
}

func backendRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
