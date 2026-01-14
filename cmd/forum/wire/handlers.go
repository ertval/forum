// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
	"html/template"
	"os"

	authAdapters "forum/internal/modules/auth/adapters"
	commentAdapters "forum/internal/modules/comment/adapters"
	moderationAdapters "forum/internal/modules/moderation/adapters"
	notificationAdapters "forum/internal/modules/notification/adapters"
	postAdapters "forum/internal/modules/post/adapters"
	reactionAdapters "forum/internal/modules/reaction/adapters"
	userAdapters "forum/internal/modules/user/adapters"
)

// Handlers holds all HTTP handler instances.
type Handlers struct {
	Auth         *authAdapters.HTTPHandler
	User         *userAdapters.HTTPHandler
	Post         *postAdapters.HTTPHandler
	Comment      *commentAdapters.HTTPHandler
	Reaction     *reactionAdapters.HTTPHandler
	Moderation   *moderationAdapters.HTTPHandler
	Notification *notificationAdapters.HTTPHandler
}

// initHandlers creates all HTTP handler instances with unified dependency injection.
// Returns error if templates directory exists but contains invalid templates.
func initHandlers(services *ServiceContainer) (*Handlers, error) {
	// Parse templates once and share between handlers that need them
	var templates *template.Template

	// Check if templates directory exists
	if info, err := os.Stat("templates"); err == nil && info.IsDir() {
		// Directory exists - parse templates (errors are fatal)
		templates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			return nil, err
		}
	}
	// If directory doesn't exist, templates remain nil (API-only mode)

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services, templates),
		User:         userAdapters.NewHTTPHandler(services, templates),
		Post:         postAdapters.NewHTTPHandler(services, templates),
		Comment:      commentAdapters.NewHTTPHandler(services, templates),
		Reaction:     reactionAdapters.NewHTTPHandler(services, templates),
		Moderation:   moderationAdapters.NewHTTPHandler(services, templates),
		Notification: notificationAdapters.NewHTTPHandler(services, templates),
	}, nil
}
