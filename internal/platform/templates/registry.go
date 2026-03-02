// Package templates provides template caching and rendering utilities.
// It eliminates redundant template parsing by caching parsed template combinations.
package templates

import (
	"fmt"
	"html/template"
	"io"
	"sync"
)

// global is the singleton template registry instance.
var global = NewRegistry()

// Registry caches parsed template combinations to avoid repeated disk I/O.
// It is safe for concurrent use.
type Registry struct {
	mu    sync.RWMutex
	cache map[string]*template.Template
}

// NewRegistry creates a new template registry.
func NewRegistry() *Registry {
	return &Registry{
		cache: make(map[string]*template.Template),
	}
}

// GetOrParse retrieves a cached template by key, or parses and caches new templates.
// The key should uniquely identify the set of template files (e.g., "base+home").
// If parsing fails, the error is returned and nothing is cached.
func (r *Registry) GetOrParse(key string, files ...string) (*template.Template, error) {
	// Fast path: check cache with read lock
	r.mu.RLock()
	tmpl, ok := r.cache[key]
	r.mu.RUnlock()
	if ok {
		return tmpl, nil
	}

	// Slow path: parse and cache with write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if tmpl, ok := r.cache[key]; ok {
		return tmpl, nil
	}

	// Parse template files
	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	r.cache[key] = tmpl
	return tmpl, nil
}

// Get is a convenience function that uses the global registry.
// It parses and caches templates on first access - subsequent calls return cached templates.
// Example usage: templates.Get("home", "templates/base.html", "templates/home.html")
func Get(key string, files ...string) (*template.Template, error) {
	return global.GetOrParse(key, files...)
}

// Lookup retrieves a cached template by key. Returns nil if not found or not yet parsed.
// Use GetOrParse to ensure the template is parsed before calling Lookup.
func (r *Registry) Lookup(key string) *template.Template {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cache[key]
}

// TemplateEntry defines a template to load: key name and file paths.
type TemplateEntry struct {
	Key   string
	Files []string
}

// LoadAll parses and caches multiple templates in one call.
// Returns an error on the first template that fails to parse.
func (r *Registry) LoadAll(entries []TemplateEntry) error {
	for _, e := range entries {
		if _, err := r.GetOrParse(e.Key, e.Files...); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTemplate looks up a cached template by key and executes the named
// template definition within it. This is a convenience method that combines
// Lookup + ExecuteTemplate for callers that don't need the raw *template.Template.
func (r *Registry) ExecuteTemplate(w io.Writer, key string, name string, data interface{}) error {
	tmpl := r.Lookup(key)
	if tmpl == nil {
		return fmt.Errorf("template %q not found in registry", key)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}
