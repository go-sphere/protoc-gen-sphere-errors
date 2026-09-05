// Package template renders the Go error-helper methods emitted for each error
// enum. It owns the data model (ErrorWrapper / ErrorInfo) and the embedded
// text/template used to produce the generated source.
package template

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
)

//go:embed template.tmpl
var defaultTemplate string

// ErrorInfo describes a single enum value rendered as an error case.
type ErrorInfo struct {
	Name  string
	Value string

	Status  int32
	Code    int32
	Reason  string
	Message string
}

// HasReason reports whether an explicit reason string was provided.
func (i *ErrorInfo) HasReason() bool {
	return i.Reason != ""
}

// ErrorWrapper is the template root: one error enum and its values, plus the
// already-qualified identifiers the generated code calls into.
type ErrorWrapper struct {
	Name           string
	Errors         []*ErrorInfo
	NewErrorFunc   string
	ErrorsJoinFunc string
}

// Renderer owns a parsed error generation template. It is immutable after
// construction and safe to reuse for every file in one plugin invocation.
type Renderer struct {
	template *template.Template
}

// NewRenderer loads and parses the embedded template, or the file at path when
// path is non-empty.
func NewRenderer(path string) (*Renderer, error) {
	source := defaultTemplate
	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read template %q: %w", path, err)
		}
		source = string(raw)
	}
	tmpl, err := template.New("errors").Funcs(template.FuncMap{
		"goString": strconv.Quote,
	}).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &Renderer{template: tmpl}, nil
}

// Execute renders the error-helper methods for an enum descriptor.
func (r *Renderer) Execute(e *ErrorWrapper) (string, error) {
	var buf strings.Builder
	if err := r.template.Execute(&buf, e); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
