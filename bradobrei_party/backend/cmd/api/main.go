package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "bradobrei/backend/docs"
	"bradobrei/backend/internal/geocoder"
	"bradobrei/backend/internal/handlers"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	reportspkg "bradobrei/backend/internal/reports"
	"bradobrei/backend/internal/repository"
	"bradobrei/backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// @title Bradobrei Party API
// @version 1.0
// @description Backend API для системы управления сетью барбершопов Bradobrei Party.
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате `Bearer <token>`.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env не найден")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не задан")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s search_path=public",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Для пустой БД или вручную пересозданной базы гарантируем наличие рабочей схемы.
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS public").Error; err != nil {
		log.Fatal("Ошибка подготовки схемы public:", err)
	}

	// Для геополей salons.location требуется расширение PostGIS.
	// Если БД создана вручную, его может не быть даже при наличии схемы public.
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		log.Fatal("Ошибка подключения расширения postgis:", err)
	}

	if err := db.AutoMigrate(
		&models.User{}, &models.EmployeeProfile{}, &models.Salon{},
		&models.Service{}, &models.Material{}, &models.ServiceMaterial{},
		&models.Inventory{}, &models.MaterialExpense{}, &models.ServiceUsage{}, &models.Booking{}, &models.BookingItem{},
		&models.Payment{}, &models.Review{},
	); err != nil {
		log.Fatal("Ошибка миграции:", err)
	}
	log.Println("БД синхронизирована")

	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	salonRepo := repository.NewSalonRepository(db)
	invRepo := repository.NewInventoryRepository(db)
	reportRepo := repository.NewReportRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	materialRepo := repository.NewMaterialRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	authSvc := services.NewAuthService(userRepo)
	bookingSvc := services.NewBookingService(bookingRepo, invRepo, db)
	geoCoder, err := geocoder.NewFromEnv()
	if err != nil {
		log.Fatal("геокодер: ", err)
	}
	salonSvc := services.NewSalonService(salonRepo, geoCoder)
	reportSvc := services.NewReportService(reportRepo)
	serviceSvc := services.NewServiceService(serviceRepo, employeeRepo, invRepo, db)
	materialSvc := services.NewMaterialService(materialRepo)
	materialExpenseSvc := services.NewMaterialExpenseService(db)
	inventorySvc := services.NewInventoryService(invRepo)
	employeeSvc := services.NewEmployeeService(employeeRepo, userRepo, db)
	paymentSvc := services.NewPaymentService(paymentRepo, bookingRepo)

	authH := handlers.NewAuthHandler(authSvc)
	bookingH := handlers.NewBookingHandler(bookingSvc)
	salonH := handlers.NewSalonHandler(salonSvc)
	reportRenderer, err := reportspkg.NewRenderer(reportspkg.NewGotenbergClientFromEnv())
	if err != nil {
		log.Fatal("Ошибка инициализации renderer отчётов:", err)
	}

	reportH := handlers.NewReportHandler(reportSvc)
	reportFileH := handlers.NewReportFileHandler(reportSvc, reportRenderer)
	reviewH := handlers.NewReviewHandler(db)
	serviceH := handlers.NewServiceHandler(serviceSvc)
	materialH := handlers.NewMaterialHandler(materialSvc)
	materialExpenseH := handlers.NewMaterialExpenseHandler(materialExpenseSvc)
	inventoryH := handlers.NewInventoryHandler(inventorySvc)
	employeeH := handlers.NewEmployeeHandler(employeeSvc)
	paymentH := handlers.NewPaymentHandler(paymentSvc)

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(middleware.RecoveryWithLog())
	r.Use(middleware.ErrorLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive", "version": "1.0", "db": "connected"})
	})
	r.GET("/", authH.DocsRedirect)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/docs/*any", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	api := v1.Group("/")
	api.Use(middleware.AuthRequired(jwtSecret))
	{
		api.GET("/me", authH.Me)

		salons := api.Group("/salons")
		{
			salons.POST("/geocode", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.GeocodeAddress)
			salons.GET("", salonH.GetAll)
			salons.GET("/:id", salonH.GetByID)
			salons.GET("/:id/masters", salonH.GetMasters)
			salons.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Create)
			salons.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Update)
			salons.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), salonH.Delete)
		}

		svcs := api.Group("/services")
		{
			svcs.GET("", serviceH.GetAll)
			svcs.GET("/:id", serviceH.GetByID)
			svcs.GET("/my", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster), serviceH.GetMy)
			svcs.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster), serviceH.Create)
			svcs.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster), serviceH.Update)
			svcs.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), serviceH.Delete)
			svcs.POST("/:id/use",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster, models.RoleBasicMaster),
				serviceH.Use)
			svcs.POST("/:id/assign-master",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster, models.RoleBasicMaster),
				serviceH.AssignToMaster)
			svcs.DELETE("/:id/assign-master/:profileId",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster),
				serviceH.RemoveFromMaster)
		}

		mats := api.Group("/materials")
		mats.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster, models.RoleAccountant))
		{
			mats.GET("", materialH.GetAll)
			mats.GET("/:id", materialH.GetByID)
			mats.POST("", middleware.RequireRoles(models.RoleAdmin), materialH.Create)
			mats.PUT("/:id", middleware.RequireRoles(models.RoleAdmin), materialH.Update)
			mats.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), materialH.Delete)
			mats.PUT("/service/:serviceId",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster),
				materialH.SetServiceMaterials)
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
			emps.PATCH("/me/schedule", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster), employeeH.UpdateMySchedule)
			emps.POST("/:id/assign-salon", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), employeeH.AssignToSalon)
			emps.DELETE("/:id/assign-salon/:salonId", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), employeeH.RemoveFromSalon)
		}

		bookings := api.Group("/bookings")
		{
			bookings.GET("", middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager), bookingH.GetAll)
			bookings.POST("", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.Create)
			bookings.GET("/my", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.GetMy)
			bookings.GET("/master", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleAdmin), bookingH.GetByMaster)
			bookings.GET("/:id", bookingH.GetByID)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		log.Printf("Сервер запущен на http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("%v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Сервер остановлен")
}
