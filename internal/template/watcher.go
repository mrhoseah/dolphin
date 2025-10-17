package template

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// TemplateWatcher watches for template file changes
type TemplateWatcher struct {
	engine *Engine
	logger *zap.Logger

	// File watcher
	watcher *fsnotify.Watcher

	// Control
	stopChan chan struct{}
	doneChan chan struct{}

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewTemplateWatcher creates a new template watcher
func NewTemplateWatcher(engine *Engine, logger *zap.Logger) (*TemplateWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	tw := &TemplateWatcher{
		engine:   engine,
		logger:   logger,
		watcher:  watcher,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	return tw, nil
}

// Watch starts watching for template changes
func (tw *TemplateWatcher) Watch() {
	defer close(tw.doneChan)

	// Add watch directories
	directories := []string{
		tw.engine.config.LayoutsDir,
		tw.engine.config.PartialsDir,
		tw.engine.config.PagesDir,
		tw.engine.config.ComponentsDir,
		tw.engine.config.EmailsDir,
	}

	for _, dir := range directories {
		if err := tw.watcher.Add(dir); err != nil {
			if tw.logger != nil {
				tw.logger.Error("Failed to add watch directory",
					zap.String("dir", dir),
					zap.Error(err))
			}
		}
	}

	// Add subdirectories
	for _, dir := range directories {
		tw.addSubdirectories(dir)
	}

	for {
		select {
		case <-tw.stopChan:
			return
		case event, ok := <-tw.watcher.Events:
			if !ok {
				return
			}
			tw.handleEvent(event)
		case err, ok := <-tw.watcher.Errors:
			if !ok {
				return
			}
			if tw.logger != nil {
				tw.logger.Error("Template watcher error", zap.Error(err))
			}
		}
	}
}

// addSubdirectories adds subdirectories to the watcher
func (tw *TemplateWatcher) addSubdirectories(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			if err := tw.watcher.Add(path); err != nil {
				if tw.logger != nil {
					tw.logger.Debug("Failed to add subdirectory",
						zap.String("dir", path),
						zap.Error(err))
				}
			}
		}

		return nil
	})
}

// handleEvent handles a file system event
func (tw *TemplateWatcher) handleEvent(event fsnotify.Event) {
	// Check if file should be processed
	if !tw.shouldProcessFile(event.Name) {
		return
	}

	if tw.engine.config.EnableLogging && tw.logger != nil {
		tw.logger.Debug("Template file changed",
			zap.String("file", event.Name),
			zap.String("op", event.Op.String()))
	}

	// Reload the specific template
	if err := tw.reloadTemplate(event.Name); err != nil {
		if tw.engine.config.EnableLogging && tw.logger != nil {
			tw.logger.Warn("Failed to reload template",
				zap.String("file", event.Name),
				zap.Error(err))
		}
	}
}

// shouldProcessFile checks if a file should be processed
func (tw *TemplateWatcher) shouldProcessFile(path string) bool {
	// Check file extension
	if !strings.HasSuffix(path, tw.engine.config.Extension) {
		return false
	}

	// Check if file is in a watched directory
	directories := []string{
		tw.engine.config.LayoutsDir,
		tw.engine.config.PartialsDir,
		tw.engine.config.PagesDir,
		tw.engine.config.ComponentsDir,
		tw.engine.config.EmailsDir,
	}

	for _, dir := range directories {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}

	return false
}

// reloadTemplate reloads a specific template
func (tw *TemplateWatcher) reloadTemplate(path string) error {
	// Determine template type from path
	var templateType TemplateType
	if strings.HasPrefix(path, tw.engine.config.LayoutsDir) {
		templateType = TypeLayout
	} else if strings.HasPrefix(path, tw.engine.config.PartialsDir) {
		templateType = TypePartial
	} else if strings.HasPrefix(path, tw.engine.config.PagesDir) {
		templateType = TypePage
	} else if strings.HasPrefix(path, tw.engine.config.ComponentsDir) {
		templateType = TypeComponent
	} else if strings.HasPrefix(path, tw.engine.config.EmailsDir) {
		templateType = TypeEmail
	} else {
		return nil // Skip unknown files
	}

	// Load the template
	tmpl, err := tw.engine.loadTemplate(path, templateType)
	if err != nil {
		return err
	}

	// Update the template in the engine
	tw.engine.mu.Lock()
	tw.engine.templates[tmpl.Name] = tmpl

	// Update type-specific map
	switch templateType {
	case TypeLayout:
		tw.engine.layouts[tmpl.Name] = tmpl
	case TypePartial:
		tw.engine.partials[tmpl.Name] = tmpl
	case TypePage:
		tw.engine.pages[tmpl.Name] = tmpl
	case TypeComponent:
		tw.engine.components[tmpl.Name] = tmpl
	case TypeEmail:
		tw.engine.emails[tmpl.Name] = tmpl
	}
	tw.engine.mu.Unlock()

	if tw.engine.config.EnableLogging && tw.logger != nil {
		tw.logger.Info("Template reloaded",
			zap.String("template", tmpl.Name),
			zap.String("type", templateType.String()),
			zap.String("path", path))
	}

	return nil
}

// Stop stops the template watcher
func (tw *TemplateWatcher) Stop() {
	close(tw.stopChan)
	<-tw.doneChan

	if tw.watcher != nil {
		tw.watcher.Close()
	}
}
