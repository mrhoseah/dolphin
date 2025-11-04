package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Task represents a scheduled task
type Task struct {
	Name        string
	Schedule    string // Cron expression or interval
	Handler     func() error
	Description string
	Enabled     bool
	LastRun     *time.Time
	NextRun     *time.Time
	ErrorCount  int
	LastError   error
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	cron     *cron.Cron
	tasks    map[string]*Task
	mu       sync.RWMutex
	logger   *zap.Logger
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewScheduler creates a new scheduler
func NewScheduler(logger *zap.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Create cron with seconds support
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(time.UTC),
	)

	return &Scheduler{
		cron:   c,
		tasks:  make(map[string]*Task),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Register registers a new task
func (s *Scheduler) Register(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("task '%s' already exists", task.Name)
	}

	s.tasks[task.Name] = task

	// Schedule if enabled
	if task.Enabled && s.running {
		return s.scheduleTask(task)
	}

	return nil
}

// scheduleTask schedules a task in the cron
func (s *Scheduler) scheduleTask(task *Task) error {
	entryID, err := s.cron.AddFunc(task.Schedule, func() {
		s.executeTask(task)
	})

	if err != nil {
		return fmt.Errorf("failed to schedule task '%s': %w", task.Name, err)
	}

	s.logger.Info("Task scheduled",
		zap.String("task", task.Name),
		zap.String("schedule", task.Schedule),
		zap.Int("entry_id", int(entryID)),
	)

	return nil
}

// executeTask executes a task
func (s *Scheduler) executeTask(task *Task) {
	now := time.Now()
	task.LastRun = &now

	s.logger.Info("Executing task", zap.String("task", task.Name))

	// Execute in goroutine with timeout
	done := make(chan error, 1)
	go func() {
		done <- task.Handler()
	}()

	// Wait for completion or timeout (default 5 minutes)
	select {
	case err := <-done:
		if err != nil {
			task.ErrorCount++
			task.LastError = err
			s.logger.Error("Task execution failed",
				zap.String("task", task.Name),
				zap.Error(err),
				zap.Int("error_count", task.ErrorCount),
			)
		} else {
			task.ErrorCount = 0
			task.LastError = nil
			s.logger.Info("Task executed successfully", zap.String("task", task.Name))
		}
	case <-time.After(5 * time.Minute):
		task.ErrorCount++
		task.LastError = fmt.Errorf("task execution timeout")
		s.logger.Error("Task execution timeout",
			zap.String("task", task.Name),
			zap.Int("error_count", task.ErrorCount),
		)
	}

	// Calculate next run time
	s.calculateNextRun(task)
}

// calculateNextRun calculates the next run time for a task
func (s *Scheduler) calculateNextRun(task *Task) {
	// Parse cron schedule
	schedule, err := cron.ParseStandard(task.Schedule)
	if err != nil {
		// Try parsing with seconds
		schedule, err = cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		).Parse(task.Schedule)
		if err != nil {
			return
		}
	}

	now := time.Now()
	if task.LastRun != nil {
		now = *task.LastRun
	}

	nextRun := schedule.Next(now)
	task.NextRun = &nextRun
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true

	// Schedule all enabled tasks
	for _, task := range s.tasks {
		if task.Enabled {
			if err := s.scheduleTask(task); err != nil {
				s.logger.Error("Failed to schedule task",
					zap.String("task", task.Name),
					zap.Error(err),
				)
			}
		}
	}

	s.cron.Start()
	s.logger.Info("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.cron.Stop()
	s.cancel()
	s.logger.Info("Scheduler stopped")
}

// RunNow executes a task immediately
func (s *Scheduler) RunNow(taskName string) error {
	s.mu.RLock()
	task, exists := s.tasks[taskName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	go s.executeTask(task)
	return nil
}

// Enable enables a task
func (s *Scheduler) Enable(taskName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskName]
	if !exists {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	task.Enabled = true

	if s.running {
		return s.scheduleTask(task)
	}

	return nil
}

// Disable disables a task
func (s *Scheduler) Disable(taskName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskName]
	if !exists {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	task.Enabled = false
	// Note: We can't easily remove from cron, but disabling prevents re-scheduling

	return nil
}

// GetTasks returns all tasks
func (s *Scheduler) GetTasks() map[string]*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*Task)
	for name, task := range s.tasks {
		taskCopy := *task
		result[name] = &taskCopy
	}
	return result
}

// GetTask returns a specific task
func (s *Scheduler) GetTask(name string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found", name)
	}

	taskCopy := *task
	return &taskCopy, nil
}

// IsRunning checks if scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

