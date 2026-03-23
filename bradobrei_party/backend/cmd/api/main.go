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

	"bradobrei/backend/internal/handlers"
	"bradobrei/backend/internal/middleware"
	"bradobrei/backend/internal/models"
	"bradobrei/backend/internal/repository"
	"bradobrei/backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 1. Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env не найден")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET не задан")
	}

	// 2. Подключение к PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Ошибка подключения к БД:", err)
	}

	// 3. Авто-миграция всех 11 таблиц
	if err := db.AutoMigrate(
		&models.User{}, &models.EmployeeProfile{}, &models.Salon{},
		&models.Service{}, &models.Material{}, &models.ServiceMaterial{},
		&models.Inventory{}, &models.Booking{}, &models.BookingItem{},
		&models.Payment{}, &models.Review{},
	); err != nil {
		log.Fatal("❌ Ошибка миграции:", err)
	}
	log.Println("✅ БД синхронизирована (11 таблиц)")

	// 4. Репозитории
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	salonRepo := repository.NewSalonRepository(db)
	invRepo := repository.NewInventoryRepository(db)
	reportRepo := repository.NewReportRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	materialRepo := repository.NewMaterialRepository(db)
	employeeRepo := repository.NewEmployeeRepository(db)

	// 5. Сервисы
	authSvc := services.NewAuthService(userRepo)
	bookingSvc := services.NewBookingService(bookingRepo, invRepo, db)
	salonSvc := services.NewSalonService(salonRepo)
	reportSvc := services.NewReportService(reportRepo)
	serviceSvc := services.NewServiceService(serviceRepo, employeeRepo)
	materialSvc := services.NewMaterialService(materialRepo)
	employeeSvc := services.NewEmployeeService(employeeRepo, userRepo)

	// 6. Хэндлеры
	authH := handlers.NewAuthHandler(authSvc)
	bookingH := handlers.NewBookingHandler(bookingSvc)
	salonH := handlers.NewSalonHandler(salonSvc)
	reportH := handlers.NewReportHandler(reportSvc)
	reviewH := handlers.NewReviewHandler(db)
	serviceH := handlers.NewServiceHandler(serviceSvc)
	materialH := handlers.NewMaterialHandler(materialSvc)
	employeeH := handlers.NewEmployeeHandler(employeeSvc)

	// 7. Роутер
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
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

	v1 := r.Group("/api/v1")

	// Публичные маршруты
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// Защищённые маршруты (JWT обязателен)
	api := v1.Group("/")
	api.Use(middleware.AuthRequired(jwtSecret))
	{
		// Салоны
		salons := api.Group("/salons")
		{
			salons.GET("", salonH.GetAll)
			salons.GET("/:id", salonH.GetByID)
			salons.GET("/:id/masters", salonH.GetMasters)
			salons.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Create)
			salons.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), salonH.Update)
			salons.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), salonH.Delete)
		}

		// Услуги
		svcs := api.Group("/services")
		{
			svcs.GET("", serviceH.GetAll)
			svcs.GET("/:id", serviceH.GetByID)
			svcs.GET("/my", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster), serviceH.GetMy)
			svcs.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster), serviceH.Create)
			svcs.PUT("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster), serviceH.Update)
			svcs.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), serviceH.Delete)
			svcs.POST("/:id/assign-master",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster, models.RoleBasicMaster),
				serviceH.AssignToMaster)
			svcs.DELETE("/:id/assign-master/:profileId",
				middleware.RequireRoles(models.RoleAdmin, models.RoleAdvancedMaster),
				serviceH.RemoveFromMaster)
		}

		// Материалы
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

		// Сотрудники
		emps := api.Group("/employees")
		{
			emps.GET("", middleware.RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleNetworkManager), employeeH.GetAll)
			emps.GET("/:id", middleware.RequireRoles(models.RoleAdmin, models.RoleHR, models.RoleNetworkManager), employeeH.GetByID)
			emps.GET("/me", middleware.RequireRoles(
				models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleHR,
				models.RoleAccountant, models.RoleNetworkManager, models.RoleAdmin,
			), employeeH.GetMe)
			emps.POST("", middleware.RequireRoles(models.RoleAdmin, models.RoleHR), employeeH.Hire)
			emps.PATCH("/me/schedule", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster), employeeH.UpdateMySchedule)
			emps.POST("/:id/assign-salon", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), employeeH.AssignToSalon)
			emps.DELETE("/:id/assign-salon/:salonId", middleware.RequireRoles(models.RoleAdmin, models.RoleNetworkManager), employeeH.RemoveFromSalon)
		}

		// Бронирования
		bookings := api.Group("/bookings")
		{
			bookings.POST("", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.Create)
			bookings.GET("/my", middleware.RequireRoles(models.RoleClient, models.RoleAdmin), bookingH.GetMy)
			bookings.GET("/master", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleAdmin), bookingH.GetByMaster)
			bookings.GET("/:id", bookingH.GetByID)
			bookings.POST("/:id/confirm", middleware.RequireRoles(models.RoleBasicMaster, models.RoleAdvancedMaster, models.RoleAdmin), bookingH.Confirm)
			bookings.POST("/:id/cancel", bookingH.Cancel)
		}

		// Отзывы
		reviews := api.Group("/reviews")
		{
			reviews.POST("", reviewH.Create)
			reviews.GET("", middleware.RequireRoles(models.RoleAdmin), reviewH.GetAll)
		}

		// Отчёты (ТЗ 2.2)
		reports := api.Group("/reports")
		reports.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleAccountant, models.RoleNetworkManager, models.RoleHR))
		{
			reports.GET("/employees", reportH.Employees)
			reports.GET("/salon-activity", reportH.SalonActivity)
			reports.GET("/service-popularity", reportH.ServicePopularity)
			reports.GET("/master-activity", reportH.MasterActivity)
			reports.GET("/reviews", reportH.Reviews)
		}
	}

	// 8. Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		log.Printf("🚀 Сервер запущен на http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ %v", err)
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("✅ Сервер остановлен")
}
