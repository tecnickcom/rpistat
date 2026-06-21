// Package sysstat gathers system usage statistics from the host. It has no
// dependency on the HTTP or metrics layers so it can be shared by the /stats
// JSON handler and the Prometheus collector without creating an import cycle.
package sysstat

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	loadShift              = 1 << 16
	defaultFileCPUTemp     = "/sys/class/thermal/thermal_zone0/temp"
	defaultFileNetworkStat = "/proc/net/dev"
	networkStatFields      = 16
)

// defaultExcludedNics are the interfaces skipped unless overridden via config.
//
//nolint:gochecknoglobals
var defaultExcludedNics = []string{"lo", "docker0"}

// NetworkStat contains network statistics for one physical interface.
type NetworkStat struct {
	// Nic is the Network Interface Card name.
	Nic string `json:"nic"`

	// Rx is the total number of bytes received.
	Rx uint64 `json:"rx"`

	// Tx is the total number of bytes transmitted.
	Tx uint64 `json:"tx"`
}

// Stats contains the collected system statistics.
type Stats struct {
	// DateTime is the human-readable date and time when the response is sent.
	DateTime string `json:"datetime"`

	// Timestamp is the machine-readable UTC timestamp in nanoseconds since EPOCH.
	Timestamp int64 `json:"timestamp"`

	// Hostname name of the host.
	Hostname string `json:"hostname"`

	// Uptime is the time elapsed since last system boot, in nanoseconds.
	Uptime time.Duration `json:"uptime"`

	// MemoryTotal is the total available memory in bytes.
	MemoryTotal uint64 `json:"memory_total"`

	// MemoryFree is the total free memory in bytes.
	MemoryFree uint64 `json:"memory_free"`

	// MemoryUsed is the total memory used in bytes.
	MemoryUsed uint64 `json:"memory_used"`

	// MemoryUsage is the used memory as a fraction in the range 0..1.
	MemoryUsage float64 `json:"memory_usage"`

	// Load1 is the 1 minute load average.
	Load1 float64 `json:"load_1m"`

	// Load5 is the 5 minutes load average.
	Load5 float64 `json:"load_5m"`

	// Load15 is the 15 minutes load average.
	Load15 float64 `json:"load_15m"`

	// TempCPU is the CPU Temperature in Celsius Degrees.
	TempCPU float64 `json:"temperature_cpu"`

	// DiskTotal is the total disk size in bytes.
	DiskTotal uint64 `json:"disk_total"`

	// DiskFree is the total free disk space in bytes.
	DiskFree uint64 `json:"disk_free"`

	// DiskUsed is the total disk used in bytes.
	DiskUsed uint64 `json:"disk_used"`

	// DiskUsage is the used disk space as a fraction in the range 0..1.
	DiskUsage float64 `json:"disk_usage"`

	// Network contains an array of network statistics, one entry for each physical interface.
	Network []NetworkStat `json:"network"`
}

// Gatherer collects system statistics from the host.
type Gatherer struct {
	log             *slog.Logger
	fileCPUTemp     string
	fileNetworkStat string
	excludedNics    map[string]struct{}
}

// NewGatherer creates a Gatherer with sensible defaults, overridable via options.
func NewGatherer(opts ...Option) *Gatherer {
	g := &Gatherer{
		log:             slog.New(slog.DiscardHandler),
		fileCPUTemp:     defaultFileCPUTemp,
		fileNetworkStat: defaultFileNetworkStat,
		excludedNics:    toSet(defaultExcludedNics),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}

	return set
}

// Gather collects a fresh snapshot of the system statistics.
func (g *Gatherer) Gather() *Stats {
	now := time.Now().UTC()

	s := &Stats{
		DateTime:  now.Format(time.RFC3339),
		Timestamp: now.UnixNano(),
	}

	g.hostname(s)
	g.sysinfo(s)
	g.cpuTemp(s)
	g.disk(s)
	g.network(s)

	return s
}

func (g *Gatherer) hostname(s *Stats) {
	hostname, err := os.Hostname()
	if err != nil {
		g.log.Debug("failed reading hostname", slog.Any("error", err))
		return
	}

	s.Hostname = hostname
}

func (g *Gatherer) sysinfo(s *Stats) {
	var u unix.Sysinfo_t

	err := unix.Sysinfo(&u)
	if err != nil {
		g.log.Debug("failed reading sysinfo", slog.Any("error", err))
		return
	}

	s.Uptime = time.Duration(u.Uptime) * time.Second
	s.MemoryTotal = uint64(u.Totalram) * uint64(u.Unit) //nolint:unconvert
	s.MemoryFree = uint64(u.Freeram) * uint64(u.Unit)   //nolint:unconvert
	s.MemoryUsed = s.MemoryTotal - s.MemoryFree
	s.MemoryUsage = usageRatio(s.MemoryUsed, s.MemoryTotal)
	s.Load1 = float64(u.Loads[0]) / loadShift
	s.Load5 = float64(u.Loads[1]) / loadShift
	s.Load15 = float64(u.Loads[2]) / loadShift
}

func (g *Gatherer) cpuTemp(s *Stats) {
	fd, err := syscall.Openat(0, g.fileCPUTemp, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil || fd < 0 {
		g.log.Debug("failed opening CPU temperature file", slog.String("path", g.fileCPUTemp), slog.Any("error", err))
		return
	}

	f := os.NewFile(uintptr(fd), g.fileCPUTemp)

	// Close through the *os.File so its finalizer is cleared; closing the raw
	// fd directly would leave the finalizer to close the (possibly recycled)
	// descriptor a second time.
	defer func() { _ = f.Close() }()

	temp, err := parseCPUTemp(f)
	if err != nil {
		g.log.Debug("failed parsing CPU temperature", slog.Any("error", err))
		return
	}

	s.TempCPU = temp
}

// parseCPUTemp reads the raw millidegree value emitted by a thermal zone file
// and returns the temperature in Celsius degrees.
func parseCPUTemp(r io.Reader) (float64, error) {
	var raw uint64

	_, err := fmt.Fscanln(r, &raw)
	if err != nil {
		return 0, fmt.Errorf("failed parsing CPU temperature: %w", err)
	}

	return float64(raw) / 1000, nil
}

func (g *Gatherer) disk(s *Stats) {
	f := syscall.Statfs_t{}

	err := syscall.Statfs("/", &f)
	if err != nil {
		g.log.Debug("failed reading disk stats", slog.Any("error", err))
		return
	}

	s.DiskTotal = f.Blocks * uint64(f.Bsize) //nolint:gosec
	s.DiskFree = f.Bfree * uint64(f.Bsize)   //nolint:gosec
	s.DiskUsed = s.DiskTotal - s.DiskFree
	s.DiskUsage = usageRatio(s.DiskUsed, s.DiskTotal)
}

func (g *Gatherer) network(s *Stats) {
	fd, err := syscall.Openat(0, g.fileNetworkStat, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil || fd < 0 {
		g.log.Debug("failed opening network stats file", slog.String("path", g.fileNetworkStat), slog.Any("error", err))
		return
	}

	f := os.NewFile(uintptr(fd), g.fileNetworkStat)

	// Close through the *os.File so its finalizer is cleared; closing the raw
	// fd directly would leave the finalizer to close the (possibly recycled)
	// descriptor a second time.
	defer func() { _ = f.Close() }()

	s.Network = parseNetworkStats(f, g.excludedNics)
}

// parseNetworkStats parses the content of /proc/net/dev and returns one entry
// per relevant physical interface. Malformed lines and excluded interfaces are
// skipped silently, matching the kernel-provided format.
func parseNetworkStats(r io.Reader, excluded map[string]struct{}) []NetworkStat {
	var stats []NetworkStat

	s := bufio.NewScanner(r)

	// skip the two header lines
	s.Scan()
	s.Scan()

	for s.Scan() {
		col := strings.SplitN(s.Text(), ":", 2)
		if len(col) != 2 {
			continue
		}

		nic := strings.TrimSpace(col[0])
		if _, ok := excluded[nic]; ok {
			continue
		}

		data := strings.Fields(col[1])
		if len(data) != networkStatFields {
			continue
		}

		rx, _ := strconv.ParseUint(data[0], 10, 64)
		tx, _ := strconv.ParseUint(data[8], 10, 64)

		stats = append(stats, NetworkStat{
			Nic: nic,
			Rx:  rx,
			Tx:  tx,
		})
	}

	return stats
}

// usageRatio returns used/total as a fraction in the range [0,1], or 0 when
// total is 0 to avoid producing NaN (which cannot be JSON-encoded).
func usageRatio(used, total uint64) float64 {
	if total == 0 {
		return 0
	}

	return float64(used) / float64(total)
}
