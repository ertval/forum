// OUTPUT PORT - Image Handler Interface
// Package ports defines the image handling output port for the post module.
package ports

// ImageHandler defines file operations for image uploads.
// This is an output port that abstracts the storage mechanism for images.
type ImageHandler interface {
	// Save validates and saves image data, returning the filename (without path prefix).
	// Returns error if image is invalid (wrong type, too large, etc.).
	Save(data []byte) (filename string, err error)

	// Delete removes an image file by filename.
	// Returns nil if file doesn't exist (idempotent).
	Delete(filename string) error
}

// ImageUploadRequest represents an image upload request with validation.
type ImageUploadRequest struct {
	Data        []byte
	RemoveImage bool
}

// IsEmpty returns true if no image data is provided and removal is not requested.
func (r ImageUploadRequest) IsEmpty() bool {
	return len(r.Data) == 0 && !r.RemoveImage
}

// HasNewImage returns true if new image data is provided.
func (r ImageUploadRequest) HasNewImage() bool {
	return len(r.Data) > 0
}
