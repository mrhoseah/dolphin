package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Job represents a job interface
type Job interface {
	Handle() error
	GetName() string
	GetPayload() map[string]interface{}
	GetAttempts() int
	GetMaxAttempts() int
	GetTimeout() time.Duration
	GetDelay() time.Duration
	GetQueue() string
	GetID() string
	SetID(id string)
	SetAttempts(attempts int)
	Failed(error)
	Processed()
}

// BaseJob provides common job functionality
type BaseJob struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Payload     map[string]interface{} `json:"payload"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Timeout     time.Duration          `json:"timeout"`
	Delay       time.Duration          `json:"delay"`
	Queue       string                 `json:"queue"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// GetName returns the job name
func (bj *BaseJob) GetName() string {
	return bj.Name
}

// GetPayload returns the job payload
func (bj *BaseJob) GetPayload() map[string]interface{} {
	return bj.Payload
}

// GetAttempts returns the current attempts
func (bj *BaseJob) GetAttempts() int {
	return bj.Attempts
}

// GetMaxAttempts returns the maximum attempts
func (bj *BaseJob) GetMaxAttempts() int {
	return bj.MaxAttempts
}

// GetTimeout returns the job timeout
func (bj *BaseJob) GetTimeout() time.Duration {
	return bj.Timeout
}

// GetDelay returns the job delay
func (bj *BaseJob) GetDelay() time.Duration {
	return bj.Delay
}

// GetQueue returns the queue name
func (bj *BaseJob) GetQueue() string {
	return bj.Queue
}

// GetID returns the job ID
func (bj *BaseJob) GetID() string {
	return bj.ID
}

// SetID sets the job ID
func (bj *BaseJob) SetID(id string) {
	bj.ID = id
}

// SetAttempts sets the attempts count
func (bj *BaseJob) SetAttempts(attempts int) {
	bj.Attempts = attempts
}

// Failed marks the job as failed
func (bj *BaseJob) Failed(err error) {
	now := time.Now()
	bj.FailedAt = &now
	bj.Error = err.Error()
}

// Processed marks the job as processed
func (bj *BaseJob) Processed() {
	now := time.Now()
	bj.ProcessedAt = &now
}

// Handle is the default implementation (should be overridden)
func (bj *BaseJob) Handle() error {
	return fmt.Errorf("job %s does not implement Handle method", bj.Name)
}

// SendEmailJob represents an email sending job
type SendEmailJob struct {
	BaseJob
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// NewSendEmailJob creates a new send email job
func NewSendEmailJob(to []string, subject, text, html string) *SendEmailJob {
	return &SendEmailJob{
		BaseJob: BaseJob{
			Name:        "SendEmail",
			MaxAttempts: 3,
			Timeout:     30 * time.Second,
			Queue:       "emails",
		},
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	}
}

// Handle handles the email sending job
func (sej *SendEmailJob) Handle() error {
	// In a real implementation, this would send the email
	fmt.Printf("Sending email to %v: %s\n", sej.To, sej.Subject)
	time.Sleep(100 * time.Millisecond) // Simulate work
	return nil
}

// ProcessPaymentJob represents a payment processing job
type ProcessPaymentJob struct {
	BaseJob
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	PaymentID string  `json:"payment_id"`
}

// NewProcessPaymentJob creates a new process payment job
func NewProcessPaymentJob(orderID string, amount float64, currency, paymentID string) *ProcessPaymentJob {
	return &ProcessPaymentJob{
		BaseJob: BaseJob{
			Name:        "ProcessPayment",
			MaxAttempts: 5,
			Timeout:     60 * time.Second,
			Queue:       "payments",
		},
		OrderID:   orderID,
		Amount:    amount,
		Currency:  currency,
		PaymentID: paymentID,
	}
}

// Handle handles the payment processing job
func (ppj *ProcessPaymentJob) Handle() error {
	// In a real implementation, this would process the payment
	fmt.Printf("Processing payment for order %s: %.2f %s\n", ppj.OrderID, ppj.Amount, ppj.Currency)
	time.Sleep(200 * time.Millisecond) // Simulate work
	return nil
}

// GenerateReportJob represents a report generation job
type GenerateReportJob struct {
	BaseJob
	ReportType string                 `json:"report_type"`
	UserID     string                 `json:"user_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

// NewGenerateReportJob creates a new generate report job
func NewGenerateReportJob(reportType, userID string, parameters map[string]interface{}) *GenerateReportJob {
	return &GenerateReportJob{
		BaseJob: BaseJob{
			Name:        "GenerateReport",
			MaxAttempts: 2,
			Timeout:     300 * time.Second,
			Queue:       "reports",
		},
		ReportType: reportType,
		UserID:     userID,
		Parameters: parameters,
	}
}

// Handle handles the report generation job
func (grj *GenerateReportJob) Handle() error {
	// In a real implementation, this would generate the report
	fmt.Printf("Generating %s report for user %s\n", grj.ReportType, grj.UserID)
	time.Sleep(500 * time.Millisecond) // Simulate work
	return nil
}

// Queue represents a queue interface
type Queue interface {
	Push(job Job) error
	Pop(queueName string) (Job, error)
	Size(queueName string) (int, error)
	Clear(queueName string) error
	GetFailedJobs() ([]Job, error)
	RetryFailedJob(jobID string) error
	DeleteFailedJob(jobID string) error
}

// MemoryQueue implements Queue using memory
type MemoryQueue struct {
	queues      map[string][]Job
	failedJobs  []Job
	mu          sync.RWMutex
	jobRegistry map[string]reflect.Type
}

// NewMemoryQueue creates a new memory queue
func NewMemoryQueue() *MemoryQueue {
	mq := &MemoryQueue{
		queues:      make(map[string][]Job),
		failedJobs:  make([]Job, 0),
		jobRegistry: make(map[string]reflect.Type),
	}

	// Register default job types
	mq.RegisterJob("SendEmail", reflect.TypeOf((*SendEmailJob)(nil)).Elem())
	mq.RegisterJob("ProcessPayment", reflect.TypeOf((*ProcessPaymentJob)(nil)).Elem())
	mq.RegisterJob("GenerateReport", reflect.TypeOf((*GenerateReportJob)(nil)).Elem())

	return mq
}

// RegisterJob registers a job type
func (mq *MemoryQueue) RegisterJob(name string, jobType reflect.Type) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.jobRegistry[name] = jobType
}

// Push pushes a job to the queue
func (mq *MemoryQueue) Push(job Job) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if job.GetID() == "" {
		job.SetID(fmt.Sprintf("job_%d", time.Now().UnixNano()))
	}

	job.SetAttempts(0)

	queueName := job.GetQueue()
	if queueName == "" {
		queueName = "default"
	}

	if mq.queues[queueName] == nil {
		mq.queues[queueName] = make([]Job, 0)
	}

	mq.queues[queueName] = append(mq.queues[queueName], job)
	return nil
}

// Pop pops a job from the queue
func (mq *MemoryQueue) Pop(queueName string) (Job, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if queueName == "" {
		queueName = "default"
	}

	queue := mq.queues[queueName]
	if len(queue) == 0 {
		return nil, fmt.Errorf("no jobs in queue %s", queueName)
	}

	job := queue[0]
	mq.queues[queueName] = queue[1:]

	return job, nil
}

// Size returns the size of a queue
func (mq *MemoryQueue) Size(queueName string) (int, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	if queueName == "" {
		queueName = "default"
	}

	return len(mq.queues[queueName]), nil
}

// Clear clears a queue
func (mq *MemoryQueue) Clear(queueName string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if queueName == "" {
		queueName = "default"
	}

	mq.queues[queueName] = make([]Job, 0)
	return nil
}

// GetFailedJobs returns all failed jobs
func (mq *MemoryQueue) GetFailedJobs() ([]Job, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	failedJobs := make([]Job, len(mq.failedJobs))
	copy(failedJobs, mq.failedJobs)

	return failedJobs, nil
}

// RetryFailedJob retries a failed job
func (mq *MemoryQueue) RetryFailedJob(jobID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for i, job := range mq.failedJobs {
		if job.GetID() == jobID {
			// Remove from failed jobs
			mq.failedJobs = append(mq.failedJobs[:i], mq.failedJobs[i+1:]...)

			// Reset attempts and add back to queue
			job.SetAttempts(0)
			queueName := job.GetQueue()
			if queueName == "" {
				queueName = "default"
			}

			if mq.queues[queueName] == nil {
				mq.queues[queueName] = make([]Job, 0)
			}

			mq.queues[queueName] = append(mq.queues[queueName], job)
			return nil
		}
	}

	return fmt.Errorf("failed job with ID %s not found", jobID)
}

// DeleteFailedJob deletes a failed job
func (mq *MemoryQueue) DeleteFailedJob(jobID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for i, job := range mq.failedJobs {
		if job.GetID() == jobID {
			mq.failedJobs = append(mq.failedJobs[:i], mq.failedJobs[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("failed job with ID %s not found", jobID)
}

// Worker represents a queue worker
type Worker struct {
	queue       Queue
	queues      []string
	concurrency int
	stopChan    chan bool
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorker creates a new worker
func NewWorker(queue Queue, queues []string, concurrency int) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		queue:       queue,
		queues:      queues,
		concurrency: concurrency,
		stopChan:    make(chan bool),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker
func (w *Worker) Start() {
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.work()
	}
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.cancel()
	close(w.stopChan)
	w.wg.Wait()
}

// work processes jobs from the queue
func (w *Worker) work() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			job, err := w.getNextJob()
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			w.processJob(job)
		}
	}
}

// getNextJob gets the next job from any queue
func (w *Worker) getNextJob() (Job, error) {
	for _, queueName := range w.queues {
		job, err := w.queue.Pop(queueName)
		if err == nil {
			return job, nil
		}
	}
	return nil, fmt.Errorf("no jobs available")
}

// processJob processes a single job
func (w *Worker) processJob(job Job) {
	ctx, cancel := context.WithTimeout(w.ctx, job.GetTimeout())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- job.Handle()
	}()

	select {
	case err := <-done:
		if err != nil {
			w.handleJobFailure(job, err)
		} else {
			job.Processed()
		}
	case <-ctx.Done():
		w.handleJobFailure(job, fmt.Errorf("job timeout"))
	}
}

// handleJobFailure handles a failed job
func (w *Worker) handleJobFailure(job Job, err error) {
	job.Failed(err)
	job.SetAttempts(job.GetAttempts() + 1)

	if job.GetAttempts() < job.GetMaxAttempts() {
		// Retry the job
		w.queue.Push(job)
	} else {
		// Add to failed jobs
		if memoryQueue, ok := w.queue.(*MemoryQueue); ok {
			memoryQueue.mu.Lock()
			memoryQueue.failedJobs = append(memoryQueue.failedJobs, job)
			memoryQueue.mu.Unlock()
		}
	}
}

// QueueManager manages queues and workers
type QueueManager struct {
	queues  map[string]Queue
	workers map[string]*Worker
}

// NewQueueManager creates a new queue manager
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues:  make(map[string]Queue),
		workers: make(map[string]*Worker),
	}
}

// RegisterQueue registers a queue
func (qm *QueueManager) RegisterQueue(name string, queue Queue) {
	qm.queues[name] = queue
}

// GetQueue returns a queue by name
func (qm *QueueManager) GetQueue(name string) Queue {
	if queue, exists := qm.queues[name]; exists {
		return queue
	}
	return nil
}

// StartWorker starts a worker for a queue
func (qm *QueueManager) StartWorker(name string, queues []string, concurrency int) {
	queue := qm.GetQueue(name)
	if queue == nil {
		return
	}

	worker := NewWorker(queue, queues, concurrency)
	qm.workers[name] = worker
	worker.Start()
}

// StopWorker stops a worker
func (qm *QueueManager) StopWorker(name string) {
	if worker, exists := qm.workers[name]; exists {
		worker.Stop()
		delete(qm.workers, name)
	}
}

// StopAllWorkers stops all workers
func (qm *QueueManager) StopAllWorkers() {
	for name := range qm.workers {
		qm.StopWorker(name)
	}
}

// Dispatch dispatches a job to a queue
func (qm *QueueManager) Dispatch(job Job, queueName string) error {
	queue := qm.GetQueue(queueName)
	if queue == nil {
		return fmt.Errorf("queue %s not found", queueName)
	}

	return queue.Push(job)
}

// DispatchToDefault dispatches a job to the default queue
func (qm *QueueManager) DispatchToDefault(job Job) error {
	return qm.Dispatch(job, "default")
}

// JobDispatcher provides a fluent interface for dispatching jobs
type JobDispatcher struct {
	queueManager *QueueManager
	job          Job
	queueName    string
	delay        time.Duration
}

// NewJobDispatcher creates a new job dispatcher
func NewJobDispatcher(queueManager *QueueManager, job Job) *JobDispatcher {
	return &JobDispatcher{
		queueManager: queueManager,
		job:          job,
		queueName:    "default",
	}
}

// OnQueue sets the queue name
func (jd *JobDispatcher) OnQueue(queueName string) *JobDispatcher {
	jd.queueName = queueName
	return jd
}

// Delay sets the delay
func (jd *JobDispatcher) Delay(delay time.Duration) *JobDispatcher {
	jd.delay = delay
	return jd
}

// Dispatch dispatches the job
func (jd *JobDispatcher) Dispatch() error {
	if jd.delay > 0 {
		// In a real implementation, this would schedule the job for later
		// For now, we'll just dispatch immediately
		time.Sleep(jd.delay)
	}

	return jd.queueManager.Dispatch(jd.job, jd.queueName)
}

// DispatchNow dispatches the job immediately
func (jd *JobDispatcher) DispatchNow() error {
	return jd.queueManager.Dispatch(jd.job, jd.queueName)
}

// SerializeJob serializes a job to JSON
func SerializeJob(job Job) ([]byte, error) {
	return json.Marshal(job)
}

// DeserializeJob deserializes a job from JSON
func DeserializeJob(data []byte, jobType reflect.Type) (Job, error) {
	job := reflect.New(jobType).Interface().(Job)
	err := json.Unmarshal(data, job)
	return job, err
}
