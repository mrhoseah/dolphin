package livereload

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Stats represents live reload statistics
type Stats struct {
	// File changes
	FileChanges    int64            `json:"file_changes"`
	FilesChanged   map[string]int64 `json:"files_changed"`
	ChangeTypes    map[string]int64 `json:"change_types"`
	
	// Reloads
	Reloads        int64     `json:"reloads"`
	LastReload     time.Time `json:"last_reload"`
	ReloadDuration time.Duration `json:"reload_duration"`
	
	// Process
	ProcessStarts  int64     `json:"process_starts"`
	ProcessStops   int64     `json:"process_stops"`
	LastStart      time.Time `json:"last_start"`
	LastStop       time.Time `json:"last_stop"`
	
	// Hot reload
	HotReloads     int64     `json:"hot_reloads"`
	LastHotReload  time.Time `json:"last_hot_reload"`
	
	// Timing
	StartTime      time.Time `json:"start_time"`
	Uptime         time.Duration `json:"uptime"`
	
	// Mutex for thread safety
	mu sync.RWMutex
}

// NewStats creates new live reload statistics
func NewStats() *Stats {
	return &Stats{
		FilesChanged: make(map[string]int64),
		ChangeTypes:  make(map[string]int64),
		StartTime:    time.Now(),
	}
}

// RecordFileChange records a file change
func (s *Stats) RecordFileChange(filename string, op fsnotify.Op) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.FileChanges++
	s.FilesChanged[filename]++
	s.ChangeTypes[op.String()]++
}

// RecordReload records a reload
func (s *Stats) RecordReload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Reloads++
	s.LastReload = time.Now()
}

// RecordProcessStart records a process start
func (s *Stats) RecordProcessStart() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ProcessStarts++
	s.LastStart = time.Now()
}

// RecordProcessStop records a process stop
func (s *Stats) RecordProcessStop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ProcessStops++
	s.LastStop = time.Now()
}

// RecordHotReload records a hot reload
func (s *Stats) RecordHotReload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.HotReloads++
	s.LastHotReload = time.Now()
}

// RecordReloadDuration records reload duration
func (s *Stats) RecordReloadDuration(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.ReloadDuration = duration
}

// GetStats returns current statistics
func (s *Stats) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	uptime := time.Since(s.StartTime)
	
	return map[string]interface{}{
		"file_changes":     s.FileChanges,
		"files_changed":    s.FilesChanged,
		"change_types":     s.ChangeTypes,
		"reloads":          s.Reloads,
		"last_reload":      s.LastReload,
		"reload_duration":  s.ReloadDuration,
		"process_starts":   s.ProcessStarts,
		"process_stops":    s.ProcessStops,
		"last_start":       s.LastStart,
		"last_stop":        s.LastStop,
		"hot_reloads":      s.HotReloads,
		"last_hot_reload":  s.LastHotReload,
		"start_time":       s.StartTime,
		"uptime":           uptime,
	}
}

// GetFileChangeRate returns the file change rate per minute
func (s *Stats) GetFileChangeRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	uptime := time.Since(s.StartTime)
	if uptime.Minutes() == 0 {
		return 0
	}
	
	return float64(s.FileChanges) / uptime.Minutes()
}

// GetReloadRate returns the reload rate per minute
func (s *Stats) GetReloadRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	uptime := time.Since(s.StartTime)
	if uptime.Minutes() == 0 {
		return 0
	}
	
	return float64(s.Reloads) / uptime.Minutes()
}

// GetMostChangedFiles returns the most changed files
func (s *Stats) GetMostChangedFiles(limit int) []FileChangeInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Convert to slice and sort
	files := make([]FileChangeInfo, 0, len(s.FilesChanged))
	for filename, count := range s.FilesChanged {
		files = append(files, FileChangeInfo{
			Filename: filename,
			Count:    count,
		})
	}
	
	// Sort by count (descending)
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].Count < files[j].Count {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	
	// Return limited results
	if limit > 0 && limit < len(files) {
		files = files[:limit]
	}
	
	return files
}

// GetChangeTypeStats returns change type statistics
func (s *Stats) GetChangeTypeStats() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Create a copy
	stats := make(map[string]int64)
	for changeType, count := range s.ChangeTypes {
		stats[changeType] = count
	}
	
	return stats
}

// Reset resets all statistics
func (s *Stats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.FileChanges = 0
	s.FilesChanged = make(map[string]int64)
	s.ChangeTypes = make(map[string]int64)
	s.Reloads = 0
	s.LastReload = time.Time{}
	s.ReloadDuration = 0
	s.ProcessStarts = 0
	s.ProcessStops = 0
	s.LastStart = time.Time{}
	s.LastStop = time.Time{}
	s.HotReloads = 0
	s.LastHotReload = time.Time{}
	s.StartTime = time.Now()
}

// FileChangeInfo represents file change information
type FileChangeInfo struct {
	Filename string `json:"filename"`
	Count    int64  `json:"count"`
}

// ReloadSummary represents a reload summary
type ReloadSummary struct {
	TotalReloads     int64         `json:"total_reloads"`
	FileChanges      int64         `json:"file_changes"`
	ReloadRate       float64       `json:"reload_rate"`
	FileChangeRate   float64       `json:"file_change_rate"`
	AverageReloadTime time.Duration `json:"average_reload_time"`
	MostChangedFiles []FileChangeInfo `json:"most_changed_files"`
	ChangeTypes      map[string]int64 `json:"change_types"`
	Uptime           time.Duration `json:"uptime"`
}
