package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonitorDashboardUsesSSEWithoutRecurringFragmentPoll(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "web", "static", "log-panel.js"))
	if err != nil {
		t.Fatalf("read log-panel.js: %v", err)
	}
	js := string(src)
	for _, want := range []string{
		"new EventSource('/api/monitor/stream')",
		"monitorBootstrap.setAttribute('hx-trigger', 'load')",
		"applyMonitorMetrics",
		"data-gpu-index",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("monitor dashboard JS missing %q", want)
		}
	}
	if strings.Contains(js, "setInterval(") {
		t.Error("monitor dashboard should use SSE, not a JavaScript polling interval")
	}
}
