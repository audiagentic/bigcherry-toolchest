package api

import (
    "bytes"
    "html/template"
    "strings"
    "testing"

    "github.com/tmac1973/llama-toolchest/internal/monitor"
    "github.com/tmac1973/llama-toolchest/web"
)

func TestMonitorBarRendersGPUName(t *testing.T) {
    tpl, err := template.New("").ParseFS(web.Templates, "templates/partials/monitor_bar.html")
    if err != nil {
        t.Fatalf("parse monitor bar: %v", err)
    }

    data := monitorBarData(monitor.Metrics{
        GPU: []monitor.GPUInfo{{
            Index:       0,
            Name:        "AMD Radeon RX 7900 XTX",
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
    for _, want := range []string{"AMD Radeon RX 7900 XTX", "42%", "12.0/24.0GiB", "57°C · 212W"} {
        if !strings.Contains(out, want) {
            t.Errorf("monitor bar missing %q", want)
        }
    }
}
