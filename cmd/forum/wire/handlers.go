// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
	"html/template"

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
func initHandlers(services *ServiceContainer) *Handlers {
	// Parse templates once and share between handlers that need them
	// Skip if templates directory doesn't exist (for tests)
	var templates *template.Template
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		// Templates are optional - API-only mode works without them
		templates = nil
	}

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services, templates),
		User:         userAdapters.NewHTTPHandler(services, templates),
		Post:         postAdapters.NewHTTPHandler(services, templates),
		Comment:      commentAdapters.NewHTTPHandler(services, templates),
		Reaction:     reactionAdapters.NewHTTPHandler(services, templates),
		Moderation:   moderationAdapters.NewHTTPHandler(services, templates),
		Notification: notificationAdapters.NewHTTPHandler(services, templates),
	}
}
