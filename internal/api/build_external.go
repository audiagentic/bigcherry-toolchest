package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/tmac1973/llama-toolchest/internal/builder"
)

// handleRegisterExternalBuild registers an already-built llama-server into
// the live Builder instance. Because the Builder is the same object used by
// router selection and benchmark jobs, a successful registration is visible
// immediately; no Toolchest restart or builds.json reload is required.
func (s *Server) handleRegisterExternalBuild(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         string `json:"id"`
		Tag        string `json:"tag"`
		Profile    string `json:"profile"`
		GitRef     string `json:"git_ref"`
		GitSHA     string `json:"git_sha"`
		BinaryPath string `json:"binary_path"`
		Replace    bool   `json:"replace"`
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.ID = r.FormValue("id")
		req.Tag = r.FormValue("tag")
		req.Profile = r.FormValue("profile")
		req.GitRef = r.FormValue("git_ref")
		req.GitSHA = r.FormValue("git_sha")
		req.BinaryPath = r.FormValue("binary_path")
		req.Replace = r.FormValue("replace") == "1" || r.FormValue("replace") == "on" || r.FormValue("replace") == "true"
	}

	result, err := s.builder.RegisterExternalBuild(builder.ExternalBuildInput{
		ID:         req.ID,
		Tag:        req.Tag,
		Profile:    req.Profile,
		GitRef:     req.GitRef,
		GitSHA:     req.GitSHA,
		BinaryPath: req.BinaryPath,
		Replace:    req.Replace,
	})
	if err != nil {
		if _, ok := err.(*builder.DuplicateBuildError); ok {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if isHTMX(r) {
		// The compiled-build list already listens for buildsChanged. This makes
		// the newly registered build appear immediately without a page reload.
		w.Header().Set("HX-Trigger", "buildsChanged")
		respondHTML(w)
		fmt.Fprintf(w, `<p><ins>Registered %s</ins></p>`, html.EscapeString(result.ID))
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, result)
}
