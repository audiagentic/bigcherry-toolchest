//go:build !linux

package monitor

func enrichGPUIdentity(_ string, _ []GPUInfo) {}
