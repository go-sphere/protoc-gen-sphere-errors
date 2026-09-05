package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	defaultRenderer, err := NewRenderer("")
	if err != nil {
		t.Fatalf("default template should load: %v", err)
	}
	if _, err := NewRenderer(filepath.Join(t.TempDir(), "does-not-exist.tmpl")); err == nil {
		t.Fatal("missing template file should return an error")
	}

	custom := "// custom template for {{.Name}}\n"
	path := filepath.Join(t.TempDir(), "custom.tmpl")
	if err := os.WriteFile(path, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	customRenderer, err := NewRenderer(path)
	if err != nil {
		t.Fatalf("valid template file should load: %v", err)
	}
	out, err := customRenderer.Execute(&ErrorWrapper{Name: "MyError"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !strings.Contains(out, "// custom template for MyError") {
		t.Errorf("custom template not applied, got: %q", out)
	}
	defaultOut, err := defaultRenderer.Execute(&ErrorWrapper{Name: "MyError"})
	if err != nil {
		t.Fatalf("default Execute failed: %v", err)
	}
	if strings.Contains(defaultOut, "// custom template") {
		t.Fatal("custom renderer must not mutate the embedded default renderer")
	}
}

func TestExecuteQuotesReasonAndMessage(t *testing.T) {
	renderer, err := NewRenderer("")
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	out, err := renderer.Execute(&ErrorWrapper{
		Name:           "UserError",
		NewErrorFunc:   "httpx.NewError",
		ErrorsJoinFunc: "errors.Join",
		Errors: []*ErrorInfo{{
			Name:    "UserError",
			Value:   "USER_ERROR_NOT_FOUND",
			Status:  404,
			Code:    1001,
			Reason:  `user "bob" missing`,
			Message: "line1\nline2",
		}},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, `return "user \"bob\" missing"`) {
		t.Errorf("reason not Go-quoted, got:\n%s", out)
	}
	if !strings.Contains(out, `return "line1\nline2"`) {
		t.Errorf("message not Go-quoted, got:\n%s", out)
	}
}
