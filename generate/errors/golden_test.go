package errors

import (
	"flag"
	"os"
	"path/filepath"
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
	wl := splitLines(want)
	gl := splitLines(got)
	n := len(wl)
	if len(gl) < n {
		n = len(gl)
	}
	for i := 0; i < n; i++ {
		if wl[i] != gl[i] {
			return "first difference at line " + itoa(i+1) + ":\n  want: " + wl[i] + "\n  got:  " + gl[i]
		}
	}
	if len(wl) != len(gl) {
		return "line count differs: want " + itoa(len(wl)) + ", got " + itoa(len(gl))
	}
	return "files differ only in trailing content"
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
