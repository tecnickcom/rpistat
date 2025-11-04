// Package httphandler handles the inbound service requests.
package httphandler

import (
	"context"
	"net/http"

	"github.com/tecnickcom/gogen/pkg/httpserver"
	"github.com/tecnickcom/gogen/pkg/httputil"
	"github.com/tecnickcom/rpistat/internal/metrics"
)

// HTTPHandler is the struct containing all the http handlers.
type HTTPHandler struct {
	metric metrics.Metrics
}

// New creates a new instance of the HTTP handler.
func New(m metrics.Metrics) *HTTPHandler {
	return &HTTPHandler{
		metric: m,
	}
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
	httputil.SendJSON(r.Context(), w, http.StatusOK, newStats(h.metric))
}
