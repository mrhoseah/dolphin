package router

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// UploadConfig configures file upload behavior
type UploadConfig struct {
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
	UploadDir    string   // Directory to save uploaded files
	Logger       *zap.Logger
}

// DefaultUploadConfig returns default upload configuration
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		MaxSize:      10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
		UploadDir:    "storage/uploads",
	}
}

// UploadedFile represents an uploaded file
type UploadedFile struct {
	OriginalName string
	StoredName   string
	Path         string
	Size         int64
	MimeType     string
	Extension    string
}

// UploadFile handles a single file upload
func UploadFile(r *http.Request, fieldName string, config *UploadConfig) (*UploadedFile, error) {
	if config == nil {
		config = DefaultUploadConfig()
	}

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from form: %w", err)
	}
	defer file.Close()

	// Check file size
	if header.Size > config.MaxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxSize)
	}

	// Check MIME type
	if len(config.AllowedTypes) > 0 {
		allowed := false
		for _, mimeType := range config.AllowedTypes {
			if strings.HasPrefix(header.Header.Get("Content-Type"), mimeType) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("file type not allowed. Allowed types: %v", config.AllowedTypes)
		}
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(header.Filename), ext)
	filePath := filepath.Join(config.UploadDir, uniqueName)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	uploadedFile := &UploadedFile{
		OriginalName: header.Filename,
		StoredName:   uniqueName,
		Path:         filePath,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		Extension:    ext,
	}

	if config.Logger != nil {
		config.Logger.Info("File uploaded successfully",
			zap.String("original_name", uploadedFile.OriginalName),
			zap.String("stored_name", uploadedFile.StoredName),
			zap.Int64("size", uploadedFile.Size),
		)
	}

	return uploadedFile, nil
}

// UploadMultipleFiles handles multiple file uploads
func UploadMultipleFiles(r *http.Request, fieldName string, config *UploadConfig) ([]*UploadedFile, error) {
	if config == nil {
		config = DefaultUploadConfig()
	}

	if err := r.ParseMultipartForm(config.MaxSize * 10); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in field: %s", fieldName)
	}

	var uploadedFiles []*UploadedFile
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		// Create a new request with single file for processing
		// This is a simplified approach - in production, you might want to refactor
		uploadedFile, err := processFileUpload(file, fileHeader, config)
		file.Close()

		if err != nil {
			if config.Logger != nil {
				config.Logger.Warn("Failed to upload file",
					zap.String("filename", fileHeader.Filename),
					zap.Error(err),
				)
			}
			continue
		}

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	if len(uploadedFiles) == 0 {
		return nil, fmt.Errorf("no files were successfully uploaded")
	}

	return uploadedFiles, nil
}

// processFileUpload processes a single file upload
func processFileUpload(file multipart.File, header *multipart.FileHeader, config *UploadConfig) (*UploadedFile, error) {
	// Check file size
	if header.Size > config.MaxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Check MIME type
	if len(config.AllowedTypes) > 0 {
		allowed := false
		for _, mimeType := range config.AllowedTypes {
			if strings.HasPrefix(header.Header.Get("Content-Type"), mimeType) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("file type not allowed")
		}
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeFilename(header.Filename), ext)
	filePath := filepath.Join(config.UploadDir, uniqueName)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &UploadedFile{
		OriginalName: header.Filename,
		StoredName:   uniqueName,
		Path:         filePath,
		Size:         header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		Extension:    ext,
	}, nil
}

// sanitizeFilename sanitizes a filename to prevent directory traversal and other security issues
func sanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	
	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")
	
	// Remove leading dots and spaces
	filename = strings.TrimLeft(filename, ". ")
	
	// Limit length
	if len(filename) > 100 {
		filename = filename[:100]
	}
	
	return filename
}

// DeleteUploadedFile deletes an uploaded file
func DeleteUploadedFile(filePath string) error {
	return os.Remove(filePath)
}

