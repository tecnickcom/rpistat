package httphandler

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tecnickcom/rpistat/internal/metrics"
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

	tests := []struct {
		name  string
		input string
		want  []NetworkStat
	}{
		{
			name:  "physical interfaces parsed, lo and docker0 skipped",
			input: procNetDev,
			want: []NetworkStat{
				{Nic: "eth0", Rx: 5000, Tx: 6000},
				{Nic: "wlan0", Rx: 700, Tx: 800},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "only headers",
			input: "header1\nheader2\n",
			want:  nil,
		},
		{
			name: "line without colon is skipped",
			input: "header1\nheader2\n" +
				"garbage line without a colon\n" +
				"  eth0: 1 2 0 0 0 0 0 0 3 4 0 0 0 0 0 0\n",
			want: []NetworkStat{{Nic: "eth0", Rx: 1, Tx: 3}},
		},
		{
			name: "line with wrong field count is skipped",
			input: "header1\nheader2\n" +
				"  bad0: 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15\n" +
				"  eth0: 9 2 0 0 0 0 0 0 8 4 0 0 0 0 0 0\n",
			want: []NetworkStat{{Nic: "eth0", Rx: 9, Tx: 8}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseNetworkStats(strings.NewReader(tt.input))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExcludedNic(t *testing.T) {
	t.Parallel()

	require.True(t, excludedNic("lo"))
	require.True(t, excludedNic("docker0"))
	require.False(t, excludedNic("eth0"))
	require.False(t, excludedNic("wlan0"))
}

// TestStatsCPUTempFromFile exercises the cpuTemp wrapper end to end against a
// real file, covering the syscall open and delegation to parseCPUTemp.
func TestStatsCPUTempFromFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "temp")
	require.NoError(t, os.WriteFile(path, []byte("48500\n"), 0o600))

	s := &Stats{metric: metrics.New(), fileCPUTemp: path}
	s.cpuTemp()

	require.InDelta(t, 48.5, s.TempCPU, 0.0001)
}

func TestStatsCPUTempMissingFile(t *testing.T) {
	t.Parallel()

	s := &Stats{metric: metrics.New(), fileCPUTemp: filepath.Join(t.TempDir(), "missing")}
	s.cpuTemp()

	require.InDelta(t, 0, s.TempCPU, 0.0001)
}

// TestStatsNetworkFromFile exercises the network wrapper end to end against a
// real file, covering the syscall open and delegation to parseNetworkStats.
func TestStatsNetworkFromFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "dev")
	require.NoError(t, os.WriteFile(path, []byte(procNetDev), 0o600))

	s := &Stats{metric: metrics.New(), fileNetworkStat: path}
	s.network()

	require.Equal(t, []NetworkStat{
		{Nic: "eth0", Rx: 5000, Tx: 6000},
		{Nic: "wlan0", Rx: 700, Tx: 800},
	}, s.Network)
}

func TestStatsNetworkMissingFile(t *testing.T) {
	t.Parallel()

	s := &Stats{metric: metrics.New(), fileNetworkStat: filepath.Join(t.TempDir(), "missing")}
	s.network()

	require.Empty(t, s.Network)
}

// TestNewStatsJSONEncodable verifies the full stats payload always encodes to
// valid JSON (NaN/Inf would make encoding/json fail) and the usage fields stay
// within the documented [0,1] range.
func TestNewStatsJSONEncodable(t *testing.T) {
	t.Parallel()

	s := newStats(metrics.New())

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

// TestStatsZeroTotalsEncode confirms that a payload built from zero totals (so
// usage ratios are guarded to 0) still encodes without a NaN error.
func TestStatsZeroTotalsEncode(t *testing.T) {
	t.Parallel()

	s := &Stats{
		MemoryUsage: usageRatio(0, 0),
		DiskUsage:   usageRatio(0, 0),
	}

	raw, err := json.Marshal(s)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"memory_usage":0`)
	require.Contains(t, string(raw), `"disk_usage":0`)
}
