// Package errors implements the code generation for protoc-gen-sphere-errors.
// It emits Go error-helper methods for every protobuf enum annotated with the
// sphere.errors extension and skips files that declare no error enums.
package errors

import (
	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
)

// errorsPackage resolves to the standard library "errors" package, used for the
// errors.Join call in the generated Join helpers.
const errorsPackage = protogen.GoImportPath("errors")

// ReplaceTemplateIfNeed overrides the built-in code template with the file at
// path when path is non-empty. It must be called once before GenerateFile. It
// is a thin wrapper over the internal template package so that callers (e.g.
// main) need not import that internal package directly.
func ReplaceTemplateIfNeed(path string) error {
	return template.ReplaceTemplateIfNeed(path)
}

// Config controls error code generation.
type Config struct {
	// NewErrorsFunc is the constructor the generated Join helpers call. It must
	// have the signature func(status, code int32, message string, err error) error.
	NewErrorsFunc protogen.GoIdent
}

// GenerateFile generates the <prefix>.errors.pb.go file for file. It returns a
// nil GeneratedFile (and nil error) when file declares no error enums.
func GenerateFile(gen *protogen.Plugin, file *protogen.File, config *Config) (*protogen.GeneratedFile, error) {
	if len(file.Enums) == 0 || !hasErrorEnums(file.Enums) {
		return nil, nil
	}
	filename := file.GeneratedFilenamePrefix + ".errors.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	generateFileHeader(gen, file, g)
	if err := generateFileContent(file, g, config); err != nil {
		return nil, err
	}
	return g, nil
}
