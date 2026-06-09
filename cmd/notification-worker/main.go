package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/database"
	applogger "pay-your-dues/internal/logger"
	"pay-your-dues/internal/messaging"
	"pay-your-dues/internal/repository"
	"pay-your-dues/internal/services"
	"pay-your-dues/internal/workers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	logger := applogger.New(level, cfg.LogFormat)
	log.Logger = logger

	logger.Info().
		Str("log_level", level.String()).
		Str("log_format", cfg.LogFormat).
		Msg("Logger initialized")

	notificationCfg := config.LoadNotificationConfig()
	rabbitCfg := config.LoadRabbitMQConfig()

	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	notificationRepo := repository.NewNotificationRepositoryGORM(db.DB)
	contactRepo := repository.NewContactRepositoryGORM(db.DB)
	debtListRepo := repository.NewDebtListRepositoryGORM(db.DB, contactRepo)
	debtItemRepo := repository.NewDebtItemRepositoryGORM(db.DB)
	userRepo := repository.NewUserRepositoryGORM(db.DB)
	userSettingsRepo := repository.NewUserSettingsRepositoryGORM(db.DB)

	publisher, err := messaging.NewRabbitMQPublisher(rabbitCfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize RabbitMQ publisher")
	}
	defer publisher.Close()

	deliveryService := services.NewNotificationDeliveryService(
		notificationRepo,
		debtListRepo,
		debtItemRepo,
		contactRepo,
		userRepo,
		db.DB,
		notificationCfg,
		logger,
	)

	consumer, err := messaging.NewConsumer(rabbitCfg, func(ctx context.Context, job messaging.NotificationJob) error {
		deliveryService.ClearCaches()
		return deliveryService.ProcessJob(ctx, job)
	}, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize RabbitMQ consumer")
	}

	scheduler := workers.NewNotificationScheduler(
		db.DB,
		notificationRepo,
		publisher,
		rabbitCfg,
		logger,
	)

	telegramSubscriptionService := services.NewTelegramSubscriptionService(
		userSettingsRepo,
		userRepo,
		notificationCfg.TelegramBotToken,
		logger,
	)

	scheduler.Start()
	if err := consumer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start consumer")
	}

	if notificationCfg.TelegramPollingMode && notificationCfg.TelegramBotToken != "" {
		if err := telegramSubscriptionService.StartLongPolling(); err != nil {
			logger.Error().Err(err).Msg("Failed to start Telegram long polling")
		} else {
			logger.Info().Msg("Telegram long polling started")
			defer telegramSubscriptionService.StopLongPolling()
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"notification-worker"}`))
	})

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.ServerHost, rabbitCfg.WorkerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info().Str("address", server.Addr).Msg("Starting notification worker health server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start health server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down notification worker...")
	scheduler.Stop()
	consumer.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Health server forced to shutdown")
	}

	logger.Info().Msg("Notification worker stopped")
}
