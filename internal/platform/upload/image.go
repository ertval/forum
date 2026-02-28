// Package upload provides image upload handling utilities.
package upload

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid/v5"
)

// Errors for image handling.
var (
	ErrInvalidImageType = errors.New("invalid image type, must be JPEG, PNG, GIF, or WebP")
	ErrEmptyImage       = errors.New("image data is empty")
	ErrPathTraversal    = errors.New("invalid filename: path traversal detected")
	ErrImageTooLarge    = errors.New("image file too large")
)

// FormatImageSizeError returns a formatted error message for image size violations.
func FormatImageSizeError(maxSize int64) string {
	return fmt.Sprintf("image file too large (max %d MB)", maxSize/(1024*1024))
}

// allowedMIMETypes maps allowed MIME types to file extensions.
var allowedMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// DetectImageType detects the image type from file data using magic bytes.
// Returns the MIME type if valid, or error if not a supported image type.
func DetectImageType(data []byte) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyImage
	}

	if isWebP(data) {
		return "image/webp", nil
	}

	// Use http.DetectContentType to detect MIME type from magic bytes
	mimeType := http.DetectContentType(data)

	// Check if it's an allowed image type
	if _, ok := allowedMIMETypes[mimeType]; ok {
		return mimeType, nil
	}

	return "", ErrInvalidImageType
}

func isWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

// MIMEToExtension converts a MIME type to a file extension.
func MIMEToExtension(mimeType string) (string, error) {
	ext, ok := allowedMIMETypes[mimeType]
	if !ok {
		return "", ErrInvalidImageType
	}
	return ext, nil
}

// ValidateImageSize checks if the image size is within the allowed limit.
func ValidateImageSize(size int64, maxSize int64) error {
	if size <= 0 {
		return ErrEmptyImage
	}
	if size > maxSize {
		return ErrImageTooLarge
	}
	return nil
}

// ValidateImage validates image data for type and size.
// Returns the detected MIME type on success.
func ValidateImage(data []byte, maxSize int64) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyImage
	}

	if err := ValidateImageSize(int64(len(data)), maxSize); err != nil {
		return "", err
	}

	return DetectImageType(data)
}

// ImageHandler handles image file operations.
type ImageHandler struct {
	uploadDir string
	maxSize   int64
}

// NewImageHandler creates a new ImageHandler with the specified upload directory and max size.
// The upload directory is created if it does not already exist.
func NewImageHandler(uploadDir string, maxSize int64) *ImageHandler {
	// Ensure the upload directory exists once at construction time.
	_ = os.MkdirAll(uploadDir, 0755)
	return &ImageHandler{
		uploadDir: uploadDir,
		maxSize:   maxSize,
	}
}

// Save saves image data to disk and returns the filename (not full path).
// The filename is a UUID with the appropriate extension based on the image type.
func (h *ImageHandler) Save(data []byte) (string, error) {
	// Validate image and get MIME type in one pass
	mimeType, err := ValidateImage(data, h.maxSize)
	if err != nil {
		return "", err
	}

	// Get extension from already-detected MIME type (no second DetectImageType call)
	ext, err := MIMEToExtension(mimeType)
	if err != nil {
		return "", err
	}

	// Generate UUID filename
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	filename := id.String() + ext

	// Build full path
	fullPath := filepath.Join(h.uploadDir, filename)

	// Verify the final path is within the upload directory (prevent path traversal)
	rel, err := filepath.Rel(h.uploadDir, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("invalid upload path")
	}

	// Write file atomically using temp file + rename
	tmpPath := fullPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		// Clean up temp file on failure
		os.Remove(tmpPath)
		return "", err
	}

	return filename, nil
}

// Delete removes an image file from disk.
// It returns nil if the file doesn't exist (idempotent).
func (h *ImageHandler) Delete(filename string) error {
	// Security: prevent path traversal
	if err := validateFilename(filename); err != nil {
		return err
	}

	fullPath := filepath.Join(h.uploadDir, filename)

	// Verify the resolved path is still within upload directory
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return err
	}
	absUploadDir, err := filepath.Abs(h.uploadDir)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absPath, absUploadDir) {
		return ErrPathTraversal
	}

	// Remove file (ignore not exist errors)
	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// FullPath returns the full filesystem path for a filename.
func (h *ImageHandler) FullPath(filename string) string {
	return filepath.Join(h.uploadDir, filename)
}

// validateFilename checks for path traversal attempts in filename.
func validateFilename(filename string) error {
	// Check for absolute paths
	if filepath.IsAbs(filename) {
		return ErrPathTraversal
	}

	// Check for path traversal patterns
	if strings.Contains(filename, "..") {
		return ErrPathTraversal
	}

	// Check for path separators that could indicate subdirectories
	if strings.ContainsAny(filename, "/\\") {
		return ErrPathTraversal
	}

	return nil
}
