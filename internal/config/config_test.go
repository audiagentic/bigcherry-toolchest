package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestConfigSearchPathsIncludesEtc guards issue #61: a system-scope install
// puts its config at /etc/llama-toolchest/llama-toolchest.yaml, but the binary
// never searched there, so a root service silently ignored it. /etc must be a
// candidate, ordered after the per-user paths so a user config still wins.
func TestConfigSearchPathsIncludesEtc(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skipf("system /etc path is Linux-specific; GOOS=%s", runtime.GOOS)
	}
	paths := configSearchPaths()
	want := filepath.Join("/etc", appName, "llama-toolchest.yaml")
	if len(paths) == 0 || paths[len(paths)-1] != want {
		t.Fatalf("expected %q as the last (lowest-priority) candidate; got %v", want, paths)
	}
}

// TestDefaultConfigPathEnvWins verifies LLAMA_TOOLCHEST_CONFIG short-circuits
// the search entirely — this is the documented escape hatch.
func TestDefaultConfigPathEnvWins(t *testing.T) {
	t.Setenv("LLAMA_TOOLCHEST_CONFIG", "/custom/spot/config.yaml")
	if got := DefaultConfigPath(); got != "/custom/spot/config.yaml" {
		t.Fatalf("DefaultConfigPath() = %q, want the env override", got)
	}
}

// TestDefaultConfigPathPrefersExistingConfig verifies the first candidate that
// actually exists on disk is returned, rather than a fixed canonical path.
func TestDefaultConfigPathPrefersExistingConfig(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skipf("uses XDG_CONFIG_HOME; GOOS=%s", runtime.GOOS)
	}
	t.Setenv("LLAMA_TOOLCHEST_CONFIG", "") // disable the env override
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	dir := filepath.Join(tmp, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "llama-toolchest.yaml")
	if err := os.WriteFile(want, []byte("data_dir: /tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := DefaultConfigPath(); got != want {
		t.Fatalf("DefaultConfigPath() = %q, want the existing XDG config %q", got, want)
	}
}

// TestSalvageModelsDir verifies that a config with a non-existent
// models_dir falls back to <DataDir>/models. This salvages registries
// poisoned by the old env-var leak, where a host bind-mount source
// got persisted as the in-container models_dir.
func TestSalvageModelsDir(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	os.WriteFile(cfgPath, []byte(`
data_dir: "`+tmp+`"
models_dir: "/nonexistent/host/path/fake"
`), 0o644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ModelsDir != "" {
		t.Fatalf("ModelsDir not salvaged: got %q, want \"\"", cfg.ModelsDir)
	}
	want := filepath.Join(tmp, "models")
	if cfg.ModelsPath() != want {
		t.Fatalf("ModelsPath() = %q, want %q", cfg.ModelsPath(), want)
	}
}

// TestKeepGoodModelsDir verifies a legitimately-set models_dir survives
// the salvage check.
func TestKeepGoodModelsDir(t *testing.T) {
	tmp := t.TempDir()
	custom := filepath.Join(tmp, "custom-models")
	if err := os.Mkdir(custom, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	os.WriteFile(cfgPath, []byte(`
data_dir: "`+tmp+`"
models_dir: "`+custom+`"
`), 0o644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ModelsDir != custom {
		t.Fatalf("ModelsDir clobbered: got %q, want %q", cfg.ModelsDir, custom)
	}
}
