package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"dolphin/internal/images"
)

func main() {
	fmt.Println("🚀 Dolphin Framework - Image Optimization Demo")
	fmt.Println("===============================================")

	// Create processor and optimizer
	processor := images.NewDefaultProcessor()
	optimizer := images.NewImageOptimizer(processor, nil, nil)

	// Demo 1: Basic Image Processing
	fmt.Println("\n=== Demo 1: Basic Image Processing ===")
	demoBasicProcessing(optimizer)

	// Demo 2: Responsive Image Generation
	fmt.Println("\n=== Demo 2: Responsive Image Generation ===")
	demoResponsiveImages(optimizer)

	// Demo 3: Format Conversion
	fmt.Println("\n=== Demo 3: Format Conversion ===")
	demoFormatConversion(optimizer)

	// Demo 4: Batch Processing
	fmt.Println("\n=== Demo 4: Batch Processing ===")
	demoBatchProcessing(optimizer)

	// Demo 5: Fin Template Helpers
	fmt.Println("\n=== Demo 5: Fin Template Helpers ===")
	demoFinHelpers()

	// Demo 6: Image Analysis
	fmt.Println("\n=== Demo 6: Image Analysis ===")
	demoImageAnalysis(processor)

	fmt.Println("\n🎉 Image optimization demo completed!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("1. Use 'dolphin images optimize <file>' to optimize single images")
	fmt.Println("2. Use 'dolphin images batch <directory>' for batch processing")
	fmt.Println("3. Use 'dolphin images responsive <file>' for responsive sets")
	fmt.Println("4. Integrate Fin helpers into your templates")
	fmt.Println("5. Set up image storage and caching for production")
}

func demoBasicProcessing(optimizer images.ImageOptimizer) {
	// Create a sample image (1x1 pixel PNG)
	sampleImage := createSampleImage()

	ctx := context.Background()
	options := &images.OptimizationOptions{
		Width:   800,
		Height:  600,
		Quality: 85,
		Format:  images.FormatJPEG,
	}

	output, metadata, err := optimizer.Optimize(ctx, sampleImage, options)
	if err != nil {
		fmt.Printf("❌ Optimization failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Image optimized successfully!\n")
	fmt.Printf("📏 Size: %dx%d\n", metadata.Size.Width, metadata.Size.Height)
	fmt.Printf("🎨 Format: %s\n", metadata.Format)
	fmt.Printf("⚡ Quality: %d\n", metadata.Quality)

	// Read output to demonstrate it worked
	var buf bytes.Buffer
	io.Copy(&buf, output)
	fmt.Printf("📦 Output size: %d bytes\n", buf.Len())
}

func demoResponsiveImages(optimizer images.ImageOptimizer) {
	sampleImage := createSampleImage()

	ctx := context.Background()
	sizes := []int{320, 640, 1024, 1920}
	options := &images.OptimizationOptions{
		Quality: 85,
		Format:  images.FormatJPEG,
	}

	responsiveSet, err := optimizer.GenerateResponsiveSet(ctx, sampleImage, sizes, options)
	if err != nil {
		fmt.Printf("❌ Responsive generation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Responsive image set generated!\n")
	fmt.Printf("📏 Generated sizes: %v\n", sizes)
	fmt.Printf("🔗 SrcSet entries: %d\n", len(responsiveSet.SrcSet))
	fmt.Printf("📐 CSS sizes: %v\n", responsiveSet.Sizes)
	fmt.Printf("🖼️ Fallback: %s\n", responsiveSet.Fallback)

	// Generate HTML example
	html := generateResponsiveHTML(responsiveSet)
	fmt.Printf("📄 HTML example generated (%d characters)\n", len(html))
}

func demoFormatConversion(optimizer images.ImageOptimizer) {
	sampleImage := createSampleImage()

	ctx := context.Background()

	// Test different formats
	formats := []images.ImageFormat{
		images.FormatJPEG,
		images.FormatPNG,
		images.FormatGIF,
	}

	for _, format := range formats {
		options := &images.OptimizationOptions{
			Width:   400,
			Height:  300,
			Quality: 85,
			Format:  format,
		}

		output, metadata, err := optimizer.Optimize(ctx, sampleImage, options)
		if err != nil {
			fmt.Printf("❌ %s conversion failed: %v\n", format, err)
			continue
		}

		var buf bytes.Buffer
		io.Copy(&buf, output)

		fmt.Printf("✅ %s: %dx%d, %d bytes\n",
			format, metadata.Size.Width, metadata.Size.Height, buf.Len())
	}
}

func demoBatchProcessing(optimizer images.ImageOptimizer) {
	// Create sample images
	sampleImages := []io.Reader{
		createSampleImage(),
		createSampleImage(),
		createSampleImage(),
	}

	ctx := context.Background()
	options := &images.OptimizationOptions{
		Width:   600,
		Height:  400,
		Quality: 80,
		Format:  images.FormatJPEG,
	}

	outputs, metadata, err := optimizer.BatchOptimize(ctx, sampleImages, options)
	if err != nil {
		fmt.Printf("❌ Batch processing failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Batch processing completed!\n")
	fmt.Printf("📦 Processed %d images\n", len(outputs))

	totalSize := 0
	for i, meta := range metadata {
		var buf bytes.Buffer
		io.Copy(&buf, outputs[i])
		totalSize += buf.Len()

		fmt.Printf("  %d: %dx%d, %d bytes\n",
			i+1, meta.Size.Width, meta.Size.Height, buf.Len())
	}

	fmt.Printf("📊 Total output size: %d bytes\n", totalSize)
}

func demoFinHelpers() {
	// Create template helpers
	helpers := images.NewFinTemplateHelpers(nil, nil, nil)

	fmt.Println("🔧 Available Fin template functions:")
	functions := helpers.GetTemplateFunctions()
	for name := range functions {
		fmt.Printf("  • %s\n", name)
	}

	// Demo responsive image component
	component := &images.ResponsiveImageComponent{
		Src:     "/images/hero.jpg",
		Alt:     "Hero image",
		Sizes:   []int{320, 640, 1024},
		Formats: []string{"webp", "avif"},
		Lazy:    true,
		Attributes: map[string]string{
			"class": "hero-image",
		},
	}

	html := component.Render()
	fmt.Printf("📄 Generated HTML (%d characters):\n", len(html))
	fmt.Printf("%s\n", html)
}

func demoImageAnalysis(processor images.ImageProcessor) {
	sampleImage := createSampleImage()

	ctx := context.Background()
	metadata, err := processor.GetMetadata(ctx, sampleImage)
	if err != nil {
		fmt.Printf("❌ Analysis failed: %v\n", err)
		return
	}

	fmt.Printf("📊 Image Analysis:\n")
	fmt.Printf("📏 Dimensions: %dx%d\n", metadata.Size.Width, metadata.Size.Height)
	fmt.Printf("🎨 Format: %s\n", metadata.Format)
	fmt.Printf("⚡ Quality: %d\n", metadata.Quality)

	// Format detection demo
	sampleData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG header
	detectedFormat := images.DetectFormat(sampleData)
	fmt.Printf("🔍 Detected format: %s\n", detectedFormat)

	// Optimization recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	if metadata.Size.Width > 1920 {
		fmt.Printf("• Consider resizing for web (max 1920px width)\n")
	}
	if metadata.Format == images.FormatPNG {
		fmt.Printf("• Consider WebP for better compression\n")
	}
	if metadata.Quality > 90 {
		fmt.Printf("• Consider reducing quality to 80-85 for web\n")
	}
}

// Helper functions

func createSampleImage() io.Reader {
	// Create a minimal 1x1 pixel PNG image
	// This is a simplified version - in production you'd use a proper image library
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 dimensions
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // bit depth, color type, etc.
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, // compressed data
		0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01, // pixel data
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, // IEND chunk
		0xAE, 0x42, 0x60, 0x82,
	}
	return bytes.NewReader(pngData)
}

func generateResponsiveHTML(set *images.ResponsiveImageSet) string {
	var srcset []string
	for width, url := range set.SrcSet {
		srcset = append(srcset, fmt.Sprintf("%s %sw", url, width))
	}

	return fmt.Sprintf(`<img src="%s" 
     srcset="%s" 
     sizes="%s" 
     alt="%s" 
     loading="lazy">`,
		set.Fallback,
		strings.Join(srcset, ", "),
		strings.Join(set.Sizes, ", "),
		set.Alt)
}

// Mock storage implementation for demo
type mockStorage struct{}

func (m *mockStorage) Store(ctx context.Context, key string, data io.Reader) error {
	return nil
}

func (m *mockStorage) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockStorage) Exists(ctx context.Context, key string) bool {
	return false
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorage) URL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("/images/%s", key), nil
}

// Mock cache implementation for demo
type mockCache struct{}

func (m *mockCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockCache) Set(ctx context.Context, key string, data io.Reader, ttl int64) error {
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCache) Clear(ctx context.Context) error {
	return nil
}
