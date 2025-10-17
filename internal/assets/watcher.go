package assets

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path string
	Op   fsnotify.Op
	Time time.Time
}

// AssetWatcher watches for file changes
type AssetWatcher struct {
	watchDir   string
	extensions []string
	logger     *zap.Logger

	// File watcher
	watcher *fsnotify.Watcher

	// Event channel
	events chan FileEvent

	// Control
	stopChan chan struct{}
	doneChan chan struct{}

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewAssetWatcher creates a new asset watcher
func NewAssetWatcher(watchDir string, extensions []string, logger *zap.Logger) (*AssetWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	aw := &AssetWatcher{
		watchDir:   watchDir,
		extensions: extensions,
		logger:     logger,
		watcher:    watcher,
		events:     make(chan FileEvent, 100),
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
	}

	// Start watching
	go aw.watch()

	return aw, nil
}

// watch starts watching for file changes
func (aw *AssetWatcher) watch() {
	defer close(aw.doneChan)
	defer close(aw.events)

	// Add watch directory
	if err := aw.watcher.Add(aw.watchDir); err != nil {
		if aw.logger != nil {
			aw.logger.Error("Failed to add watch directory",
				zap.String("dir", aw.watchDir),
				zap.Error(err))
		}
		return
	}

	// Add subdirectories
	aw.addSubdirectories(aw.watchDir)

	for {
		select {
		case <-aw.stopChan:
			return
		case event, ok := <-aw.watcher.Events:
			if !ok {
				return
			}
			aw.handleEvent(event)
		case err, ok := <-aw.watcher.Errors:
			if !ok {
				return
			}
			if aw.logger != nil {
				aw.logger.Error("File watcher error", zap.Error(err))
			}
		}
	}
}

// addSubdirectories adds subdirectories to the watcher
func (aw *AssetWatcher) addSubdirectories(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			if err := aw.watcher.Add(path); err != nil {
				if aw.logger != nil {
					aw.logger.Debug("Failed to add subdirectory",
						zap.String("dir", path),
						zap.Error(err))
				}
			}
		}

		return nil
	})
}

// handleEvent handles a file system event
func (aw *AssetWatcher) handleEvent(event fsnotify.Event) {
	// Check if file extension is watched
	if !aw.isWatchedFile(event.Name) {
		return
	}

	// Create file event
	fileEvent := FileEvent{
		Path: event.Name,
		Op:   event.Op,
		Time: time.Now(),
	}

	// Send event to channel
	select {
	case aw.events <- fileEvent:
	default:
		// Channel is full, skip event
		if aw.logger != nil {
			aw.logger.Warn("Event channel full, skipping event",
				zap.String("file", event.Name))
		}
	}
}

// isWatchedFile checks if a file should be watched
func (aw *AssetWatcher) isWatchedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	for _, watchExt := range aw.extensions {
		if ext == watchExt {
			return true
		}
	}

	return false
}

// Events returns the event channel
func (aw *AssetWatcher) Events() <-chan FileEvent {
	return aw.events
}

// Stop stops the watcher
func (aw *AssetWatcher) Stop() {
	close(aw.stopChan)
	<-aw.doneChan

	if aw.watcher != nil {
		aw.watcher.Close()
	}
}
