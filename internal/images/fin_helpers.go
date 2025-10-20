package images

import (
	"fmt"
	"html/template"
	"strings"
)

// FinTemplateHelpers provides Fin template functions for image optimization
type FinTemplateHelpers struct {
	optimizer ImageOptimizer
	storage   ImageStorage
	cache     ImageCache
}

// NewFinTemplateHelpers creates new Fin template helpers
func NewFinTemplateHelpers(optimizer ImageOptimizer, storage ImageStorage, cache ImageCache) *FinTemplateHelpers {
	return &FinTemplateHelpers{
		optimizer: optimizer,
		storage:   storage,
		cache:     cache,
	}
}

// GetTemplateFunctions returns template functions for Fin
func (h *FinTemplateHelpers) GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"image":          h.image,
		"responsive_img": h.responsiveImg,
		"image_url":      h.imageUrl,
		"image_srcset":   h.imageSrcset,
		"image_sizes":    h.imageSizes,
		"lazy_image":     h.lazyImage,
		"webp_image":     h.webpImage,
		"avif_image":     h.avifImage,
		"image_metadata": h.imageMetadata,
		"optimize_image": h.optimizeImage,
	}
}

// image generates a basic img tag
func (h *FinTemplateHelpers) image(src string, alt string, attrs ...map[string]string) template.HTML {
	attributes := make(map[string]string)
	if len(attrs) > 0 {
		attributes = attrs[0]
	}

	// Set default attributes
	if alt == "" {
		alt = "Image"
	}
	attributes["src"] = src
	attributes["alt"] = alt

	return template.HTML(h.buildImgTag(attributes))
}

// responsiveImg generates a responsive img tag with srcset
func (h *FinTemplateHelpers) responsiveImg(src string, alt string, sizes []int, attrs ...map[string]string) template.HTML {
	attributes := make(map[string]string)
	if len(attrs) > 0 {
		attributes = attrs[0]
	}

	// Generate srcset
	srcset := h.generateSrcset(src, sizes)
	if srcset != "" {
		attributes["srcset"] = srcset
	}

	// Generate sizes attribute
	if len(sizes) > 0 {
		sizesAttr := h.generateSizes(sizes)
		attributes["sizes"] = sizesAttr
	}

	attributes["src"] = src
	attributes["alt"] = alt

	return template.HTML(h.buildImgTag(attributes))
}

// lazyImage generates a lazy-loaded img tag
func (h *FinTemplateHelpers) lazyImage(src string, alt string, placeholder string, attrs ...map[string]string) template.HTML {
	attributes := make(map[string]string)
	if len(attrs) > 0 {
		attributes = attrs[0]
	}

	// Add lazy loading attributes
	attributes["src"] = placeholder
	attributes["data-src"] = src
	attributes["alt"] = alt
	attributes["loading"] = "lazy"
	attributes["class"] = "lazy-image"

	return template.HTML(h.buildImgTag(attributes))
}

// webpImage generates a picture tag with WebP fallback
func (h *FinTemplateHelpers) webpImage(src string, alt string, attrs ...map[string]string) template.HTML {
	attributes := make(map[string]string)
	if len(attrs) > 0 {
		attributes = attrs[0]
	}

	// Generate WebP version
	webpSrc := h.generateWebPSrc(src)

	var sb strings.Builder
	sb.WriteString("<picture>")

	// WebP source
	if webpSrc != "" {
		sb.WriteString(fmt.Sprintf(`<source srcset="%s" type="image/webp">`, webpSrc))
	}

	// Fallback img tag
	attributes["src"] = src
	attributes["alt"] = alt
	sb.WriteString(h.buildImgTag(attributes))
	sb.WriteString("</picture>")

	return template.HTML(sb.String())
}

// avifImage generates a picture tag with AVIF fallback
func (h *FinTemplateHelpers) avifImage(src string, alt string, attrs ...map[string]string) template.HTML {
	attributes := make(map[string]string)
	if len(attrs) > 0 {
		attributes = attrs[0]
	}

	// Generate AVIF and WebP versions
	avifSrc := h.generateAVIFSrc(src)
	webpSrc := h.generateWebPSrc(src)

	var sb strings.Builder
	sb.WriteString("<picture>")

	// AVIF source
	if avifSrc != "" {
		sb.WriteString(fmt.Sprintf(`<source srcset="%s" type="image/avif">`, avifSrc))
	}

	// WebP source
	if webpSrc != "" {
		sb.WriteString(fmt.Sprintf(`<source srcset="%s" type="image/webp">`, webpSrc))
	}

	// Fallback img tag
	attributes["src"] = src
	attributes["alt"] = alt
	sb.WriteString(h.buildImgTag(attributes))
	sb.WriteString("</picture>")

	return template.HTML(sb.String())
}

// imageUrl generates an optimized image URL
func (h *FinTemplateHelpers) imageUrl(src string, width int, height int, format string) string {
	// This would integrate with your image optimization service
	// For now, return a placeholder URL structure
	if width > 0 || height > 0 || format != "" {
		params := []string{}
		if width > 0 {
			params = append(params, fmt.Sprintf("w=%d", width))
		}
		if height > 0 {
			params = append(params, fmt.Sprintf("h=%d", height))
		}
		if format != "" {
			params = append(params, fmt.Sprintf("f=%s", format))
		}

		if len(params) > 0 {
			return fmt.Sprintf("%s?%s", src, strings.Join(params, "&"))
		}
	}

	return src
}

// imageSrcset generates srcset attribute
func (h *FinTemplateHelpers) imageSrcset(src string, sizes []int) string {
	return h.generateSrcset(src, sizes)
}

// imageSizes generates sizes attribute
func (h *FinTemplateHelpers) imageSizes(sizes []int) string {
	return h.generateSizes(sizes)
}

// imageMetadata returns image metadata (placeholder)
func (h *FinTemplateHelpers) imageMetadata(src string) map[string]interface{} {
	// This would fetch actual metadata from the image
	return map[string]interface{}{
		"width":  0,
		"height": 0,
		"format": "unknown",
		"size":   0,
	}
}

// optimizeImage returns optimization info (placeholder)
func (h *FinTemplateHelpers) optimizeImage(src string) map[string]interface{} {
	return map[string]interface{}{
		"original_size":   0,
		"optimized_size":  0,
		"savings_percent": 0,
		"format":          "unknown",
	}
}

// Helper methods

// buildImgTag builds an img tag from attributes
func (h *FinTemplateHelpers) buildImgTag(attrs map[string]string) string {
	var parts []string
	for key, value := range attrs {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, template.HTMLEscapeString(value)))
	}

	return fmt.Sprintf("<img %s>", strings.Join(parts, " "))
}

// generateSrcset generates srcset attribute
func (h *FinTemplateHelpers) generateSrcset(baseSrc string, sizes []int) string {
	var srcset []string

	for _, size := range sizes {
		optimizedSrc := h.imageUrl(baseSrc, size, 0, "")
		srcset = append(srcset, fmt.Sprintf("%s %dw", optimizedSrc, size))
	}

	return strings.Join(srcset, ", ")
}

// generateSizes generates sizes attribute
func (h *FinTemplateHelpers) generateSizes(sizes []int) string {
	var sizeDescriptors []string

	for _, size := range sizes {
		sizeDescriptors = append(sizeDescriptors, fmt.Sprintf("(max-width: %dpx) %dpx", size, size))
	}

	// Add default size
	if len(sizes) > 0 {
		sizeDescriptors = append(sizeDescriptors, fmt.Sprintf("%dpx", sizes[len(sizes)-1]))
	}

	return strings.Join(sizeDescriptors, ", ")
}

// generateWebPSrc generates WebP version URL
func (h *FinTemplateHelpers) generateWebPSrc(src string) string {
	return h.imageUrl(src, 0, 0, "webp")
}

// generateAVIFSrc generates AVIF version URL
func (h *FinTemplateHelpers) generateAVIFSrc(src string) string {
	return h.imageUrl(src, 0, 0, "avif")
}

// ResponsiveImageComponent represents a responsive image component
type ResponsiveImageComponent struct {
	Src         string            `json:"src"`
	Alt         string            `json:"alt"`
	Sizes       []int             `json:"sizes"`
	Formats     []string          `json:"formats"`
	Lazy        bool              `json:"lazy"`
	Placeholder string            `json:"placeholder"`
	Attributes  map[string]string `json:"attributes"`
}

// Render renders the responsive image component
func (c *ResponsiveImageComponent) Render() template.HTML {
	helpers := NewFinTemplateHelpers(nil, nil, nil)

	if len(c.Formats) > 0 && contains(c.Formats, "avif") {
		return helpers.avifImage(c.Src, c.Alt, c.Attributes)
	} else if len(c.Formats) > 0 && contains(c.Formats, "webp") {
		return helpers.webpImage(c.Src, c.Alt, c.Attributes)
	} else if c.Lazy {
		return helpers.lazyImage(c.Src, c.Alt, c.Placeholder, c.Attributes)
	} else if len(c.Sizes) > 0 {
		return helpers.responsiveImg(c.Src, c.Alt, c.Sizes, c.Attributes)
	} else {
		return helpers.image(c.Src, c.Alt, c.Attributes)
	}
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
