//go:build linux

package monitor

import (
	"encoding/csv"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tmac1973/llama-toolchest/internal/builder"
)

type gpuStaticIdentity struct {
	Name string
	Arch string
	BDF  string
}

var (
	rocmIdentityOnce sync.Once
	rocmIdentity     map[int]gpuStaticIdentity
)

// enrichGPUIdentity overlays cached, stable hardware identity onto dynamic
// telemetry. ROCm's identity scan is intentionally separate from the 3-second
// metric poll: product name, BDF and architecture do not change while the
// process is running, and spawning an identity command on every poll is waste.
func enrichGPUIdentity(backend string, gpus []GPUInfo) {
	if backend != "rocm" || len(gpus) == 0 {
		return
	}

	rocmIdentityOnce.Do(func() {
		rocmIdentity = discoverROCmIdentity()
	})

	for i := range gpus {
		id, ok := rocmIdentity[gpus[i].Index]
		if !ok {
			continue
		}
		if id.Name != "" {
			gpus[i].Name = id.Name
		}
		if id.Arch != "" {
			gpus[i].Arch = id.Arch
		}
		if id.BDF != "" {
			gpus[i].BDF = id.BDF
		}
	}
}

func discoverROCmIdentity() map[int]gpuStaticIdentity {
	dirs := listAMDGPUDirs()
	byBDF := kfdIndexByBDF(dirs)
	ids := make(map[int]gpuStaticIdentity, len(dirs))

	// Even if rocm-smi is unavailable, retain the BDF identity obtained from
	// the same KFD-ordered sysfs directories llama-server uses for ROCm<N>.
	for idx, dir := range dirs {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			bdf := strings.ToLower(filepath.Base(resolved))
			ids[idx] = gpuStaticIdentity{BDF: bdf}
		}
	}

	smi := builder.FindROCmTool("rocm-smi")
	if smi == "" {
		return ids
	}
	out, err := exec.Command(smi, "--showbus", "--showproductname", "--csv").Output()
	if err != nil {
		return ids
	}
	for idx, discovered := range parseROCmProductCSV(string(out), byBDF) {
		base := ids[idx]
		if discovered.Name != "" {
			base.Name = discovered.Name
		}
		if discovered.Arch != "" {
			base.Arch = discovered.Arch
		}
		if discovered.BDF != "" {
			base.BDF = discovered.BDF
		}
		ids[idx] = base
	}
	return ids
}

// parseROCmProductCSV maps rocm-smi rows to KFD/llama-server indices by PCI
// BDF. This deliberately does not trust rocm-smi's cardN label because that
// enumeration is not guaranteed to match ROCm<N> on mixed/multi-GPU hosts.
func parseROCmProductCSV(out string, byBDF map[string]int) map[int]gpuStaticIdentity {
	result := make(map[int]gpuStaticIdentity)
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(out)))
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return result
	}

	cols := make(map[string]int, len(records[0]))
	for i, h := range records[0] {
		cols[strings.TrimSpace(h)] = i
	}
	busCol, ok := cols["PCI Bus"]
	if !ok {
		return result
	}

	for _, row := range records[1:] {
		if busCol >= len(row) {
			continue
		}
		bdf := strings.ToLower(strings.TrimSpace(row[busCol]))
		idx, ok := byBDF[bdf]
		if !ok {
			continue
		}
		id := gpuStaticIdentity{BDF: bdf}
		if col, ok := cols["Card Series"]; ok && col < len(row) {
			name := strings.TrimSpace(row[col])
			if name != "" && !strings.EqualFold(name, "N/A") {
				id.Name = name
			}
		}
		if col, ok := cols["GFX Version"]; ok && col < len(row) {
			arch := strings.TrimSpace(row[col])
			if arch != "" && !strings.EqualFold(arch, "N/A") {
				id.Arch = arch
			}
		}
		result[idx] = id
	}
	return result
}
