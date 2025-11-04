package performance

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Budget represents a performance budget
type Budget struct {
	Name        string        `json:"name"`
	Metric      string        `json:"metric"`      // response_time, memory, cpu, etc.
	Threshold   float64       `json:"threshold"`   // Threshold value
	Window      time.Duration `json:"window"`      // Time window for evaluation
	Description string        `json:"description"`
	AlertOn     string        `json:"alert_on"`    // exceed, below, equal
}

// Measurement represents a performance measurement
type Measurement struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Context   string    `json:"context,omitempty"` // e.g., route path
}

// Monitor monitors performance budgets
type Monitor struct {
	budgets     map[string]*Budget
	measurements []Measurement
	mu          sync.RWMutex
	logger      *zap.Logger
	filePath    string
	maxHistory  int
	alerts      []Alert
}

// Alert represents a performance budget alert
type Alert struct {
	Budget      string    `json:"budget"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // warning, critical
}

// NewMonitor creates a new performance budget monitor
func NewMonitor(logger *zap.Logger, filePath string) *Monitor {
	m := &Monitor{
		budgets:     make(map[string]*Budget),
		measurements: make([]Measurement, 0),
		logger:      logger,
		filePath:    filePath,
		maxHistory:  10000, // Keep last 10k measurements
		alerts:      make([]Alert, 0),
	}
	m.loadFromFile()
	return m
}

// Register registers a performance budget
func (m *Monitor) Register(budget *Budget) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.budgets[budget.Name]; exists {
		return fmt.Errorf("budget '%s' already exists", budget.Name)
	}

	m.budgets[budget.Name] = budget

	if m.filePath != "" {
		return m.saveToFile()
	}
	return nil
}

// Record records a performance measurement
func (m *Monitor) Record(metric string, value float64, context string) {
	measurement := Measurement{
		Metric:    metric,
		Value:     value,
		Timestamp: time.Now(),
		Context:   context,
	}

	m.mu.Lock()
	m.measurements = append(m.measurements, measurement)

	// Keep only recent measurements
	if len(m.measurements) > m.maxHistory {
		m.measurements = m.measurements[len(m.measurements)-m.maxHistory:]
	}

	// Check budgets
	m.checkBudgets(measurement)
	m.mu.Unlock()
}

// checkBudgets checks if measurement violates any budgets
func (m *Monitor) checkBudgets(measurement Measurement) {
	for name, budget := range m.budgets {
		if budget.Metric != measurement.Metric {
			continue
		}

		violated := false
		severity := "warning"

		switch budget.AlertOn {
		case "exceed":
			violated = measurement.Value > budget.Threshold
			if violated {
				severity = "critical"
			}
		case "below":
			violated = measurement.Value < budget.Threshold
		case "equal":
			violated = measurement.Value == budget.Threshold
		}

		if violated {
			alert := Alert{
				Budget:    name,
				Metric:    budget.Metric,
				Value:     measurement.Value,
				Threshold: budget.Threshold,
				Timestamp: measurement.Timestamp,
				Severity:  severity,
			}

			m.alerts = append(m.alerts, alert)

			// Keep only recent alerts (last 1000)
			if len(m.alerts) > 1000 {
				m.alerts = m.alerts[len(m.alerts)-1000:]
			}

			m.logger.Warn("Performance budget violated",
				zap.String("budget", name),
				zap.String("metric", budget.Metric),
				zap.Float64("value", measurement.Value),
				zap.Float64("threshold", budget.Threshold),
				zap.String("severity", severity),
			)
		}
	}
}

// GetStats returns statistics for a metric
func (m *Monitor) GetStats(metric string, window time.Duration) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	var relevant []Measurement

	for _, m := range m.measurements {
		if m.Metric == metric && m.Timestamp.After(cutoff) {
			relevant = append(relevant, m)
		}
	}

	if len(relevant) == 0 {
		return map[string]interface{}{
			"metric":  metric,
			"count":   0,
			"average": 0.0,
			"min":     0.0,
			"max":     0.0,
		}
	}

	sum := 0.0
	min := relevant[0].Value
	max := relevant[0].Value

	for _, m := range relevant {
		sum += m.Value
		if m.Value < min {
			min = m.Value
		}
		if m.Value > max {
			max = m.Value
		}
	}

	return map[string]interface{}{
		"metric":  metric,
		"count":   len(relevant),
		"average": sum / float64(len(relevant)),
		"min":     min,
		"max":     max,
	}
}

// GetAlerts returns recent alerts
func (m *Monitor) GetAlerts(limit int) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	result := make([]Alert, limit)
	copy(result, m.alerts[len(m.alerts)-limit:])
	return result
}

// GetBudgets returns all budgets
func (m *Monitor) GetBudgets() map[string]*Budget {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Budget)
	for name, budget := range m.budgets {
		budgetCopy := *budget
		result[name] = &budgetCopy
	}
	return result
}

// loadFromFile loads budgets from file
func (m *Monitor) loadFromFile() error {
	if m.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var budgets map[string]*Budget
	if err := json.Unmarshal(data, &budgets); err != nil {
		return err
	}

	m.mu.Lock()
	m.budgets = budgets
	m.mu.Unlock()

	return nil
}

// saveToFile saves budgets to file
func (m *Monitor) saveToFile() error {
	if m.filePath == "" {
		return nil
	}

	m.mu.RLock()
	data, err := json.MarshalIndent(m.budgets, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

