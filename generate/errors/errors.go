// Package errors implements the code generation for protoc-gen-sphere-errors.
// It emits Go error-helper methods for every protobuf enum annotated with the
// sphere.errors extension and skips files that declare no error enums.
package errors

import (
	"google.golang.org/protobuf/compiler/protogen"
)

// errorsPackage resolves to the standard library "errors" package, used for the
// errors.Join call in the generated Join helpers.
const errorsPackage = protogen.GoImportPath("errors")

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
