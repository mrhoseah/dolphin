package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Driver defines the interface for storage drivers
type Driver interface {
	Put(path string, content io.Reader) error
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
	Exists(path string) bool
	URL(path string) string
	Size(path string) (int64, error)
	List(prefix string) ([]FileInfo, error)
	Copy(src, dest string) error
	Move(src, dest string) error
}

// FileInfo represents file information
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// LocalDriver implements local file system storage
type LocalDriver struct {
	rootPath string
	baseURL  string
}

// NewLocalDriver creates a new local storage driver
func NewLocalDriver(rootPath, baseURL string) *LocalDriver {
	return &LocalDriver{
		rootPath: rootPath,
		baseURL:  baseURL,
	}
}

func (d *LocalDriver) Put(path string, content io.Reader) error {
	fullPath := filepath.Join(d.rootPath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, content)
	return err
}

func (d *LocalDriver) Get(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(d.rootPath, path)
	return os.Open(fullPath)
}

func (d *LocalDriver) Delete(path string) error {
	fullPath := filepath.Join(d.rootPath, path)
	return os.Remove(fullPath)
}

func (d *LocalDriver) Exists(path string) bool {
	fullPath := filepath.Join(d.rootPath, path)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

func (d *LocalDriver) URL(path string) string {
	return d.baseURL + "/" + strings.TrimPrefix(path, "/")
}

func (d *LocalDriver) Size(path string) (int64, error) {
	fullPath := filepath.Join(d.rootPath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (d *LocalDriver) List(prefix string) ([]FileInfo, error) {
	fullPath := filepath.Join(d.rootPath, prefix)

	var files []FileInfo
	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(d.rootPath, path)
		files = append(files, FileInfo{
			Name:    info.Name(),
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})

		return nil
	})

	return files, err
}

func (d *LocalDriver) Copy(src, dest string) error {
	srcPath := filepath.Join(d.rootPath, src)
	destPath := filepath.Join(d.rootPath, dest)

	// Create destination directory
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func (d *LocalDriver) Move(src, dest string) error {
	if err := d.Copy(src, dest); err != nil {
		return err
	}
	return d.Delete(src)
}

// S3Driver implements AWS S3 storage
type S3Driver struct {
	bucket   string
	region   string
	baseURL  string
	endpoint string // For S3-compatible services
}

// NewS3Driver creates a new S3 storage driver
func NewS3Driver(bucket, region, baseURL, endpoint string) *S3Driver {
	return &S3Driver{
		bucket:   bucket,
		region:   region,
		baseURL:  baseURL,
		endpoint: endpoint,
	}
}

func (d *S3Driver) Put(path string, content io.Reader) error {
	// Implementation would use AWS SDK
	// For now, return a placeholder
	return fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) Get(path string) (io.ReadCloser, error) {
	// Implementation would use AWS SDK
	return nil, fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) Delete(path string) error {
	return fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) Exists(path string) bool {
	// Implementation would use AWS SDK
	return false
}

func (d *S3Driver) URL(path string) string {
	if d.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", d.endpoint, d.bucket, path)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", d.bucket, d.region, path)
}

func (d *S3Driver) Size(path string) (int64, error) {
	return 0, fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) List(prefix string) ([]FileInfo, error) {
	return nil, fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) Copy(src, dest string) error {
	return fmt.Errorf("S3 driver not implemented yet")
}

func (d *S3Driver) Move(src, dest string) error {
	return fmt.Errorf("S3 driver not implemented yet")
}

// StorageManager manages storage operations
type StorageManager struct {
	driver Driver
}

// NewStorageManager creates a new storage manager
func NewStorageManager(driver Driver) *StorageManager {
	return &StorageManager{driver: driver}
}

// Put stores content at the given path
func (m *StorageManager) Put(path string, content io.Reader) error {
	return m.driver.Put(path, content)
}

// PutString stores string content
func (m *StorageManager) PutString(path, content string) error {
	return m.Put(path, strings.NewReader(content))
}

// PutBytes stores byte content
func (m *StorageManager) PutBytes(path string, content []byte) error {
	return m.Put(path, bytes.NewReader(content))
}

// Get retrieves content from the given path
func (m *StorageManager) Get(path string) (io.ReadCloser, error) {
	return m.driver.Get(path)
}

// GetString retrieves content as string
func (m *StorageManager) GetString(path string) (string, error) {
	reader, err := m.Get(path)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	return buf.String(), err
}

// Delete removes content at the given path
func (m *StorageManager) Delete(path string) error {
	return m.driver.Delete(path)
}

// Exists checks if content exists at the given path
func (m *StorageManager) Exists(path string) bool {
	return m.driver.Exists(path)
}

// URL generates a public URL for the given path
func (m *StorageManager) URL(path string) string {
	return m.driver.URL(path)
}

// Size returns the size of content at the given path
func (m *StorageManager) Size(path string) (int64, error) {
	return m.driver.Size(path)
}

// List returns a list of files with the given prefix
func (m *StorageManager) List(prefix string) ([]FileInfo, error) {
	return m.driver.List(prefix)
}

// Copy copies content from src to dest
func (m *StorageManager) Copy(src, dest string) error {
	return m.driver.Copy(src, dest)
}

// Move moves content from src to dest
func (m *StorageManager) Move(src, dest string) error {
	return m.driver.Move(src, dest)
}

// UploadFile uploads a file from the local filesystem
func (m *StorageManager) UploadFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return m.Put(remotePath, file)
}

// DownloadFile downloads content to the local filesystem
func (m *StorageManager) DownloadFile(remotePath, localPath string) error {
	reader, err := m.Get(remotePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Default string                `yaml:"default"`
	Disks   map[string]DiskConfig `yaml:"disks"`
}

// DiskConfig holds disk-specific configuration
type DiskConfig struct {
	Driver  string            `yaml:"driver"`
	Options map[string]string `yaml:"options"`
}

// DefaultStorageConfig returns default storage configuration
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Default: "local",
		Disks: map[string]DiskConfig{
			"local": {
				Driver: "local",
				Options: map[string]string{
					"root":     "./storage/app",
					"base_url": "/storage",
				},
			},
			"s3": {
				Driver: "s3",
				Options: map[string]string{
					"bucket":   "",
					"region":   "us-east-1",
					"base_url": "",
					"endpoint": "",
				},
			},
		},
	}
}
