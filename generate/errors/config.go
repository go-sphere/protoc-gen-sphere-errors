package errors

import (
	"errors"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

const (
	defaultErrorsPackage = "github.com/go-sphere/httpx"

	// DefaultNewErrorFunc is the default error constructor in
	// "import/path;Ident" form.
	DefaultNewErrorFunc = defaultErrorsPackage + ";NewError"
)

// Config controls error-helper generation.
type Config struct {
	TemplateFile string

	// NewErrorFunc is the constructor the generated Join helpers call. It must
	// have the signature func(status, code int32, message string, err error) error.
	NewErrorFunc protogen.GoIdent
}

// ParseGoIdent parses an "import/path;Ident" string into a protogen.GoIdent.
func ParseGoIdent(raw string) (protogen.GoIdent, error) {
	importPath, goName, ok := strings.Cut(raw, ";")
	if !ok || importPath == "" || goName == "" || strings.Contains(goName, ";") {
		return protogen.GoIdent{}, errors.New("invalid GoIdent format, expected 'import/path;Ident'")
	}
	return protogen.GoIdent{
		GoName:       goName,
		GoImportPath: protogen.GoImportPath(importPath),
	}, nil
}

// Validate checks that the required error constructor is configured.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.NewErrorFunc.GoImportPath == "" || c.NewErrorFunc.GoName == "" {
		return errors.New("new_errors_func is required (format: 'import/path;Ident')")
	}
	return nil
}

// DefaultConfig returns the plugin's real defaults.
func DefaultConfig() *Config {
	ident, err := ParseGoIdent(DefaultNewErrorFunc)
	if err != nil {
		panic(err)
	}
	return &Config{NewErrorFunc: ident}
}
