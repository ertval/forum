// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
	"html/template"
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
	Templates    *template.Template
}

// initHandlers creates all HTTP handler instances with unified dependency injection.
// Returns error if templates directory exists but contains invalid templates.
func initHandlers(services *ServiceContainer) (*Handlers, error) {
	// Parse HTML templates once for handlers that still use *html/template.Template
	var htmlTemplates *template.Template

	// Create shared template registry for handlers using platform/templates.Registry
	templateRegistry := platformTemplates.NewRegistry()

	// Check if templates directory exists
	if info, err := os.Stat("templates"); err == nil && info.IsDir() {
		// Directory exists - parse templates (errors are fatal)
		htmlTemplates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			return nil, err
		}

		if _, err = templateRegistry.GetOrParse("post_detail", "templates/base.html", "templates/post_detail.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("home", "templates/base.html", "templates/home.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("board", "templates/base.html", "templates/board.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("post_create", "templates/base.html", "templates/post_create.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("post_edit", "templates/base.html", "templates/post_edit.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("activity", "templates/base.html", "templates/activity.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("comments", "templates/base.html", "templates/comments.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("settings", "templates/base.html", "templates/settings.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("login", "templates/base.html", "templates/login.html"); err != nil {
			return nil, err
		}
		if _, err = templateRegistry.GetOrParse("register", "templates/base.html", "templates/register.html"); err != nil {
			return nil, err
		}
	}
	// If directory doesn't exist, htmlTemplates remain nil (API-only mode)

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services, templateRegistry),
		User:         userAdapters.NewHTTPHandler(services, templateRegistry),
		Post:         postAdapters.NewHTTPHandler(services, templateRegistry),
		Comment:      commentAdapters.NewHTTPHandler(services, templateRegistry),
		Reaction:     reactionAdapters.NewHTTPHandler(services, templateRegistry),
		Moderation:   moderationAdapters.NewHTTPHandler(services, templateRegistry),
		Notification: notificationAdapters.NewHTTPHandler(services, templateRegistry),
		Templates:    htmlTemplates,
	}, nil
}
