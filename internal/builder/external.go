package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ExternalBuildInput describes an already-built llama.cpp server that should
// participate in Toolchest exactly like a locally compiled build.
//
// Native BuildResult fields intentionally carry the provenance so existing
// benchmark snapshots keep it without a parallel result schema:
//   - ID: stable lane/build identity
//   - Tag: optional human label (for example bigcherry-mmq)
//   - Profile: Toolchest backend/profile (rocm, vulkan, ...)
//   - GitRef: source/base ref (for example b10227 or tuning-code-rebase)
//   - GitSHA: exact source commit when known
//   - BinaryPath: absolute llama-server executable path
//
// RegisterExternalBuild never copies, moves, chmods, or deletes the supplied
// executable. Toolchest stores only the reference in builds.json.
type ExternalBuildInput struct {
	ID         string
	Tag        string
	Profile    string
	GitRef     string
	GitSHA     string
	BinaryPath string
	Replace    bool
}

var validExternalBuildIDRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// RegisterExternalBuild adds an existing llama-server executable to the
// native build registry so current build selection, router lifecycle,
// benchmark jobs and BuildSnapshot persistence work without special cases.
func (b *Builder) RegisterExternalBuild(in ExternalBuildInput) (*BuildResult, error) {
	in.ID = strings.TrimSpace(in.ID)
	in.Tag = strings.TrimSpace(in.Tag)
	in.Profile = strings.TrimSpace(in.Profile)
	in.GitRef = strings.TrimSpace(in.GitRef)
	in.GitSHA = strings.TrimSpace(in.GitSHA)
	in.BinaryPath = strings.TrimSpace(in.BinaryPath)

	if in.ID == "" {
		return nil, fmt.Errorf("external build id is required")
	}
	if !validExternalBuildIDRE.MatchString(in.ID) {
		return nil, fmt.Errorf("invalid external build id %q: use letters, digits, '.', '_' or '-'", in.ID)
	}
	if in.Profile == "" {
		return nil, fmt.Errorf("external build profile is required")
	}
	if _, ok := FindProfile(in.Profile); !ok {
		return nil, fmt.Errorf("unknown profile: %s", in.Profile)
	}

	binaryPath, err := resolveExternalServerBinary(in.BinaryPath)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := BuildResult{
		ID:         in.ID,
		Tag:        in.Tag,
		Profile:    in.Profile,
		GitRef:     in.GitRef,
		GitSHA:     in.GitSHA,
		Status:     BuildStatusSuccess,
		BinaryPath: binaryPath,
		StartedAt:  now,
		FinishedAt: now,
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for i, existing := range b.builds {
		if existing.ID != result.ID {
			continue
		}
		if !in.Replace {
			return nil, &DuplicateBuildError{ID: result.ID}
		}
		b.builds[i] = result
		b.saveBuilds()
		out := result
		return &out, nil
	}

	b.builds = append(b.builds, result)
	b.saveBuilds()
	out := result
	return &out, nil
}

// resolveExternalServerBinary accepts either the llama-server executable or
// a directory containing it. A directory form is useful for BigCherry lanes,
// where each build generally has its own bin/output directory and co-located
// shared libraries. Process Manager already adds the executable directory to
// the child library search path.
func resolveExternalServerBinary(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("external build binary path is required")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve external build path: %w", err)
	}
	abs = filepath.Clean(abs)

	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("external build path %q: %w", abs, err)
	}
	if st.IsDir() {
		candidates := []string{
			filepath.Join(abs, "llama-server"),
			filepath.Join(abs, "llama-server.exe"),
			filepath.Join(abs, "server"),
			filepath.Join(abs, "server.exe"),
		}
		found := ""
		for _, candidate := range candidates {
			if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
				found = candidate
				st = info
				break
			}
		}
		if found == "" {
			return "", fmt.Errorf("external build directory %q does not contain llama-server", abs)
		}
		abs = found
	}

	if runtime.GOOS != "windows" && st.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("external build binary %q is not executable", abs)
	}
	return abs, nil
}
