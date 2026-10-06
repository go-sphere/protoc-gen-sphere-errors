package main

import (
	"flag"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-sphere/protoc-gen-sphere-errors/generate/errors"
)

func TestExtractConfig(t *testing.T) {
	want := errors.DefaultConfig()
	want.PluginVersion = "v" + version
	got, err := extractConfig()
	if err != nil {
		t.Fatalf("extractConfig() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractConfig() = %#v, want %#v", got, want)
	}

	original := *newErrorFunc
	t.Cleanup(func() { *newErrorFunc = original })
	*newErrorFunc = ";NewError"
	if _, err := extractConfig(); err == nil {
		t.Fatal("extractConfig() with invalid new_errors_func error = nil")
	}
}

// TestExtractConfigHTTPXOptIn covers the documented opt-out from the default
// constructor: new_errors_func=github.com/go-sphere/httpx;NewError, the
// pre-0.0.4 default, must still be accepted through the plugin parameter path.
func TestExtractConfigHTTPXOptIn(t *testing.T) {
	original := *newErrorFunc
	t.Cleanup(func() { *newErrorFunc = original })
	if err := flag.CommandLine.Set("new_errors_func", "github.com/go-sphere/httpx;NewError"); err != nil {
		t.Fatalf("set new_errors_func: %v", err)
	}
	got, err := extractConfig()
	if err != nil {
		t.Fatalf("extractConfig() error = %v", err)
	}
	if got.NewErrorFunc.GoImportPath != "github.com/go-sphere/httpx" || got.NewErrorFunc.GoName != "NewError" {
		t.Errorf("NewErrorFunc = %v, want github.com/go-sphere/httpx;NewError", got.NewErrorFunc)
	}
}

func TestVersionFlag(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "protoc-gen-sphere-errors")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, output)
	}

	command := exec.CommandContext(t.Context(), binPath, "-version")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run -version: %v\n%s", err, output)
	}
	if got, want := string(output), "protoc-gen-sphere-errors "+version+"\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}
