package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
)

// CronService handles scheduling and managing cron jobs for notifications
type CronService struct {
	scheduler *gocron.Scheduler
	jobs      map[string]*gocron.Job
	mu        sync.RWMutex
}

// NewCronService creates a new cron service instance
func NewCronService() *CronService {
	// Create scheduler with location
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()

	return &CronService{
		scheduler: scheduler,
		jobs:      make(map[string]*gocron.Job),
	}
}

// ScheduleAt schedules a job to run at a specific time
func (cs *CronService) ScheduleAt(scheduledTime time.Time, task func()) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Schedule job
	job, err := cs.scheduler.Every(1).Day().At(scheduledTime).Do(task)
	if err != nil {
		return "", fmt.Errorf("failed to schedule job: %w", err)
	}

	// Store job reference
	cs.jobs[jobID] = job

	return jobID, nil
}

// ScheduleOneTime schedules a one-time job to run at a specific datetime
func (cs *CronService) ScheduleOneTime(scheduledTime time.Time, task func()) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Calculate duration until scheduled time
	duration := time.Until(scheduledTime)
	if duration < 0 {
		return "", fmt.Errorf("scheduled time is in the past")
	}

	// Schedule one-time job
	job, err := cs.scheduler.Every(duration).LimitRunsTo(1).Do(task)
	if err != nil {
		return "", fmt.Errorf("failed to schedule one-time job: %w", err)
	}

	// Store job reference
	cs.jobs[jobID] = job

	return jobID, nil
}

// ScheduleRecurring schedules a recurring job with a specific interval
func (cs *CronService) ScheduleRecurring(interval time.Duration, task func()) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Schedule recurring job
	job, err := cs.scheduler.Every(interval).Do(task)
	if err != nil {
		return "", fmt.Errorf("failed to schedule recurring job: %w", err)
	}

	// Store job reference
	cs.jobs[jobID] = job

	return jobID, nil
}

// ScheduleCron schedules a job using cron expression
func (cs *CronService) ScheduleCron(cronExpr string, task func()) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Schedule with cron expression
	job, err := cs.scheduler.Cron(cronExpr).Do(task)
	if err != nil {
		return "", fmt.Errorf("failed to schedule cron job: %w", err)
	}

	// Store job reference
	cs.jobs[jobID] = job

	return jobID, nil
}

// ScheduleDaily schedules a job to run daily at a specific time
func (cs *CronService) ScheduleDaily(hour, minute int, task func()) (string, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Generate unique job ID
	jobID := uuid.New().String()

	// Schedule daily job
	timeStr := fmt.Sprintf("%02d:%02d", hour, minute)
	job, err := cs.scheduler.Every(1).Day().At(timeStr).Do(task)
	if err != nil {
		return "", fmt.Errorf("failed to schedule daily job: %w", err)
	}

	// Store job reference
	cs.jobs[jobID] = job

	return jobID, nil
}

// Remove removes a scheduled job by ID
func (cs *CronService) Remove(jobID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	job, exists := cs.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Remove job from scheduler
	cs.scheduler.RemoveByReference(job)

	// Remove from map
	delete(cs.jobs, jobID)

	return nil
}

// RemoveAll removes all scheduled jobs
func (cs *CronService) RemoveAll() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Clear scheduler
	cs.scheduler.Clear()

	// Clear jobs map
	cs.jobs = make(map[string]*gocron.Job)
}

// GetJob returns a job by ID
func (cs *CronService) GetJob(jobID string) (*gocron.Job, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	job, exists := cs.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// GetJobCount returns the number of scheduled jobs
func (cs *CronService) GetJobCount() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return len(cs.jobs)
}

// IsJobScheduled checks if a job with the given ID is scheduled
func (cs *CronService) IsJobScheduled(jobID string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	_, exists := cs.jobs[jobID]
	return exists
}

// GetNextRun returns the next run time for a job
func (cs *CronService) GetNextRun(jobID string) (time.Time, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	job, exists := cs.jobs[jobID]
	if !exists {
		return time.Time{}, fmt.Errorf("job not found: %s", jobID)
	}

	nextRun := job.NextRun()
	return nextRun, nil
}

// GetLastRun returns the last run time for a job
func (cs *CronService) GetLastRun(jobID string) (time.Time, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	job, exists := cs.jobs[jobID]
	if !exists {
		return time.Time{}, fmt.Errorf("job not found: %s", jobID)
	}

	lastRun := job.LastRun()
	return lastRun, nil
}

// Stop stops the scheduler
func (cs *CronService) Stop() {
	cs.scheduler.Stop()
}

// Start starts the scheduler if it's stopped
func (cs *CronService) Start() {
	cs.scheduler.StartAsync()
}

// IsRunning returns whether the scheduler is running
func (cs *CronService) IsRunning() bool {
	return cs.scheduler.IsRunning()
}

// GetAllJobIDs returns all job IDs
func (cs *CronService) GetAllJobIDs() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	ids := make([]string, 0, len(cs.jobs))
	for id := range cs.jobs {
		ids = append(ids, id)
	}

	return ids
}

