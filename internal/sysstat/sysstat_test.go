package sysstat

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsageRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		used  uint64
		total uint64
		want  float64
	}{
		{name: "half", used: 50, total: 100, want: 0.5},
		{name: "full", used: 100, total: 100, want: 1},
		{name: "empty", used: 0, total: 100, want: 0},
		{name: "zero total returns zero (no NaN)", used: 10, total: 0, want: 0},
		{name: "both zero returns zero (no NaN)", used: 0, total: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := usageRatio(tt.used, tt.total)
			require.False(t, math.IsNaN(got), "ratio must never be NaN")
			require.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestParseCPUTemp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{name: "valid", input: "52123\n", want: 52.123},
		{name: "valid no newline", input: "40000", want: 40},
		{name: "zero", input: "0\n", want: 0},
		{name: "empty", input: "", wantErr: true},
		{name: "non numeric", input: "hot\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCPUTemp(strings.NewReader(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

// procNetDev is a representative /proc/net/dev sample: two header lines
// followed by one line per interface. Receive bytes are column 0 and transmit
// bytes column 8 of the 16 fields that follow the interface name.
const procNetDev = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:  100  1 0 0 0 0 0 0  100  1 0 0 0 0 0 0
docker0:  200  2 0 0 0 0 0 0  300  3 0 0 0 0 0 0
  eth0: 5000 50 0 0 0 0 0 0 6000 60 0 0 0 0 0 0
 wlan0:  700  7 0 0 0 0 0 0  800  8 0 0 0 0 0 0
`

func TestParseNetworkStats(t *testing.T) {
	t.Parallel()

	defaultExcluded := toSet(defaultExcludedNics)

	tests := []struct {
		name     string
		input    string
		excluded map[string]struct{}
		want     []NetworkStat
	}{
		{
			name:     "physical interfaces parsed, lo and docker0 skipped",
			input:    procNetDev,
			excluded: defaultExcluded,
			want: []NetworkStat{
				{Nic: "eth0", Rx: 5000, Tx: 6000},
				{Nic: "wlan0", Rx: 700, Tx: 800},
			},
		},
		{
			name:     "custom exclusion list skips eth0",
			input:    procNetDev,
			excluded: toSet([]string{"eth0"}),
			want: []NetworkStat{
				{Nic: "lo", Rx: 100, Tx: 100},
				{Nic: "docker0", Rx: 200, Tx: 300},
				{Nic: "wlan0", Rx: 700, Tx: 800},
			},
		},
		{
			name:     "empty exclusion list keeps everything",
			input:    procNetDev,
			excluded: toSet(nil),
			want: []NetworkStat{
				{Nic: "lo", Rx: 100, Tx: 100},
				{Nic: "docker0", Rx: 200, Tx: 300},
				{Nic: "eth0", Rx: 5000, Tx: 6000},
				{Nic: "wlan0", Rx: 700, Tx: 800},
			},
		},
		{
			name:     "empty input",
			input:    "",
			excluded: defaultExcluded,
			want:     nil,
		},
		{
			name:     "only headers",
			input:    "header1\nheader2\n",
			excluded: defaultExcluded,
			want:     nil,
		},
		{
			name:     "line without colon is skipped",
			input:    "header1\nheader2\ngarbage line without a colon\n  eth0: 1 2 0 0 0 0 0 0 3 4 0 0 0 0 0 0\n",
			excluded: defaultExcluded,
			want:     []NetworkStat{{Nic: "eth0", Rx: 1, Tx: 3}},
		},
		{
			name:     "line with wrong field count is skipped",
			input:    "header1\nheader2\n  bad0: 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15\n  eth0: 9 2 0 0 0 0 0 0 8 4 0 0 0 0 0 0\n",
			excluded: defaultExcluded,
			want:     []NetworkStat{{Nic: "eth0", Rx: 9, Tx: 8}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseNetworkStats(strings.NewReader(tt.input), tt.excluded)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestGatherCPUTempFromFile exercises the cpuTemp path end to end against a
// real file, covering the syscall open and delegation to parseCPUTemp.
func TestGatherCPUTempFromFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "temp")
	require.NoError(t, os.WriteFile(path, []byte("48500\n"), 0o600))

	g := NewGatherer(WithCPUTempFile(path))
	s := g.Gather()

	require.InDelta(t, 48.5, s.TempCPU, 0.0001)
}

// TestGatherNetworkFromFile exercises the network path end to end against a
// real file, covering the syscall open, exclusions, and parsing.
func TestGatherNetworkFromFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "dev")
	require.NoError(t, os.WriteFile(path, []byte(procNetDev), 0o600))

	g := NewGatherer(WithNetworkStatFile(path), WithExcludedNics([]string{"lo", "docker0"}))
	s := g.Gather()

	require.Equal(t, []NetworkStat{
		{Nic: "eth0", Rx: 5000, Tx: 6000},
		{Nic: "wlan0", Rx: 700, Tx: 800},
	}, s.Network)
}

// TestGatherCPUTempParseError covers the path where the file opens but its
// content cannot be parsed: the temperature stays zero and the error is logged.
func TestGatherCPUTempParseError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "temp")
	require.NoError(t, os.WriteFile(path, []byte("not-a-number\n"), 0o600))

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	g := NewGatherer(WithLogger(logger), WithCPUTempFile(path))
	s := g.Gather()

	require.InDelta(t, 0, s.TempCPU, 0.0001)
	require.Contains(t, buf.String(), "failed parsing CPU temperature")
}

// TestGatherJSONEncodable verifies the gathered snapshot always encodes to
// valid JSON (NaN/Inf would make encoding/json fail) and the usage fields stay
// within the documented [0,1] range.
func TestGatherJSONEncodable(t *testing.T) {
	t.Parallel()

	s := NewGatherer().Gather()

	raw, err := json.Marshal(s)
	require.NoError(t, err)

	var decoded Stats

	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.False(t, math.IsNaN(s.MemoryUsage))
	require.False(t, math.IsNaN(s.DiskUsage))
	require.GreaterOrEqual(t, s.MemoryUsage, 0.0)
	require.LessOrEqual(t, s.MemoryUsage, 1.0)
	require.GreaterOrEqual(t, s.DiskUsage, 0.0)
	require.LessOrEqual(t, s.DiskUsage, 1.0)
}

// TestGatherLogsErrors verifies that otherwise-swallowed collection errors are
// logged at debug level when a logger is provided.
func TestGatherLogsErrors(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	missingDir := t.TempDir()
	g := NewGatherer(
		WithLogger(logger),
		WithCPUTempFile(filepath.Join(missingDir, "missing-temp")),
		WithNetworkStatFile(filepath.Join(missingDir, "missing-dev")),
	)

	g.Gather()

	out := buf.String()
	require.Contains(t, out, "failed opening CPU temperature file")
	require.Contains(t, out, "failed opening network stats file")
}
