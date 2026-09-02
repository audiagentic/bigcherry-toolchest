package api

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/tmac1973/llama-toolchest/internal/monitor"
	"github.com/tmac1973/llama-toolchest/web"
)

func TestMonitorBarRendersHardwareDashboard(t *testing.T) {
	tpl, err := template.New("").ParseFS(web.Templates, "templates/partials/monitor_bar.html")
	if err != nil {
		t.Fatalf("parse monitor bar: %v", err)
	}

	data := monitorBarData(monitor.Metrics{
		Backend: "rocm",
		GPU: []monitor.GPUInfo{{
			Index:       0,
			Name:        "AMD Radeon RX 7900 XTX",
			BDF:         "0000:03:00.0",
			Arch:        "gfx1100",
			UtilPercent: 42,
			VRAMUsedMB:  12288,
			VRAMTotalMB: 24576,
			TempC:       57,
			PowerW:      212,
		}},
		CPU:    monitor.CPUInfo{UsagePercent: 17},
		Memory: monitor.MemoryInfo{UsedMB: 32768, TotalMB: 98304},
	})

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "monitor_bar", data); err != nil {
		t.Fatalf("execute monitor bar: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"Hardware",
		"1 GPUs",
		"ROCm0",
		"RX 7900 XTX",
		"AMD Radeon RX 7900 XTX",
		"42%",
		"12.0 / 24.0 GiB",
		"57°C",
		"212W",
		"gfx1100",
		"0000:03:00.0",
		`class="monitor-gpu-grid"`,
		`class="monitor-gpu-card"`,
		`data-gpu-index="0"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("monitor dashboard missing %q", want)
		}
	}
}

func TestMonitorDeviceLabels(t *testing.T) {
	cases := []struct {
		backend string
		index   int
		want    string
	}{
		{"rocm", 2, "ROCm2"},
		{"nvidia", 1, "CUDA1"},
		{"cuda", 3, "CUDA3"},
		{"", 4, "GPU4"},
	}
	for _, tc := range cases {
		if got := monitorDeviceLabel(tc.backend, tc.index); got != tc.want {
			t.Errorf("monitorDeviceLabel(%q, %d) = %q, want %q", tc.backend, tc.index, got, tc.want)
		}
	}
}

func TestShortGPUName(t *testing.T) {
	cases := map[string]string{
		"AMD Radeon RX 7900 XTX": "RX 7900 XTX",
		"Radeon AI PRO R9700":    "AI PRO R9700",
		"NVIDIA GeForce RTX 5090": "RTX 5090",
		"gfx1100":                 "gfx1100",
	}
	for input, want := range cases {
		if got := shortGPUName(input, 0); got != want {
			t.Errorf("shortGPUName(%q) = %q, want %q", input, got, want)
		}
	}
}
