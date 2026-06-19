package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "llama-toolchest"

// DefaultDataDir returns the platform-appropriate data directory for a host
// install. Containers override this via the YAML's data_dir field.
func DefaultDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", appName)
		}
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, appName)
		}
	default:
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return filepath.Join(dir, appName)
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", appName)
		}
	}
	return "/data"
}

// DefaultConfigPath returns the config file path to load. It honors
// LLAMA_TOOLCHEST_CONFIG if set; otherwise it returns the first candidate
// location that exists, falling back to the preferred user-scope path for
// writing when none exist.
//
// The existence check is what lets a system-scope install find
// /etc/llama-toolchest/llama-toolchest.yaml without the systemd unit having to
// pass --config or set an env var. Before this, the default resolved only to
// ~/.config (or /data), so a root service silently ignored the documented
// /etc config and used built-in defaults instead (issue #61).
func DefaultConfigPath() string {
	if path := os.Getenv("LLAMA_TOOLCHEST_CONFIG"); path != "" {
		return path
	}
	candidates := configSearchPaths()
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Nothing exists yet — return the preferred writable default (the first
	// candidate), so a later save lands in the conventional per-user location.
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "/data/config/llama-toolchest.yaml"
}

// configSearchPaths returns the candidate config locations for the current OS,
// in priority order. The first existing one wins in DefaultConfigPath.
func configSearchPaths() []string {
	const file = "llama-toolchest.yaml"
	switch runtime.GOOS {
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return []string{filepath.Join(home, "Library", "Application Support", appName, file)}
		}
	case "windows":
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return []string{filepath.Join(dir, appName, file)}
		}
	default:
		var paths []string
		if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
			paths = append(paths, filepath.Join(dir, appName, file))
		}
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(home, ".config", appName, file))
		}
		// System-scope install location, checked after the per-user paths so a
		// user config still takes precedence when both are present.
		paths = append(paths, filepath.Join("/etc", appName, file))
		return paths
	}
	return []string{"/data/config/llama-toolchest.yaml"}
}
