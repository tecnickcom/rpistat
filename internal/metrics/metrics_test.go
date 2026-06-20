package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tecnickcom/rpistat/internal/sysstat"
)

type fakeGatherer struct {
	stats *sysstat.Stats
}

func (f fakeGatherer) Gather() *sysstat.Stats { return f.stats }

func testStats() *sysstat.Stats {
	return &sysstat.Stats{
		Uptime:      5, // nanoseconds; small value keeps the text exposition tidy
		MemoryTotal: 2000,
		MemoryFree:  500,
		Load1:       1.5,
		Load5:       2.5,
		Load15:      3.5,
		TempCPU:     48.5,
		DiskTotal:   10000,
		DiskFree:    4000,
		Network: []sysstat.NetworkStat{
			{Nic: "eth0", Rx: 100, Tx: 200},
			{Nic: "wlan0", Rx: 300, Tx: 400},
		},
	}
}

func newTestClient() *Client {
	return New(fakeGatherer{stats: testStats()})
}

func TestNew(t *testing.T) {
	t.Parallel()

	m := newTestClient()
	require.NotNil(t, m, "Metrics should not be nil")
}

func TestCreateMetricsClientFunc(t *testing.T) {
	t.Parallel()

	m := newTestClient()
	c, err := m.CreateMetricsClientFunc()
	require.NoError(t, err, "CreateMetricsClientFunc() unexpected error = %v", err)
	require.NotNil(t, c, "metrics.Client should not be nil")
}

// TestCollectScalars checks representative scalar metrics, including the uptime
// help text which must state the nanosecond unit.
func TestCollectScalars(t *testing.T) {
	t.Parallel()

	m := newTestClient()

	want := `# HELP rpistat_uptime Time elapsed since last system boot, in nanoseconds.
# TYPE rpistat_uptime gauge
rpistat_uptime 5
`
	require.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(want), nameUptime))

	want = `# HELP rpistat_memory_total Total available memory in bytes.
# TYPE rpistat_memory_total gauge
rpistat_memory_total 2000
`
	require.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(want), nameMemoryTotal))

	want = `# HELP rpistat_temperature_cpu CPU Temperature in Celsius Degrees.
# TYPE rpistat_temperature_cpu gauge
rpistat_temperature_cpu 48.5
`
	require.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(want), nameTempCPU))
}

// TestCollectNetworkHelpText guards against regressing the (previously swapped)
// help text of the network gauges and verifies per-NIC labels and values.
func TestCollectNetworkHelpText(t *testing.T) {
	t.Parallel()

	m := newTestClient()

	wantRx := `# HELP rpistat_network_rx Network received bytes.
# TYPE rpistat_network_rx gauge
rpistat_network_rx{rpistat_nic="eth0"} 100
rpistat_network_rx{rpistat_nic="wlan0"} 300
`
	require.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(wantRx), nameNetworkRx))

	wantTx := `# HELP rpistat_network_tx Network transmitted bytes.
# TYPE rpistat_network_tx gauge
rpistat_network_tx{rpistat_nic="eth0"} 200
rpistat_network_tx{rpistat_nic="wlan0"} 400
`
	require.NoError(t, testutil.CollectAndCompare(m, strings.NewReader(wantTx), nameNetworkTx))
}

// TestCollectSeriesCount verifies the total number of emitted series: 9 scalar
// metrics plus rx/tx per network interface.
func TestCollectSeriesCount(t *testing.T) {
	t.Parallel()

	m := newTestClient()
	require.Equal(t, 9+2*2, testutil.CollectAndCount(m))
}

// TestCollectNoNetwork verifies that interfaces that are absent emit no series,
// which is what makes stale NICs disappear on the next scrape.
func TestCollectNoNetwork(t *testing.T) {
	t.Parallel()

	stats := testStats()
	stats.Network = nil
	m := New(fakeGatherer{stats: stats})

	require.Equal(t, 9, testutil.CollectAndCount(m))
	require.Zero(t, testutil.CollectAndCount(m, nameNetworkRx))
}
