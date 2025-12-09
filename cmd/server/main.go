package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/database"
	"pay-your-dues/internal/handlers"
	"pay-your-dues/internal/middleware"
	"pay-your-dues/internal/repository"
	"pay-your-dues/internal/services"
	"pay-your-dues/internal/workers"
)

func main() {
	// Initialize logger with structured format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Caller().
		Logger()

	// Set global logger
	log.Logger = logger

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Warn().Str("level", cfg.LogLevel).Msg("Invalid log level, defaulting to info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	logger.Info().Str("log_level", level.String()).Msg("Logger initialized")

	// Initialize database with GORM
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	logger.Info().Msg("Database connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepositoryGORM(db.DB)
	contactRepo := repository.NewContactRepositoryGORM(db.DB)
	debtListRepo := repository.NewDebtListRepositoryGORM(db.DB, contactRepo)
	debtItemRepo := repository.NewDebtItemRepositoryGORM(db.DB)
	notificationRepo := repository.NewNotificationRepositoryGORM(db.DB)
	notificationTemplateRepo := repository.NewNotificationTemplateRepositoryGORM(db.DB)
	userSettingsRepo := repository.NewUserSettingsRepositoryGORM(db.DB)

	// Initialize services with dependency injection
	paymentScheduleService := services.NewPaymentScheduleService()
	contactService := services.NewContactService(contactRepo, userRepo)
	
	// Initialize S3 service for file storage
	s3Service, err := services.NewS3Service(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize S3 service")
	}
	
	// Initialize notification service (Phase 4)
	notificationConfig := config.LoadNotificationConfig()
	notificationService := services.NewNotificationService(
		notificationRepo,
		notificationTemplateRepo,
		debtListRepo,
		debtItemRepo,
		contactRepo,
		userRepo,
		userSettingsRepo,
		db.DB,
		notificationConfig,
		logger,
	)
	
	// Initialize notification worker (Phase 5)
	notificationWorker := workers.NewNotificationWorker(
		notificationRepo,
		debtListRepo,
		contactRepo,
		userRepo,
		db.DB,
		notificationConfig,
		logger,
	)
	
	debtService := services.NewDebtService(debtListRepo, debtItemRepo, contactRepo, paymentScheduleService, s3Service, notificationService)

	// Initialize user settings service
	userSettingsService := services.NewUserSettingsService(userSettingsRepo, userRepo, notificationRepo, logger)

	// Initialize Telegram subscription service (for bot-based subscription flow)
	telegramSubscriptionService := services.NewTelegramSubscriptionService(
		userSettingsRepo,
		userRepo,
		notificationConfig.TelegramBotToken,
		logger,
	)

	// Initialize auth service with all dependencies
	authService, err := services.NewAuthService(userRepo, contactService, userSettingsService, cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize auth service")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	contactHandler := handlers.NewContactHandler(contactService, logger)
	debtHandler := handlers.NewDebtHandler(debtService, s3Service, logger)
	notificationHandler := handlers.NewNotificationHandler(notificationService, logger)
	userSettingsHandler := handlers.NewUserSettingsHandler(userSettingsService, logger)
	telegramHandler := handlers.NewTelegramHandler(telegramSubscriptionService, notificationConfig.TelegramWebhookSecret, logger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	// Initialize Gin router with production settings
	if level == zerolog.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add global middleware
	router.Use(loggingMiddleware.Recovery())
	router.Use(loggingMiddleware.CORS())
	router.Use(loggingMiddleware.LogRequests())

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "pay-your-dues-api",
			"version":   "1.0.0",
			"timestamp": time.Now(),
		})
	})

	// Telegram webhook endpoint (public, called by Telegram servers)
	router.POST("/api/telegram/webhook", telegramHandler.HandleWebhook)

	// API routes
	apiV1 := router.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes (auth required)
		protected := apiV1.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			// Health check for authenticated users
		protected.GET("/health", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				c.JSON(http.StatusOK, gin.H{
					"status":    "ok",
					"user_id":   userID,
					"timestamp": time.Now(),
				})
		})

					// Contact routes
			contacts := protected.Group("/contacts")
			{
				contacts.POST("", contactHandler.CreateContact)
				contacts.GET("", contactHandler.GetUserContacts)
				contacts.GET("/:id", contactHandler.GetContact)
				contacts.PUT("/:id", contactHandler.UpdateContact)
				contacts.DELETE("/:id", contactHandler.DeleteContact)
			}

			// Debt management routes
			debts := protected.Group("/debts")
			{
				// Debt list operations
				debts.POST("", debtHandler.CreateDebtList)
				debts.GET("", debtHandler.GetUserDebtLists)
				debts.GET("/:id", debtHandler.GetDebtList)
				debts.PUT("/:id", debtHandler.UpdateDebtList)
				debts.DELETE("/:id", debtHandler.DeleteDebtList)

				// Debt item (payment) operations
				debts.POST("/payments", debtHandler.CreateDebtItem)
				debts.GET("/:id/payments", debtHandler.GetDebtListItems)
				debts.DELETE("/payments/:id", debtHandler.DeleteDebtItem)

				// Payment verification operations
				debts.GET("/verifications/pending", debtHandler.GetPendingVerifications)
				debts.POST("/payments/:id/verify", debtHandler.VerifyDebtItem)
				debts.POST("/payments/:id/reject", debtHandler.RejectDebtItem)
				debts.POST("/payments/:id/receipt", debtHandler.UploadReceipt)

				// Receipt photo serving
				debts.GET("/:id/receipts/:filename", debtHandler.GetReceiptPhoto)

				// Analytics and reporting
				debts.GET("/overdue", debtHandler.GetOverdueItems)
				debts.GET("/due-soon", debtHandler.GetDueSoonItems)
				debts.GET("/:id/schedule", debtHandler.GetPaymentSchedule)
				debts.GET("/:id/summary", debtHandler.GetTotalPaymentsForDebtList)
			}

			// Notification routes (Phase 4)
			notifications := protected.Group("/notifications")
			{
				notifications.POST("", notificationHandler.CreateNotification)
				notifications.GET("/:id", notificationHandler.GetNotification)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)
				notifications.POST("/:id/enable", notificationHandler.EnableNotification)
				notifications.POST("/:id/disable", notificationHandler.DisableNotification)
				
				// Debt list notification operations
				notifications.POST("/schedule", notificationHandler.ScheduleNotifications)
				notifications.GET("/debt-lists/:debt_list_id", notificationHandler.GetNotificationsByDebtList)
				
				// Manual notification sending
				notifications.POST("/send", notificationHandler.SendManualNotification)
			}

			// User settings routes
			settings := protected.Group("/settings")
			{
				settings.GET("", userSettingsHandler.GetUserSettings)
				settings.PUT("", userSettingsHandler.UpdateUserSettings)

				// Telegram subscription routes
				settings.GET("/telegram/status", telegramHandler.GetTelegramStatus)
				settings.POST("/telegram/link", telegramHandler.GenerateLinkCode)
				settings.DELETE("/telegram/link", telegramHandler.UnlinkTelegram)
			}

			// Additional analytics routes
			protected.GET("/upcoming-payments", debtHandler.GetUpcomingPayments)
		}
	}

	// Start server with graceful shutdown
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start notification worker (Phase 5)
	notificationWorker.Start()
	logger.Info().Msg("Notification worker started")
	defer func() {
		logger.Info().Msg("Stopping notification worker...")
		notificationWorker.Stop()
		logger.Info().Msg("Notification worker stopped")
	}()

	// Start Telegram long polling if configured (for development without a domain)
	if notificationConfig.TelegramPollingMode && notificationConfig.TelegramBotToken != "" {
		if err := telegramSubscriptionService.StartLongPolling(); err != nil {
			logger.Error().Err(err).Msg("Failed to start Telegram long polling")
		} else {
			logger.Info().Msg("Telegram long polling started (development mode)")
			defer func() {
				logger.Info().Msg("Stopping Telegram long polling...")
				telegramSubscriptionService.StopLongPolling()
			}()
		}
	} else if notificationConfig.TelegramBotToken != "" {
		logger.Info().Msg("Telegram webhook mode enabled (set TELEGRAM_POLLING_MODE=true for development)")
	}

	// Start server in a goroutine
	go func() {
		logger.Info().Str("address", addr).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		logger.Info().Msg("Server shutdown completed")
	}
} 