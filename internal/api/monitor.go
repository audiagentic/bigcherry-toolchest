package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tmac1973/llama-toolchest/internal/monitor"
)

func (s *Server) handleMonitorStream(w http.ResponseWriter, r *http.Request) {
	sse, err := NewSSEWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ch := s.monitor.Subscribe()
	defer s.monitor.Unsubscribe(ch)

	// Send current state immediately.
	data, _ := json.Marshal(s.monitor.Current())
	sse.SendEvent("metrics", string(data))

	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(m)
			sse.SendEvent("metrics", string(data))
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleMonitorStatus(w http.ResponseWriter, r *http.Request) {
	m := s.monitor.Current()

	if isHTMX(r) {
		respondHTML(w)
		s.renderPartial(w, "monitor_bar", monitorBarData(m))
		return
	}

	respondJSON(w, m)
}

// monitorBarGPU holds precomputed static/display values for the initial
// dashboard render. Dynamic values are subsequently updated in place from the
// monitor SSE stream, so the cards are not replaced every three seconds.
type monitorBarGPU struct {
	Index       int
	DeviceLabel string
	Name        string
	ShortName   string
	Arch        string
	BDF         string
	UtilPercent int
	VRAMPercent int
	VRAMGB      string
	TempC       int
	PowerW      float64
	IsIGPU      bool
}

type monitorBarView struct {
	Backend    string
	GPUs       []monitorBarGPU
	CPUPercent float64
	RAMPercent int
	RAMGB      string
}

// monitorBarData prepares monitor metrics for the monitor_bar template.
func monitorBarData(m monitor.Metrics) monitorBarView {
	gpus := make([]monitorBarGPU, len(m.GPU))
	for i, gpu := range m.GPU {
		vramPct := 0
		if gpu.VRAMTotalMB > 0 {
			vramPct = gpu.VRAMUsedMB * 100 / gpu.VRAMTotalMB
		}
		gpus[i] = monitorBarGPU{
			Index:       gpu.Index,
			DeviceLabel: monitorDeviceLabel(m.Backend, gpu.Index),
			Name:        gpu.Name,
			ShortName:   shortGPUName(gpu.Name, gpu.Index),
			Arch:        gpu.Arch,
			BDF:         gpu.BDF,
			UtilPercent: gpu.UtilPercent,
			VRAMPercent: vramPct,
			VRAMGB:      fmt.Sprintf("%.1f / %.1f GiB", float64(gpu.VRAMUsedMB)/1024, float64(gpu.VRAMTotalMB)/1024),
			TempC:       gpu.TempC,
			PowerW:      gpu.PowerW,
			IsIGPU:      gpu.IsIGPU,
		}
	}
	ramPct := 0
	if m.Memory.TotalMB > 0 {
		ramPct = m.Memory.UsedMB * 100 / m.Memory.TotalMB
	}
	return monitorBarView{
		Backend:    m.Backend,
		GPUs:       gpus,
		CPUPercent: m.CPU.UsagePercent,
		RAMPercent: ramPct,
		RAMGB:      fmt.Sprintf("%.1f / %.1f GiB", float64(m.Memory.UsedMB)/1024, float64(m.Memory.TotalMB)/1024),
	}
}

func monitorDeviceLabel(backend string, index int) string {
	switch strings.ToLower(backend) {
	case "rocm":
		return fmt.Sprintf("ROCm%d", index)
	case "nvidia", "cuda":
		return fmt.Sprintf("CUDA%d", index)
	default:
		return fmt.Sprintf("GPU%d", index)
	}
}

func shortGPUName(name string, index int) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Sprintf("GPU %d", index)
	}
	for _, prefix := range []string{
		"Advanced Micro Devices, Inc. ",
		"AMD Radeon ",
		"Radeon ",
		"NVIDIA GeForce ",
		"NVIDIA ",
	} {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimSpace(strings.TrimPrefix(name, prefix))
			break
		}
	}
	return name
}
