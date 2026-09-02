package api

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/tmac1973/llama-toolchest/web"
)

func TestWorkstationFontAvoidsVisibleSwap(t *testing.T) {
	b, err := fs.ReadFile(web.Static, "static/workstation-v3.css")
	if err != nil {
		t.Fatalf("read workstation stylesheet: %v", err)
	}
	css := string(b)

	for _, want := range []string{
		"font-family: 'IBM Plex Sans Workstation'",
		"font-display: block",
		"--pico-font-family-sans-serif: 'IBM Plex Sans Workstation'",
		"--pico-font-family: var(--pico-font-family-sans-serif) !important",
		"font-family: var(--pico-font-family) !important",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("workstation stylesheet missing no-swap font contract %q", want)
		}
	}

	// The base layout still carries an older IBM Plex Sans face with
	// font-display: swap for non-workstation themes. The workstation sheet must
	// use a distinct family name so that face can never win font matching here.
	if strings.Contains(css, "font-family: 'IBM Plex Sans';") {
		t.Error("workstation stylesheet must not reuse the legacy swap family")
	}
}
