package http

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// CorrelationIDGenerator generates correlation IDs
type CorrelationIDGenerator struct {
	// Configuration
	prefix    string
	length    int
	useTime   bool
	useRandom bool

	// Counter for sequential IDs
	counter int64

	// Mutex for thread safety
	mu sync.Mutex
}

// NewCorrelationIDGenerator creates a new correlation ID generator
func NewCorrelationIDGenerator() *CorrelationIDGenerator {
	return &CorrelationIDGenerator{
		prefix:    "dolphin",
		length:    16,
		useTime:   true,
		useRandom: true,
		counter:   0,
	}
}

// Generate generates a new correlation ID
func (cig *CorrelationIDGenerator) Generate() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	var parts []string

	// Add prefix if configured
	if cig.prefix != "" {
		parts = append(parts, cig.prefix)
	}

	// Add timestamp if configured
	if cig.useTime {
		timestamp := time.Now().UnixNano()
		parts = append(parts, fmt.Sprintf("%d", timestamp))
	}

	// Add counter
	cig.counter++
	parts = append(parts, fmt.Sprintf("%d", cig.counter))

	// Add random bytes if configured
	if cig.useRandom {
		randomBytes := make([]byte, cig.length/2)
		rand.Read(randomBytes)
		parts = append(parts, hex.EncodeToString(randomBytes))
	}

	// Join parts with separator
	correlationID := ""
	for i, part := range parts {
		if i > 0 {
			correlationID += "-"
		}
		correlationID += part
	}

	return correlationID
}

// GenerateWithPrefix generates a correlation ID with a custom prefix
func (cig *CorrelationIDGenerator) GenerateWithPrefix(prefix string) string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	var parts []string

	// Add custom prefix
	if prefix != "" {
		parts = append(parts, prefix)
	}

	// Add timestamp if configured
	if cig.useTime {
		timestamp := time.Now().UnixNano()
		parts = append(parts, fmt.Sprintf("%d", timestamp))
	}

	// Add counter
	cig.counter++
	parts = append(parts, fmt.Sprintf("%d", cig.counter))

	// Add random bytes if configured
	if cig.useRandom {
		randomBytes := make([]byte, cig.length/2)
		rand.Read(randomBytes)
		parts = append(parts, hex.EncodeToString(randomBytes))
	}

	// Join parts with separator
	correlationID := ""
	for i, part := range parts {
		if i > 0 {
			correlationID += "-"
		}
		correlationID += part
	}

	return correlationID
}

// GenerateShort generates a short correlation ID
func (cig *CorrelationIDGenerator) GenerateShort() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	// Generate random bytes
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	// Add counter for uniqueness
	cig.counter++

	// Create short ID
	correlationID := fmt.Sprintf("%s-%d", hex.EncodeToString(randomBytes), cig.counter)

	return correlationID
}

// GenerateLong generates a long correlation ID
func (cig *CorrelationIDGenerator) GenerateLong() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	var parts []string

	// Add prefix
	if cig.prefix != "" {
		parts = append(parts, cig.prefix)
	}

	// Add timestamp
	timestamp := time.Now().UnixNano()
	parts = append(parts, fmt.Sprintf("%d", timestamp))

	// Add counter
	cig.counter++
	parts = append(parts, fmt.Sprintf("%d", cig.counter))

	// Add random bytes
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	parts = append(parts, hex.EncodeToString(randomBytes))

	// Add process ID (simulated)
	parts = append(parts, fmt.Sprintf("%d", time.Now().Unix()%10000))

	// Join parts with separator
	correlationID := ""
	for i, part := range parts {
		if i > 0 {
			correlationID += "-"
		}
		correlationID += part
	}

	return correlationID
}

// GenerateUUID generates a UUID-style correlation ID
func (cig *CorrelationIDGenerator) GenerateUUID() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	// Generate random bytes
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// Set version (4) and variant bits
	randomBytes[6] = (randomBytes[6] & 0x0f) | 0x40
	randomBytes[8] = (randomBytes[8] & 0x3f) | 0x80

	// Format as UUID
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		randomBytes[0:4],
		randomBytes[4:6],
		randomBytes[6:8],
		randomBytes[8:10],
		randomBytes[10:16])

	return uuid
}

// GenerateTimestamp generates a timestamp-based correlation ID
func (cig *CorrelationIDGenerator) GenerateTimestamp() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	// Add counter for uniqueness
	cig.counter++

	// Create timestamp-based ID
	timestamp := time.Now().UnixNano()
	correlationID := fmt.Sprintf("%d-%d", timestamp, cig.counter)

	return correlationID
}

// GenerateRandom generates a random correlation ID
func (cig *CorrelationIDGenerator) GenerateRandom() string {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	// Generate random bytes
	randomBytes := make([]byte, cig.length)
	rand.Read(randomBytes)

	// Add counter for uniqueness
	cig.counter++

	// Create random ID
	correlationID := fmt.Sprintf("%s-%d", hex.EncodeToString(randomBytes), cig.counter)

	return correlationID
}

// SetPrefix sets the prefix for correlation IDs
func (cig *CorrelationIDGenerator) SetPrefix(prefix string) {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	cig.prefix = prefix
}

// SetLength sets the length for random parts
func (cig *CorrelationIDGenerator) SetLength(length int) {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	cig.length = length
}

// SetUseTime sets whether to include timestamp
func (cig *CorrelationIDGenerator) SetUseTime(useTime bool) {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	cig.useTime = useTime
}

// SetUseRandom sets whether to include random bytes
func (cig *CorrelationIDGenerator) SetUseRandom(useRandom bool) {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	cig.useRandom = useRandom
}

// GetConfig returns the current configuration
func (cig *CorrelationIDGenerator) GetConfig() map[string]interface{} {
	cig.mu.RLock()
	defer cig.mu.RUnlock()

	return map[string]interface{}{
		"prefix":     cig.prefix,
		"length":     cig.length,
		"use_time":   cig.useTime,
		"use_random": cig.useRandom,
		"counter":    cig.counter,
	}
}

// GetStats returns generator statistics
func (cig *CorrelationIDGenerator) GetStats() map[string]interface{} {
	cig.mu.RLock()
	defer cig.mu.RUnlock()

	return map[string]interface{}{
		"total_generated": cig.counter,
		"prefix":          cig.prefix,
		"length":          cig.length,
		"use_time":        cig.useTime,
		"use_random":      cig.useRandom,
	}
}

// Reset resets the counter
func (cig *CorrelationIDGenerator) Reset() {
	cig.mu.Lock()
	defer cig.mu.Unlock()

	cig.counter = 0
}

// GetCounter returns the current counter value
func (cig *CorrelationIDGenerator) GetCounter() int64 {
	cig.mu.RLock()
	defer cig.mu.RUnlock()

	return cig.counter
}

// Validate validates a correlation ID format
func (cig *CorrelationIDGenerator) Validate(correlationID string) bool {
	if correlationID == "" {
		return false
	}

	// Basic validation - should contain at least one separator
	if len(correlationID) < 8 {
		return false
	}

	// Check for valid characters (alphanumeric and hyphens)
	for _, char := range correlationID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return false
		}
	}

	return true
}

// Parse parses a correlation ID and returns its components
func (cig *CorrelationIDGenerator) Parse(correlationID string) map[string]interface{} {
	if !cig.Validate(correlationID) {
		return nil
	}

	parts := strings.Split(correlationID, "-")

	result := map[string]interface{}{
		"correlation_id": correlationID,
		"parts":          parts,
		"part_count":     len(parts),
		"is_valid":       true,
	}

	// Try to identify components
	if len(parts) >= 2 {
		result["prefix"] = parts[0]
	}

	if len(parts) >= 3 {
		result["timestamp"] = parts[1]
	}

	if len(parts) >= 4 {
		result["counter"] = parts[2]
	}

	if len(parts) >= 5 {
		result["random"] = parts[3]
	}

	return result
}

// GetInfo returns detailed information about the generator
func (cig *CorrelationIDGenerator) GetInfo() map[string]interface{} {
	cig.mu.RLock()
	defer cig.mu.RUnlock()

	return map[string]interface{}{
		"prefix":          cig.prefix,
		"length":          cig.length,
		"use_time":        cig.useTime,
		"use_random":      cig.useRandom,
		"counter":         cig.counter,
		"total_generated": cig.counter,
		"config":          cig.GetConfig(),
		"stats":           cig.GetStats(),
	}
}
