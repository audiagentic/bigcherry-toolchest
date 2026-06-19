//go:build linux

package monitor

import (
	"reflect"
	"testing"
)

// TestParseROCmGPUNamesSkipsCPU guards issue #68: rocminfo lists the host CPU
// as an agent, and the old name-blacklist let non-AMD CPUs (e.g. an Intel
// Xeon) leak through and get labelled as "GPU 0". Selecting by Device Type
// must drop the CPU agent regardless of its vendor.
func TestParseROCmGPUNamesSkipsCPU(t *testing.T) {
	// Abridged rocminfo output: an Intel Xeon CPU agent followed by an AMD GPU.
	out := `ROCk module is loaded
=====================
HSA System Attributes
=====================
Runtime Version:         1.1
*******
Agent 1
*******
  Name:                    Intel(R) Xeon(R) W-2225 CPU @ 4.10GHz
  Uuid:                    CPU-XX
  Marketing Name:          Intel(R) Xeon(R) W-2225 CPU @ 4.10GHz
  Device Type:             CPU
  Pool Info:
    Pool 1
      Segment:               GLOBAL; FLAGS: FINE GRAINED
*******
Agent 2
*******
  Name:                    gfx1100
  Uuid:                    GPU-XX
  Marketing Name:          AMD Radeon RX 7900 XTX
  Device Type:             GPU
  Pool Info:
    Pool 1
      Segment:               GLOBAL; FLAGS: COARSE GRAINED
`
	got := parseROCmGPUNames(out)
	want := []string{"AMD Radeon RX 7900 XTX"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseROCmGPUNames = %v; want %v", got, want)
	}
}

// TestParseROCmGPUNamesMultiGPU verifies GPU agents are returned in order and
// the CPU between/around them is skipped.
func TestParseROCmGPUNamesMultiGPU(t *testing.T) {
	out := `*******
Agent 1
*******
  Name:                    AMD Ryzen 9 7950X
  Marketing Name:          AMD Ryzen 9 7950X 16-Core Processor
  Device Type:             CPU
*******
Agent 2
*******
  Name:                    gfx1100
  Marketing Name:          AMD Radeon RX 7900 XTX
  Device Type:             GPU
*******
Agent 3
*******
  Name:                    gfx1100
  Marketing Name:          AMD Radeon RX 7900 XTX
  Device Type:             GPU
`
	got := parseROCmGPUNames(out)
	want := []string{"AMD Radeon RX 7900 XTX", "AMD Radeon RX 7900 XTX"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseROCmGPUNames = %v; want %v", got, want)
	}
}

// TestParseROCmGPUNamesFallbackToName covers a GPU agent with no marketing
// name — the gfx id stands in so the entry is never blank.
func TestParseROCmGPUNamesFallbackToName(t *testing.T) {
	out := `*******
Agent 1
*******
  Name:                    AMD Eng Sample
  Marketing Name:          AMD Eng Sample
  Device Type:             CPU
*******
Agent 2
*******
  Name:                    gfx1151
  Marketing Name:
  Device Type:             GPU
`
	got := parseROCmGPUNames(out)
	want := []string{"gfx1151"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseROCmGPUNames = %v; want %v", got, want)
	}
}

// TestParseROCmGPUNamesNoGPU returns nil when only a CPU agent is present, so
// readGPUNameSysfs falls back to a generic label rather than the CPU name.
func TestParseROCmGPUNamesNoGPU(t *testing.T) {
	out := `*******
Agent 1
*******
  Name:                    Intel(R) Xeon(R) W-2225 CPU @ 4.10GHz
  Marketing Name:          Intel(R) Xeon(R) W-2225 CPU @ 4.10GHz
  Device Type:             CPU
`
	if got := parseROCmGPUNames(out); len(got) != 0 {
		t.Errorf("parseROCmGPUNames = %v; want empty", got)
	}
}
