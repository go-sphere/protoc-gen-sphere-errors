// Package template renders the Go error-helper methods emitted for each error
// enum. It owns the data model (ErrorWrapper / ErrorInfo) and the embedded
// text/template used to produce the generated source.
package template

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed template.tmpl
var errorsTemplate string

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
	NewErrorsFunc  string
	ErrorsJoinFunc string
}

// Execute renders the error-helper methods for the wrapped enum.
func (e *ErrorWrapper) Execute() (string, error) {
	tmpl, err := template.New("errors").Parse(errorsTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, e); err != nil {
		return "", err
	}
	return buf.String(), nil
}
