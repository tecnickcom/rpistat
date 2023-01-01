// Package main contains the rpistat program to return Raspberry Pi system statistics.
// Nicola Asuni - 2022-12-31
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	programName             = "rpistat"
	serverAddress           = ":65501"
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 10 * time.Second
	serverWriteTimeout      = 10 * time.Second
	loadShift               = 1 << 16
	fileCPUTemp             = "/sys/class/thermal/thermal_zone0/temp"
	fileNetworkStat         = "/proc/net/dev"
)

var (
	// programVersion contains the version of the application injected at compile time.
	programVersion = "0.0.0" //nolint:gochecknoglobals

	// programRelease contains the release of the application injected at compile time.
	programRelease = "0" //nolint:gochecknoglobals
)

// Stats contains the information to be returned.
type Stats struct {
	// Program is the application name.
	Program string `json:"program"`

	// Version is the program semantic version (e.g. 1.2.3).
	Version string `json:"version"`

	// Release is the program build number that is appended to the version.
	Release string `json:"release"`

	// DateTime is the human-readable date and time when the response is sent.
	DateTime string `json:"datetime"`

	// Timestamp is the machine-readable UTC timestamp in nanoseconds since EPOCH.
	Timestamp int64 `json:"timestamp"`

	// Hostname name of the host.
	Hostname string `json:"hostname"`

	// Uptime time since last system boot.
	Uptime time.Duration `json:"uptime"`

	// Total Memory in bytes.
	MemoryTotal uint64 `json:"memory_total"`

	// Total free memory in bytes.
	MemoryFree uint64 `json:"memory_free"`

	// Total memory used in bytes.
	MemoryUsed uint64 `json:"memory_used"`

	// Total memory used in percentage
	MemoryUsage float64 `json:"memory_usage"`

	// Load is the 1 minute load average.
	Load1 float64 `json:"load_1m"`

	// Load is the 5 minute load average.
	Load5 float64 `json:"load_5m"`

	// Load is the 15 minute load average.
	Load15 float64 `json:"load_15m"`

	// CPU Temperature in Celsius Degrees.
	TempCPU float64 `json:"temperature_cpu"`

	// Total Disk size in bytes.
	DiskTotal uint64 `json:"disk_total"`

	// Total free disk space in bytes.
	DiskFree uint64 `json:"disk_free"`

	// Total disk used in bytes.
	DiskUsed uint64 `json:"disk_used"`

	// Total disk used in percentage
	DiskUsage float64 `json:"disk_usage"`

	// Network contains an array of network statistics, one entry for each physical interface.
	Network []NetworkStat `json:"network"`
}

// NetworkStat contains network statistics for one phisical interface.
type NetworkStat struct {
	Nic string `json:"nic"`
	Rx  uint64 `json:"rx"`
	Tx  uint64 `json:"tx"`
}

func (t *Stats) hostname() {
	hostname, err := os.Hostname()
	if err == nil {
		t.Hostname = hostname
	}
}

func (t *Stats) sysinfo() {
	var u unix.Sysinfo_t
	if err := unix.Sysinfo(&u); err != nil {
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
}

func (t *Stats) disk() {
	f := syscall.Statfs_t{}

	err := syscall.Statfs("/", &f)
	if err != nil {
		return
	}

	t.DiskTotal = f.Blocks * uint64(f.Bsize)
	t.DiskFree = f.Bfree * uint64(f.Bsize)
	t.DiskUsed = t.DiskTotal - t.DiskFree
	t.DiskUsage = (float64(t.DiskUsed) / float64(t.DiskTotal))
}

//nolint:gocognit
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
	}
}

func newStats() *Stats {
	now := time.Now().UTC()

	t := &Stats{
		Program:   programName,
		Version:   programVersion,
		Release:   programRelease,
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

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newStats()) //nolint:errchkjson
}

func main() {
	handler := http.NewServeMux()
	handler.HandleFunc("/", statsHandler)

	srv := &http.Server{
		Addr:              serverAddress,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
	}

	log.Println(programName + " listening at " + serverAddress)
	log.Fatal(srv.ListenAndServe())
}
