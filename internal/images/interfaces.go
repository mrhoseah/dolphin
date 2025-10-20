package images

import (
	"context"
	"io"
)

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
	FormatWebP ImageFormat = "webp"
	FormatAVIF ImageFormat = "avif"
	FormatGIF  ImageFormat = "gif"
)

// ImageSize represents image dimensions
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ImageMetadata contains image information
type ImageMetadata struct {
	Format   ImageFormat `json:"format"`
	Size     ImageSize   `json:"size"`
	FileSize int64       `json:"file_size"`
	Quality  int         `json:"quality"`
}

// OptimizationOptions configures image processing
type OptimizationOptions struct {
	Format      ImageFormat `json:"format"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	Quality     int         `json:"quality"`
	Progressive bool        `json:"progressive"`
	Strip       bool        `json:"strip"`
	Lossless    bool        `json:"lossless"`
}

// ResponsiveImageSet represents multiple image sizes
type ResponsiveImageSet struct {
	SrcSet   map[string]string `json:"src_set"`  // width -> url
	Sizes    []string          `json:"sizes"`    // CSS sizes
	Fallback string            `json:"fallback"` // Default image
	Alt      string            `json:"alt"`
	Lazy     bool              `json:"lazy"`
}

// ImageProcessor processes images
type ImageProcessor interface {
	// Process processes an image with given options
	Process(ctx context.Context, input io.Reader, options *OptimizationOptions) (io.Reader, *ImageMetadata, error)

	// GetMetadata extracts metadata from image
	GetMetadata(ctx context.Context, input io.Reader) (*ImageMetadata, error)

	// Supports checks if format is supported
	Supports(format ImageFormat) bool
}

// ImageOptimizer provides high-level optimization features
type ImageOptimizer interface {
	// Optimize optimizes a single image
	Optimize(ctx context.Context, input io.Reader, options *OptimizationOptions) (io.Reader, *ImageMetadata, error)

	// GenerateResponsiveSet creates multiple sizes for responsive images
	GenerateResponsiveSet(ctx context.Context, input io.Reader, sizes []int, options *OptimizationOptions) (*ResponsiveImageSet, error)

	// BatchOptimize processes multiple images
	BatchOptimize(ctx context.Context, inputs []io.Reader, options *OptimizationOptions) ([]io.Reader, []*ImageMetadata, error)
}

// ImageStorage handles image storage and retrieval
type ImageStorage interface {
	// Store stores an optimized image
	Store(ctx context.Context, key string, data io.Reader) error

	// Retrieve retrieves a stored image
	Retrieve(ctx context.Context, key string) (io.ReadCloser, error)

	// Exists checks if image exists
	Exists(ctx context.Context, key string) bool

	// Delete removes an image
	Delete(ctx context.Context, key string) error

	// URL generates a public URL for the image
	URL(ctx context.Context, key string) (string, error)
}

// ImageCache provides caching for optimized images
type ImageCache interface {
	// Get retrieves cached image
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Set stores image in cache
	Set(ctx context.Context, key string, data io.Reader, ttl int64) error

	// Delete removes from cache
	Delete(ctx context.Context, key string) error

	// Clear clears all cached images
	Clear(ctx context.Context) error
}
