package fakes

import (
	"fmt"
	"sync"
	"time"
)

// Fakeable defines the interface for fakeable services
type Fakeable interface {
	// Fake replaces the real implementation with a fake
	Fake()

	// Restore restores the real implementation
	Restore()

	// IsFaked returns whether the service is currently faked
	IsFaked() bool
}

// FakeManager manages all fakes in the application
type FakeManager struct {
	fakes map[string]Fakeable
	mu    sync.RWMutex
}

// NewFakeManager creates a new fake manager
func NewFakeManager() *FakeManager {
	return &FakeManager{
		fakes: make(map[string]Fakeable),
	}
}

// Register registers a fakeable service
func (fm *FakeManager) Register(name string, fakeable Fakeable) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.fakes[name] = fakeable
}

// Fake fakes a specific service
func (fm *FakeManager) Fake(name string) error {
	fm.mu.RLock()
	fakeable, exists := fm.fakes[name]
	fm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service '%s' not found", name)
	}

	fakeable.Fake()
	return nil
}

// Restore restores a specific service
func (fm *FakeManager) Restore(name string) error {
	fm.mu.RLock()
	fakeable, exists := fm.fakes[name]
	fm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service '%s' not found", name)
	}

	fakeable.Restore()
	return nil
}

// FakeAll fakes all registered services
func (fm *FakeManager) FakeAll() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, fakeable := range fm.fakes {
		fakeable.Fake()
	}
}

// RestoreAll restores all registered services
func (fm *FakeManager) RestoreAll() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, fakeable := range fm.fakes {
		fakeable.Restore()
	}
}

// IsFaked checks if a service is faked
func (fm *FakeManager) IsFaked(name string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	fakeable, exists := fm.fakes[name]
	if !exists {
		return false
	}

	return fakeable.IsFaked()
}

// GetFakedServices returns a list of currently faked services
func (fm *FakeManager) GetFakedServices() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var faked []string
	for name, fakeable := range fm.fakes {
		if fakeable.IsFaked() {
			faked = append(faked, name)
		}
	}

	return faked
}

// FakeCache implements a fake cache service
type FakeCache struct {
	data      map[string]interface{}
	ttl       map[string]time.Time
	mu        sync.RWMutex
	isFaked   bool
	realCache interface{} // Store reference to real cache
}

// NewFakeCache creates a new fake cache
func NewFakeCache() *FakeCache {
	return &FakeCache{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
}

// Fake replaces the real cache with fake implementation
func (fc *FakeCache) Fake() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.isFaked = true
}

// Restore restores the real cache implementation
func (fc *FakeCache) Restore() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.isFaked = false
	fc.data = make(map[string]interface{})
	fc.ttl = make(map[string]time.Time)
}

// IsFaked returns whether the cache is faked
func (fc *FakeCache) IsFaked() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.isFaked
}

// Get retrieves a value from fake cache
func (fc *FakeCache) Get(key string) (interface{}, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if !fc.isFaked {
		return nil, false
	}

	// Check TTL
	if ttl, exists := fc.ttl[key]; exists && time.Now().After(ttl) {
		delete(fc.data, key)
		delete(fc.ttl, key)
		return nil, false
	}

	value, exists := fc.data[key]
	return value, exists
}

// Set stores a value in fake cache
func (fc *FakeCache) Set(key string, value interface{}, ttl time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if !fc.isFaked {
		return
	}

	fc.data[key] = value
	if ttl > 0 {
		fc.ttl[key] = time.Now().Add(ttl)
	}
}

// Delete removes a value from fake cache
func (fc *FakeCache) Delete(key string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if !fc.isFaked {
		return
	}

	delete(fc.data, key)
	delete(fc.ttl, key)
}

// Clear clears all fake cache data
func (fc *FakeCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if !fc.isFaked {
		return
	}

	fc.data = make(map[string]interface{})
	fc.ttl = make(map[string]time.Time)
}

// Has checks if a key exists in fake cache
func (fc *FakeCache) Has(key string) bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if !fc.isFaked {
		return false
	}

	// Check TTL
	if ttl, exists := fc.ttl[key]; exists && time.Now().After(ttl) {
		delete(fc.data, key)
		delete(fc.ttl, key)
		return false
	}

	_, exists := fc.data[key]
	return exists
}

// FakeFileStorage implements a fake file storage service
type FakeFileStorage struct {
	files       map[string][]byte
	mu          sync.RWMutex
	isFaked     bool
	realStorage interface{} // Store reference to real storage
}

// NewFakeFileStorage creates a new fake file storage
func NewFakeFileStorage() *FakeFileStorage {
	return &FakeFileStorage{
		files: make(map[string][]byte),
	}
}

// Fake replaces the real storage with fake implementation
func (ffs *FakeFileStorage) Fake() {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()
	ffs.isFaked = true
}

// Restore restores the real storage implementation
func (ffs *FakeFileStorage) Restore() {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()
	ffs.isFaked = false
	ffs.files = make(map[string][]byte)
}

// IsFaked returns whether the storage is faked
func (ffs *FakeFileStorage) IsFaked() bool {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()
	return ffs.isFaked
}

// Put stores a file in fake storage
func (ffs *FakeFileStorage) Put(path string, content []byte) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if !ffs.isFaked {
		return fmt.Errorf("storage is not faked")
	}

	ffs.files[path] = content
	return nil
}

// Get retrieves a file from fake storage
func (ffs *FakeFileStorage) Get(path string) ([]byte, error) {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()

	if !ffs.isFaked {
		return nil, fmt.Errorf("storage is not faked")
	}

	content, exists := ffs.files[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	return content, nil
}

// Delete removes a file from fake storage
func (ffs *FakeFileStorage) Delete(path string) error {
	ffs.mu.Lock()
	defer ffs.mu.Unlock()

	if !ffs.isFaked {
		return fmt.Errorf("storage is not faked")
	}

	delete(ffs.files, path)
	return nil
}

// Exists checks if a file exists in fake storage
func (ffs *FakeFileStorage) Exists(path string) bool {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()

	if !ffs.isFaked {
		return false
	}

	_, exists := ffs.files[path]
	return exists
}

// List lists all files in fake storage
func (ffs *FakeFileStorage) List(prefix string) []string {
	ffs.mu.RLock()
	defer ffs.mu.RUnlock()

	if !ffs.isFaked {
		return nil
	}

	var files []string
	for path := range ffs.files {
		if prefix == "" || len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			files = append(files, path)
		}
	}

	return files
}

// FakeEventDispatcher implements a fake event dispatcher
type FakeEventDispatcher struct {
	events         []Event
	listeners      map[string][]func(Event)
	mu             sync.RWMutex
	isFaked        bool
	realDispatcher interface{} // Store reference to real dispatcher
}

// Event represents a fake event
type Event struct {
	Name      string
	Payload   interface{}
	Timestamp time.Time
}

// NewFakeEventDispatcher creates a new fake event dispatcher
func NewFakeEventDispatcher() *FakeEventDispatcher {
	return &FakeEventDispatcher{
		events:    make([]Event, 0),
		listeners: make(map[string][]func(Event)),
	}
}

// Fake replaces the real dispatcher with fake implementation
func (fed *FakeEventDispatcher) Fake() {
	fed.mu.Lock()
	defer fed.mu.Unlock()
	fed.isFaked = true
}

// Restore restores the real dispatcher implementation
func (fed *FakeEventDispatcher) Restore() {
	fed.mu.Lock()
	defer fed.mu.Unlock()
	fed.isFaked = false
	fed.events = make([]Event, 0)
	fed.listeners = make(map[string][]func(Event))
}

// IsFaked returns whether the dispatcher is faked
func (fed *FakeEventDispatcher) IsFaked() bool {
	fed.mu.RLock()
	defer fed.mu.RUnlock()
	return fed.isFaked
}

// Dispatch dispatches an event in fake dispatcher
func (fed *FakeEventDispatcher) Dispatch(name string, payload interface{}) {
	fed.mu.Lock()
	defer fed.mu.Unlock()

	if !fed.isFaked {
		return
	}

	event := Event{
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	fed.events = append(fed.events, event)

	// Call listeners
	if listeners, exists := fed.listeners[name]; exists {
		for _, listener := range listeners {
			go listener(event)
		}
	}
}

// Listen adds an event listener in fake dispatcher
func (fed *FakeEventDispatcher) Listen(name string, listener func(Event)) {
	fed.mu.Lock()
	defer fed.mu.Unlock()

	if !fed.isFaked {
		return
	}

	if fed.listeners[name] == nil {
		fed.listeners[name] = make([]func(Event), 0)
	}
	fed.listeners[name] = append(fed.listeners[name], listener)
}

// GetEvents returns all dispatched events
func (fed *FakeEventDispatcher) GetEvents() []Event {
	fed.mu.RLock()
	defer fed.mu.RUnlock()

	if !fed.isFaked {
		return nil
	}

	// Return a copy
	events := make([]Event, len(fed.events))
	copy(events, fed.events)
	return events
}

// GetEventsByName returns events by name
func (fed *FakeEventDispatcher) GetEventsByName(name string) []Event {
	fed.mu.RLock()
	defer fed.mu.RUnlock()

	if !fed.isFaked {
		return nil
	}

	var events []Event
	for _, event := range fed.events {
		if event.Name == name {
			events = append(events, event)
		}
	}

	return events
}

// ClearEvents clears all events
func (fed *FakeEventDispatcher) ClearEvents() {
	fed.mu.Lock()
	defer fed.mu.Unlock()

	if !fed.isFaked {
		return
	}

	fed.events = make([]Event, 0)
}

// FakeMailer implements a fake mailer service
type FakeMailer struct {
	sentMails  []Mail
	mu         sync.RWMutex
	isFaked    bool
	realMailer interface{} // Store reference to real mailer
}

// Mail represents a fake mail
type Mail struct {
	To      []string
	Subject string
	Body    string
	HTML    string
	SentAt  time.Time
}

// NewFakeMailer creates a new fake mailer
func NewFakeMailer() *FakeMailer {
	return &FakeMailer{
		sentMails: make([]Mail, 0),
	}
}

// Fake replaces the real mailer with fake implementation
func (fm *FakeMailer) Fake() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.isFaked = true
}

// Restore restores the real mailer implementation
func (fm *FakeMailer) Restore() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.isFaked = false
	fm.sentMails = make([]Mail, 0)
}

// IsFaked returns whether the mailer is faked
func (fm *FakeMailer) IsFaked() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.isFaked
}

// Send sends a mail in fake mailer
func (fm *FakeMailer) Send(to []string, subject, body, html string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isFaked {
		return fmt.Errorf("mailer is not faked")
	}

	mail := Mail{
		To:      to,
		Subject: subject,
		Body:    body,
		HTML:    html,
		SentAt:  time.Now(),
	}

	fm.sentMails = append(fm.sentMails, mail)
	return nil
}

// GetSentMails returns all sent mails
func (fm *FakeMailer) GetSentMails() []Mail {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if !fm.isFaked {
		return nil
	}

	// Return a copy
	mails := make([]Mail, len(fm.sentMails))
	copy(mails, fm.sentMails)
	return mails
}

// ClearSentMails clears all sent mails
func (fm *FakeMailer) ClearSentMails() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.isFaked {
		return
	}

	fm.sentMails = make([]Mail, 0)
}

// FakeQueue implements a fake queue service
type FakeQueue struct {
	jobs      []Job
	mu        sync.RWMutex
	isFaked   bool
	realQueue interface{} // Store reference to real queue
}

// Job represents a fake job
type Job struct {
	Name      string
	Payload   interface{}
	Attempts  int
	CreatedAt time.Time
}

// NewFakeQueue creates a new fake queue
func NewFakeQueue() *FakeQueue {
	return &FakeQueue{
		jobs: make([]Job, 0),
	}
}

// Fake replaces the real queue with fake implementation
func (fq *FakeQueue) Fake() {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	fq.isFaked = true
}

// Restore restores the real queue implementation
func (fq *FakeQueue) Restore() {
	fq.mu.Lock()
	defer fq.mu.Unlock()
	fq.isFaked = false
	fq.jobs = make([]Job, 0)
}

// IsFaked returns whether the queue is faked
func (fq *FakeQueue) IsFaked() bool {
	fq.mu.RLock()
	defer fq.mu.RUnlock()
	return fq.isFaked
}

// Push pushes a job to fake queue
func (fq *FakeQueue) Push(name string, payload interface{}) error {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	if !fq.isFaked {
		return fmt.Errorf("queue is not faked")
	}

	job := Job{
		Name:      name,
		Payload:   payload,
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	fq.jobs = append(fq.jobs, job)
	return nil
}

// GetJobs returns all jobs
func (fq *FakeQueue) GetJobs() []Job {
	fq.mu.RLock()
	defer fq.mu.RUnlock()

	if !fq.isFaked {
		return nil
	}

	// Return a copy
	jobs := make([]Job, len(fq.jobs))
	copy(jobs, fq.jobs)
	return jobs
}

// GetJobsByName returns jobs by name
func (fq *FakeQueue) GetJobsByName(name string) []Job {
	fq.mu.RLock()
	defer fq.mu.RUnlock()

	if !fq.isFaked {
		return nil
	}

	var jobs []Job
	for _, job := range fq.jobs {
		if job.Name == name {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// ClearJobs clears all jobs
func (fq *FakeQueue) ClearJobs() {
	fq.mu.Lock()
	defer fq.mu.Unlock()

	if !fq.isFaked {
		return
	}

	fq.jobs = make([]Job, 0)
}
