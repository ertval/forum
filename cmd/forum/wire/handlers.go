// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
	"os"

	platformTemplates "forum/internal/platform/templates"

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
	Templates    *platformTemplates.Registry
}

// pageTemplates lists all page templates that use base.html as a layout.
// Adding a new page template only requires adding an entry here.
var pageTemplates = []platformTemplates.TemplateEntry{
	{Key: "home", Files: []string{"templates/base.html", "templates/home.html"}},
	{Key: "board", Files: []string{"templates/base.html", "templates/board.html"}},
	{Key: "post_detail", Files: []string{"templates/base.html", "templates/post_detail.html"}},
	{Key: "post_create", Files: []string{"templates/base.html", "templates/post_create.html"}},
	{Key: "post_edit", Files: []string{"templates/base.html", "templates/post_edit.html"}},
	{Key: "activity", Files: []string{"templates/base.html", "templates/activity.html"}},
	{Key: "comments", Files: []string{"templates/base.html", "templates/comments.html"}},
	{Key: "settings", Files: []string{"templates/base.html", "templates/settings.html"}},
	{Key: "login", Files: []string{"templates/base.html", "templates/login.html"}},
	{Key: "register", Files: []string{"templates/base.html", "templates/register.html"}},
	{Key: "health", Files: []string{"templates/base.html", "templates/health.html"}},
}

// initHandlers creates all HTTP handler instances with unified dependency injection.
// Returns error if templates directory exists but contains invalid templates.
func initHandlers(services *ServiceContainer) (*Handlers, error) {
	registry := platformTemplates.NewRegistry()

	// Parse all page templates if the templates directory exists (skip for API-only mode)
	if info, err := os.Stat("templates"); err == nil && info.IsDir() {
		if err := registry.LoadAll(pageTemplates); err != nil {
			return nil, err
		}
	}

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services, registry),
		User:         userAdapters.NewHTTPHandler(services, registry),
		Post:         postAdapters.NewHTTPHandler(services, registry),
		Comment:      commentAdapters.NewHTTPHandler(services, registry),
		Reaction:     reactionAdapters.NewHTTPHandler(services, registry),
		Moderation:   moderationAdapters.NewHTTPHandler(services, registry),
		Notification: notificationAdapters.NewHTTPHandler(services, registry),
		Templates:    registry,
	}, nil
}
