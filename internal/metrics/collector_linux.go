//go:build linux

package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Collector reads system metrics from /proc.
type Collector struct{}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Collect gathers current system metrics.
func (c *Collector) Collect() (*Metrics, error) {
	m := &Metrics{}

	cpu, err := c.readCPU()
	if err == nil {
		m.CPU = cpu
	}

	ramUsed, ramTotal, err := c.readMemInfo()
	if err == nil {
		m.RAMUsed = ramUsed
		m.RAMTotal = ramTotal
	}

	diskUsed, diskTotal, err := c.readDisk()
	if err == nil {
		m.DiskUsed = diskUsed
		m.DiskTotal = diskTotal
	}

	load, err := c.readLoadAvg()
	if err == nil {
		m.LoadAvg = load
	}

	rx, tx, err := c.readNetwork()
	if err == nil {
		m.NetworkRx = rx
		m.NetworkTx = tx
	}

	m.Processes = c.countProcesses()

	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err == nil {
		m.Uptime = uint64(info.Uptime)
	}

	return m, nil
}

func (c *Collector) readCPU() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		var user, nice, system, idle uint64
		fmt.Sscanf(fields[1], "%d", &user)
		fmt.Sscanf(fields[2], "%d", &nice)
		fmt.Sscanf(fields[3], "%d", &system)
		fmt.Sscanf(fields[4], "%d", &idle)
		total := user + nice + system + idle
		if total == 0 {
			return 0, fmt.Errorf("zero cpu total")
		}
		return float64(user+nice+system) * 100.0 / float64(total), nil
	}
	return 0, fmt.Errorf("cpu line not found")
}

func (c *Collector) readMemInfo() (used, total uint64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %d", &memTotal)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %d", &memAvailable)
		}
	}
	if memTotal == 0 {
		return 0, 0, fmt.Errorf("memtotal not found")
	}
	return (memTotal - memAvailable) * 1024, memTotal * 1024, nil
}

func (c *Collector) readDisk() (used, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0, err
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	return total - free, total, nil
}

func (c *Collector) readLoadAvg() ([3]float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	parts := strings.Fields(string(data))
	if len(parts) < 3 {
		return [3]float64{}, fmt.Errorf("invalid loadavg")
	}
	var load [3]float64
	for i := 0; i < 3; i++ {
		load[i], _ = strconv.ParseFloat(parts[i], 64)
	}
	return load, nil
}

func (c *Collector) readNetwork() (rx, tx uint64, err error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}
		iface := strings.TrimSuffix(parts[0], ":")
		if iface == "lo" {
			continue
		}
		r, _ := strconv.ParseUint(parts[1], 10, 64)
		t, _ := strconv.ParseUint(parts[9], 10, 64)
		rx += r
		tx += t
	}
	return rx, tx, nil
}

func (c *Collector) countProcesses() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err == nil {
			count++
		}
	}
	return count
}
