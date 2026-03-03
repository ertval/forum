// INPUT ADAPTER - HTTP Handler Base
// Package adapters implements the HTTP handlers for user endpoints.
package adapters

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	userDomain "forum/internal/modules/user/domain"
	"forum/internal/modules/user/ports"
	platformTemplates "forum/internal/platform/templates"
	"forum/internal/platform/upload"
)

const maxAvatarImageSize int64 = 1 * 1024 * 1024 // 1MB

// HTTPHandler handles HTTP requests for user operations.
type HTTPHandler struct {
	userService        ports.UserService
	middlewareProvider authPorts.AuthMiddleware
	templates          *platformTemplates.Registry
	avatarImageHandler *upload.ImageHandler
	maxAvatarSize      int64
}

// ServiceContainer defines the minimal interface needed by this handler.
type ServiceContainer interface {
	User() ports.UserService
	AuthMiddleware() authPorts.AuthMiddleware
	UploadDir() string
}

// NewHTTPHandler creates a new HTTP handler for users with unified dependency injection.
func NewHTTPHandler(services ServiceContainer, templates *platformTemplates.Registry) *HTTPHandler {
	dir := services.UploadDir()
	if dir == "" {
		dir = "./static/uploads"
	}

	return &HTTPHandler{
		userService:        services.User(),
		middlewareProvider: services.AuthMiddleware(),
		templates:          templates,
		avatarImageHandler: upload.NewImageHandler(dir, maxAvatarImageSize),
		maxAvatarSize:      maxAvatarImageSize,
	}
}

// RegisterRoutes registers all user routes.
func (h *HTTPHandler) RegisterRoutes(router *http.ServeMux) {
	// Register API routes
	h.RegisterAPIRoutes(router)

	// Register page routes (none yet)
	h.RegisterPageRoutes(router)
}

// settingsUpdateInput holds parsed settings form/request data.
type settingsUpdateInput struct {
	Username      string
	Email         string
	NewPassword   string
	Confirm       string
	AvatarData    []byte
	HasAvatarFile bool
	DeleteAvatar  bool
}

// updateCurrentUserSettings processes the settings update for a given user.
// Shared by both UpdateSettingsPage and UpdateSettingsAPI.
func (h *HTTPHandler) updateCurrentUserSettings(r *http.Request, userPublicID string) (*userDomain.User, int, string) {
	input, statusCode, errMessage := parseSettingsUpdateInput(r)
	if errMessage != "" {
		currentUser, _ := h.userService.GetByPublicID(r.Context(), userPublicID)
		return currentUser, statusCode, errMessage
	}

	if input.NewPassword != input.Confirm {
		currentUser, _ := h.userService.GetByPublicID(r.Context(), userPublicID)
		return currentUser, http.StatusBadRequest, "password confirmation does not match"
	}

	currentUser, err := h.userService.GetByPublicID(r.Context(), userPublicID)
	if err != nil || currentUser == nil {
		return nil, http.StatusNotFound, "user not found"
	}

	avatarPath := currentUser.AvatarPath
	newAvatarFilename := ""
	deleteOldAvatar := false

	if input.DeleteAvatar {
		avatarPath = ""
		deleteOldAvatar = true
	} else if input.HasAvatarFile {
		if _, err := upload.ValidateImage(input.AvatarData, h.maxAvatarSize); err != nil {
			if errors.Is(err, upload.ErrImageTooLarge) {
				return currentUser, http.StatusRequestEntityTooLarge, upload.FormatImageSizeError(h.maxAvatarSize)
			}
			return currentUser, http.StatusBadRequest, "invalid image type, must be JPEG, PNG, GIF, or WebP"
		}

		savedFilename, saveErr := h.avatarImageHandler.Save(input.AvatarData)
		if saveErr != nil {
			return currentUser, http.StatusInternalServerError, "failed to store avatar image"
		}
		avatarPath = savedFilename
		newAvatarFilename = savedFilename
	}

	updatedUser, err := h.userService.UpdateSettings(
		r.Context(),
		userPublicID,
		strings.TrimSpace(input.Username),
		strings.TrimSpace(input.Email),
		strings.TrimSpace(input.NewPassword),
		avatarPath,
	)
	if err != nil {
		if newAvatarFilename != "" {
			_ = h.avatarImageHandler.Delete(newAvatarFilename)
		}
		switch {
		case userDomain.IsPasswordValidationError(err):
			return currentUser, http.StatusBadRequest, err.Error()
		case err == userDomain.ErrInvalidEmail || err == userDomain.ErrInvalidUsername || err == userDomain.ErrWeakPassword:
			return currentUser, http.StatusBadRequest, err.Error()
		case err == userDomain.ErrEmailAlreadyExists || err == userDomain.ErrUsernameAlreadyExists:
			return currentUser, http.StatusConflict, err.Error()
		case err == userDomain.ErrUserNotFound:
			return nil, http.StatusNotFound, "user not found"
		default:
			return currentUser, http.StatusInternalServerError, "failed to update settings"
		}
	}

	if currentUser.AvatarPath != "" {
		if newAvatarFilename != "" && currentUser.AvatarPath != newAvatarFilename {
			_ = h.avatarImageHandler.Delete(currentUser.AvatarPath)
		} else if deleteOldAvatar {
			_ = h.avatarImageHandler.Delete(currentUser.AvatarPath)
		}
	}

	return updatedUser, http.StatusOK, ""
}

// parseSettingsUpdateInput parses either a multipart form or a regular form
// into a settingsUpdateInput struct.
func parseSettingsUpdateInput(r *http.Request) (*settingsUpdateInput, int, string) {
	input := &settingsUpdateInput{}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxAvatarImageSize); err != nil {
			return nil, http.StatusBadRequest, "invalid form data"
		}

		input.Username = r.FormValue("username")
		input.Email = r.FormValue("email")
		input.NewPassword = r.FormValue("new_password")
		input.Confirm = r.FormValue("confirm_password")
		input.DeleteAvatar = r.FormValue("delete_avatar") == "true"

		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			if header.Size > maxAvatarImageSize {
				return nil, http.StatusRequestEntityTooLarge, upload.FormatImageSizeError(maxAvatarImageSize)
			}
			data, readErr := io.ReadAll(io.LimitReader(file, maxAvatarImageSize))
			if readErr != nil {
				return nil, http.StatusBadRequest, "failed to read avatar image"
			}
			if int64(len(data)) > maxAvatarImageSize {
				return nil, http.StatusRequestEntityTooLarge, upload.FormatImageSizeError(maxAvatarImageSize)
			}
			input.AvatarData = data
			input.HasAvatarFile = len(data) > 0
		} else if err != http.ErrMissingFile {
			return nil, http.StatusBadRequest, "invalid avatar upload"
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return nil, http.StatusBadRequest, "invalid form data"
		}
		input.Username = r.FormValue("username")
		input.Email = r.FormValue("email")
		input.NewPassword = r.FormValue("new_password")
		input.Confirm = r.FormValue("confirm_password")
		input.DeleteAvatar = r.FormValue("delete_avatar") == "true"
	}

	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.Email) == "" {
		return nil, http.StatusBadRequest, "username and email are required"
	}

	return input, http.StatusOK, ""
}

// renderSettingsPage renders the settings HTML page with the given state.
func (h *HTTPHandler) renderSettingsPage(w http.ResponseWriter, statusCode int, currentUser *userDomain.User, errMessage string, updated bool) {
	if currentUser == nil {
		currentUser = &userDomain.User{}
	}

	if currentUser.AvatarPath != "" && currentUser.AvatarURL == "" {
		currentUser.AvatarURL = userDomain.AvatarURLPrefix + filepath.Base(currentUser.AvatarPath)
	}

	data := map[string]any{
		"Title":             "Settings",
		"User":              currentUser,
		"ErrorMessage":      errMessage,
		"Updated":           updated,
		"ShowFilter":        false,
		"ShowSidebar":       false,
		"MaxAvatarSizeMB":   h.maxAvatarSize / (1024 * 1024),
		"MaxAvatarSizeByte": h.maxAvatarSize,
	}

	if h.templates == nil {
		http.Error(w, "templates not configured", http.StatusInternalServerError)
		return
	}
	tmpl := h.templates.Lookup("settings")
	if tmpl == nil {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	buf.WriteTo(w)
}
