package livereload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// ReloadStrategy represents the reload strategy
type ReloadStrategy int

const (
	StrategyRestart ReloadStrategy = iota
	StrategyRebuild
	StrategyHotReload
)

func (s ReloadStrategy) String() string {
	switch s {
	case StrategyRestart:
		return "restart"
	case StrategyRebuild:
		return "rebuild"
	case StrategyHotReload:
		return "hot_reload"
	default:
		return "unknown"
	}
}

// Config represents live reload configuration
type Config struct {
	// Watch configuration
	WatchPaths     []string `yaml:"watch_paths" json:"watch_paths"`
	IgnorePaths    []string `yaml:"ignore_paths" json:"ignore_paths"`
	FileExtensions []string `yaml:"file_extensions" json:"file_extensions"`

	// Reload strategy
	Strategy ReloadStrategy `yaml:"strategy" json:"strategy"`

	// Build configuration
	BuildCommand string        `yaml:"build_command" json:"build_command"`
	RunCommand   string        `yaml:"run_command" json:"run_command"`
	BuildTimeout time.Duration `yaml:"build_timeout" json:"build_timeout"`
	RestartDelay time.Duration `yaml:"restart_delay" json:"restart_delay"`

	// Hot reload configuration
	EnableHotReload bool     `yaml:"enable_hot_reload" json:"enable_hot_reload"`
	HotReloadPort   int      `yaml:"hot_reload_port" json:"hot_reload_port"`
	HotReloadPaths  []string `yaml:"hot_reload_paths" json:"hot_reload_paths"`

	// Debouncing
	DebounceDelay time.Duration `yaml:"debounce_delay" json:"debounce_delay"`
	MaxDebounce   time.Duration `yaml:"max_debounce" json:"max_debounce"`

	// Logging
	EnableLogging  bool `yaml:"enable_logging" json:"enable_logging"`
	VerboseLogging bool `yaml:"verbose_logging" json:"verbose_logging"`
}

// DefaultConfig returns default live reload configuration
func DefaultConfig() *Config {
	return &Config{
		WatchPaths: []string{
			".",
			"cmd",
			"internal",
			"app",
			"ui",
			"public",
		},
		IgnorePaths: []string{
			".git",
			"node_modules",
			"vendor",
			"*.log",
			"*.tmp",
			".env",
		},
		FileExtensions: []string{
			".go",
			".html",
			".css",
			".js",
			".json",
			".yaml",
			".yml",
		},
		Strategy:        StrategyRestart,
		BuildCommand:    "go build -o bin/app cmd/dolphin/main.go",
		RunCommand:      "./bin/app serve",
		BuildTimeout:    30 * time.Second,
		RestartDelay:    1 * time.Second,
		EnableHotReload: false,
		HotReloadPort:   35729,
		HotReloadPaths:  []string{"/"},
		DebounceDelay:   500 * time.Millisecond,
		MaxDebounce:     5 * time.Second,
		EnableLogging:   true,
		VerboseLogging:  false,
	}
}

// LiveReloadManager manages live reload functionality
type LiveReloadManager struct {
	config *Config
	logger *zap.Logger

	// File watching
	watcher  *fsnotify.Watcher
	watchMap map[string]bool
	watchMu  sync.RWMutex

	// Process management
	process   *exec.Cmd
	processMu sync.RWMutex
	isRunning bool

	// Hot reload
	hotReloadServer *HotReloadServer
	hotReloadMu     sync.RWMutex

	// Debouncing
	debounceTimer *time.Timer
	debounceMu    sync.Mutex

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	// Statistics
	stats *Stats
}

// NewLiveReloadManager creates a new live reload manager
func NewLiveReloadManager(config *Config, logger *zap.Logger) (*LiveReloadManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	lrm := &LiveReloadManager{
		config:   config,
		logger:   logger,
		watcher:  watcher,
		watchMap: make(map[string]bool),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		stats:    NewStats(),
	}

	// Initialize hot reload server if enabled
	if config.EnableHotReload {
		hotReloadServer, err := NewHotReloadServer(config.HotReloadPort, config.HotReloadPaths, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create hot reload server: %w", err)
		}
		lrm.hotReloadServer = hotReloadServer
	}

	return lrm, nil
}

// Start starts the live reload manager
func (lrm *LiveReloadManager) Start() error {
	// Start file watching
	if err := lrm.startWatching(); err != nil {
		return fmt.Errorf("failed to start file watching: %w", err)
	}

	// Start hot reload server if enabled
	if lrm.config.EnableHotReload {
		if err := lrm.startHotReloadServer(); err != nil {
			return fmt.Errorf("failed to start hot reload server: %w", err)
		}
	}

	// Start the main process
	if err := lrm.startProcess(); err != nil {
		return fmt.Errorf("failed to start main process: %w", err)
	}

	// Start the main loop
	go lrm.mainLoop()

	if lrm.config.EnableLogging && lrm.logger != nil {
		lrm.logger.Info("Live reload manager started",
			zap.String("strategy", lrm.config.Strategy.String()),
			zap.Strings("watch_paths", lrm.config.WatchPaths),
			zap.Bool("hot_reload_enabled", lrm.config.EnableHotReload))
	}

	return nil
}

// Stop stops the live reload manager
func (lrm *LiveReloadManager) Stop() error {
	// Cancel context
	lrm.cancel()

	// Stop file watcher
	if lrm.watcher != nil {
		lrm.watcher.Close()
	}

	// Stop hot reload server
	if lrm.hotReloadServer != nil {
		lrm.hotReloadServer.Stop()
	}

	// Stop main process
	lrm.stopProcess()

	// Wait for main loop to finish
	<-lrm.done

	if lrm.config.EnableLogging && lrm.logger != nil {
		lrm.logger.Info("Live reload manager stopped")
	}

	return nil
}

// startWatching starts watching files for changes
func (lrm *LiveReloadManager) startWatching() error {
	lrm.watchMu.Lock()
	defer lrm.watchMu.Unlock()

	// Add watch paths
	for _, path := range lrm.config.WatchPaths {
		if err := lrm.addWatchPath(path); err != nil {
			return fmt.Errorf("failed to watch path %s: %w", path, err)
		}
	}

	return nil
}

// addWatchPath adds a path to watch
func (lrm *LiveReloadManager) addWatchPath(path string) error {
	// Check if path should be ignored
	if lrm.shouldIgnorePath(path) {
		return nil
	}

	// Add to watcher
	if err := lrm.watcher.Add(path); err != nil {
		return err
	}

	// Add to watch map
	lrm.watchMap[path] = true

	// Recursively add subdirectories
	return filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if should be ignored
		if lrm.shouldIgnorePath(subPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Add directory to watcher
		if info.IsDir() && subPath != path {
			if err := lrm.watcher.Add(subPath); err != nil {
				// Log error but continue
				if lrm.config.EnableLogging && lrm.logger != nil {
					lrm.logger.Debug("Failed to watch directory",
						zap.String("path", subPath),
						zap.Error(err))
				}
			} else {
				lrm.watchMap[subPath] = true
			}
		}

		return nil
	})
}

// shouldIgnorePath checks if a path should be ignored
func (lrm *LiveReloadManager) shouldIgnorePath(path string) bool {
	// Check ignore paths
	for _, ignorePath := range lrm.config.IgnorePaths {
		if strings.Contains(path, ignorePath) {
			return true
		}
	}

	// Check file extensions
	if !lrm.isWatchedFile(path) {
		return true
	}

	return false
}

// isWatchedFile checks if a file should be watched based on extension
func (lrm *LiveReloadManager) isWatchedFile(path string) bool {
	ext := filepath.Ext(path)
	for _, watchedExt := range lrm.config.FileExtensions {
		if ext == watchedExt {
			return true
		}
	}
	return false
}

// startHotReloadServer starts the hot reload server
func (lrm *LiveReloadManager) startHotReloadServer() error {
	lrm.hotReloadMu.Lock()
	defer lrm.hotReloadMu.Unlock()

	if lrm.hotReloadServer == nil {
		return fmt.Errorf("hot reload server not initialized")
	}

	return lrm.hotReloadServer.Start()
}

// startProcess starts the main process
func (lrm *LiveReloadManager) startProcess() error {
	lrm.processMu.Lock()
	defer lrm.processMu.Unlock()

	// Build the application
	if err := lrm.buildProcess(); err != nil {
		return fmt.Errorf("failed to build process: %w", err)
	}

	// Start the process
	cmd := exec.CommandContext(lrm.ctx, "sh", "-c", lrm.config.RunCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	lrm.process = cmd
	lrm.isRunning = true

	if lrm.config.EnableLogging && lrm.logger != nil {
		lrm.logger.Info("Process started",
			zap.Int("pid", cmd.Process.Pid),
			zap.String("command", lrm.config.RunCommand))
	}

	return nil
}

// buildProcess builds the application
func (lrm *LiveReloadManager) buildProcess() error {
	ctx, cancel := context.WithTimeout(lrm.ctx, lrm.config.BuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", lrm.config.BuildCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if lrm.config.EnableLogging && lrm.logger != nil {
		lrm.logger.Info("Build completed",
			zap.String("command", lrm.config.BuildCommand))
	}

	return nil
}

// stopProcess stops the main process
func (lrm *LiveReloadManager) stopProcess() {
	lrm.processMu.Lock()
	defer lrm.processMu.Unlock()

	if lrm.process != nil && lrm.isRunning {
		// Send SIGTERM
		if err := lrm.process.Process.Signal(os.Interrupt); err != nil {
			if lrm.config.EnableLogging && lrm.logger != nil {
				lrm.logger.Warn("Failed to send interrupt signal",
					zap.Error(err))
			}
		}

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- lrm.process.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't exit
			lrm.process.Process.Kill()
			if lrm.config.EnableLogging && lrm.logger != nil {
				lrm.logger.Warn("Process force killed")
			}
		case <-done:
			// Process exited normally
		}

		lrm.isRunning = false
		lrm.process = nil
	}
}

// mainLoop is the main event loop
func (lrm *LiveReloadManager) mainLoop() {
	defer close(lrm.done)

	for {
		select {
		case <-lrm.ctx.Done():
			return
		case event, ok := <-lrm.watcher.Events:
			if !ok {
				return
			}
			lrm.handleFileEvent(event)
		case err, ok := <-lrm.watcher.Errors:
			if !ok {
				return
			}
			if lrm.config.EnableLogging && lrm.logger != nil {
				lrm.logger.Error("File watcher error", zap.Error(err))
			}
		}
	}
}

// handleFileEvent handles file system events
func (lrm *LiveReloadManager) handleFileEvent(event fsnotify.Event) {
	// Check if file should be watched
	if !lrm.isWatchedFile(event.Name) {
		return
	}

	// Check if path should be ignored
	if lrm.shouldIgnorePath(event.Name) {
		return
	}

	// Log file change
	if lrm.config.VerboseLogging && lrm.logger != nil {
		lrm.logger.Debug("File changed",
			zap.String("file", event.Name),
			zap.String("op", event.Op.String()))
	}

	// Update statistics
	lrm.stats.RecordFileChange(event.Name, event.Op)

	// Trigger reload with debouncing
	lrm.triggerReload()
}

// triggerReload triggers a reload with debouncing
func (lrm *LiveReloadManager) triggerReload() {
	lrm.debounceMu.Lock()
	defer lrm.debounceMu.Unlock()

	// Cancel existing timer
	if lrm.debounceTimer != nil {
		lrm.debounceTimer.Stop()
	}

	// Create new timer
	lrm.debounceTimer = time.AfterFunc(lrm.config.DebounceDelay, func() {
		lrm.performReload()
	})
}

// performReload performs the actual reload
func (lrm *LiveReloadManager) performReload() {
	if lrm.config.EnableLogging && lrm.logger != nil {
		lrm.logger.Info("Performing reload",
			zap.String("strategy", lrm.config.Strategy.String()))
	}

	// Update statistics
	lrm.stats.RecordReload()

	switch lrm.config.Strategy {
	case StrategyRestart:
		lrm.performRestart()
	case StrategyRebuild:
		lrm.performRebuild()
	case StrategyHotReload:
		lrm.performHotReload()
	}
}

// performRestart performs a restart
func (lrm *LiveReloadManager) performRestart() {
	// Stop current process
	lrm.stopProcess()

	// Wait for restart delay
	time.Sleep(lrm.config.RestartDelay)

	// Start new process
	if err := lrm.startProcess(); err != nil {
		if lrm.config.EnableLogging && lrm.logger != nil {
			lrm.logger.Error("Failed to restart process", zap.Error(err))
		}
	}
}

// performRebuild performs a rebuild
func (lrm *LiveReloadManager) performRebuild() {
	// Build the application
	if err := lrm.buildProcess(); err != nil {
		if lrm.config.EnableLogging && lrm.logger != nil {
			lrm.logger.Error("Failed to rebuild process", zap.Error(err))
		}
		return
	}

	// Restart the process
	lrm.performRestart()
}

// performHotReload performs hot reload
func (lrm *LiveReloadManager) performHotReload() {
	lrm.hotReloadMu.RLock()
	defer lrm.hotReloadMu.RUnlock()

	if lrm.hotReloadServer != nil {
		lrm.hotReloadServer.NotifyReload()
	}
}

// GetStats returns current statistics
func (lrm *LiveReloadManager) GetStats() *Stats {
	return lrm.stats
}

// IsRunning returns true if the process is running
func (lrm *LiveReloadManager) IsRunning() bool {
	lrm.processMu.RLock()
	defer lrm.processMu.RUnlock()
	return lrm.isRunning
}

// GetWatchedPaths returns the currently watched paths
func (lrm *LiveReloadManager) GetWatchedPaths() []string {
	lrm.watchMu.RLock()
	defer lrm.watchMu.RUnlock()

	paths := make([]string, 0, len(lrm.watchMap))
	for path := range lrm.watchMap {
		paths = append(paths, path)
	}
	return paths
}

// ConfigFromEnv creates config from environment variables
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	// Override with environment variables if present
	// This would typically read from environment variables
	// For now, return the default config
	return config
}
