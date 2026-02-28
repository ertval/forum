// Package templates provides template caching and rendering utilities.
// It eliminates redundant template parsing by caching parsed template combinations.
package templates

import (
	"html/template"
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
