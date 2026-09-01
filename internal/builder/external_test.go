package builder

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRegisterExternalBuildPersists(t *testing.T) {
	dataDir := t.TempDir()
	binary := fakeExternalServer(t)

	b := NewBuilder(dataDir)
	got, err := b.RegisterExternalBuild(ExternalBuildInput{
		ID:         "bigcherry-b10227-hip",
		Tag:        "bigcherry-mmq",
		Profile:    "rocm",
		GitRef:     "b10227",
		GitSHA:     "0123456789abcdef",
		BinaryPath: binary,
	})
	if err != nil {
		t.Fatalf("RegisterExternalBuild: %v", err)
	}
	if got.Status != BuildStatusSuccess {
		t.Fatalf("status = %q, want %q", got.Status, BuildStatusSuccess)
	}
	if !filepath.IsAbs(got.BinaryPath) {
		t.Fatalf("binary path = %q, want absolute path", got.BinaryPath)
	}

	reloaded := NewBuilder(dataDir)
	persisted, ok := reloaded.Find("bigcherry-b10227-hip")
	if !ok {
		t.Fatal("registered external build missing after reload")
	}
	if persisted.BinaryPath != got.BinaryPath {
		t.Fatalf("persisted binary = %q, want %q", persisted.BinaryPath, got.BinaryPath)
	}
	if persisted.GitSHA != "0123456789abcdef" || persisted.GitRef != "b10227" || persisted.Tag != "bigcherry-mmq" {
		t.Fatalf("provenance did not persist: %+v", persisted)
	}
}

func TestRegisterExternalBuildAcceptsDirectory(t *testing.T) {
	dataDir := t.TempDir()
	binary := fakeExternalServer(t)

	b := NewBuilder(dataDir)
	got, err := b.RegisterExternalBuild(ExternalBuildInput{
		ID:         "bigcherry-vulkan",
		Profile:    "vulkan",
		BinaryPath: filepath.Dir(binary),
	})
	if err != nil {
		t.Fatalf("RegisterExternalBuild directory: %v", err)
	}
	if got.BinaryPath != binary {
		t.Fatalf("resolved binary = %q, want %q", got.BinaryPath, binary)
	}
}

func TestRegisterExternalBuildDuplicateRequiresReplace(t *testing.T) {
	dataDir := t.TempDir()
	binary := fakeExternalServer(t)
	b := NewBuilder(dataDir)

	in := ExternalBuildInput{ID: "bigcherry-hip", Profile: "rocm", BinaryPath: binary, GitSHA: "old"}
	if _, err := b.RegisterExternalBuild(in); err != nil {
		t.Fatalf("first register: %v", err)
	}
	in.GitSHA = "new"
	if _, err := b.RegisterExternalBuild(in); err == nil {
		t.Fatal("duplicate registration succeeded without Replace")
	} else if _, ok := err.(*DuplicateBuildError); !ok {
		t.Fatalf("duplicate error = %T %v, want *DuplicateBuildError", err, err)
	}

	in.Replace = true
	if _, err := b.RegisterExternalBuild(in); err != nil {
		t.Fatalf("replace register: %v", err)
	}
	got, _ := b.Find(in.ID)
	if got.GitSHA != "new" {
		t.Fatalf("replaced SHA = %q, want new", got.GitSHA)
	}
}

func TestDeleteExternalBuildDoesNotDeleteSourceBinary(t *testing.T) {
	dataDir := t.TempDir()
	binary := fakeExternalServer(t)
	b := NewBuilder(dataDir)

	if _, err := b.RegisterExternalBuild(ExternalBuildInput{
		ID:         "bigcherry-safe-delete",
		Profile:    "rocm",
		BinaryPath: binary,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := b.Delete("bigcherry-safe-delete"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("source binary was touched by registry delete: %v", err)
	}
}

func TestRegisterExternalBuildRejectsMissingBinary(t *testing.T) {
	b := NewBuilder(t.TempDir())
	_, err := b.RegisterExternalBuild(ExternalBuildInput{
		ID:         "missing",
		Profile:    "rocm",
		BinaryPath: filepath.Join(t.TempDir(), "does-not-exist"),
	})
	if err == nil {
		t.Fatal("missing binary accepted")
	}
}

func fakeExternalServer(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "llama-server"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o644
	}
	if err := os.WriteFile(path, []byte("external test server"), mode); err != nil {
		t.Fatalf("write fake server: %v", err)
	}
	return path
}
