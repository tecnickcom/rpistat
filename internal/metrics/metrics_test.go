package metrics

import (
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
	require.Equal(t, float64(v), i)
}

func TestSetMemoryTotal(t *testing.T) {
	t.Parallel()

	v := uint64(1699)

	m := New()

	m.SetMemoryTotal(v)
	i := testutil.ToFloat64(m.collectorMemoryTotal)
	require.Equal(t, float64(v), i)
}

func TestSetMemoryFree(t *testing.T) {
	t.Parallel()

	v := uint64(1709)

	m := New()

	m.SetMemoryFree(v)
	i := testutil.ToFloat64(m.collectorMemoryFree)
	require.Equal(t, float64(v), i)
}

func TestSetLoad1(t *testing.T) {
	t.Parallel()

	v := float64(1721)

	m := New()

	m.SetLoad1(v)
	i := testutil.ToFloat64(m.collectorLoad1)
	require.Equal(t, v, i)
}

func TestSetLoad5(t *testing.T) {
	t.Parallel()

	v := float64(1723)

	m := New()

	m.SetLoad5(v)
	i := testutil.ToFloat64(m.collectorLoad5)
	require.Equal(t, v, i)
}

func TestSetLoad15(t *testing.T) {
	t.Parallel()

	v := float64(1733)

	m := New()

	m.SetLoad15(v)
	i := testutil.ToFloat64(m.collectorLoad15)
	require.Equal(t, v, i)
}

func TestSetTempCPU(t *testing.T) {
	t.Parallel()

	v := float64(1733)

	m := New()

	m.SetTempCPU(v)
	i := testutil.ToFloat64(m.collectorTempCPU)
	require.Equal(t, v, i)
}

func TestSetDiskTotal(t *testing.T) {
	t.Parallel()

	v := uint64(1741)

	m := New()

	m.SetDiskTotal(v)
	i := testutil.ToFloat64(m.collectorDiskTotal)
	require.Equal(t, float64(v), i)
}

func TestSetDiskFree(t *testing.T) {
	t.Parallel()

	v := uint64(1747)

	m := New()

	m.SetDiskFree(v)
	i := testutil.ToFloat64(m.collectorDiskFree)
	require.Equal(t, float64(v), i)
}

func TestSetNetwork(t *testing.T) {
	t.Parallel()

	rx := uint64(2683)
	tx := uint64(3169)

	m := New()

	m.SetNetwork("test", rx, tx)

	rc := testutil.ToFloat64(m.collectorNetworkRx)
	require.Equal(t, float64(rx), rc)

	tc := testutil.ToFloat64(m.collectorNetworkTx)
	require.Equal(t, float64(tx), tc)
}
