// Package httphandler handles the inbound service requests.
package httphandler

import (
	"context"
	"net/http"

	"github.com/Vonage/gosrvlib/pkg/httpserver"
	"github.com/Vonage/gosrvlib/pkg/httputil"
)

// Service is the interface representing the business logic of the service.
type Service interface {
	// NOTE
	// This is a sample Service interface.
	// It is meant to demonstrate where the business logic of a service should reside.
	// It adds the capability of mocking the HTTP Handler independently from the rest of the code.
	// Add service functions here.
}

// New creates a new instance of the HTTP handler.
func New(s Service) *HTTPHandler {
	return &HTTPHandler{
		service: s,
	}
}

// HTTPHandler is the struct containing all the http handlers.
type HTTPHandler struct {
	service Service
}

// BindHTTP implements the function to bind the handler to a server.
func (h *HTTPHandler) BindHTTP(_ context.Context) []httpserver.Route {
	return []httpserver.Route{
		{
			Method:      http.MethodGet,
			Path:        "/stats",
			Description: "Returns system statistics",
			Handler:     h.handleStats,
		},
	}
}

func (h *HTTPHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	httputil.SendJSON(r.Context(), w, http.StatusOK, newStats())
}
