// Package metrics defines the instrumentation metrics for this program.
//
// Metrics are produced by a Prometheus collector that reads a fresh system
// snapshot on every scrape, so /metrics is correct independently of /stats
// traffic and interfaces that disappear stop being reported automatically.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tecnickcom/nurago/pkg/metrics"
	prom "github.com/tecnickcom/nurago/pkg/metrics/prometheus"
	"github.com/tecnickcom/rpistat/internal/sysstat"
)

// Metric names.
const (
	namePrefix      = "rpistat_"
	nameUptime      = namePrefix + "uptime"
	nameMemoryTotal = namePrefix + "memory_total"
	nameMemoryFree  = namePrefix + "memory_free"
	nameLoad1       = namePrefix + "load_1m"
	nameLoad5       = namePrefix + "load_5m"
	nameLoad15      = namePrefix + "load_15m"
	nameTempCPU     = namePrefix + "temperature_cpu"
	nameDiskTotal   = namePrefix + "disk_total"
	nameDiskFree    = namePrefix + "disk_free"
	nameNetworkRx   = namePrefix + "network_rx"
	nameNetworkTx   = namePrefix + "network_tx"
)

// Labels.
const (
	labelPrefix = "rpistat_"
	labelNic    = labelPrefix + "nic" // NIC (Network Interface Card)
)

// StatsGatherer provides a fresh system snapshot on demand.
type StatsGatherer interface {
	Gather() *sysstat.Stats
}

// Client is a Prometheus collector that exposes the system statistics.
type Client struct {
	gatherer StatsGatherer

	descUptime      *prometheus.Desc
	descMemoryTotal *prometheus.Desc
	descMemoryFree  *prometheus.Desc
	descLoad1       *prometheus.Desc
	descLoad5       *prometheus.Desc
	descLoad15      *prometheus.Desc
	descTempCPU     *prometheus.Desc
	descDiskTotal   *prometheus.Desc
	descDiskFree    *prometheus.Desc
	descNetworkRx   *prometheus.Desc
	descNetworkTx   *prometheus.Desc
}

// New creates a new Metrics collector backed by the given stats gatherer.
func New(gatherer StatsGatherer) *Client {
	return &Client{
		gatherer:        gatherer,
		descUptime:      prometheus.NewDesc(nameUptime, "Time elapsed since last system boot, in nanoseconds.", nil, nil),
		descMemoryTotal: prometheus.NewDesc(nameMemoryTotal, "Total available memory in bytes.", nil, nil),
		descMemoryFree:  prometheus.NewDesc(nameMemoryFree, "Total free memory in bytes.", nil, nil),
		descLoad1:       prometheus.NewDesc(nameLoad1, "1 minute CPU load average.", nil, nil),
		descLoad5:       prometheus.NewDesc(nameLoad5, "5 minutes CPU load average.", nil, nil),
		descLoad15:      prometheus.NewDesc(nameLoad15, "15 minutes CPU load average.", nil, nil),
		descTempCPU:     prometheus.NewDesc(nameTempCPU, "CPU Temperature in Celsius Degrees.", nil, nil),
		descDiskTotal:   prometheus.NewDesc(nameDiskTotal, "Total disk size in bytes.", nil, nil),
		descDiskFree:    prometheus.NewDesc(nameDiskFree, "Total free disk space in bytes.", nil, nil),
		descNetworkRx:   prometheus.NewDesc(nameNetworkRx, "Network received bytes.", []string{labelNic}, nil),
		descNetworkTx:   prometheus.NewDesc(nameNetworkTx, "Network transmitted bytes.", []string{labelNic}, nil),
	}
}

// Describe implements prometheus.Collector. It is intentionally side-effect
// free (no system snapshot is taken here) so registration stays cheap.
func (m *Client) Describe(ch chan<- *prometheus.Desc) {
	send := func(desc *prometheus.Desc) { ch <- desc }

	send(m.descUptime)
	send(m.descMemoryTotal)
	send(m.descMemoryFree)
	send(m.descLoad1)
	send(m.descLoad5)
	send(m.descLoad15)
	send(m.descTempCPU)
	send(m.descDiskTotal)
	send(m.descDiskFree)
	send(m.descNetworkRx)
	send(m.descNetworkTx)
}

// Collect implements prometheus.Collector, reading a fresh snapshot per scrape.
func (m *Client) Collect(ch chan<- prometheus.Metric) {
	s := m.gatherer.Gather()

	gauge := func(desc *prometheus.Desc, value float64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labels...)
	}

	gauge(m.descUptime, float64(s.Uptime))
	gauge(m.descMemoryTotal, float64(s.MemoryTotal))
	gauge(m.descMemoryFree, float64(s.MemoryFree))
	gauge(m.descLoad1, s.Load1)
	gauge(m.descLoad5, s.Load5)
	gauge(m.descLoad15, s.Load15)
	gauge(m.descTempCPU, s.TempCPU)
	gauge(m.descDiskTotal, float64(s.DiskTotal))
	gauge(m.descDiskFree, float64(s.DiskFree))

	for _, n := range s.Network {
		gauge(m.descNetworkRx, float64(n.Rx), n.Nic)
		gauge(m.descNetworkTx, float64(n.Tx), n.Nic)
	}
}

// CreateMetricsClientFunc returns the metrics Client with this collector registered.
func (m *Client) CreateMetricsClientFunc() (metrics.Client, error) {
	return prom.New(prom.WithCollector(m)) //nolint:wrapcheck
}
