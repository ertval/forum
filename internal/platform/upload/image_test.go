// Package upload provides image upload handling utilities.
package upload

import (
	"os"
	"path/filepath"
	"testing"
)

const testMaxImageSize = 20 * 1024 * 1024 // 20MB for tests

// Sample image magic bytes for testing
var (
	// JPEG magic bytes: FF D8 FF
	jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	// GIF magic bytes: 47 49 46 38
	gifMagic = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	// WebP magic bytes: RIFF....WEBP
	webpMagic = []byte{0x52, 0x49, 0x46, 0x46, 0x24, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}
	// Invalid file (plain text)
	textFile = []byte("This is not an image file")
)

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantType string
		wantErr  bool
	}{
		{
			name:     "valid JPEG",
			data:     jpegMagic,
			wantType: "image/jpeg",
			wantErr:  false,
		},
		{
			name:     "valid PNG",
			data:     pngMagic,
			wantType: "image/png",
			wantErr:  false,
		},
		{
			name:     "valid GIF",
			data:     gifMagic,
			wantType: "image/gif",
			wantErr:  false,
		},
		{
			name:     "valid WebP",
			data:     webpMagic,
			wantType: "image/webp",
			wantErr:  false,
		},
		{
			name:     "invalid - text file",
			data:     textFile,
			wantType: "",
			wantErr:  true,
		},
		{
			name:     "invalid - empty data",
			data:     []byte{},
			wantType: "",
			wantErr:  true,
		},
		{
			name:     "invalid - nil data",
			data:     nil,
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := DetectImageType(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectImageType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("DetectImageType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestMIMEToExtension(t *testing.T) {
	tests := []struct {
		mime    string
		wantExt string
		wantErr bool
	}{
		{"image/jpeg", ".jpg", false},
		{"image/png", ".png", false},
		{"image/gif", ".gif", false},
		{"image/webp", ".webp", false},
		{"text/plain", "", true},
		{"application/pdf", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			gotExt, err := MIMEToExtension(tt.mime)
			if (err != nil) != tt.wantErr {
				t.Errorf("MIMEToExtension(%q) error = %v, wantErr %v", tt.mime, err, tt.wantErr)
				return
			}
			if gotExt != tt.wantExt {
				t.Errorf("MIMEToExtension(%q) = %v, want %v", tt.mime, gotExt, tt.wantExt)
			}
		})
	}
}

func TestValidateImageSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"valid - 1 byte", 1, false},
		{"valid - 1KB", 1024, false},
		{"valid - 1MB", 1024 * 1024, false},
		{"valid - 19MB", 19 * 1024 * 1024, false},
		{"valid - exactly 20MB", testMaxImageSize, false},
		{"invalid - 21MB", 21 * 1024 * 1024, true},
		{"invalid - 100MB", 100 * 1024 * 1024, true},
		{"invalid - 0 bytes", 0, true},
		{"invalid - negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageSize(tt.size, testMaxImageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageSize(%d) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
		})
	}
}

func TestImageHandler_Save(t *testing.T) {
	// Create temporary directory for test uploads
	tmpDir := t.TempDir()
	handler := NewImageHandler(tmpDir, testMaxImageSize)

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "save valid JPEG",
			data:    jpegMagic,
			wantErr: false,
		},
		{
			name:    "save valid PNG",
			data:    pngMagic,
			wantErr: false,
		},
		{
			name:    "save valid GIF",
			data:    gifMagic,
			wantErr: false,
		},
		{
			name:    "save valid WebP",
			data:    webpMagic,
			wantErr: false,
		},
		{
			name:    "reject invalid file",
			data:    textFile,
			wantErr: true,
		},
		{
			name:    "reject empty file",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename, err := handler.Save(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageHandler.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				fullPath := filepath.Join(tmpDir, filename)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("ImageHandler.Save() file not created at %s", fullPath)
				}

				// Verify filename format (UUID.ext)
				ext := filepath.Ext(filename)
				if ext != ".jpg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
					t.Errorf("ImageHandler.Save() unexpected extension: %s", ext)
				}
			}
		})
	}
}

func TestImageHandler_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewImageHandler(tmpDir, testMaxImageSize)

	// Save an image first
	filename, err := handler.Save(jpegMagic)
	if err != nil {
		t.Fatalf("Failed to save test image: %v", err)
	}

	// Verify file exists
	fullPath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("Test image not created at %s", fullPath)
	}

	// Delete the image
	err = handler.Delete(filename)
	if err != nil {
		t.Errorf("ImageHandler.Delete() error = %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Errorf("ImageHandler.Delete() file still exists at %s", fullPath)
	}
}

func TestImageHandler_Delete_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewImageHandler(tmpDir, testMaxImageSize)

	// Deleting non-existent file should not error
	err := handler.Delete("non-existent.jpg")
	if err != nil {
		t.Errorf("ImageHandler.Delete() should not error on non-existent file, got: %v", err)
	}
}

func TestImageHandler_Delete_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewImageHandler(tmpDir, testMaxImageSize)

	// Attempt path traversal should be blocked
	tests := []string{
		"../secret.txt",
		"../../etc/passwd",
		"/etc/passwd",
		"subdir/../../../secret.txt",
	}

	for _, filename := range tests {
		t.Run(filename, func(t *testing.T) {
			err := handler.Delete(filename)
			if err == nil {
				t.Errorf("ImageHandler.Delete(%q) should error on path traversal", filename)
			}
		})
	}
}

func TestImageHandler_Save_CreatesDirectory(t *testing.T) {
	// Use a path that doesn't exist yet
	tmpDir := filepath.Join(t.TempDir(), "nested", "uploads")
	handler := NewImageHandler(tmpDir, testMaxImageSize)

	filename, err := handler.Save(jpegMagic)
	if err != nil {
		t.Errorf("ImageHandler.Save() error = %v", err)
		return
	}

	// Verify directory was created and file exists
	fullPath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("ImageHandler.Save() file not created at %s", fullPath)
	}
}

func TestValidateImage(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid JPEG",
			data:    jpegMagic,
			wantErr: false,
		},
		{
			name:    "valid PNG",
			data:    pngMagic,
			wantErr: false,
		},
		{
			name:    "valid GIF",
			data:    gifMagic,
			wantErr: false,
		},
		{
			name:    "valid WebP",
			data:    webpMagic,
			wantErr: false,
		},
		{
			name:    "invalid type",
			data:    textFile,
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImage(tt.data, testMaxImageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageHandler_FullPath(t *testing.T) {
	baseDir := filepath.Join("var", "uploads")
	handler := NewImageHandler(baseDir, testMaxImageSize)

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "simple filename",
			filename: "image.png",
			want:     filepath.Join(baseDir, "image.png"),
		},
		{
			name:     "uuid filename",
			filename: "550e8400-e29b-41d4-a716-446655440000.jpg",
			want:     filepath.Join(baseDir, "550e8400-e29b-41d4-a716-446655440000.jpg"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.FullPath(tt.filename)
			if got != tt.want {
				t.Errorf("FullPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid filename",
			filename: "image.png",
			wantErr:  false,
		},
		{
			name:     "uuid filename",
			filename: "550e8400-e29b-41d4-a716-446655440000.jpg",
			wantErr:  false,
		},
		{
			name:     "absolute path unix",
			filename: "/etc/passwd",
			wantErr:  true,
		},
		{
			name:     "path traversal with double dots",
			filename: "../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "path with forward slash",
			filename: "subdir/image.png",
			wantErr:  true,
		},
		{
			name:     "path with backslash",
			filename: "subdir\\image.png",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
