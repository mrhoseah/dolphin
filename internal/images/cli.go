package images

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// CLIOptimizer provides CLI commands for image optimization
type CLIOptimizer struct {
	optimizer ImageOptimizer
	processor ImageProcessor
}

// NewCLIOptimizer creates a new CLI optimizer
func NewCLIOptimizer(optimizer ImageOptimizer, processor ImageProcessor) *CLIOptimizer {
	return &CLIOptimizer{
		optimizer: optimizer,
		processor: processor,
	}
}

// GetCommands returns CLI commands for image optimization
func (cli *CLIOptimizer) GetCommands() []*cobra.Command {
	return []*cobra.Command{
		cli.optimizeCommand(),
		cli.batchCommand(),
		cli.responsiveCommand(),
		cli.analyzeCommand(),
		cli.convertCommand(),
	}
}

// optimizeCommand creates the optimize command
func (cli *CLIOptimizer) optimizeCommand() *cobra.Command {
	var (
		width       int
		height      int
		quality     int
		format      string
		output      string
		progressive bool
		strip       bool
	)

	cmd := &cobra.Command{
		Use:   "optimize <input>",
		Short: "Optimize a single image",
		Long:  "Optimize a single image with specified parameters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Open input file
			inputFile, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer inputFile.Close()

			// Create options
			options := &OptimizationOptions{
				Width:       width,
				Height:      height,
				Quality:     quality,
				Progressive: progressive,
				Strip:       strip,
			}

			// Set format if specified
			if format != "" {
				options.Format = ImageFormat(format)
			}

			// Optimize image
			ctx := context.Background()
			outputReader, metadata, err := cli.optimizer.Optimize(ctx, inputFile, options)
			if err != nil {
				return fmt.Errorf("failed to optimize image: %w", err)
			}

			// Determine output path
			outputPath := output
			if outputPath == "" {
				ext := filepath.Ext(inputPath)
				base := strings.TrimSuffix(inputPath, ext)
				outputPath = fmt.Sprintf("%s_optimized%s", base, ext)
			}

			// Write output file
			outputFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, outputReader)
			if err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			// Print results
			fmt.Printf("✅ Image optimized successfully!\n")
			fmt.Printf("📁 Output: %s\n", outputPath)
			fmt.Printf("📏 Size: %dx%d\n", metadata.Size.Width, metadata.Size.Height)
			fmt.Printf("🎨 Format: %s\n", metadata.Format)
			fmt.Printf("⚡ Quality: %d\n", metadata.Quality)

			return nil
		},
	}

	cmd.Flags().IntVarP(&width, "width", "w", 0, "Target width (0 = maintain aspect ratio)")
	cmd.Flags().IntVarP(&height, "height", "h", 0, "Target height (0 = maintain aspect ratio)")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Image quality (1-100)")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Output format (jpeg, png, webp, avif)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().BoolVar(&progressive, "progressive", false, "Use progressive encoding")
	cmd.Flags().BoolVar(&strip, "strip", false, "Strip metadata")

	return cmd
}

// batchCommand creates the batch optimize command
func (cli *CLIOptimizer) batchCommand() *cobra.Command {
	var (
		inputDir    string
		outputDir   string
		width       int
		height      int
		quality     int
		format      string
		recursive   bool
		progressive bool
		strip       bool
	)

	cmd := &cobra.Command{
		Use:   "batch <input-directory>",
		Short: "Batch optimize images in a directory",
		Long:  "Optimize all images in a directory with specified parameters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputDir = args[0]

			// Find image files
			var imageFiles []string
			err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !recursive && path != inputDir && filepath.Dir(path) != inputDir {
					return nil
				}

				if info.IsDir() {
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				if isImageFile(ext) {
					imageFiles = append(imageFiles, path)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to scan directory: %w", err)
			}

			if len(imageFiles) == 0 {
				fmt.Println("No image files found in directory")
				return nil
			}

			// Create output directory
			if outputDir == "" {
				outputDir = filepath.Join(inputDir, "optimized")
			}

			err = os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Process images
			ctx := context.Background()
			options := &OptimizationOptions{
				Width:       width,
				Height:      height,
				Quality:     quality,
				Progressive: progressive,
				Strip:       strip,
			}

			if format != "" {
				options.Format = ImageFormat(format)
			}

			fmt.Printf("🔄 Processing %d images...\n", len(imageFiles))

			successCount := 0
			for i, imagePath := range imageFiles {
				fmt.Printf("Processing %d/%d: %s\n", i+1, len(imageFiles), filepath.Base(imagePath))

				// Open input file
				inputFile, err := os.Open(imagePath)
				if err != nil {
					fmt.Printf("❌ Failed to open %s: %v\n", imagePath, err)
					continue
				}

				// Optimize image
				outputReader, _, err := cli.optimizer.Optimize(ctx, inputFile, options)
				inputFile.Close()

				if err != nil {
					fmt.Printf("❌ Failed to optimize %s: %v\n", imagePath, err)
					continue
				}

				// Generate output path
				relPath, _ := filepath.Rel(inputDir, imagePath)
				outputPath := filepath.Join(outputDir, relPath)

				// Create output directory if needed
				err = os.MkdirAll(filepath.Dir(outputPath), 0755)
				if err != nil {
					fmt.Printf("❌ Failed to create output directory for %s: %v\n", imagePath, err)
					continue
				}

				// Write output file
				outputFile, err := os.Create(outputPath)
				if err != nil {
					fmt.Printf("❌ Failed to create output file for %s: %v\n", imagePath, err)
					continue
				}

				_, err = io.Copy(outputFile, outputReader)
				outputFile.Close()

				if err != nil {
					fmt.Printf("❌ Failed to write output file for %s: %v\n", imagePath, err)
					continue
				}

				successCount++
				fmt.Printf("✅ Optimized: %s\n", relPath)
			}

			fmt.Printf("\n🎉 Batch optimization complete!\n")
			fmt.Printf("✅ Successfully processed: %d/%d images\n", successCount, len(imageFiles))
			fmt.Printf("📁 Output directory: %s\n", outputDir)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: input/optimized)")
	cmd.Flags().IntVarP(&width, "width", "w", 0, "Target width")
	cmd.Flags().IntVarP(&height, "height", "h", 0, "Target height")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Image quality")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Output format")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Process subdirectories recursively")
	cmd.Flags().BoolVar(&progressive, "progressive", false, "Use progressive encoding")
	cmd.Flags().BoolVar(&strip, "strip", false, "Strip metadata")

	return cmd
}

// responsiveCommand creates the responsive images command
func (cli *CLIOptimizer) responsiveCommand() *cobra.Command {
	var (
		inputPath string
		outputDir string
		sizes     []int
		quality   int
		format    string
	)

	cmd := &cobra.Command{
		Use:   "responsive <input>",
		Short: "Generate responsive image set",
		Long:  "Generate multiple sizes of an image for responsive design",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath = args[0]

			// Open input file
			inputFile, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer inputFile.Close()

			// Create output directory
			if outputDir == "" {
				base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
				outputDir = filepath.Join(filepath.Dir(inputPath), base+"_responsive")
			}

			err = os.MkdirAll(outputDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Create options
			options := &OptimizationOptions{
				Quality: quality,
			}

			if format != "" {
				options.Format = ImageFormat(format)
			}

			// Generate responsive set
			ctx := context.Background()
			responsiveSet, err := cli.optimizer.GenerateResponsiveSet(ctx, inputFile, sizes, options)
			if err != nil {
				return fmt.Errorf("failed to generate responsive set: %w", err)
			}

			// Generate HTML example
			htmlExample := generateResponsiveHTML(responsiveSet)

			// Write HTML example
			htmlPath := filepath.Join(outputDir, "example.html")
			err = os.WriteFile(htmlPath, []byte(htmlExample), 0644)
			if err != nil {
				return fmt.Errorf("failed to write HTML example: %w", err)
			}

			fmt.Printf("✅ Responsive image set generated!\n")
			fmt.Printf("📁 Output directory: %s\n", outputDir)
			fmt.Printf("📏 Generated sizes: %v\n", sizes)
			fmt.Printf("📄 HTML example: %s\n", htmlPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	cmd.Flags().IntSliceVarP(&sizes, "sizes", "s", []int{320, 640, 1024, 1920}, "Image sizes to generate")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Image quality")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Output format")

	return cmd
}

// analyzeCommand creates the analyze command
func (cli *CLIOptimizer) analyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <input>",
		Short: "Analyze image metadata and optimization potential",
		Long:  "Analyze an image to show metadata and optimization recommendations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Open input file
			inputFile, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer inputFile.Close()

			// Get file info
			fileInfo, err := inputFile.Stat()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}

			// Get metadata
			ctx := context.Background()
			metadata, err := cli.processor.GetMetadata(ctx, inputFile)
			if err != nil {
				return fmt.Errorf("failed to get metadata: %w", err)
			}

			// Print analysis
			fmt.Printf("📊 Image Analysis: %s\n", filepath.Base(inputPath))
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📏 Dimensions: %dx%d\n", metadata.Size.Width, metadata.Size.Height)
			fmt.Printf("🎨 Format: %s\n", metadata.Format)
			fmt.Printf("💾 File size: %s\n", formatFileSize(fileInfo.Size()))
			fmt.Printf("⚡ Quality: %d\n", metadata.Quality)

			// Optimization recommendations
			fmt.Printf("\n💡 Optimization Recommendations:\n")

			if metadata.Size.Width > 1920 || metadata.Size.Height > 1080 {
				fmt.Printf("• Consider resizing for web (max 1920x1080)\n")
			}

			if metadata.Format == FormatPNG && fileInfo.Size() > 100*1024 {
				fmt.Printf("• Consider converting to WebP for better compression\n")
			}

			if metadata.Quality > 90 {
				fmt.Printf("• Consider reducing quality to 80-85 for web\n")
			}

			// Calculate potential savings
			webpSize := CalculateFileSize(fileInfo.Size(), FormatWebP, 85)
			avifSize := CalculateFileSize(fileInfo.Size(), FormatAVIF, 80)

			fmt.Printf("\n📈 Potential Savings:\n")
			fmt.Printf("• WebP (85%): %s (%.1f%% reduction)\n",
				formatFileSize(webpSize),
				float64(fileInfo.Size()-webpSize)/float64(fileInfo.Size())*100)
			fmt.Printf("• AVIF (80%): %s (%.1f%% reduction)\n",
				formatFileSize(avifSize),
				float64(fileInfo.Size()-avifSize)/float64(fileInfo.Size())*100)

			return nil
		},
	}

	return cmd
}

// convertCommand creates the convert command
func (cli *CLIOptimizer) convertCommand() *cobra.Command {
	var (
		inputPath  string
		outputPath string
		format     string
		quality    int
	)

	cmd := &cobra.Command{
		Use:   "convert <input> <output>",
		Short: "Convert image to different format",
		Long:  "Convert an image from one format to another",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath = args[0]
			outputPath = args[1]

			// Open input file
			inputFile, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer inputFile.Close()

			// Create options
			options := &OptimizationOptions{
				Format:  ImageFormat(format),
				Quality: quality,
			}

			// Convert image
			ctx := context.Background()
			outputReader, metadata, err := cli.optimizer.Optimize(ctx, inputFile, options)
			if err != nil {
				return fmt.Errorf("failed to convert image: %w", err)
			}

			// Write output file
			outputFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, outputReader)
			if err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			fmt.Printf("✅ Image converted successfully!\n")
			fmt.Printf("📁 Output: %s\n", outputPath)
			fmt.Printf("🎨 Format: %s\n", metadata.Format)
			fmt.Printf("📏 Size: %dx%d\n", metadata.Size.Width, metadata.Size.Height)

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "webp", "Target format")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "Image quality")

	return cmd
}

// Helper functions

// isImageFile checks if file extension is a supported image format
func isImageFile(ext string) bool {
	supportedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif"}
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// generateResponsiveHTML generates HTML example for responsive images
func generateResponsiveHTML(set *ResponsiveImageSet) string {
	var srcset []string
	for width, url := range set.SrcSet {
		srcset = append(srcset, fmt.Sprintf("%s %sw", url, width))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Responsive Image Example</title>
</head>
<body>
    <h1>Responsive Image Example</h1>
    
    <!-- Basic responsive image -->
    <img src="%s" 
         srcset="%s" 
         sizes="%s" 
         alt="%s" 
         loading="lazy">
    
    <!-- Picture element with format fallbacks -->
    <picture>
        <source srcset="%s" type="image/avif">
        <source srcset="%s" type="image/webp">
        <img src="%s" alt="%s" loading="lazy">
    </picture>
</body>
</html>`,
		set.Fallback,
		strings.Join(srcset, ", "),
		strings.Join(set.Sizes, ", "),
		set.Alt,
		set.Fallback, // AVIF version (placeholder)
		set.Fallback, // WebP version (placeholder)
		set.Fallback,
		set.Alt)
}
