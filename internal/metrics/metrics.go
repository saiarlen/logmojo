package metrics

import (
	"local-monitor/internal/db"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type HostMetrics struct {
	CPUPercent  float64 `json:"cpu_percent"`
	CPUCores    int     `json:"cpu_cores"`
	BusyCores   int     `json:"busy_cores"`
	RAMPercent  float64 `json:"ram_percent"`
	RAMTotal    uint64  `json:"ram_total"`
	RAMUsed     uint64  `json:"ram_used"`
	DiskPercent float64 `json:"disk_percent"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	Uptime      uint64  `json:"uptime"`
	LoadAvg     float64 `json:"load_avg"`
	NetSent     uint64  `json:"net_sent"`
	NetRecv     uint64  `json:"net_recv"`
}

func GetHostMetrics() (HostMetrics, error) {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	perCore, _ := cpu.Percent(0, true)
	cores, _ := cpu.Counts(true)
	d, _ := disk.Usage("/")
	u, _ := host.Info()
	l, _ := load.Avg()
	n, _ := net.IOCounters(false)

	netSent := uint64(0)
	netRecv := uint64(0)
	if len(n) > 0 {
		netSent = n[0].BytesSent
		netRecv = n[0].BytesRecv
	}

	busyCores := 0
	for _, p := range perCore {
		if p > 5.0 { // Consider a core busy if usage > 5%
			busyCores++
		}
	}

	return HostMetrics{
		CPUPercent:  c[0],
		CPUCores:    cores,
		BusyCores:   busyCores,
		RAMPercent:  v.UsedPercent,
		RAMTotal:    v.Total,
		RAMUsed:     v.Used,
		DiskPercent: d.UsedPercent,
		DiskTotal:   d.Total,
		DiskUsed:    d.Used,
		Uptime:      u.Uptime,
		LoadAvg:     l.Load1,
		NetSent:     netSent,
		NetRecv:     netRecv,
	}, nil
}

func StartHistoryRecorder() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			d, err := disk.Usage("/")
			v, _ := mem.VirtualMemory()
			c, _ := cpu.Percent(0, false)

			if err == nil && v != nil && len(c) > 0 {
				db.RecordMetricsHistory(c[0], v.UsedPercent, v.Used, d.UsedPercent)
			}
		}
	}()
}
