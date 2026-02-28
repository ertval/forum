// INPUT ADAPTER - Settings Update Handlers
// Package adapters implements settings update handlers for user endpoints.
package adapters

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	authPorts "forum/internal/modules/auth/ports"
	"forum/internal/modules/user/domain"
	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/templates"
	"forum/internal/platform/upload"
)

const maxAvatarImageSize int64 = 1 * 1024 * 1024 // 1MB

type settingsUpdateInput struct {
	Username      string
	Email         string
	NewPassword   string
	Confirm       string
	AvatarData    []byte
	HasAvatarFile bool
	DeleteAvatar  bool
}

func (h *HTTPHandler) UpdateSettingsPage(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updatedUser, statusCode, errMessage := h.updateCurrentUserSettings(r, userPublicID)
	if errMessage != "" {
		h.renderSettingsPage(w, statusCode, updatedUser, errMessage, false)
		return
	}

	http.Redirect(w, r, "/settings?updated=1", http.StatusSeeOther)
}

func (h *HTTPHandler) UpdateSettingsAPI(w http.ResponseWriter, r *http.Request) {
	userPublicID := authPorts.GetUserID(r.Context())
	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "authentication required")
		return
	}

	updatedUser, statusCode, errMessage := h.updateCurrentUserSettings(r, userPublicID)
	if errMessage != "" {
		platformErrors.WriteErrorJSON(w, statusCode, errMessage)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "settings updated successfully",
		"user":    updatedUser,
	})
}

func (h *HTTPHandler) updateCurrentUserSettings(r *http.Request, userPublicID string) (*domain.User, int, string) {
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
		// User wants to remove their avatar
		avatarPath = ""
		deleteOldAvatar = true
	} else if input.HasAvatarFile {
		if err := upload.ValidateImage(input.AvatarData, h.maxAvatarSize); err != nil {
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
		case domain.IsPasswordValidationError(err):
			return currentUser, http.StatusBadRequest, err.Error()
		case err == domain.ErrInvalidEmail || err == domain.ErrInvalidUsername || err == domain.ErrWeakPassword:
			return currentUser, http.StatusBadRequest, err.Error()
		case err == domain.ErrEmailAlreadyExists || err == domain.ErrUsernameAlreadyExists:
			return currentUser, http.StatusConflict, err.Error()
		case err == domain.ErrUserNotFound:
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

func (h *HTTPHandler) renderSettingsPage(w http.ResponseWriter, statusCode int, currentUser *domain.User, errMessage string, updated bool) {
	if currentUser == nil {
		currentUser = &domain.User{}
	}

	if currentUser != nil && currentUser.AvatarPath != "" && currentUser.AvatarURL == "" {
		currentUser.AvatarURL = domain.AvatarURLPrefix + filepath.Base(currentUser.AvatarPath)
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

	tmpl, err := templates.Get("settings", "templates/base.html", "templates/settings.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
