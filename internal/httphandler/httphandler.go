// Package httphandler handles the inbound service requests.
package httphandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tecnickcom/gogen/pkg/httpserver"
	"github.com/tecnickcom/gogen/pkg/httputil"
	"github.com/tecnickcom/rpistat/internal/sysstat"
)

// StatsGatherer provides a fresh system snapshot on demand.
type StatsGatherer interface {
	Gather() *sysstat.Stats
}

// HTTPHandler is the struct containing all the http handlers.
type HTTPHandler struct {
	httpres  *httputil.HTTPResp
	gatherer StatsGatherer
}

// New creates a new instance of the HTTP handler.
func New(l *slog.Logger, gatherer StatsGatherer) *HTTPHandler {
	return &HTTPHandler{
		httpres:  httputil.NewHTTPResp(l),
		gatherer: gatherer,
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
	h.httpres.SendJSON(r.Context(), w, http.StatusOK, h.gatherer.Gather())
}
