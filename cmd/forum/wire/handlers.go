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

// initHandlers creates all HTTP handler instances.
func initHandlers(services *Services) *Handlers {
	// Parse templates once and share between handlers that need them
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		panic(err)
	}

	postHandler := postAdapters.NewHTTPHandler(services.Post)
	// Set the shared templates
	postHandler.SetTemplates(templates)
	postHandler.SetCategoryService(services.Category)

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services.Auth),
		User:         userAdapters.NewHTTPHandler(services.User),
		Post:         postHandler,
		Comment:      commentAdapters.NewHTTPHandler(services.Comment),
		Reaction:     reactionAdapters.NewHTTPHandler(services.Reaction),
		Moderation:   moderationAdapters.NewHTTPHandler(services.Moderation),
		Notification: notificationAdapters.NewHTTPHandler(services.Notification),
	}
}
