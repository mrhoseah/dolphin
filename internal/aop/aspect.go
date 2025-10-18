package aop

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Aspect defines the interface for cross-cutting concerns
type Aspect interface {
	Before(ctx context.Context, method reflect.Method, args []reflect.Value) error
	After(ctx context.Context, method reflect.Method, args []reflect.Value, result []reflect.Value) error
	OnError(ctx context.Context, method reflect.Method, args []reflect.Value, err error) error
}

// LoggingAspect provides automatic logging for method calls
type LoggingAspect struct {
	logger Logger
}

func NewLoggingAspect(logger Logger) *LoggingAspect {
	return &LoggingAspect{logger: logger}
}

func (a *LoggingAspect) Before(ctx context.Context, method reflect.Method, args []reflect.Value) error {
	a.logger.Info("Calling method", "method", method.Name, "args", args)
	return nil
}

func (a *LoggingAspect) After(ctx context.Context, method reflect.Method, args []reflect.Value, result []reflect.Value) error {
	a.logger.Info("Method completed", "method", method.Name, "result", result)
	return nil
}

func (a *LoggingAspect) OnError(ctx context.Context, method reflect.Method, args []reflect.Value, err error) error {
	a.logger.Error("Method failed", "method", method.Name, "error", err)
	return err
}

// CachingAspect provides automatic caching for method results
type CachingAspect struct {
	cache Cache
}

func NewCachingAspect(cache Cache) *CachingAspect {
	return &CachingAspect{cache: cache}
}

func (a *CachingAspect) Before(ctx context.Context, method reflect.Method, args []reflect.Value) error {
	cacheKey := a.generateCacheKey(method.Name, args)

	if cached, exists := a.cache.Get(cacheKey); exists {
		// Store cached result in context for After method to use
		ctx = context.WithValue(ctx, "cached_result", cached)
		return nil
	}

	return nil
}

func (a *CachingAspect) After(ctx context.Context, method reflect.Method, args []reflect.Value, result []reflect.Value) error {
	// Check if we already have a cached result
	if cached := ctx.Value("cached_result"); cached != nil {
		// Replace result with cached value
		cachedValue := reflect.ValueOf(cached)
		if len(result) > 0 {
			result[0] = cachedValue
		}
		return nil
	}

	cacheKey := a.generateCacheKey(method.Name, args)
	a.cache.Set(cacheKey, result, 5*time.Minute)
	return nil
}

func (a *CachingAspect) OnError(ctx context.Context, method reflect.Method, args []reflect.Value, err error) error {
	return err
}

func (a *CachingAspect) generateCacheKey(methodName string, args []reflect.Value) string {
	// Generate cache key based on method name and arguments
	return fmt.Sprintf("%s:%v", methodName, args)
}

// TransactionAspect provides automatic transaction management
type TransactionAspect struct {
	db *sql.DB
}

func NewTransactionAspect(db *sql.DB) *TransactionAspect {
	return &TransactionAspect{db: db}
}

func (a *TransactionAspect) Before(ctx context.Context, method reflect.Method, args []reflect.Value) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Store transaction in context
	ctx = context.WithValue(ctx, "tx", tx)
	return nil
}

func (a *TransactionAspect) After(ctx context.Context, method reflect.Method, args []reflect.Value, result []reflect.Value) error {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		return tx.Commit()
	}
	return nil
}

func (a *TransactionAspect) OnError(ctx context.Context, method reflect.Method, args []reflect.Value, err error) error {
	if tx, ok := ctx.Value("tx").(*sql.Tx); ok {
		tx.Rollback()
	}
	return err
}

// PerformanceAspect provides automatic performance monitoring
type PerformanceAspect struct {
	metrics MetricsCollector
}

func NewPerformanceAspect(metrics MetricsCollector) *PerformanceAspect {
	return &PerformanceAspect{metrics: metrics}
}

func (a *PerformanceAspect) Before(ctx context.Context, method reflect.Method, args []reflect.Value) error {
	// Store start time in context
	ctx = context.WithValue(ctx, "start_time", time.Now())
	return nil
}

func (a *PerformanceAspect) After(ctx context.Context, method reflect.Method, args []reflect.Value, result []reflect.Value) error {
	if startTime, ok := ctx.Value("start_time").(time.Time); ok {
		duration := time.Since(startTime)
		a.metrics.RecordMethodDuration(method.Name, duration)
	}
	return nil
}

func (a *PerformanceAspect) OnError(ctx context.Context, method reflect.Method, args []reflect.Value, err error) error {
	if startTime, ok := ctx.Value("start_time").(time.Time); ok {
		duration := time.Since(startTime)
		a.metrics.RecordMethodError(method.Name, duration, err)
	}
	return err
}

// AspectWeaver applies aspects to methods
type AspectWeaver struct {
	aspects []Aspect
	mu      sync.RWMutex
}

func NewAspectWeaver(aspects ...Aspect) *AspectWeaver {
	return &AspectWeaver{aspects: aspects}
}

func (w *AspectWeaver) AddAspect(aspect Aspect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.aspects = append(w.aspects, aspect)
}

func (w *AspectWeaver) Weave(target interface{}) interface{} {
	targetType := reflect.TypeOf(target)
	targetValue := reflect.ValueOf(target)

	// Create proxy that intercepts method calls
	proxy := reflect.New(targetType)

	for i := 0; i < targetType.NumMethod(); i++ {
		method := targetType.Method(i)
		originalMethod := targetValue.Method(i)

		// Create wrapped method
		wrappedMethod := func(args []reflect.Value) []reflect.Value {
			ctx := context.Background()

			// Execute Before aspects
			for _, aspect := range w.aspects {
				if err := aspect.Before(ctx, method, args); err != nil {
					return []reflect.Value{reflect.ValueOf(err)}
				}
			}

			// Execute original method
			result := originalMethod.Call(args)

			// Check for errors
			if len(result) > 0 && result[len(result)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if err := result[len(result)-1].Interface().(error); err != nil {
					// Execute OnError aspects
					for _, aspect := range w.aspects {
						aspect.OnError(ctx, method, args, err)
					}
					return result
				}
			}

			// Execute After aspects
			for _, aspect := range w.aspects {
				aspect.After(ctx, method, args, result)
			}

			return result
		}

		// Set wrapped method
		proxy.Elem().Field(i).Set(reflect.MakeFunc(method.Type, wrappedMethod))
	}

	return proxy.Interface()
}

// Interfaces for dependencies
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
}

type MetricsCollector interface {
	RecordMethodDuration(methodName string, duration time.Duration)
	RecordMethodError(methodName string, duration time.Duration, err error)
}

// AspectRegistry manages aspect registration and application
type AspectRegistry struct {
	aspects map[string][]Aspect
	mu      sync.RWMutex
}

func NewAspectRegistry() *AspectRegistry {
	return &AspectRegistry{
		aspects: make(map[string][]Aspect),
	}
}

func (r *AspectRegistry) RegisterAspect(targetType reflect.Type, aspect Aspect) {
	r.mu.Lock()
	defer r.mu.Unlock()

	typeName := targetType.String()
	r.aspects[typeName] = append(r.aspects[typeName], aspect)
}

func (r *AspectRegistry) GetAspects(targetType reflect.Type) []Aspect {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeName := targetType.String()
	return r.aspects[typeName]
}

func (r *AspectRegistry) ApplyAspects(target interface{}) interface{} {
	targetType := reflect.TypeOf(target)
	aspects := r.GetAspects(targetType)

	if len(aspects) == 0 {
		return target
	}

	weaver := NewAspectWeaver(aspects...)
	return weaver.Weave(target)
}
