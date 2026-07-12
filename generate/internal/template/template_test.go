package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceTemplateIfNeed verifies ENC-24: the errors plugin can override its
// embedded template with a --template_file, matching the sphere and route
// plugins.
func TestReplaceTemplateIfNeed(t *testing.T) {
	original := errorsTemplate
	t.Cleanup(func() { errorsTemplate = original })

	// An empty path is a no-op and leaves the embedded default in place.
	if err := ReplaceTemplateIfNeed(""); err != nil {
		t.Fatalf("empty path should be a no-op: %v", err)
	}
	if errorsTemplate != original {
		t.Fatal("empty path must not change the template")
	}

	// A missing file surfaces the read error.
	if err := ReplaceTemplateIfNeed(filepath.Join(t.TempDir(), "does-not-exist.tmpl")); err == nil {
		t.Fatal("missing template file should return an error")
	}

	// A real file replaces the template and is used when rendering.
	custom := "// custom template for {{.Name}}\n"
	path := filepath.Join(t.TempDir(), "custom.tmpl")
	if err := os.WriteFile(path, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceTemplateIfNeed(path); err != nil {
		t.Fatalf("valid template file should load: %v", err)
	}
	out, err := (&ErrorWrapper{Name: "MyError"}).Execute()
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !strings.Contains(out, "// custom template for MyError") {
		t.Errorf("custom template not applied, got: %q", out)
	}
}
