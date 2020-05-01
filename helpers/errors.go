package helpers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// ErrorResponse is the standard ContracterAPI error format
type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// Render sets the error status code.
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Status)
	return nil
}

// ErrBadRequest returns a 400 status code response.
func ErrBadRequest(err error) render.Renderer {
	return &ErrorResponse{Message: err.Error(), Status: 400}
}

// ErrUnauthorized returns a 401 status code response.
func ErrUnauthorized(err error) render.Renderer {
	return &ErrorResponse{Message: err.Error(), Status: 401}
}

// ErrNotFound returns a 404 status code response.
func ErrNotFound(resource string, key string) render.Renderer {
	m := fmt.Sprintf("%v (%v) not found", resource, key)
	return &ErrorResponse{Message: m, Status: 404}
}

// ErrConflict returns a 409 status code response.
func ErrConflict(err error) render.Renderer {
	return &ErrorResponse{Message: err.Error(), Status: 409}
}
