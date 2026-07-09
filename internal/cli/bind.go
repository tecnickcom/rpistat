package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/tecnickcom/gogen/pkg/bootstrap"
	"github.com/tecnickcom/gogen/pkg/healthcheck"
	"github.com/tecnickcom/gogen/pkg/httpclient"
	"github.com/tecnickcom/gogen/pkg/httpserver"
	"github.com/tecnickcom/gogen/pkg/httputil"
	"github.com/tecnickcom/gogen/pkg/httputil/jsendx"
	"github.com/tecnickcom/gogen/pkg/ipify"
	"github.com/tecnickcom/gogen/pkg/metrics"
	"github.com/tecnickcom/gogen/pkg/traceid"
	"github.com/tecnickcom/rpistat/internal/httphandler"
	"github.com/tecnickcom/rpistat/internal/sysstat"
)

// bind is the entry point of the service, this is where the wiring of all
// components happens. The wiring is split into small, named helpers
// (newIpifyClient, bindServiceHandlers) so each concern stays readable and
// independently testable.
func bind(cfg *appConfig, appInfo *jsendx.AppInfo, gatherer *sysstat.Gatherer, wg *sync.WaitGroup, sc chan struct{}) bootstrap.BindFunc {
	return func(ctx context.Context, l *slog.Logger, m metrics.Client) error {
		jsx := jsendx.NewJSXResp(httputil.NewHTTPResp(l))

		// Common outbound HTTP client options shared by every external client
		// (structured logging, trace propagation, and metrics instrumentation).
		// This base is built once and reused: each client constructor appends its
		// own timeout on top of it (see newIpifyClient).
		httpClientOpts := []httpclient.Option{
			httpclient.WithLogger(l),
			httpclient.WithRoundTripper(m.InstrumentRoundTripper),
			httpclient.WithTraceIDHeaderName(traceid.DefaultHeader),
			httpclient.WithComponent(appInfo.ProgramName),
		}

		// ipify is used only as a diagnostic (the monitoring /ip route); it is
		// intentionally not part of the health checks.
		ipifyClient, err := newIpifyClient(cfg, httpClientOpts)
		if err != nil {
			return err
		}

		serviceBinder, statusHandler := bindServiceHandlers(cfg, appInfo, jsx, l, gatherer)

		middleware := func(args httpserver.MiddlewareArgs, next http.Handler) http.Handler {
			return m.InstrumentHandler(args.Path, next.ServeHTTP)
		}

		// start monitoring server
		httpMonitoringOpts := []httpserver.Option{
			httpserver.WithLogger(l),
			httpserver.WithServerAddr(cfg.Servers.Monitoring.Address),
			httpserver.WithRequestTimeout(time.Duration(cfg.Servers.Monitoring.Timeout) * time.Second),
			httpserver.WithMetricsHandlerFunc(m.MetricsHandlerFunc()),
			httpserver.WithTraceIDHeaderName(traceid.DefaultHeader),
			httpserver.WithMiddlewareFn(middleware),
			httpserver.WithNotFoundHandlerFunc(jsx.DefaultNotFoundHandlerFunc(appInfo)),
			httpserver.WithMethodNotAllowedHandlerFunc(jsx.DefaultMethodNotAllowedHandlerFunc(appInfo)),
			httpserver.WithPanicHandlerFunc(jsx.DefaultPanicHandlerFunc(appInfo)),
			httpserver.WithEnableAllDefaultRoutes(),
			httpserver.WithIndexHandlerFunc(jsx.DefaultIndexHandler(appInfo)),
			httpserver.WithIPHandlerFunc(jsx.DefaultIPHandler(appInfo, ipifyClient.GetPublicIP)),
			httpserver.WithPingHandlerFunc(jsx.DefaultPingHandler(appInfo)),
			httpserver.WithStatusHandlerFunc(statusHandler),
			httpserver.WithShutdownWaitGroup(wg),
			httpserver.WithShutdownSignalChan(sc),
		}

		httpMonitoringServer, err := httpserver.New(ctx, serviceBinder, httpMonitoringOpts...)
		if err != nil {
			return fmt.Errorf("error creating monitoring HTTP server: %w", err)
		}

		httpMonitoringServer.StartServer()

		return nil
	}
}

// newIpifyClient builds the ipify client used by the monitoring server's /ip
// diagnostic route.
//
// It takes the shared base HTTP client options and appends the ipify-specific
// timeout, keeping every outbound client consistent (logging, tracing, metrics)
// while each keeps its own timeout. The base slice is left untouched: it has no
// spare capacity, so the append always copies rather than mutating the caller's
// slice, which keeps it safe to reuse for the next client.
func newIpifyClient(cfg *appConfig, baseHTTPClientOpts []httpclient.Option) (*ipify.Client, error) {
	ipifyTimeout := time.Duration(cfg.Clients.Ipify.Timeout) * time.Second

	ipifyHTTPClient := httpclient.New(append(
		baseHTTPClientOpts,
		httpclient.WithTimeout(ipifyTimeout),
	)...)

	ipifyClient, err := ipify.New(
		ipify.WithHTTPClient(ipifyHTTPClient),
		ipify.WithTimeout(ipifyTimeout),
		ipify.WithURL(cfg.Clients.Ipify.Address),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build ipify client: %w", err)
	}

	return ipifyClient, nil
}

// bindServiceHandlers wires the service binder together with the status handler.
//
// When the service is disabled it returns a no-op binder and the default status
// handler. When enabled it attaches the real system-statistics handler and
// upgrades the status handler to a health check.
func bindServiceHandlers(
	cfg *appConfig,
	appInfo *jsendx.AppInfo,
	jsx *jsendx.JSXResp,
	l *slog.Logger,
	gatherer *sysstat.Gatherer,
) (httpserver.Binder, http.HandlerFunc) {
	if !cfg.Enabled {
		return httpserver.NopBinder(), jsx.DefaultStatusHandler(appInfo)
	}

	serviceBinder := httphandler.New(l, gatherer)

	// override the default status handler with a health check
	healthCheckHandler := healthcheck.NewHandler(
		[]healthcheck.HealthCheck{
			// healthcheck.New("<ID>", < HANDLER >),
			// healthcheck.NewWithTimeout("<ID>", <HANDLER>, <TIMEOUT>),
		},
		healthcheck.WithLogger(l),
		healthcheck.WithResultWriter(jsx.HealthCheckResultWriter(appInfo)),
	)

	return serviceBinder, healthCheckHandler.ServeHTTP
}
