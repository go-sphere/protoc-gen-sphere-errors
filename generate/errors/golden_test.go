package errors

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/testutil"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files instead of comparing")

func TestGolden(t *testing.T) {
	tests := []struct {
		name       string
		pbFile     string
		protoName  string
		wantFile   bool
		goldenFile string
	}{
		{
			name:       "basic_errors",
			pbFile:     "testdata/pb/basic_errors.pb",
			protoName:  "basic_errors.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/basic_errors.errors.pb.go",
		},
		{
			name:       "mixed_enums",
			pbFile:     "testdata/pb/mixed_enums.pb",
			protoName:  "mixed_enums.proto",
			wantFile:   true,
			goldenFile: "testdata/golden/mixed_enums.errors.pb.go",
		},
		{
			name:      "no_errors",
			pbFile:    "testdata/pb/no_errors.pb",
			protoName: "no_errors.proto",
			wantFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := testutil.PluginFromPB(t, tt.pbFile, tt.protoName)
			genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
			if err != nil {
				t.Fatalf("GenerateFile failed: %v", err)
			}

			if !tt.wantFile {
				if genFile != nil {
					t.Error("expected no generated file, got one")
				}
				return
			}
			if genFile == nil {
				t.Fatal("expected generated file, got nil")
			}
			content := mustContent(t, genFile)

			if *updateGolden {
				if err := os.MkdirAll(filepath.Dir(tt.goldenFile), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tt.goldenFile, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated golden file: %s", tt.goldenFile)
				return
			}

			want, err := os.ReadFile(tt.goldenFile)
			if err != nil {
				t.Fatalf("read golden file (run `make update-golden` to create it): %v", err)
			}
			if string(want) != content {
				t.Errorf("generated content mismatch for %s.\n%s", tt.goldenFile, firstDiff(string(want), content))
			}
		})
	}
}

// firstDiff returns a short description of the first line that differs between
// want and got, which is enough to locate golden drift without a diff library.
func firstDiff(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	for i := 0; i < min(len(wl), len(gl)); i++ {
		if wl[i] != gl[i] {
			return "first difference at line " + strconv.Itoa(i+1) + ":\n  want: " + wl[i] + "\n  got:  " + gl[i]
		}
	}
	if len(wl) != len(gl) {
		return "line count differs: want " + strconv.Itoa(len(wl)) + ", got " + strconv.Itoa(len(gl))
	}
	return "files differ only in trailing content"
}
