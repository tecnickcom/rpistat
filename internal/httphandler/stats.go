package httphandler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tecnickcom/rpistat/internal/metrics"
	"golang.org/x/sys/unix"
)

const (
	loadShift              = 1 << 16
	defaultFileCPUTemp     = "/sys/class/thermal/thermal_zone0/temp"
	defaultFileNetworkStat = "/proc/net/dev"
	networkStatFields      = 16
)

// NetworkStat contains network statistics for one physical interface.
type NetworkStat struct {
	// NIC is the Network Interface Card name.
	Nic string `json:"nic"`

	// Rx is the total number of bytes received.
	Rx uint64 `json:"rx"`

	// Tx is the total number of bytes transmitted.
	Tx uint64 `json:"tx"`
}

// Stats contains the information to be returned.
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

	metric metrics.Metrics

	// fileCPUTemp is the path to the CPU temperature file (overridable for testing).
	fileCPUTemp string

	// fileNetworkStat is the path to the network statistics file (overridable for testing).
	fileNetworkStat string
}

func newStats(m metrics.Metrics) *Stats {
	now := time.Now().UTC()

	t := &Stats{
		metric:          m,
		fileCPUTemp:     defaultFileCPUTemp,
		fileNetworkStat: defaultFileNetworkStat,
		DateTime:        now.Format(time.RFC3339),
		Timestamp:       now.UnixNano(),
	}

	t.hostname()
	t.sysinfo()
	t.cpuTemp()
	t.disk()
	t.network()

	return t
}

func (t *Stats) hostname() {
	hostname, err := os.Hostname()
	if err == nil {
		t.Hostname = hostname
	}
}

func (t *Stats) sysinfo() {
	var u unix.Sysinfo_t

	err := unix.Sysinfo(&u)
	if err != nil {
		return
	}

	t.Uptime = time.Duration(u.Uptime) * time.Second
	t.MemoryTotal = uint64(u.Totalram) * uint64(u.Unit) //nolint:unconvert
	t.MemoryFree = uint64(u.Freeram) * uint64(u.Unit)   //nolint:unconvert
	t.MemoryUsed = t.MemoryTotal - t.MemoryFree
	t.MemoryUsage = usageRatio(t.MemoryUsed, t.MemoryTotal)
	t.Load1 = float64(u.Loads[0]) / loadShift
	t.Load5 = float64(u.Loads[1]) / loadShift
	t.Load15 = float64(u.Loads[2]) / loadShift

	// metrics
	t.metric.SetUptime(t.Uptime)
	t.metric.SetMemoryTotal(t.MemoryTotal)
	t.metric.SetMemoryFree(t.MemoryFree)
	t.metric.SetLoad1(t.Load1)
	t.metric.SetLoad5(t.Load5)
	t.metric.SetLoad15(t.Load15)
}

func (t *Stats) cpuTemp() {
	fd, err := syscall.Openat(0, t.fileCPUTemp, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil || fd < 0 {
		return
	}

	f := os.NewFile(uintptr(fd), t.fileCPUTemp)

	defer func() { _ = syscall.Close(fd) }()

	temp, err := parseCPUTemp(f)
	if err != nil {
		return
	}

	t.TempCPU = temp

	t.metric.SetTempCPU(t.TempCPU)
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

func (t *Stats) disk() {
	f := syscall.Statfs_t{}

	err := syscall.Statfs("/", &f)
	if err != nil {
		return
	}

	t.DiskTotal = f.Blocks * uint64(f.Bsize) //nolint:gosec
	t.DiskFree = f.Bfree * uint64(f.Bsize)   //nolint:gosec
	t.DiskUsed = t.DiskTotal - t.DiskFree
	t.DiskUsage = usageRatio(t.DiskUsed, t.DiskTotal)

	t.metric.SetDiskTotal(t.DiskTotal)
	t.metric.SetDiskFree(t.DiskFree)
}

func (t *Stats) network() {
	fd, err := syscall.Openat(0, t.fileNetworkStat, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil || fd < 0 {
		return
	}

	f := os.NewFile(uintptr(fd), t.fileNetworkStat)

	defer func() { _ = syscall.Close(fd) }()

	t.Network = parseNetworkStats(f)

	for _, ns := range t.Network {
		t.metric.SetNetwork(ns.Nic, ns.Rx, ns.Tx)
	}
}

// excludedNic reports whether the named interface should be skipped (loopback
// and the default docker bridge are not physical interfaces of interest).
func excludedNic(nic string) bool {
	return nic == "lo" || nic == "docker0"
}

// parseNetworkStats parses the content of /proc/net/dev and returns one entry
// per relevant physical interface. Malformed lines and excluded interfaces are
// skipped silently, matching the kernel-provided format.
func parseNetworkStats(r io.Reader) []NetworkStat {
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
		if excludedNic(nic) {
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
