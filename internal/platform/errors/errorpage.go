// PLATFORM UTILITY - Error Page Rendering
// Provides HTML error page rendering for browser-facing page handlers.
package errors

import (
	"fmt"
	"net/http"

	"forum/internal/platform/logger"
	"forum/internal/platform/templates"
)

// statusTitles maps HTTP status codes to user-friendly titles.
var statusTitles = map[int]string{
	http.StatusBadRequest:          "Bad Request",
	http.StatusUnauthorized:        "Unauthorized",
	http.StatusForbidden:           "Forbidden",
	http.StatusNotFound:            "Page Not Found",
	http.StatusMethodNotAllowed:    "Method Not Allowed",
	http.StatusConflict:            "Conflict",
	http.StatusTooManyRequests:     "Too Many Requests",
	http.StatusInternalServerError: "Internal Server Error",
}

// statusMessages maps HTTP status codes to default user-friendly messages.
var statusMessages = map[int]string{
	http.StatusBadRequest:          "The request could not be understood or was missing required parameters.",
	http.StatusUnauthorized:        "You need to be logged in to access this page.",
	http.StatusForbidden:           "You don't have permission to access this page.",
	http.StatusNotFound:            "The page you're looking for doesn't exist or has been moved.",
	http.StatusMethodNotAllowed:    "The request method is not supported for this page.",
	http.StatusConflict:            "There was a conflict with the current state of the resource.",
	http.StatusTooManyRequests:     "You've made too many requests. Please try again later.",
	http.StatusInternalServerError: "Something went wrong on our end. Please try again later.",
}

// RenderErrorPage renders a styled HTML error page using the base template.
// If message is empty, a default message for the status code is used.
// The user parameter should be the current user data (or nil if not logged in)
// in the same format used by other page handlers.
func RenderErrorPage(w http.ResponseWriter, statusCode int, message string, user interface{}) {
	title := statusTitles[statusCode]
	if title == "" {
		title = http.StatusText(statusCode)
	}
	if title == "" {
		title = "Error"
	}

	if message == "" {
		message = statusMessages[statusCode]
		if message == "" {
			message = "An unexpected error occurred."
		}
	}

	data := map[string]interface{}{
		"Title":        fmt.Sprintf("Error %d", statusCode),
		"ErrorCode":    statusCode,
		"ErrorTitle":   title,
		"ErrorMessage": message,
		"User":         user,
	}

	tmpl, err := templates.Get("error", "templates/base.html", "templates/error.html")
	if err != nil {
		// Fallback to plain text if template fails
		http.Error(w, fmt.Sprintf("%d - %s", statusCode, title), statusCode)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		// Already wrote header, just log
		errLogger.Error("failed to render error page", logger.Error(err))
	}
}
