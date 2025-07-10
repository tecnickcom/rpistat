package httphandler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tecnickcom/rpistat/internal/metrics"
	"golang.org/x/sys/unix"
)

const (
	loadShift       = 1 << 16
	fileCPUTemp     = "/sys/class/thermal/thermal_zone0/temp"
	fileNetworkStat = "/proc/net/dev"
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

	// Uptime time elapsed since last system boot.
	Uptime time.Duration `json:"uptime"`

	// MemoryTotal is the total available memory in bytes.
	MemoryTotal uint64 `json:"memory_total"`

	// MemoryFree is the total free memory in bytes.
	MemoryFree uint64 `json:"memory_free"`

	// MemoryUsed is the total memory used in bytes.
	MemoryUsed uint64 `json:"memory_used"`

	// MemoryUsage is the total memory used in percentage
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

	// DiskUsage is the total disk used in percentage
	DiskUsage float64 `json:"disk_usage"`

	// Network contains an array of network statistics, one entry for each physical interface.
	Network []NetworkStat `json:"network"`

	metric metrics.Metrics
}

func newStats(m metrics.Metrics) *Stats {
	now := time.Now().UTC()

	t := &Stats{
		metric:    m,
		DateTime:  now.Format(time.RFC3339),
		Timestamp: now.UnixNano(),
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
	t.MemoryUsage = (float64(t.MemoryUsed) / float64(t.MemoryTotal))
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
	fd, err := syscall.Openat(0, fileCPUTemp, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return
	}

	f := os.NewFile(uintptr(fd), fileCPUTemp)

	defer func() { _ = syscall.Close(fd) }()

	var raw uint64

	_, err = fmt.Fscanln(f, &raw)
	if err != nil {
		return
	}

	t.TempCPU = float64(raw) / 1000

	t.metric.SetTempCPU(t.TempCPU)
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
	t.DiskUsage = (float64(t.DiskUsed) / float64(t.DiskTotal))

	t.metric.SetDiskTotal(t.DiskTotal)
	t.metric.SetDiskFree(t.DiskFree)
}

func (t *Stats) network() {
	fd, err := syscall.Openat(0, fileNetworkStat, syscall.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return
	}

	f := os.NewFile(uintptr(fd), fileNetworkStat)

	defer func() { _ = syscall.Close(fd) }()

	s := bufio.NewScanner(f)

	s.Scan()
	s.Scan()

	for s.Scan() {
		col := strings.SplitN(s.Text(), ":", 2)
		if len(col) != 2 {
			continue
		}

		nic := strings.TrimSpace(col[0])
		if nic == "lo" || nic == "docker0" {
			continue
		}

		data := strings.Fields(col[1])
		if len(data) != 16 {
			continue
		}

		rx, _ := strconv.ParseUint(data[0], 10, 64)
		tx, _ := strconv.ParseUint(data[8], 10, 64)

		t.Network = append(
			t.Network,
			NetworkStat{
				Nic: nic,
				Rx:  rx,
				Tx:  tx,
			},
		)

		t.metric.SetNetwork(nic, rx, tx)
	}
}
