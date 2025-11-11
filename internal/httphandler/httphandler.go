// Package httphandler handles the inbound service requests.
package httphandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tecnickcom/gogen/pkg/httpserver"
	"github.com/tecnickcom/gogen/pkg/httputil"
	"github.com/tecnickcom/rpistat/internal/metrics"
)

// HTTPHandler is the struct containing all the http handlers.
type HTTPHandler struct {
	httpres *httputil.HTTPResp
	metric  metrics.Metrics
}

// New creates a new instance of the HTTP handler.
func New(l *slog.Logger, m metrics.Metrics) *HTTPHandler {
	return &HTTPHandler{
		httpres: httputil.NewHTTPResp(l),
		metric:  m,
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
	h.httpres.SendJSON(r.Context(), w, http.StatusOK, newStats(h.metric))
}
