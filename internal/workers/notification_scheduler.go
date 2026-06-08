package workers

import (
	"sync"
	"time"

	"pay-your-dues/internal/config"
	"pay-your-dues/internal/domain/interfaces"
	"pay-your-dues/internal/messaging"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// NotificationScheduler polls the database for due notifications and publishes them to RabbitMQ.
type NotificationScheduler struct {
	db               *gorm.DB
	notificationRepo interfaces.NotificationRepository
	publisher        *messaging.RabbitMQPublisher
	rabbitCfg        *config.RabbitMQConfig
	logger           zerolog.Logger

	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
}

func NewNotificationScheduler(
	db *gorm.DB,
	notificationRepo interfaces.NotificationRepository,
	publisher *messaging.RabbitMQPublisher,
	rabbitCfg *config.RabbitMQConfig,
	logger zerolog.Logger,
) *NotificationScheduler {
	return &NotificationScheduler{
		db:               db,
		notificationRepo: notificationRepo,
		publisher:        publisher,
		rabbitCfg:        rabbitCfg,
		logger:           logger.With().Str("component", "notification_scheduler").Logger(),
		stopChan:         make(chan struct{}),
	}
}

func (s *NotificationScheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	s.wg.Add(1)
	go s.run()
	s.logger.Info().Msg("Notification scheduler started")
}

func (s *NotificationScheduler) Stop() {
	if !s.running {
		return
	}
	close(s.stopChan)
	s.wg.Wait()
	s.running = false
	s.logger.Info().Msg("Notification scheduler stopped")
}

func (s *NotificationScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.rabbitCfg.SchedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *NotificationScheduler) tick() {
	if !s.tryAcquireLock() {
		return
	}
	defer s.releaseLock()

	now := time.Now()
	notifications, err := s.notificationRepo.ClaimDueNotifications(now, s.rabbitCfg.SchedulerBatchSize)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to claim due notifications")
		return
	}
	if len(notifications) == 0 {
		return
	}

	s.logger.Info().Int("count", len(notifications)).Msg("Publishing due notifications")

	for _, notif := range notifications {
		jobType := messaging.JobTypeScheduled
		if notif.ScheduleType == "manual" {
			jobType = messaging.JobTypeManual
		} else if notif.ScheduleType == "event" {
			jobType = messaging.JobTypeEvent
		}

		job := messaging.NotificationJob{
			NotificationID: notif.ID,
			JobType:        jobType,
			Attempt:        1,
		}

		if err := s.publisher.Publish(job); err != nil {
			s.logger.Error().
				Err(err).
				Str("notification_id", notif.ID.String()).
				Msg("Failed to publish notification job, reverting to pending")
			if revertErr := s.notificationRepo.RevertToPending(notif.ID); revertErr != nil {
				s.logger.Error().Err(revertErr).Str("notification_id", notif.ID.String()).Msg("Failed to revert notification to pending")
			}
		}
	}
}

func (s *NotificationScheduler) tryAcquireLock() bool {
	var acquired bool
	err := s.db.Raw("SELECT pg_try_advisory_lock(?)", s.rabbitCfg.SchedulerLockID).Scan(&acquired).Error
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to acquire scheduler advisory lock")
		return false
	}
	return acquired
}

func (s *NotificationScheduler) releaseLock() {
	if err := s.db.Exec("SELECT pg_advisory_unlock(?)", s.rabbitCfg.SchedulerLockID).Error; err != nil {
		s.logger.Error().Err(err).Msg("Failed to release scheduler advisory lock")
	}
}
