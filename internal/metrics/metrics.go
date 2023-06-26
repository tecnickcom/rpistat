// Package metrics defines the instrumentation metrics for this program.
package metrics

import (
	"time"

	"github.com/Vonage/gosrvlib/pkg/metrics"
	prom "github.com/Vonage/gosrvlib/pkg/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
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

// Metrics is the interface for the custom metrics.
//
//nolint:interfacebloat
type Metrics interface {
	CreateMetricsClientFunc() (metrics.Client, error)
	SetUptime(v time.Duration)
	SetMemoryTotal(v uint64)
	SetMemoryFree(v uint64)
	SetLoad1(v float64)
	SetLoad5(v float64)
	SetLoad15(v float64)
	SetTempCPU(v float64)
	SetDiskTotal(v uint64)
	SetDiskFree(v uint64)
	SetNetwork(nic string, rx uint64, tx uint64)
}

// Client groups the custom collectors to be shared with other packages.
type Client struct {
	collectorUptime      prometheus.Gauge
	collectorMemoryTotal prometheus.Gauge
	collectorMemoryFree  prometheus.Gauge
	collectorLoad1       prometheus.Gauge
	collectorLoad5       prometheus.Gauge
	collectorLoad15      prometheus.Gauge
	collectorTempCPU     prometheus.Gauge
	collectorDiskTotal   prometheus.Gauge
	collectorDiskFree    prometheus.Gauge
	collectorNetworkRx   *prometheus.GaugeVec
	collectorNetworkTx   *prometheus.GaugeVec
}

// New creates a new Metrics instance.
//
//nolint:promlinter
func New() *Client {
	return &Client{
		collectorUptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameUptime,
				Help: "Time elapsed since last system boot.",
			},
		),
		collectorMemoryTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameMemoryTotal,
				Help: "Total available memory in bytes.",
			},
		),
		collectorMemoryFree: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameMemoryFree,
				Help: "Total free memory in bytes.",
			},
		),
		collectorLoad1: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameLoad1,
				Help: "1 minute CPU load average.",
			},
		),
		collectorLoad5: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameLoad5,
				Help: "5 minutes CPU load average.",
			},
		),
		collectorLoad15: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameLoad15,
				Help: "15 minutes CPU load average.",
			},
		),
		collectorTempCPU: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameTempCPU,
				Help: "CPU Temperature in Celsius Degrees.",
			},
		),
		collectorDiskTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameDiskTotal,
				Help: "Total disk size in bytes.",
			},
		),
		collectorDiskFree: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: nameDiskFree,
				Help: "Total free disk space in bytes",
			},
		),
		collectorNetworkRx: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: nameNetworkRx,
				Help: "Network transmitted bytes.",
			},
			[]string{labelNic},
		),
		collectorNetworkTx: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: nameNetworkTx,
				Help: "Network of received bytes.",
			},
			[]string{labelNic},
		),
	}
}

// CreateMetricsClientFunc returns the metrics Client.
func (m *Client) CreateMetricsClientFunc() (metrics.Client, error) {
	opts := []prom.Option{
		prom.WithCollector(m.collectorUptime),
		prom.WithCollector(m.collectorMemoryTotal),
		prom.WithCollector(m.collectorMemoryFree),
		prom.WithCollector(m.collectorLoad1),
		prom.WithCollector(m.collectorLoad5),
		prom.WithCollector(m.collectorLoad15),
		prom.WithCollector(m.collectorTempCPU),
		prom.WithCollector(m.collectorDiskTotal),
		prom.WithCollector(m.collectorDiskFree),
		prom.WithCollector(m.collectorNetworkRx),
		prom.WithCollector(m.collectorNetworkTx),
	}

	return prom.New(opts...) //nolint:wrapcheck
}

// SetUptime sets the time elapsed since last system boot.
func (m *Client) SetUptime(v time.Duration) {
	m.collectorUptime.Set(float64(v))
}

// SetMemoryTotal sets the total available memory in bytes.
func (m *Client) SetMemoryTotal(v uint64) {
	m.collectorMemoryTotal.Set(float64(v))
}

// SetMemoryFree sets the total free memory in bytes.
func (m *Client) SetMemoryFree(v uint64) {
	m.collectorMemoryFree.Set(float64(v))
}

// SetLoad1 sets the 1 minute CPU load average.
func (m *Client) SetLoad1(v float64) {
	m.collectorLoad1.Set(v)
}

// SetLoad5 sets the 5 minutes CPU load average.
func (m *Client) SetLoad5(v float64) {
	m.collectorLoad5.Set(v)
}

// SetLoad15 sets the 15 minutes CPU load average.
func (m *Client) SetLoad15(v float64) {
	m.collectorLoad15.Set(v)
}

// SetTempCPU sets the CPU Temperature in Celsius Degrees.
func (m *Client) SetTempCPU(v float64) {
	m.collectorTempCPU.Set(v)
}

// SetDiskTotal sets the total disk size in bytes.
func (m *Client) SetDiskTotal(v uint64) {
	m.collectorDiskTotal.Set(float64(v))
}

// SetDiskFree sets the total free disk space in bytes.
func (m *Client) SetDiskFree(v uint64) {
	m.collectorDiskFree.Set(float64(v))
}

// SetNetwork sets the network received and transmitted bytes for the specified NIC.
func (m *Client) SetNetwork(nic string, rx uint64, tx uint64) {
	plabels := prometheus.Labels{
		labelNic: nic,
	}

	m.collectorNetworkRx.With(plabels).Set(float64(rx))
	m.collectorNetworkTx.With(plabels).Set(float64(tx))
}
