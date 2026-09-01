package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tmac1973/llama-toolchest/internal/builder"
)

func externalTestBinary(t *testing.T) string {
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
	if err := os.WriteFile(path, []byte("test server"), mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHandleRegisterExternalBuildJSONIsImmediatelyVisible(t *testing.T) {
	dataDir := t.TempDir()
	s := &Server{builder: builder.NewBuilder(dataDir)}
	binary := externalTestBinary(t)
	body, _ := json.Marshal(map[string]any{
		"id":          "bigcherry-control-linux",
		"profile":     "rocm",
		"binary_path": binary,
		"git_ref":     "deadbeef-upstream",
		"git_sha":     "0123456789abcdef",
		"tag":         "eb-1234-rb-5678",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/builds?external=1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rr := httptest.NewRecorder()

	s.handleTriggerBuild(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	got, ok := s.builder.Find("bigcherry-control-linux")
	if !ok {
		t.Fatal("build was not visible through live builder after registration")
	}
	if got.BinaryPath != binary || got.GitRef != "deadbeef-upstream" || got.GitSHA != "0123456789abcdef" {
		t.Fatalf("registered build=%+v", got)
	}
}

func TestHandleRegisterExternalBuildDuplicateAndReplace(t *testing.T) {
	s := &Server{builder: builder.NewBuilder(t.TempDir())}
	binary := externalTestBinary(t)
	post := func(replace bool, gitSHA string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{
			"id":          "same-id",
			"profile":     "rocm",
			"binary_path": binary,
			"git_sha":     gitSHA,
			"replace":     replace,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/builds?external=1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		s.handleTriggerBuild(rr, req)
		return rr
	}

	if rr := post(false, "old"); rr.Code != http.StatusCreated {
		t.Fatalf("first status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := post(false, "new"); rr.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := post(true, "new"); rr.Code != http.StatusCreated {
		t.Fatalf("replace status=%d body=%s", rr.Code, rr.Body.String())
	}
	got, _ := s.builder.Find("same-id")
	if got.GitSHA != "new" {
		t.Fatalf("GitSHA=%q want new", got.GitSHA)
	}
}

func TestHandleRegisterExternalBuildHTMXTriggersListRefresh(t *testing.T) {
	s := &Server{builder: builder.NewBuilder(t.TempDir())}
	binary := externalTestBinary(t)
	form := "id=ui-build&profile=rocm&binary_path=" + strings.ReplaceAll(binary, " ", "+")
	req := httptest.NewRequest(http.MethodPost, "/api/builds?external=1", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()

	s.handleTriggerBuild(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("HX-Trigger") != "buildsChanged" {
		t.Fatalf("HX-Trigger=%q", rr.Header().Get("HX-Trigger"))
	}
}
