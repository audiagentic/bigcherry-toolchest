package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tmac1973/llama-toolchest/internal/builder"
)

func main() {
	var in builder.ExternalBuildInput
	var dataDir string

	flag.StringVar(&dataDir, "data-dir", "", "Toolchest data directory containing config/builds.json")
	flag.StringVar(&in.ID, "id", "", "stable build/lane id, e.g. bigcherry-b10227-hip")
	flag.StringVar(&in.Profile, "profile", "", "Toolchest build profile, e.g. rocm or vulkan")
	flag.StringVar(&in.BinaryPath, "binary", "", "llama-server executable or directory containing it")
	flag.StringVar(&in.GitRef, "git-ref", "", "source/base ref, e.g. b10227 or tuning-code-rebase")
	flag.StringVar(&in.GitSHA, "git-sha", "", "exact source commit SHA when known")
	flag.StringVar(&in.Tag, "tag", "", "optional human label, e.g. bigcherry-mmq")
	flag.BoolVar(&in.Replace, "replace", false, "replace an existing registry entry with the same id")
	flag.Parse()

	if dataDir == "" {
		fatalf("--data-dir is required")
	}

	b := builder.NewBuilder(dataDir)
	result, err := b.RegisterExternalBuild(in)
	if err != nil {
		fatalf("%v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fatalf("encode result: %v", err)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "toolchest-import-build: "+format+"\n", args...)
	os.Exit(1)
}
