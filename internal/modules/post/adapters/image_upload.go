// INPUT ADAPTER - Image Upload Handler
// Package adapters implements HTTP handlers for image upload operations.
package adapters

import (
	"io"
	"mime/multipart"
	"net/http"

	"forum/internal/platform/upload"
)

// MaxImageUploadSize is the maximum allowed image upload size (20MB).
const MaxImageUploadSize = 20 << 20 // 20MB

// ImageUploadResult contains the result of parsing an image upload.
type ImageUploadResult struct {
	Data        []byte
	RemoveImage bool
	HasFile     bool
}

// ParseImageUpload extracts image data from a multipart form.
// It handles:
// - File uploads (with size validation)
// - Remove image flags
// - Missing file (optional image)
//
// Returns:
// - ImageUploadResult with parsed data
// - error if file exists but cannot be read or exceeds size limit
func ParseImageUpload(r *http.Request, fieldName string) (*ImageUploadResult, error) {
	result := &ImageUploadResult{
		RemoveImage: r.FormValue("remove_image") == "true",
	}

	file, header, err := r.FormFile(fieldName)
	if err == http.ErrMissingFile {
		// No file uploaded - this is OK for optional images
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result.HasFile = true

	// Validate size before reading
	if header.Size > MaxImageUploadSize {
		return nil, upload.ErrImageTooLarge
	}

	// Read file data with limit
	data, err := io.ReadAll(io.LimitReader(file, MaxImageUploadSize))
	if err != nil {
		return nil, err
	}

	result.Data = data
	return result, nil
}

// ParseMultipartImageUpload extracts image data from an already-parsed multipart form.
// Use this when r.ParseMultipartForm has already been called.
func ParseMultipartImageUpload(form *multipart.Form, fieldName string) (*ImageUploadResult, error) {
	result := &ImageUploadResult{}

	// Check for remove_image flag
	if values, ok := form.Value["remove_image"]; ok && len(values) > 0 {
		result.RemoveImage = values[0] == "true"
	}

	// Get file headers
	files, ok := form.File[fieldName]
	if !ok || len(files) == 0 {
		return result, nil
	}

	header := files[0]
	result.HasFile = true

	// Validate size before reading
	if header.Size > MaxImageUploadSize {
		return nil, upload.ErrImageTooLarge
	}

	file, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file data with limit
	data, err := io.ReadAll(io.LimitReader(file, MaxImageUploadSize))
	if err != nil {
		return nil, err
	}

	result.Data = data
	return result, nil
}

// ValidateImageType validates image data and returns a user-friendly error message.
// Returns nil if image is valid or empty.
func ValidateImageType(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	_, err := upload.DetectImageType(data)
	return err
}

// FormatImageError returns a user-friendly error message for image upload errors.
func FormatImageError(err error) string {
	switch err {
	case upload.ErrInvalidImageType:
		return "Invalid image type, must be JPEG, PNG, or GIF"
	case upload.ErrImageTooLarge:
		return "Image file too large (max 20MB)"
	case upload.ErrEmptyImage:
		return "Image data is empty"
	default:
		return "Failed to process image upload"
	}
}
