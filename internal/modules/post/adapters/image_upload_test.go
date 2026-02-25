package adapters

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/internal/platform/upload"
)

const testMaxImageSize = 20 * 1024 * 1024 // 20MB for tests

// Helper to create a multipart request with a file
func createMultipartRequest(t *testing.T, fieldName string, fileName string, fileContent []byte, extraFields map[string]string) *http.Request {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add extra form fields first
	for key, value := range extraFields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("Failed to write field %s: %v", key, err)
		}
	}

	// Add file if content is provided
	if fileContent != nil {
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(fileContent)); err != nil {
			t.Fatalf("Failed to copy file content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestParseImageUpload(t *testing.T) {
	// Valid PNG file bytes
	validPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tests := []struct {
		name         string
		setupRequest func(t *testing.T) *http.Request
		wantHasFile  bool
		wantRemove   bool
		wantErr      error
	}{
		{
			name: "no file uploaded",
			setupRequest: func(t *testing.T) *http.Request {
				return createMultipartRequest(t, "other_field", "", nil, nil)
			},
			wantHasFile: false,
			wantRemove:  false,
			wantErr:     nil,
		},
		{
			name: "valid image file uploaded",
			setupRequest: func(t *testing.T) *http.Request {
				return createMultipartRequest(t, "image", "test.png", validPNG, nil)
			},
			wantHasFile: true,
			wantRemove:  false,
			wantErr:     nil,
		},
		{
			name: "remove_image flag set",
			setupRequest: func(t *testing.T) *http.Request {
				return createMultipartRequest(t, "image", "test.png", validPNG, map[string]string{"remove_image": "true"})
			},
			wantHasFile: true,
			wantRemove:  true,
			wantErr:     nil,
		},
		{
			name: "remove_image flag false",
			setupRequest: func(t *testing.T) *http.Request {
				return createMultipartRequest(t, "image", "test.png", validPNG, map[string]string{"remove_image": "false"})
			},
			wantHasFile: true,
			wantRemove:  false,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest(t)
			result, err := ParseImageUpload(req, "image", testMaxImageSize)

			if tt.wantErr != nil {
				if err == nil || err != tt.wantErr {
					t.Errorf("ParseImageUpload() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseImageUpload() unexpected error = %v", err)
				return
			}

			if result.HasFile != tt.wantHasFile {
				t.Errorf("ParseImageUpload() HasFile = %v, want %v", result.HasFile, tt.wantHasFile)
			}

			if result.RemoveImage != tt.wantRemove {
				t.Errorf("ParseImageUpload() RemoveImage = %v, want %v", result.RemoveImage, tt.wantRemove)
			}
		})
	}
}

func TestParseMultipartImageUpload(t *testing.T) {
	// Valid PNG file bytes
	validPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tests := []struct {
		name        string
		setupForm   func(t *testing.T) *multipart.Form
		fieldName   string
		wantHasFile bool
		wantRemove  bool
		wantErr     error
	}{
		{
			name: "no file in form",
			setupForm: func(t *testing.T) *multipart.Form {
				return &multipart.Form{
					Value: map[string][]string{},
					File:  map[string][]*multipart.FileHeader{},
				}
			},
			fieldName:   "image",
			wantHasFile: false,
			wantRemove:  false,
			wantErr:     nil,
		},
		{
			name: "remove_image flag set",
			setupForm: func(t *testing.T) *multipart.Form {
				return &multipart.Form{
					Value: map[string][]string{
						"remove_image": {"true"},
					},
					File: map[string][]*multipart.FileHeader{},
				}
			},
			fieldName:   "image",
			wantHasFile: false,
			wantRemove:  true,
			wantErr:     nil,
		},
		{
			name: "valid file in form",
			setupForm: func(t *testing.T) *multipart.Form {
				// Create a request with file to get proper FileHeader
				req := createMultipartRequest(t, "image", "test.png", validPNG, nil)
				if err := req.ParseMultipartForm(testMaxImageSize); err != nil {
					t.Fatalf("Failed to parse multipart form: %v", err)
				}
				return req.MultipartForm
			},
			fieldName:   "image",
			wantHasFile: true,
			wantRemove:  false,
			wantErr:     nil,
		},
		{
			name: "file with remove flag",
			setupForm: func(t *testing.T) *multipart.Form {
				req := createMultipartRequest(t, "image", "test.png", validPNG, map[string]string{"remove_image": "true"})
				if err := req.ParseMultipartForm(testMaxImageSize); err != nil {
					t.Fatalf("Failed to parse multipart form: %v", err)
				}
				return req.MultipartForm
			},
			fieldName:   "image",
			wantHasFile: true,
			wantRemove:  true,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := tt.setupForm(t)
			result, err := ParseMultipartImageUpload(form, tt.fieldName, testMaxImageSize)

			if tt.wantErr != nil {
				if err == nil || err != tt.wantErr {
					t.Errorf("ParseMultipartImageUpload() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMultipartImageUpload() unexpected error = %v", err)
				return
			}

			if result.HasFile != tt.wantHasFile {
				t.Errorf("ParseMultipartImageUpload() HasFile = %v, want %v", result.HasFile, tt.wantHasFile)
			}

			if result.RemoveImage != tt.wantRemove {
				t.Errorf("ParseMultipartImageUpload() RemoveImage = %v, want %v", result.RemoveImage, tt.wantRemove)
			}
		})
	}
}

func TestValidateImageType(t *testing.T) {
	// Valid PNG magic bytes
	validPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	// Valid JPEG magic bytes
	validJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE0}

	// Valid GIF magic bytes
	validGIF := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}

	// Invalid image bytes
	invalidImage := []byte{0x00, 0x00, 0x00, 0x00}

	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: nil,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: nil,
		},
		{
			name:    "valid PNG",
			data:    validPNG,
			wantErr: nil,
		},
		{
			name:    "valid JPEG",
			data:    validJPEG,
			wantErr: nil,
		},
		{
			name:    "valid GIF",
			data:    validGIF,
			wantErr: nil,
		},
		{
			name:    "invalid image data",
			data:    invalidImage,
			wantErr: upload.ErrInvalidImageType,
		},
		{
			name:    "text data",
			data:    []byte("this is not an image"),
			wantErr: upload.ErrInvalidImageType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageType(tt.data)
			if err != tt.wantErr {
				t.Errorf("ValidateImageType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatImageError(t *testing.T) {
	const testMaxSize = 20 * 1024 * 1024
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "invalid image type error",
			err:     upload.ErrInvalidImageType,
			wantMsg: "Invalid image type, must be JPEG, PNG, or GIF",
		},
		{
			name:    "image too large error",
			err:     upload.ErrImageTooLarge,
			wantMsg: "image file too large (max 20 MB)",
		},
		{
			name:    "empty image error",
			err:     upload.ErrEmptyImage,
			wantMsg: "Image data is empty",
		},
		{
			name:    "unknown error",
			err:     io.EOF,
			wantMsg: "Failed to process image upload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMsg := FormatImageError(tt.err, testMaxSize)
			if gotMsg != tt.wantMsg {
				t.Errorf("FormatImageError() = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}
