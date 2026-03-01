package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"renderowl-api/internal/domain/social"
)

// Job represents a scheduled job
type Job struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
	Delay     time.Duration   `json:"delay"`
	Attempts  int             `json:"attempts"`
	MaxRetries int            `json:"maxRetries"`
	Status    JobStatus       `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	RunAt     time.Time       `json:"runAt"`
	Error     string          `json:"error,omitempty"`
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusActive    JobStatus = "active"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusDelayed   JobStatus = "delayed"
)

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// Scheduler manages job scheduling and execution
type Scheduler struct {
	client   *redis.Client
	handlers map[string]JobHandler
	quit     chan bool
}

// NewScheduler creates a new scheduler instance
func NewScheduler(redisAddr string, redisPass string, redisDB int) *Scheduler {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})

	return &Scheduler{
		client:   client,
		handlers: make(map[string]JobHandler),
		quit:     make(chan bool),
	}
}

// RegisterHandler registers a handler for a job type
func (s *Scheduler) RegisterHandler(name string, handler JobHandler) {
	s.handlers[name] = handler
}

// AddJob adds a job to the queue
func (s *Scheduler) AddJob(ctx context.Context, job *Job) error {
	if job.ID == "" {
		job.ID = generateJobID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.RunAt.IsZero() {
		job.RunAt = time.Now().Add(job.Delay)
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	job.Status = JobStatusDelayed

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to delayed set (sorted by run time)
	score := float64(job.RunAt.Unix())
	return s.client.ZAdd(ctx, "scheduler:delayed", &redis.Z{
		Score:  score,
		Member: string(jobData),
	}).Err()
}

// AddRecurringJob adds a recurring job
func (s *Scheduler) AddRecurringJob(ctx context.Context, name string, data interface{}, rule *social.RecurringRule, handler JobHandler) error {
	// Store recurring job definition
	recurringData := map[string]interface{}{
		"name":     name,
		"data":     data,
		"rule":     rule,
		"last_run": time.Now(),
	}

	jsonData, err := json.Marshal(recurringData)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, "scheduler:recurring:"+name, jsonData, 0).Err()
}

// ProcessJobs starts processing jobs (blocking)
func (s *Scheduler) ProcessJobs(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			s.processPendingJobs(ctx)
			s.processRecurringJobs(ctx)
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.quit)
}

// GetJobStatus returns the status of a job
func (s *Scheduler) GetJobStatus(ctx context.Context, jobID string) (*Job, error) {
	// Check in delayed set
	jobs, err := s.client.ZRange(ctx, "scheduler:delayed", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	for _, jobData := range jobs {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}
		if job.ID == jobID {
			return &job, nil
		}
	}

	// Check in completed/failed sets
	for _, set := range []string{"scheduler:completed", "scheduler:failed"} {
		data, err := s.client.HGet(ctx, set, jobID).Result()
		if err == nil {
			var job Job
			if err := json.Unmarshal([]byte(data), &job); err != nil {
				continue
			}
			return &job, nil
		}
	}

	return nil, fmt.Errorf("job not found")
}

// GetQueueStats returns statistics about the job queue
func (s *Scheduler) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count delayed jobs
	delayed, err := s.client.ZCard(ctx, "scheduler:delayed").Result()
	if err != nil {
		return nil, err
	}
	stats["delayed"] = delayed

	// Count active jobs
	active, err := s.client.LLen(ctx, "scheduler:active").Result()
	if err != nil {
		return nil, err
	}
	stats["active"] = active

	// Count completed jobs
	completed, err := s.client.HLen(ctx, "scheduler:completed").Result()
	if err != nil {
		return nil, err
	}
	stats["completed"] = completed

	// Count failed jobs
	failed, err := s.client.HLen(ctx, "scheduler:failed").Result()
	if err != nil {
		return nil, err
	}
	stats["failed"] = failed

	return stats, nil
}

// Private methods

func (s *Scheduler) processPendingJobs(ctx context.Context) {
	now := float64(time.Now().Unix())

	// Get jobs that are ready to run
	jobs, err := s.client.ZRangeByScore(ctx, "scheduler:delayed", &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%f", now),
	}).Result()
	if err != nil {
		log.Printf("Failed to get pending jobs: %v", err)
		return
	}

	for _, jobData := range jobs {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		// Remove from delayed set
		s.client.ZRem(ctx, "scheduler:delayed", jobData)

		// Add to active queue
		s.client.RPush(ctx, "scheduler:active", jobData)

		// Process job
		go s.executeJob(ctx, &job)
	}
}

func (s *Scheduler) processRecurringJobs(ctx context.Context) {
	// Get all recurring jobs
	keys, err := s.client.Keys(ctx, "scheduler:recurring:*").Result()
	if err != nil {
		return
	}

	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var recurring struct {
			Name    string                `json:"name"`
			Data    json.RawMessage       `json:"data"`
			Rule    *social.RecurringRule `json:"rule"`
			LastRun time.Time             `json:"last_run"`
		}

		if err := json.Unmarshal([]byte(data), &recurring); err != nil {
			continue
		}

		// Check if job should run
		nextRun := calculateNextRun(recurring.LastRun, recurring.Rule)
		if time.Now().After(nextRun) {
			// Create job
			job := &Job{
				Name:  recurring.Name,
				Data:  recurring.Data,
				RunAt: nextRun,
			}

			if err := s.AddJob(ctx, job); err != nil {
				log.Printf("Failed to add recurring job: %v", err)
			}

			// Update last run
			recurring.LastRun = time.Now()
			updatedData, _ := json.Marshal(recurring)
			s.client.Set(ctx, key, updatedData, 0)
		}
	}
}

func (s *Scheduler) executeJob(ctx context.Context, job *Job) {
	handler, ok := s.handlers[job.Name]
	if !ok {
		job.Status = JobStatusFailed
		job.Error = "no handler registered"
		s.saveJobResult(ctx, job)
		return
	}

	job.Status = JobStatusActive
	job.Attempts++

	// Execute with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	err := handler(jobCtx, job)

	if err != nil {
		job.Error = err.Error()

		if job.Attempts < job.MaxRetries {
			// Retry with exponential backoff
			delay := time.Duration(job.Attempts) * time.Minute * 5
			job.RunAt = time.Now().Add(delay)
			job.Status = JobStatusDelayed
			s.AddJob(ctx, job)
			return
		}

		job.Status = JobStatusFailed
	} else {
		job.Status = JobStatusCompleted
	}

	s.saveJobResult(ctx, job)
}

func (s *Scheduler) saveJobResult(ctx context.Context, job *Job) {
	// Remove from active queue
	s.client.LRem(ctx, "scheduler:active", 0, job.ID)

	// Save to appropriate set
	jobData, _ := json.Marshal(job)

	if job.Status == JobStatusCompleted {
		s.client.HSet(ctx, "scheduler:completed", job.ID, jobData)
	} else if job.Status == JobStatusFailed {
		s.client.HSet(ctx, "scheduler:failed", job.ID, jobData)
	}
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func calculateNextRun(lastRun time.Time, rule *social.RecurringRule) time.Time {
	if rule == nil {
		return lastRun.Add(24 * time.Hour)
	}

	switch rule.Frequency {
	case "daily":
		return lastRun.AddDate(0, 0, rule.Interval)
	case "weekly":
		return lastRun.AddDate(0, 0, 7*rule.Interval)
	case "monthly":
		return lastRun.AddDate(0, rule.Interval, 0)
	default:
		return lastRun.Add(24 * time.Hour)
	}
}
