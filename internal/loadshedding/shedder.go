package loadshedding

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SheddingStrategy represents the load shedding strategy
type SheddingStrategy int

const (
	StrategyCPU SheddingStrategy = iota
	StrategyMemory
	StrategyGoroutines
	StrategyRequestRate
	StrategyResponseTime
	StrategyCombined
)

func (s SheddingStrategy) String() string {
	switch s {
	case StrategyCPU:
		return "CPU"
	case StrategyMemory:
		return "Memory"
	case StrategyGoroutines:
		return "Goroutines"
	case StrategyRequestRate:
		return "RequestRate"
	case StrategyResponseTime:
		return "ResponseTime"
	case StrategyCombined:
		return "Combined"
	default:
		return "Unknown"
	}
}

// SheddingLevel represents the current shedding level
type SheddingLevel int

const (
	LevelNone SheddingLevel = iota
	LevelLight
	LevelModerate
	LevelHeavy
	LevelCritical
)

func (l SheddingLevel) String() string {
	switch l {
	case LevelNone:
		return "None"
	case LevelLight:
		return "Light"
	case LevelModerate:
		return "Moderate"
	case LevelHeavy:
		return "Heavy"
	case LevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// Config represents load shedding configuration
type Config struct {
	// Shedding strategy
	Strategy SheddingStrategy `yaml:"strategy" json:"strategy"`

	// Thresholds for different levels
	LightThreshold    float64 `yaml:"light_threshold" json:"light_threshold"`
	ModerateThreshold float64 `yaml:"moderate_threshold" json:"moderate_threshold"`
	HeavyThreshold    float64 `yaml:"heavy_threshold" json:"heavy_threshold"`
	CriticalThreshold float64 `yaml:"critical_threshold" json:"critical_threshold"`

	// Shedding rates for each level
	LightShedRate    float64 `yaml:"light_shed_rate" json:"light_shed_rate"`
	ModerateShedRate float64 `yaml:"moderate_shed_rate" json:"moderate_shed_rate"`
	HeavyShedRate    float64 `yaml:"heavy_shed_rate" json:"heavy_shed_rate"`
	CriticalShedRate float64 `yaml:"critical_shed_rate" json:"critical_shed_rate"`

	// Monitoring intervals
	CheckInterval    time.Duration `yaml:"check_interval" json:"check_interval"`
	AdaptiveInterval time.Duration `yaml:"adaptive_interval" json:"adaptive_interval"`

	// Hysteresis to prevent oscillation
	Hysteresis float64 `yaml:"hysteresis" json:"hysteresis"`

	// Minimum and maximum shedding rates
	MinShedRate float64 `yaml:"min_shed_rate" json:"min_shed_rate"`
	MaxShedRate float64 `yaml:"max_shed_rate" json:"max_shed_rate"`

	// Enable adaptive adjustment
	EnableAdaptive bool `yaml:"enable_adaptive" json:"enable_adaptive"`

	// Enable logging
	EnableLogging bool `yaml:"enable_logging" json:"enable_logging"`
}

// DefaultConfig returns default load shedding configuration
func DefaultConfig() *Config {
	return &Config{
		Strategy:          StrategyCombined,
		LightThreshold:    0.6,
		ModerateThreshold: 0.75,
		HeavyThreshold:    0.85,
		CriticalThreshold: 0.95,
		LightShedRate:     0.1,
		ModerateShedRate:  0.3,
		HeavyShedRate:     0.6,
		CriticalShedRate:  0.9,
		CheckInterval:     1 * time.Second,
		AdaptiveInterval:  5 * time.Second,
		Hysteresis:        0.05,
		MinShedRate:       0.0,
		MaxShedRate:       0.95,
		EnableAdaptive:    true,
		EnableLogging:     true,
	}
}

// LoadShedder implements adaptive load shedding
type LoadShedder struct {
	config *Config
	logger *zap.Logger

	// Current state
	currentLevel    SheddingLevel
	currentShedRate float64
	mu              sync.RWMutex

	// Metrics
	metrics *Metrics

	// System monitoring
	cpuUsage     float64
	memoryUsage  float64
	goroutines   int
	requestRate  float64
	responseTime time.Duration

	// Adaptive adjustment
	adaptiveEnabled bool
	lastAdjustment  time.Time
	adjustmentCount int

	// Control channels
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewLoadShedder creates a new load shedder
func NewLoadShedder(config *Config, logger *zap.Logger) *LoadShedder {
	if config == nil {
		config = DefaultConfig()
	}

	ls := &LoadShedder{
		config:          config,
		logger:          logger,
		currentLevel:    LevelNone,
		currentShedRate: 0.0,
		metrics:         NewMetrics("loadshedder"),
		adaptiveEnabled: config.EnableAdaptive,
		stopChan:        make(chan struct{}),
		doneChan:        make(chan struct{}),
	}

	// Start monitoring
	go ls.startMonitoring()

	return ls
}

// ShouldShed determines if a request should be shed
func (ls *LoadShedder) ShouldShed(ctx context.Context) bool {
	ls.mu.RLock()
	shouldShed := ls.shouldShedInternal()
	ls.mu.RUnlock()

	// Record metrics
	ls.metrics.RecordRequest(shouldShed)

	return shouldShed
}

// shouldShedInternal determines if a request should be shed (internal method)
func (ls *LoadShedder) shouldShedInternal() bool {
	// Always shed if in critical level
	if ls.currentLevel >= LevelCritical {
		return true
	}

	// Use current shed rate for probabilistic shedding
	if ls.currentShedRate > 0 {
		// Simple probabilistic shedding based on current rate
		// In a real implementation, you might use more sophisticated algorithms
		return ls.currentShedRate > 0.5
	}

	return false
}

// GetCurrentLevel returns the current shedding level
func (ls *LoadShedder) GetCurrentLevel() SheddingLevel {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.currentLevel
}

// GetCurrentShedRate returns the current shedding rate
func (ls *LoadShedder) GetCurrentShedRate() float64 {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.currentShedRate
}

// GetStats returns current load shedding statistics
func (ls *LoadShedder) GetStats() Stats {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return Stats{
		CurrentLevel:    ls.currentLevel,
		CurrentShedRate: ls.currentShedRate,
		CPUUsage:        ls.cpuUsage,
		MemoryUsage:     ls.memoryUsage,
		Goroutines:      ls.goroutines,
		RequestRate:     ls.requestRate,
		ResponseTime:    ls.responseTime,
		LastAdjustment:  ls.lastAdjustment,
		AdjustmentCount: ls.adjustmentCount,
	}
}

// ForceLevel forces a specific shedding level
func (ls *LoadShedder) ForceLevel(level SheddingLevel) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.currentLevel = level
	ls.currentShedRate = ls.getShedRateForLevel(level)

	if ls.config.EnableLogging && ls.logger != nil {
		ls.logger.Info("Load shedding level forced",
			zap.String("level", level.String()),
			zap.Float64("shed_rate", ls.currentShedRate))
	}
}

// Reset resets the load shedder to normal operation
func (ls *LoadShedder) Reset() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.currentLevel = LevelNone
	ls.currentShedRate = 0.0
	ls.lastAdjustment = time.Time{}
	ls.adjustmentCount = 0

	if ls.config.EnableLogging && ls.logger != nil {
		ls.logger.Info("Load shedder reset")
	}
}

// Stop stops the load shedder
func (ls *LoadShedder) Stop() {
	close(ls.stopChan)
	<-ls.doneChan
}

// startMonitoring starts the monitoring goroutine
func (ls *LoadShedder) startMonitoring() {
	defer close(ls.doneChan)

	ticker := time.NewTicker(ls.config.CheckInterval)
	defer ticker.Stop()

	adaptiveTicker := time.NewTicker(ls.config.AdaptiveInterval)
	defer adaptiveTicker.Stop()

	for {
		select {
		case <-ls.stopChan:
			return
		case <-ticker.C:
			ls.updateSystemMetrics()
			ls.adjustSheddingLevel()
		case <-adaptiveTicker.C:
			if ls.adaptiveEnabled {
				ls.performAdaptiveAdjustment()
			}
		}
	}
}

// updateSystemMetrics updates system metrics
func (ls *LoadShedder) updateSystemMetrics() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Update CPU usage
	ls.cpuUsage = ls.getCPUUsage()

	// Update memory usage
	ls.memoryUsage = ls.getMemoryUsage()

	// Update goroutine count
	ls.goroutines = runtime.NumGoroutine()

	// Update request rate (this would be provided by the application)
	ls.requestRate = ls.metrics.GetRequestRate()

	// Update response time (this would be provided by the application)
	ls.responseTime = ls.metrics.GetAverageResponseTime()
}

// adjustSheddingLevel adjusts the shedding level based on current metrics
func (ls *LoadShedder) adjustSheddingLevel() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	oldLevel := ls.currentLevel
	oldShedRate := ls.currentShedRate

	// Determine new level based on strategy
	newLevel := ls.determineSheddingLevel()

	// Apply hysteresis to prevent oscillation
	if ls.shouldApplyHysteresis(oldLevel, newLevel) {
		return
	}

	// Update level and shed rate
	ls.currentLevel = newLevel
	ls.currentShedRate = ls.getShedRateForLevel(newLevel)

	// Log level change
	if oldLevel != newLevel && ls.config.EnableLogging && ls.logger != nil {
		ls.logger.Info("Load shedding level changed",
			zap.String("old_level", oldLevel.String()),
			zap.String("new_level", newLevel.String()),
			zap.Float64("old_shed_rate", oldShedRate),
			zap.Float64("new_shed_rate", ls.currentShedRate),
			zap.Float64("cpu_usage", ls.cpuUsage),
			zap.Float64("memory_usage", ls.memoryUsage),
			zap.Int("goroutines", ls.goroutines))
	}
}

// determineSheddingLevel determines the appropriate shedding level
func (ls *LoadShedder) determineSheddingLevel() SheddingLevel {
	var loadValue float64

	switch ls.config.Strategy {
	case StrategyCPU:
		loadValue = ls.cpuUsage
	case StrategyMemory:
		loadValue = ls.memoryUsage
	case StrategyGoroutines:
		loadValue = float64(ls.goroutines) / 1000.0 // Normalize to 0-1 range
	case StrategyRequestRate:
		loadValue = ls.requestRate / 1000.0 // Normalize to 0-1 range
	case StrategyResponseTime:
		loadValue = float64(ls.responseTime.Milliseconds()) / 1000.0 // Normalize to 0-1 range
	case StrategyCombined:
		loadValue = ls.calculateCombinedLoad()
	default:
		loadValue = ls.calculateCombinedLoad()
	}

	// Determine level based on thresholds
	if loadValue >= ls.config.CriticalThreshold {
		return LevelCritical
	} else if loadValue >= ls.config.HeavyThreshold {
		return LevelHeavy
	} else if loadValue >= ls.config.ModerateThreshold {
		return LevelModerate
	} else if loadValue >= ls.config.LightThreshold {
		return LevelLight
	}

	return LevelNone
}

// calculateCombinedLoad calculates a combined load metric
func (ls *LoadShedder) calculateCombinedLoad() float64 {
	// Weighted combination of different metrics
	cpuWeight := 0.4
	memoryWeight := 0.3
	goroutineWeight := 0.2
	responseTimeWeight := 0.1

	// Normalize goroutines (assume 1000 is high)
	normalizedGoroutines := math.Min(float64(ls.goroutines)/1000.0, 1.0)

	// Normalize response time (assume 1 second is high)
	normalizedResponseTime := math.Min(float64(ls.responseTime.Milliseconds())/1000.0, 1.0)

	return cpuWeight*ls.cpuUsage +
		memoryWeight*ls.memoryUsage +
		goroutineWeight*normalizedGoroutines +
		responseTimeWeight*normalizedResponseTime
}

// shouldApplyHysteresis determines if hysteresis should be applied
func (ls *LoadShedder) shouldApplyHysteresis(oldLevel, newLevel SheddingLevel) bool {
	if oldLevel == newLevel {
		return false
	}

	// Apply hysteresis to prevent oscillation
	levelDiff := math.Abs(float64(newLevel) - float64(oldLevel))
	return levelDiff == 1 && ls.config.Hysteresis > 0
}

// getShedRateForLevel returns the shedding rate for a given level
func (ls *LoadShedder) getShedRateForLevel(level SheddingLevel) float64 {
	var rate float64

	switch level {
	case LevelNone:
		rate = 0.0
	case LevelLight:
		rate = ls.config.LightShedRate
	case LevelModerate:
		rate = ls.config.ModerateShedRate
	case LevelHeavy:
		rate = ls.config.HeavyShedRate
	case LevelCritical:
		rate = ls.config.CriticalShedRate
	default:
		rate = 0.0
	}

	// Apply min/max constraints
	rate = math.Max(rate, ls.config.MinShedRate)
	rate = math.Min(rate, ls.config.MaxShedRate)

	return rate
}

// performAdaptiveAdjustment performs adaptive adjustment of thresholds
func (ls *LoadShedder) performAdaptiveAdjustment() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.lastAdjustment = time.Now()
	ls.adjustmentCount++

	// Simple adaptive adjustment based on recent performance
	// In a real implementation, this would be more sophisticated

	if ls.config.EnableLogging && ls.logger != nil {
		ls.logger.Debug("Performing adaptive adjustment",
			zap.Int("adjustment_count", ls.adjustmentCount),
			zap.Float64("current_shed_rate", ls.currentShedRate))
	}
}

// getCPUUsage returns current CPU usage (simplified)
func (ls *LoadShedder) getCPUUsage() float64 {
	// This is a simplified implementation
	// In a real implementation, you would use proper CPU monitoring
	return 0.5 // Placeholder
}

// getMemoryUsage returns current memory usage
func (ls *LoadShedder) getMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage as a percentage
	totalMemory := uint64(8 * 1024 * 1024 * 1024) // Assume 8GB total
	return float64(m.Alloc) / float64(totalMemory)
}

// Stats represents load shedding statistics
type Stats struct {
	CurrentLevel    SheddingLevel `json:"current_level"`
	CurrentShedRate float64       `json:"current_shed_rate"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     float64       `json:"memory_usage"`
	Goroutines      int           `json:"goroutines"`
	RequestRate     float64       `json:"request_rate"`
	ResponseTime    time.Duration `json:"response_time"`
	LastAdjustment  time.Time     `json:"last_adjustment"`
	AdjustmentCount int           `json:"adjustment_count"`
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
