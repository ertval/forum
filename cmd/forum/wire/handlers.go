// INPUT ADAPTERS - HTTP Handler Initialization
package wire

import (
	"html/template"
	"os"

	"forum/internal/platform/config"
	logger "forum/internal/platform/logger"
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
func initHandlers(services *ServiceContainer, cfg *config.Config) (*Handlers, error) {
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
	}
	// If directory doesn't exist, htmlTemplates remain nil (API-only mode)

	// Cookie security is determined by config (from environment)
	// In production, cfg.Session.Secure should be true
	secureCookies := cfg.Session.Secure
	sessionCookieName := cfg.Session.CookieName

	return &Handlers{
		Auth:         authAdapters.NewHTTPHandler(services, htmlTemplates, secureCookies, sessionCookieName),
		User:         userAdapters.NewHTTPHandler(services, templateRegistry),
		Post:         postAdapters.NewHTTPHandler(services, templateRegistry, logger.New(logger.InfoLevel, os.Stderr)),
		Comment:      commentAdapters.NewHTTPHandler(services, templateRegistry),
		Reaction:     reactionAdapters.NewHTTPHandler(services, htmlTemplates),
		Moderation:   moderationAdapters.NewHTTPHandler(services, htmlTemplates),
		Notification: notificationAdapters.NewHTTPHandler(services, htmlTemplates),
		Templates:    htmlTemplates,
	}, nil
}
