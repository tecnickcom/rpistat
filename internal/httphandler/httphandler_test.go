package httphandler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tecnickcom/rpistat/internal/metrics"
)

func TestNew(t *testing.T) {
	t.Parallel()

	hh := New(nil, nil)
	require.NotNil(t, hh)
}

func TestHTTPHandler_BindHTTP(t *testing.T) {
	t.Parallel()

	mtr := metrics.New()
	h := &HTTPHandler{nil, mtr}
	got := h.BindHTTP(t.Context())
	require.Len(t, got, 1)
}

func TestHTTPHandler_handleStats(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	mtr := metrics.New()

	hh := New(nil, mtr)
	require.NotNil(t, hh)

	hh.handleStats(rr, req)

	resp := rr.Result()
	require.NotNil(t, resp)

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err, "error closing resp.Body")
	}()

	body, _ := io.ReadAll(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
	require.NotEmpty(t, string(body))
}
