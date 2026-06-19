package process

import (
	"strings"
	"testing"
)

// countCUDAOrder returns how many CUDA_DEVICE_ORDER entries are in env and the
// value of the last one (the one that wins).
func countCUDAOrder(env []string) (n int, last string) {
	for _, kv := range env {
		if strings.HasPrefix(kv, "CUDA_DEVICE_ORDER=") {
			n++
			last = strings.TrimPrefix(kv, "CUDA_DEVICE_ORDER=")
		}
	}
	return n, last
}

// TestPinCUDADeviceOrderAdds pins issue #68: when the user hasn't set
// CUDA_DEVICE_ORDER, we add PCI_BUS_ID so the CUDA backend's device indices
// match nvidia-smi (and thus the web UI's "GPU N" labels).
func TestPinCUDADeviceOrderAdds(t *testing.T) {
	env := pinCUDADeviceOrder([]string{"PATH=/usr/bin", "HOME=/root"})
	n, last := countCUDAOrder(env)
	if n != 1 || last != "PCI_BUS_ID" {
		t.Fatalf("expected exactly one CUDA_DEVICE_ORDER=PCI_BUS_ID; got count=%d last=%q env=%v", n, last, env)
	}
}

// TestPinCUDADeviceOrderRespectsExisting verifies a user-provided value is
// left untouched — we never override an explicit choice.
func TestPinCUDADeviceOrderRespectsExisting(t *testing.T) {
	env := pinCUDADeviceOrder([]string{"PATH=/usr/bin", "CUDA_DEVICE_ORDER=FASTEST_FIRST"})
	n, last := countCUDAOrder(env)
	if n != 1 || last != "FASTEST_FIRST" {
		t.Fatalf("expected the existing FASTEST_FIRST to survive untouched; got count=%d last=%q env=%v", n, last, env)
	}
}
