package sysstat

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGathererDefaults(t *testing.T) {
	t.Parallel()

	g := NewGatherer()

	require.NotNil(t, g.log)
	require.Equal(t, defaultFileCPUTemp, g.fileCPUTemp)
	require.Equal(t, defaultFileNetworkStat, g.fileNetworkStat)
	require.Equal(t, toSet(defaultExcludedNics), g.excludedNics)
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	g := NewGatherer(WithLogger(logger))
	require.Same(t, logger, g.log)

	// A nil logger keeps the default non-nil logger.
	dflt := NewGatherer()
	defaultLog := dflt.log
	WithLogger(nil)(dflt)
	require.Same(t, defaultLog, dflt.log)
	require.NotNil(t, dflt.log)
}

func TestWithExcludedNics(t *testing.T) {
	t.Parallel()

	// Explicit list replaces the defaults.
	g := NewGatherer(WithExcludedNics([]string{"eth1", "br0"}))
	require.Equal(t, toSet([]string{"eth1", "br0"}), g.excludedNics)

	// Nil keeps the defaults.
	g = NewGatherer(WithExcludedNics(nil))
	require.Equal(t, toSet(defaultExcludedNics), g.excludedNics)

	// An explicitly empty slice excludes nothing.
	g = NewGatherer(WithExcludedNics([]string{}))
	require.Empty(t, g.excludedNics)
}

func TestWithCPUTempFile(t *testing.T) {
	t.Parallel()

	g := NewGatherer(WithCPUTempFile("/custom/temp"))
	require.Equal(t, "/custom/temp", g.fileCPUTemp)
}

func TestWithNetworkStatFile(t *testing.T) {
	t.Parallel()

	g := NewGatherer(WithNetworkStatFile("/custom/dev"))
	require.Equal(t, "/custom/dev", g.fileNetworkStat)
}
