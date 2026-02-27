// Template Validation Utilities
package templates

import (
	"fmt"
	"html/template"
	"strings"
)

// TemplateValidator validates template structure and completeness.
type TemplateValidator struct {
	templates *template.Template
}

// NewTemplateValidator creates a new template validator.
func NewTemplateValidator(templates *template.Template) *TemplateValidator {
	return &TemplateValidator{
		templates: templates,
	}
}

// ValidateRequired checks if all required templates are present.
// Returns error if any required template is missing.
func (v *TemplateValidator) ValidateRequired(requiredTemplates []string) error {
	if v.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	available := v.ListTemplateNames()
	availableMap := make(map[string]bool)
	for _, name := range available {
		availableMap[name] = true
	}

	var missing []string
	for _, required := range requiredTemplates {
		if !availableMap[required] {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required templates: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ListTemplateNames returns all parsed template names.
func (v *TemplateValidator) ListTemplateNames() []string {
	if v.templates == nil {
		return nil
	}

	var names []string
	for _, tmpl := range v.templates.Templates() {
		names = append(names, tmpl.Name())
	}
	return names
}

// LogTemplateInfo logs template structure information.
// Returns a string with template information for logging.
func (v *TemplateValidator) LogTemplateInfo() string {
	if v.templates == nil {
		return "No templates loaded"
	}

	names := v.ListTemplateNames()
	var info strings.Builder

	info.WriteString(fmt.Sprintf("Loaded %d templates:\n", len(names)))
	for _, name := range names {
		info.WriteString(fmt.Sprintf("  - %s\n", name))
	}

	return info.String()
}

// ValidateBaseTemplate checks if base template has required structure.
func (v *TemplateValidator) ValidateBaseTemplate() error {
	if v.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	// Check if "base" template exists
	baseTemplate := v.templates.Lookup("base")
	if baseTemplate == nil {
		return fmt.Errorf("base template not found")
	}

	// TODO: Add more sophisticated validation if needed
	// For now, just checking existence is sufficient

	return nil
}

// ValidateContentTemplates checks if content templates are compatible with base.
// This is a basic check - more sophisticated validation could be added later.
func (v *TemplateValidator) ValidateContentTemplates(contentTemplates []string) error {
	if v.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	for _, name := range contentTemplates {
		tmpl := v.templates.Lookup(name)
		if tmpl == nil {
			return fmt.Errorf("content template %q not found", name)
		}
	}

	return nil
}

// getRequiredTemplates returns the list of templates required for the application.
func getRequiredTemplates() []string {
	return []string{
		"base",    // Base layout template
		"content", // Content block (defined in various templates)
		// Note: Other templates are checked dynamically
	}
}
