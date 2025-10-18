package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TransactionManager provides advanced transaction management
type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// ExecuteInTransaction executes a function within a transaction
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ExecuteInNestedTransaction supports nested transactions using savepoints
func (tm *TransactionManager) ExecuteInNestedTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	// Check if we're already in a transaction
	if existingTx := ctx.Value("tx"); existingTx != nil {
		tx := existingTx.(*sql.Tx)

		// Create savepoint
		savepoint := fmt.Sprintf("sp_%d", time.Now().UnixNano())
		if _, err := tx.ExecContext(ctx, "SAVEPOINT "+savepoint); err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint)
				panic(p)
			}
		}()

		if err := fn(tx); err != nil {
			tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint)
			return err
		}

		return nil
	}

	// Start new transaction
	return tm.ExecuteInTransaction(ctx, fn)
}

// TransactionTemplate provides a template for common transaction patterns
type TransactionTemplate struct {
	transactionManager *TransactionManager
}

func NewTransactionTemplate(tm *TransactionManager) *TransactionTemplate {
	return &TransactionTemplate{transactionManager: tm}
}

// ExecuteWithRetry executes a transaction with retry logic
func (tt *TransactionTemplate) ExecuteWithRetry(ctx context.Context, maxRetries int, fn func(tx *sql.Tx) error) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		err := tt.transactionManager.ExecuteInTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !tt.isRetryableError(err) {
			return err
		}

		// Wait before retry
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

func (tt *TransactionTemplate) isRetryableError(err error) bool {
	// Check for deadlock, timeout, or connection errors
	errStr := err.Error()
	return strings.Contains(errStr, "deadlock") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection")
}

// TransactionIsolationLevel represents different isolation levels
type TransactionIsolationLevel int

const (
	ReadUncommitted TransactionIsolationLevel = iota
	ReadCommitted
	RepeatableRead
	Serializable
)

// ExecuteWithIsolation executes a transaction with specific isolation level
func (tm *TransactionManager) ExecuteWithIsolation(ctx context.Context, level TransactionIsolationLevel, fn func(tx *sql.Tx) error) error {
	return tm.ExecuteInTransaction(ctx, func(tx *sql.Tx) error {
		// Set isolation level
		var isolationSQL string
		switch level {
		case ReadUncommitted:
			isolationSQL = "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED"
		case ReadCommitted:
			isolationSQL = "SET TRANSACTION ISOLATION LEVEL READ COMMITTED"
		case RepeatableRead:
			isolationSQL = "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"
		case Serializable:
			isolationSQL = "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"
		}

		if _, err := tx.ExecContext(ctx, isolationSQL); err != nil {
			return err
		}

		return fn(tx)
	})
}

// TransactionOptions provides options for transaction execution
type TransactionOptions struct {
	IsolationLevel TransactionIsolationLevel
	ReadOnly       bool
	Timeout        time.Duration
}

// ExecuteWithOptions executes a transaction with specific options
func (tm *TransactionManager) ExecuteWithOptions(ctx context.Context, options TransactionOptions, fn func(tx *sql.Tx) error) error {
	// Create context with timeout if specified
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Set transaction options
	txOptions := &sql.TxOptions{
		Isolation: sql.IsolationLevel(options.IsolationLevel),
		ReadOnly:  options.ReadOnly,
	}

	tx, err := tm.db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// TransactionCallback defines a callback for transaction events
type TransactionCallback interface {
	BeforeTransaction(ctx context.Context) error
	AfterTransaction(ctx context.Context, err error) error
}

// TransactionCallbackManager manages transaction callbacks
type TransactionCallbackManager struct {
	callbacks []TransactionCallback
}

func NewTransactionCallbackManager() *TransactionCallbackManager {
	return &TransactionCallbackManager{
		callbacks: make([]TransactionCallback, 0),
	}
}

func (tcm *TransactionCallbackManager) AddCallback(callback TransactionCallback) {
	tcm.callbacks = append(tcm.callbacks, callback)
}

func (tcm *TransactionCallbackManager) ExecuteWithCallbacks(ctx context.Context, tm *TransactionManager, fn func(tx *sql.Tx) error) error {
	// Execute before callbacks
	for _, callback := range tcm.callbacks {
		if err := callback.BeforeTransaction(ctx); err != nil {
			return err
		}
	}

	// Execute transaction
	err := tm.ExecuteInTransaction(ctx, fn)

	// Execute after callbacks
	for _, callback := range tcm.callbacks {
		if afterErr := callback.AfterTransaction(ctx, err); afterErr != nil {
			return afterErr
		}
	}

	return err
}

// LoggingTransactionCallback provides logging for transactions
type LoggingTransactionCallback struct {
	logger Logger
}

func NewLoggingTransactionCallback(logger Logger) *LoggingTransactionCallback {
	return &LoggingTransactionCallback{logger: logger}
}

func (ltc *LoggingTransactionCallback) BeforeTransaction(ctx context.Context) error {
	ltc.logger.Info("Starting transaction")
	return nil
}

func (ltc *LoggingTransactionCallback) AfterTransaction(ctx context.Context, err error) error {
	if err != nil {
		ltc.logger.Error("Transaction failed", "error", err)
	} else {
		ltc.logger.Info("Transaction completed successfully")
	}
	return nil
}

// MetricsTransactionCallback provides metrics for transactions
type MetricsTransactionCallback struct {
	metrics MetricsCollector
}

func NewMetricsTransactionCallback(metrics MetricsCollector) *MetricsTransactionCallback {
	return &MetricsTransactionCallback{metrics: metrics}
}

func (mtc *MetricsTransactionCallback) BeforeTransaction(ctx context.Context) error {
	mtc.metrics.IncrementTransactionCount()
	return nil
}

func (mtc *MetricsTransactionCallback) AfterTransaction(ctx context.Context, err error) error {
	if err != nil {
		mtc.metrics.IncrementTransactionErrorCount()
	} else {
		mtc.metrics.IncrementTransactionSuccessCount()
	}
	return nil
}

// Interfaces for dependencies
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

type MetricsCollector interface {
	IncrementTransactionCount()
	IncrementTransactionSuccessCount()
	IncrementTransactionErrorCount()
}

// TransactionContext provides transaction context information
type TransactionContext struct {
	ID            string
	StartTime     time.Time
	IsolationLevel TransactionIsolationLevel
	ReadOnly      bool
}

// TransactionContextManager manages transaction context
type TransactionContextManager struct {
	contexts map[string]*TransactionContext
	mu       sync.RWMutex
}

func NewTransactionContextManager() *TransactionContextManager {
	return &TransactionContextManager{
		contexts: make(map[string]*TransactionContext),
	}
}

func (tcm *TransactionContextManager) CreateContext(id string, options TransactionOptions) *TransactionContext {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	ctx := &TransactionContext{
		ID:             id,
		StartTime:      time.Now(),
		IsolationLevel: options.IsolationLevel,
		ReadOnly:       options.ReadOnly,
	}

	tcm.contexts[id] = ctx
	return ctx
}

func (tcm *TransactionContextManager) GetContext(id string) (*TransactionContext, bool) {
	tcm.mu.RLock()
	defer tcm.mu.RUnlock()

	ctx, exists := tcm.contexts[id]
	return ctx, exists
}

func (tcm *TransactionContextManager) RemoveContext(id string) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	delete(tcm.contexts, id)
}

// DistributedTransactionManager manages distributed transactions
type DistributedTransactionManager struct {
	managers map[string]*TransactionManager
	mu       sync.RWMutex
}

func NewDistributedTransactionManager() *DistributedTransactionManager {
	return &DistributedTransactionManager{
		managers: make(map[string]*TransactionManager),
	}
}

func (dtm *DistributedTransactionManager) AddManager(name string, manager *TransactionManager) {
	dtm.mu.Lock()
	defer dtm.mu.Unlock()

	dtm.managers[name] = manager
}

func (dtm *DistributedTransactionManager) ExecuteDistributedTransaction(ctx context.Context, operations map[string]func(tx *sql.Tx) error) error {
	// Start transactions on all managers
	transactions := make(map[string]*sql.Tx)
	managers := make(map[string]*TransactionManager)

	dtm.mu.RLock()
	for name, manager := range dtm.managers {
		managers[name] = manager
	}
	dtm.mu.RUnlock()

	// Begin transactions
	for name, manager := range managers {
		tx, err := manager.db.BeginTx(ctx, nil)
		if err != nil {
			// Rollback already started transactions
			for _, existingTx := range transactions {
				existingTx.Rollback()
			}
			return fmt.Errorf("failed to begin transaction on %s: %w", name, err)
		}
		transactions[name] = tx
	}

	// Execute operations
	for name, operation := range operations {
		if tx, exists := transactions[name]; exists {
			if err := operation(tx); err != nil {
				// Rollback all transactions
				for _, tx := range transactions {
					tx.Rollback()
				}
				return fmt.Errorf("operation failed on %s: %w", name, err)
			}
		}
	}

	// Commit all transactions
	for name, tx := range transactions {
		if err := tx.Commit(); err != nil {
			// Rollback remaining transactions
			for remainingName, remainingTx := range transactions {
				if remainingName != name {
					remainingTx.Rollback()
				}
			}
			return fmt.Errorf("failed to commit transaction on %s: %w", name, err)
		}
	}

	return nil
}
