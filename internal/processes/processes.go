package processes

import (
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPU        float64 `json:"cpu"`
	Memory     float32 `json:"memory"`
	Command    string  `json:"command"`
	Status     string  `json:"status"`
	Username   string  `json:"username"`
	CreateTime int64   `json:"create_time"`
}

func ListProcesses() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var result []ProcessInfo
	for _, p := range procs {
		name, _ := p.Name()
		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		cmd, _ := p.Cmdline()
		statusSlice, _ := p.Status()
		username, _ := p.Username()
		createTime, _ := p.CreateTime()

		// Convert status slice to string
		status := "unknown"
		if len(statusSlice) > 0 {
			status = statusSlice[0]
		}

		// Limit command length for display
		if len(cmd) > 100 {
			cmd = cmd[:100] + "..."
		}

		result = append(result, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPU:        cpu,
			Memory:     mem,
			Command:    cmd,
			Status:     status,
			Username:   username,
			CreateTime: createTime,
		})
	}
	return result, nil
}

func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
