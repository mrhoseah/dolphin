package assets

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// Optimizer represents an asset optimizer
type Optimizer struct {
	config *Config
	logger *zap.Logger
}

// NewOptimizer creates a new asset optimizer
func NewOptimizer(config *Config, logger *zap.Logger) *Optimizer {
	return &Optimizer{
		config: config,
		logger: logger,
	}
}

// OptimizeAsset optimizes a single asset
func (o *Optimizer) OptimizeAsset(asset *Asset) error {
	if !o.config.EnableOptimization {
		return nil
	}

	switch asset.Type {
	case TypeCSS:
		return o.optimizeCSS(asset)
	case TypeJS:
		return o.optimizeJS(asset)
	case TypeImage:
		return o.optimizeImage(asset)
	default:
		return nil // No optimization for other types
	}
}

// optimizeCSS optimizes CSS assets
func (o *Optimizer) optimizeCSS(asset *Asset) error {
	if !o.config.OptimizeCSS {
		return nil
	}

	// Read CSS content
	content, err := o.readFile(asset.Path)
	if err != nil {
		return err
	}

	// Basic CSS minification
	optimized := o.minifyCSS(content)

	// Write optimized content
	outputPath := o.getOptimizedPath(asset)
	if err := o.writeFile(outputPath, optimized); err != nil {
		return err
	}

	// Update asset
	asset.Size = int64(len(optimized))

	if o.config.EnableLogging && o.logger != nil {
		o.logger.Debug("CSS optimized",
			zap.String("file", asset.Path),
			zap.Int64("original_size", int64(len(content))),
			zap.Int64("optimized_size", int64(len(optimized))))
	}

	return nil
}

// optimizeJS optimizes JavaScript assets
func (o *Optimizer) optimizeJS(asset *Asset) error {
	if !o.config.OptimizeJS {
		return nil
	}

	// Read JS content
	content, err := o.readFile(asset.Path)
	if err != nil {
		return err
	}

	// Basic JS minification
	optimized := o.minifyJS(content)

	// Write optimized content
	outputPath := o.getOptimizedPath(asset)
	if err := o.writeFile(outputPath, optimized); err != nil {
		return err
	}

	// Update asset
	asset.Size = int64(len(optimized))

	if o.config.EnableLogging && o.logger != nil {
		o.logger.Debug("JS optimized",
			zap.String("file", asset.Path),
			zap.Int64("original_size", int64(len(content))),
			zap.Int64("optimized_size", int64(len(optimized))))
	}

	return nil
}

// optimizeImage optimizes image assets
func (o *Optimizer) optimizeImage(asset *Asset) error {
	if !o.config.OptimizeImages {
		return nil
	}

	// For now, just copy the image
	// In a real implementation, you would use image optimization libraries
	// like imageoptim, pngquant, jpegoptim, etc.

	inputFile, err := os.Open(asset.Path)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputPath := o.getOptimizedPath(asset)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return err
	}

	// Get optimized file size
	info, err := outputFile.Stat()
	if err != nil {
		return err
	}

	asset.Size = info.Size()

	if o.config.EnableLogging && o.logger != nil {
		o.logger.Debug("Image optimized",
			zap.String("file", asset.Path),
			zap.Int64("optimized_size", asset.Size))
	}

	return nil
}

// minifyCSS performs basic CSS minification
func (o *Optimizer) minifyCSS(content string) string {
	// Remove comments
	content = o.removeCSSComments(content)

	// Remove unnecessary whitespace
	content = o.removeWhitespace(content)

	// Remove unnecessary semicolons
	content = o.removeUnnecessarySemicolons(content)

	return content
}

// minifyJS performs basic JavaScript minification
func (o *Optimizer) minifyJS(content string) string {
	// Remove single-line comments
	content = o.removeSingleLineComments(content)

	// Remove multi-line comments
	content = o.removeMultiLineComments(content)

	// Remove unnecessary whitespace
	content = o.removeWhitespace(content)

	return content
}

// removeCSSComments removes CSS comments
func (o *Optimizer) removeCSSComments(content string) string {
	// Remove /* ... */ comments
	for {
		start := strings.Index(content, "/*")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "*/")
		if end == -1 {
			break
		}

		content = content[:start] + content[start+end+2:]
	}

	return content
}

// removeSingleLineComments removes single-line comments
func (o *Optimizer) removeSingleLineComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Find // comment
		commentIndex := strings.Index(line, "//")
		if commentIndex != -1 {
			// Check if it's inside a string
			beforeComment := line[:commentIndex]
			quoteCount := strings.Count(beforeComment, "\"") - strings.Count(beforeComment, "\\\"")
			if quoteCount%2 == 0 {
				line = line[:commentIndex]
			}
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// removeMultiLineComments removes multi-line comments
func (o *Optimizer) removeMultiLineComments(content string) string {
	// Remove /* ... */ comments
	for {
		start := strings.Index(content, "/*")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "*/")
		if end == -1 {
			break
		}

		content = content[:start] + content[start+end+2:]
	}

	return content
}

// removeWhitespace removes unnecessary whitespace
func (o *Optimizer) removeWhitespace(content string) string {
	// Replace multiple spaces with single space
	content = strings.ReplaceAll(content, "  ", " ")
	content = strings.ReplaceAll(content, "   ", " ")
	content = strings.ReplaceAll(content, "    ", " ")

	// Remove leading/trailing whitespace from lines
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// removeUnnecessarySemicolons removes unnecessary semicolons
func (o *Optimizer) removeUnnecessarySemicolons(content string) string {
	// Remove semicolons before closing braces
	content = strings.ReplaceAll(content, ";\n}", "}")
	content = strings.ReplaceAll(content, ";\n  }", "}")
	content = strings.ReplaceAll(content, ";\n    }", "}")

	return content
}

// readFile reads a file and returns its content
func (o *Optimizer) readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// writeFile writes content to a file
func (o *Optimizer) writeFile(path string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// getOptimizedPath returns the path for the optimized asset
func (o *Optimizer) getOptimizedPath(asset *Asset) string {
	// Get relative path from source directory
	relPath, err := filepath.Rel(o.config.SourceDir, asset.Path)
	if err != nil {
		relPath = asset.Path
	}

	// Add .min to filename
	ext := filepath.Ext(relPath)
	name := strings.TrimSuffix(relPath, ext)
	minifiedName := name + ".min" + ext

	return filepath.Join(o.config.OutputDir, minifiedName)
}

// OptimizeBundle optimizes a bundle
func (o *Optimizer) OptimizeBundle(bundle *Bundle) error {
	if !o.config.EnableOptimization {
		return nil
	}

	// Optimize individual assets
	for _, asset := range bundle.Assets {
		if err := o.OptimizeAsset(asset); err != nil {
			if o.config.EnableLogging && o.logger != nil {
				o.logger.Warn("Failed to optimize asset",
					zap.String("asset", asset.Path),
					zap.Error(err))
			}
		}
	}

	// Optimize combined file if it exists
	if bundle.CombinedPath != "" {
		if err := o.optimizeCombinedFile(bundle); err != nil {
			return err
		}
	}

	return nil
}

// optimizeCombinedFile optimizes a combined file
func (o *Optimizer) optimizeCombinedFile(bundle *Bundle) error {
	// Read combined file
	content, err := o.readFile(bundle.CombinedPath)
	if err != nil {
		return err
	}

	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(bundle.CombinedPath))

	var optimized string
	switch ext {
	case ".css":
		optimized = o.minifyCSS(content)
	case ".js":
		optimized = o.minifyJS(content)
	default:
		optimized = content
	}

	// Write optimized content
	optimizedPath := strings.TrimSuffix(bundle.CombinedPath, ext) + ".min" + ext
	if err := o.writeFile(optimizedPath, optimized); err != nil {
		return err
	}

	// Update bundle
	bundle.CombinedPath = optimizedPath
	bundle.Size = int64(len(optimized))

	if o.config.EnableLogging && o.logger != nil {
		o.logger.Debug("Bundle optimized",
			zap.String("bundle", bundle.Name),
			zap.Int64("original_size", int64(len(content))),
			zap.Int64("optimized_size", int64(len(optimized))))
	}

	return nil
}

// GetOptimizationStats returns optimization statistics
func (o *Optimizer) GetOptimizationStats() map[string]interface{} {
	return map[string]interface{}{
		"css_optimization":     o.config.OptimizeCSS,
		"js_optimization":      o.config.OptimizeJS,
		"image_optimization":   o.config.OptimizeImages,
		"minification":         o.config.MinifyAssets,
		"optimization_enabled": o.config.EnableOptimization,
	}
}
