package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	m := New()
	require.NotNil(t, m, "Metrics should not be nil")
}

func TestCreateMetricsClientFunc(t *testing.T) {
	t.Parallel()

	m := New()
	c, err := m.CreateMetricsClientFunc()
	require.NoError(t, err, "CreateMetricsClientFunc() unexpected error = %v", err)
	require.NotNil(t, c, "metrics.Client should not be nil")
}

func TestSetUptime(t *testing.T) {
	t.Parallel()

	v := 3 * time.Second

	m := New()

	m.SetUptime(v)
	i := testutil.ToFloat64(m.collectorUptime)
	require.InDelta(t, float64(v), i, 0.001)
}

func TestSetMemoryTotal(t *testing.T) {
	t.Parallel()

	v := uint64(1699)

	m := New()

	m.SetMemoryTotal(v)
	i := testutil.ToFloat64(m.collectorMemoryTotal)
	require.InDelta(t, float64(v), i, 0.001)
}

func TestSetMemoryFree(t *testing.T) {
	t.Parallel()

	v := uint64(1709)

	m := New()

	m.SetMemoryFree(v)
	i := testutil.ToFloat64(m.collectorMemoryFree)
	require.InDelta(t, float64(v), i, 0.001)
}

func TestSetLoad1(t *testing.T) {
	t.Parallel()

	v := float64(1721)

	m := New()

	m.SetLoad1(v)
	i := testutil.ToFloat64(m.collectorLoad1)
	require.InDelta(t, v, i, 0.001)
}

func TestSetLoad5(t *testing.T) {
	t.Parallel()

	v := float64(1723)

	m := New()

	m.SetLoad5(v)
	i := testutil.ToFloat64(m.collectorLoad5)
	require.InDelta(t, v, i, 0.001)
}

func TestSetLoad15(t *testing.T) {
	t.Parallel()

	v := float64(1733)

	m := New()

	m.SetLoad15(v)
	i := testutil.ToFloat64(m.collectorLoad15)
	require.InDelta(t, v, i, 0.001)
}

func TestSetTempCPU(t *testing.T) {
	t.Parallel()

	v := float64(1733)

	m := New()

	m.SetTempCPU(v)
	i := testutil.ToFloat64(m.collectorTempCPU)
	require.InDelta(t, v, i, 0.001)
}

func TestSetDiskTotal(t *testing.T) {
	t.Parallel()

	v := uint64(1741)

	m := New()

	m.SetDiskTotal(v)
	i := testutil.ToFloat64(m.collectorDiskTotal)
	require.InDelta(t, float64(v), i, 0.001)
}

func TestSetDiskFree(t *testing.T) {
	t.Parallel()

	v := uint64(1747)

	m := New()

	m.SetDiskFree(v)
	i := testutil.ToFloat64(m.collectorDiskFree)
	require.InDelta(t, float64(v), i, 0.001)
}

func TestSetNetwork(t *testing.T) {
	t.Parallel()

	rx := uint64(2683)
	tx := uint64(3169)

	m := New()

	m.SetNetwork("test", rx, tx)

	rc := testutil.ToFloat64(m.collectorNetworkRx)
	require.InDelta(t, float64(rx), rc, 0.001)

	tc := testutil.ToFloat64(m.collectorNetworkTx)
	require.InDelta(t, float64(tx), tc, 0.001)
}

// TestNetworkHelpText guards against regressing the (previously swapped) help
// text of the network gauges: rx must read "received", tx "transmitted".
func TestNetworkHelpText(t *testing.T) {
	t.Parallel()

	m := New()
	m.SetNetwork("eth0", 100, 200)

	wantRx := `# HELP rpistat_network_rx Network received bytes.
# TYPE rpistat_network_rx gauge
rpistat_network_rx{rpistat_nic="eth0"} 100
`
	err := testutil.CollectAndCompare(m.collectorNetworkRx, strings.NewReader(wantRx), nameNetworkRx)
	require.NoError(t, err)

	wantTx := `# HELP rpistat_network_tx Network transmitted bytes.
# TYPE rpistat_network_tx gauge
rpistat_network_tx{rpistat_nic="eth0"} 200
`
	err = testutil.CollectAndCompare(m.collectorNetworkTx, strings.NewReader(wantTx), nameNetworkTx)
	require.NoError(t, err)
}
