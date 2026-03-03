// SHARED ADAPTER - JSON HTTP Utilities
// Package httpjson provides shared JSON request/response utilities for module HTTP handlers.
package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
)

// ErrEmptyBody is returned when the request body is nil.
var ErrEmptyBody = errors.New("request body is empty")

// WriteJSON writes a JSON response with the given status code and data.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// WriteError writes a JSON error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// ParseJSON decodes a JSON request body into the given target.
// It validates that the Content-Type is application/json and rejects
// unknown fields. Returns an error if decoding fails.
func ParseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return ErrEmptyBody
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return fmt.Errorf("content type is not application/json")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
