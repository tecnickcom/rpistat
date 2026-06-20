package sysstat

import "log/slog"

// Option configures a Gatherer.
type Option func(*Gatherer)

// WithLogger sets the logger used to report otherwise-swallowed errors. A nil
// logger keeps the default (a no-op discard logger).
func WithLogger(l *slog.Logger) Option {
	return func(g *Gatherer) {
		if l != nil {
			g.log = l
		}
	}
}

// WithExcludedNics sets the network interfaces to skip. A nil slice keeps the
// defaults; an explicitly empty slice excludes nothing.
func WithExcludedNics(nics []string) Option {
	return func(g *Gatherer) {
		if nics != nil {
			g.excludedNics = toSet(nics)
		}
	}
}

// WithCPUTempFile overrides the CPU temperature source path (used for testing).
func WithCPUTempFile(path string) Option {
	return func(g *Gatherer) { g.fileCPUTemp = path }
}

// WithNetworkStatFile overrides the network statistics source path (used for testing).
func WithNetworkStatFile(path string) Option {
	return func(g *Gatherer) { g.fileNetworkStat = path }
}
