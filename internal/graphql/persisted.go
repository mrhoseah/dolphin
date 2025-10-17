package graphql

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PersistedQuery represents a persisted GraphQL query
type PersistedQuery struct {
	ID          string    `json:"id"`
	Query       string    `json:"query"`
	Variables   string    `json:"variables,omitempty"`
	Operation   string    `json:"operation,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	UseCount    int       `json:"use_count"`
	Description string    `json:"description,omitempty"`
}

// PersistedQueryManager manages persisted queries
type PersistedQueryManager struct {
	queries   map[string]*PersistedQuery
	queryHash map[string]string // hash -> query ID
	mu        sync.RWMutex
	logger    *zap.Logger
	storage   PersistedQueryStorage
}

// PersistedQueryStorage defines the interface for persisted query storage
type PersistedQueryStorage interface {
	Save(query *PersistedQuery) error
	Load(id string) (*PersistedQuery, error)
	LoadByHash(hash string) (*PersistedQuery, error)
	Delete(id string) error
	List() ([]*PersistedQuery, error)
}

// MemoryPersistedQueryStorage implements in-memory storage for persisted queries
type MemoryPersistedQueryStorage struct {
	queries   map[string]*PersistedQuery
	queryHash map[string]string
	mu        sync.RWMutex
}

// NewMemoryPersistedQueryStorage creates a new in-memory storage
func NewMemoryPersistedQueryStorage() *MemoryPersistedQueryStorage {
	return &MemoryPersistedQueryStorage{
		queries:   make(map[string]*PersistedQuery),
		queryHash: make(map[string]string),
	}
}

// Save saves a persisted query
func (s *MemoryPersistedQueryStorage) Save(query *PersistedQuery) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queries[query.ID] = query
	hash := s.generateHash(query.Query)
	s.queryHash[hash] = query.ID

	return nil
}

// Load loads a persisted query by ID
func (s *MemoryPersistedQueryStorage) Load(id string) (*PersistedQuery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query, exists := s.queries[id]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", id)
	}

	return query, nil
}

// LoadByHash loads a persisted query by hash
func (s *MemoryPersistedQueryStorage) LoadByHash(hash string) (*PersistedQuery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queryID, exists := s.queryHash[hash]
	if !exists {
		return nil, fmt.Errorf("query not found for hash: %s", hash)
	}

	query, exists := s.queries[queryID]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", queryID)
	}

	return query, nil
}

// Delete deletes a persisted query
func (s *MemoryPersistedQueryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query, exists := s.queries[id]
	if !exists {
		return fmt.Errorf("query not found: %s", id)
	}

	// Remove from hash map
	hash := s.generateHash(query.Query)
	delete(s.queryHash, hash)
	delete(s.queries, id)

	return nil
}

// List lists all persisted queries
func (s *MemoryPersistedQueryStorage) List() ([]*PersistedQuery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	queries := make([]*PersistedQuery, 0, len(s.queries))
	for _, query := range s.queries {
		queries = append(queries, query)
	}

	return queries, nil
}

// generateHash generates a hash for a query
func (s *MemoryPersistedQueryStorage) generateHash(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

// NewPersistedQueryManager creates a new persisted query manager
func NewPersistedQueryManager(logger *zap.Logger) *PersistedQueryManager {
	return &PersistedQueryManager{
		queries:   make(map[string]*PersistedQuery),
		queryHash: make(map[string]string),
		logger:    logger,
		storage:   NewMemoryPersistedQueryStorage(),
	}
}

// SetStorage sets the storage backend
func (pqm *PersistedQueryManager) SetStorage(storage PersistedQueryStorage) {
	pqm.storage = storage
}

// PersistQuery persists a GraphQL query
func (pqm *PersistedQueryManager) PersistQuery(query, operation, description string) (*PersistedQuery, error) {
	// Generate query ID
	queryID := pqm.generateQueryID(query)

	// Check if query already exists
	if existing, err := pqm.storage.Load(queryID); err == nil {
		pqm.logger.Info("Query already persisted", zap.String("id", queryID))
		return existing, nil
	}

	// Create persisted query
	persistedQuery := &PersistedQuery{
		ID:          queryID,
		Query:       query,
		Operation:   operation,
		Description: description,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		UseCount:    0,
	}

	// Save to storage
	if err := pqm.storage.Save(persistedQuery); err != nil {
		return nil, fmt.Errorf("failed to save persisted query: %w", err)
	}

	pqm.logger.Info("Query persisted",
		zap.String("id", queryID),
		zap.String("operation", operation),
	)

	return persistedQuery, nil
}

// LoadQuery loads a persisted query by ID
func (pqm *PersistedQueryManager) LoadQuery(id string) (*PersistedQuery, error) {
	query, err := pqm.storage.Load(id)
	if err != nil {
		return nil, err
	}

	// Update usage statistics
	query.LastUsed = time.Now()
	query.UseCount++

	// Save updated query
	pqm.storage.Save(query)

	return query, nil
}

// LoadQueryByHash loads a persisted query by hash
func (pqm *PersistedQueryManager) LoadQueryByHash(hash string) (*PersistedQuery, error) {
	query, err := pqm.storage.LoadByHash(hash)
	if err != nil {
		return nil, err
	}

	// Update usage statistics
	query.LastUsed = time.Now()
	query.UseCount++

	// Save updated query
	pqm.storage.Save(query)

	return query, nil
}

// DeleteQuery deletes a persisted query
func (pqm *PersistedQueryManager) DeleteQuery(id string) error {
	if err := pqm.storage.Delete(id); err != nil {
		return err
	}

	pqm.logger.Info("Query deleted", zap.String("id", id))
	return nil
}

// ListQueries lists all persisted queries
func (pqm *PersistedQueryManager) ListQueries() ([]*PersistedQuery, error) {
	return pqm.storage.List()
}

// generateQueryID generates a unique ID for a query
func (pqm *PersistedQueryManager) generateQueryID(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

// PersistedQueryRequest represents a request with persisted query
type PersistedQueryRequest struct {
	ID        string                 `json:"id,omitempty"`
	Hash      string                 `json:"hash,omitempty"`
	Query     string                 `json:"query,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// HandlePersistedQuery handles a persisted query request
func (pqm *PersistedQueryManager) HandlePersistedQuery(req *PersistedQueryRequest) (*PersistedQuery, error) {
	var query *PersistedQuery
	var err error

	// Try to load by ID first
	if req.ID != "" {
		query, err = pqm.LoadQuery(req.ID)
		if err == nil {
			return query, nil
		}
		pqm.logger.Warn("Failed to load query by ID", zap.String("id", req.ID), zap.Error(err))
	}

	// Try to load by hash
	if req.Hash != "" {
		query, err = pqm.LoadQueryByHash(req.Hash)
		if err == nil {
			return query, nil
		}
		pqm.logger.Warn("Failed to load query by hash", zap.String("hash", req.Hash), zap.Error(err))
	}

	// If we have a query string, persist it
	if req.Query != "" {
		query, err = pqm.PersistQuery(req.Query, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to persist query: %w", err)
		}
		return query, nil
	}

	return nil, fmt.Errorf("no valid query identifier provided")
}

// GetStats returns persisted query statistics
func (pqm *PersistedQueryManager) GetStats() map[string]interface{} {
	queries, err := pqm.ListQueries()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	totalQueries := len(queries)
	totalUses := 0
	oldestQuery := time.Now()
	newestQuery := time.Time{}

	for _, query := range queries {
		totalUses += query.UseCount
		if query.CreatedAt.Before(oldestQuery) {
			oldestQuery = query.CreatedAt
		}
		if query.CreatedAt.After(newestQuery) {
			newestQuery = query.CreatedAt
		}
	}

	return map[string]interface{}{
		"total_queries": totalQueries,
		"total_uses":    totalUses,
		"oldest_query":  oldestQuery,
		"newest_query":  newestQuery,
	}
}

// PersistedQueryMiddleware creates middleware for persisted queries
func PersistedQueryMiddleware(pqm *PersistedQueryManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only handle GraphQL requests
			if r.URL.Path != "/graphql" {
				next.ServeHTTP(w, r)
				return
			}

			// Parse request body
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Check for persisted query
			if id, ok := req["id"].(string); ok {
				query, err := pqm.LoadQuery(id)
				if err == nil {
					// Replace the request with the persisted query
					req["query"] = query.Query
					delete(req, "id")
				}
			}

			// Re-encode the request
			body, err := json.Marshal(req)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Create new request with updated body
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			r.ContentLength = int64(len(body))

			next.ServeHTTP(w, r)
		})
	}
}
