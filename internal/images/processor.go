package images

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

// DefaultProcessor provides basic image processing using Go's standard library
type DefaultProcessor struct {
	supportedFormats map[ImageFormat]bool
}

// NewDefaultProcessor creates a new default image processor
func NewDefaultProcessor() *DefaultProcessor {
	return &DefaultProcessor{
		supportedFormats: map[ImageFormat]bool{
			FormatJPEG: true,
			FormatPNG:  true,
			FormatGIF:  true,
			FormatWebP: false, // Would need external library
			FormatAVIF: false, // Would need external library
		},
	}
}

// Process processes an image with given options
func (p *DefaultProcessor) Process(ctx context.Context, input io.Reader, options *OptimizationOptions) (io.Reader, *ImageMetadata, error) {
	// Decode image
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original metadata (for future use)
	_ = ImageSize{
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}

	// Resize if needed
	if options.Width > 0 || options.Height > 0 {
		img = p.resizeImage(img, options.Width, options.Height)
	}

	// Encode with new options
	var buf bytes.Buffer
	outputFormat := options.Format
	if outputFormat == "" {
		outputFormat = ImageFormat(format)
	}

	err = p.encodeImage(&buf, img, outputFormat, options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode image: %w", err)
	}

	metadata := &ImageMetadata{
		Format:  outputFormat,
		Size:    ImageSize{Width: img.Bounds().Dx(), Height: img.Bounds().Dy()},
		Quality: options.Quality,
	}

	return &buf, metadata, nil
}

// GetMetadata extracts metadata from image
func (p *DefaultProcessor) GetMetadata(ctx context.Context, input io.Reader) (*ImageMetadata, error) {
	// Read the image to get metadata
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &ImageMetadata{
		Format: ImageFormat(format),
		Size: ImageSize{
			Width:  img.Bounds().Dx(),
			Height: img.Bounds().Dy(),
		},
	}, nil
}

// Supports checks if format is supported
func (p *DefaultProcessor) Supports(format ImageFormat) bool {
	return p.supportedFormats[format]
}

// resizeImage resizes an image maintaining aspect ratio
func (p *DefaultProcessor) resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if targetWidth > 0 && targetHeight > 0 {
		// Both dimensions specified - crop to fit
		aspectRatio := float64(originalWidth) / float64(originalHeight)
		targetAspectRatio := float64(targetWidth) / float64(targetHeight)

		if aspectRatio > targetAspectRatio {
			// Image is wider - fit height
			newHeight = targetHeight
			newWidth = int(float64(targetHeight) * aspectRatio)
		} else {
			// Image is taller - fit width
			newWidth = targetWidth
			newHeight = int(float64(targetWidth) / aspectRatio)
		}
	} else if targetWidth > 0 {
		// Only width specified
		newWidth = targetWidth
		newHeight = int(float64(targetWidth) * float64(originalHeight) / float64(originalWidth))
	} else if targetHeight > 0 {
		// Only height specified
		newHeight = targetHeight
		newWidth = int(float64(targetHeight) * float64(originalWidth) / float64(originalHeight))
	} else {
		// No resize needed
		return img
	}

	// Simple nearest neighbor resize (in production, use better algorithms)
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * originalWidth / newWidth
			srcY := y * originalHeight / newHeight
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	return resized
}

// encodeImage encodes image with specified format and options
func (p *DefaultProcessor) encodeImage(w io.Writer, img image.Image, format ImageFormat, options *OptimizationOptions) error {
	switch format {
	case FormatJPEG:
		quality := options.Quality
		if quality == 0 {
			quality = 85 // Default quality
		}

		opts := &jpeg.Options{Quality: quality}
		return jpeg.Encode(w, img, opts)

	case FormatPNG:
		return png.Encode(w, img)

	case FormatGIF:
		// For GIF, we'd need to handle animated GIFs differently
		// This is a simplified version
		return gif.Encode(w, img, nil)

	case FormatWebP:
		return fmt.Errorf("WebP encoding not supported in default processor")

	case FormatAVIF:
		return fmt.Errorf("AVIF encoding not supported in default processor")

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ImageOptimizerImpl provides high-level optimization features
type ImageOptimizerImpl struct {
	processor ImageProcessor
	storage   ImageStorage
	cache     ImageCache
}

// NewImageOptimizer creates a new image optimizer
func NewImageOptimizer(processor ImageProcessor, storage ImageStorage, cache ImageCache) *ImageOptimizerImpl {
	return &ImageOptimizerImpl{
		processor: processor,
		storage:   storage,
		cache:     cache,
	}
}

// Optimize optimizes a single image
func (o *ImageOptimizerImpl) Optimize(ctx context.Context, input io.Reader, options *OptimizationOptions) (io.Reader, *ImageMetadata, error) {
	return o.processor.Process(ctx, input, options)
}

// GenerateResponsiveSet creates multiple sizes for responsive images
func (o *ImageOptimizerImpl) GenerateResponsiveSet(ctx context.Context, input io.Reader, sizes []int, options *OptimizationOptions) (*ResponsiveImageSet, error) {
	srcSet := make(map[string]string)
	var cssSizes []string

	for _, width := range sizes {
		// Create options for this size
		sizeOptions := *options
		sizeOptions.Width = width
		sizeOptions.Height = 0 // Maintain aspect ratio

		// Process image
		output, metadata, err := o.processor.Process(ctx, input, &sizeOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to process image for width %d: %w", width, err)
		}

		// Generate key for storage
		key := fmt.Sprintf("responsive_%dx%d_%s", width, metadata.Size.Height, metadata.Format)

		// Store optimized image
		if o.storage != nil {
			err = o.storage.Store(ctx, key, output)
			if err != nil {
				return nil, fmt.Errorf("failed to store image: %w", err)
			}

			// Get URL
			url, err := o.storage.URL(ctx, key)
			if err != nil {
				return nil, fmt.Errorf("failed to get URL: %w", err)
			}

			srcSet[fmt.Sprintf("%d", width)] = url
		}

		// Add CSS size descriptor
		cssSizes = append(cssSizes, fmt.Sprintf("(max-width: %dpx) %dpx", width, width))
	}

	return &ResponsiveImageSet{
		SrcSet:   srcSet,
		Sizes:    cssSizes,
		Fallback: srcSet[fmt.Sprintf("%d", sizes[len(sizes)-1])], // Use largest as fallback
		Alt:      "Responsive image",
		Lazy:     true,
	}, nil
}

// BatchOptimize processes multiple images
func (o *ImageOptimizerImpl) BatchOptimize(ctx context.Context, inputs []io.Reader, options *OptimizationOptions) ([]io.Reader, []*ImageMetadata, error) {
	var outputs []io.Reader
	var metadata []*ImageMetadata

	for i, input := range inputs {
		output, meta, err := o.processor.Process(ctx, input, options)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process image %d: %w", i, err)
		}

		outputs = append(outputs, output)
		metadata = append(metadata, meta)
	}

	return outputs, metadata, nil
}

// Utility functions

// DetectFormat detects image format from content
func DetectFormat(data []byte) ImageFormat {
	if len(data) < 4 {
		return ""
	}

	// Check magic bytes
	if data[0] == 0xFF && data[1] == 0xD8 {
		return FormatJPEG
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return FormatPNG
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return FormatGIF
	}
	if len(data) >= 12 && string(data[8:12]) == "WEBP" {
		return FormatWebP
	}
	if len(data) >= 12 && string(data[4:8]) == "ftyp" && string(data[8:12]) == "avif" {
		return FormatAVIF
	}

	return ""
}

// GetOptimalFormat returns the best format for given requirements
func GetOptimalFormat(originalFormat ImageFormat, supportsWebP, supportsAVIF bool) ImageFormat {
	if supportsAVIF {
		return FormatAVIF
	}
	if supportsWebP {
		return FormatWebP
	}
	return originalFormat
}

// CalculateFileSize estimates file size for given options
func CalculateFileSize(originalSize int64, format ImageFormat, quality int) int64 {
	// Rough estimates based on format and quality
	var compressionRatio float64

	switch format {
	case FormatAVIF:
		compressionRatio = 0.3
	case FormatWebP:
		compressionRatio = 0.4
	case FormatJPEG:
		compressionRatio = float64(quality) / 100.0
	case FormatPNG:
		compressionRatio = 0.8
	case FormatGIF:
		compressionRatio = 0.9
	default:
		compressionRatio = 1.0
	}

	return int64(float64(originalSize) * compressionRatio)
}
