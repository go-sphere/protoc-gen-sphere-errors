// Package errors implements the code generation for protoc-gen-sphere-errors.
// It emits Go error-helper methods for every protobuf enum annotated with the
// sphere.errors extension and skips files that declare no error enums.
package errors

import (
	stderrors "errors"

	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
)

// errorsPackage resolves to the standard library "errors" package, used for the
// errors.Join call in the generated Join helpers.
const errorsPackage = protogen.GoImportPath("errors")

// Generator owns validated configuration and an immutable parsed template.
type Generator struct {
	cfg      *Config
	renderer *template.Renderer
}

// NewGenerator validates cfg and loads its template once for reuse across all
// files in a protoc invocation.
func NewGenerator(cfg *Config) (*Generator, error) {
	if cfg == nil {
		return nil, stderrors.New("config is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	renderer, err := template.NewRenderer(cfg.TemplateFile)
	if err != nil {
		return nil, err
	}
	return &Generator{cfg: new(*cfg), renderer: renderer}, nil
}

// GenerateFile is a convenience wrapper for generating one file. Callers that
// generate multiple files should construct a Generator and reuse it.
func GenerateFile(plugin *protogen.Plugin, file *protogen.File, cfg *Config) (*protogen.GeneratedFile, error) {
	generator, err := NewGenerator(cfg)
	if err != nil {
		return nil, err
	}
	return generator.GenerateFile(plugin, file)
}

// GenerateFile generates the <prefix>.errors.pb.go file for file. It returns a
// nil GeneratedFile (and nil error) when file declares no error enums.
func (g *Generator) GenerateFile(plugin *protogen.Plugin, file *protogen.File) (*protogen.GeneratedFile, error) {
	if len(file.Enums) == 0 || !hasErrorEnums(file.Enums) {
		return nil, nil
	}
	filename := file.GeneratedFilenamePrefix + ".errors.pb.go"
	generated := plugin.NewGeneratedFile(filename, file.GoImportPath)
	generateFileHeader(plugin, file, generated)
	if err := generateFileContent(file, generated, g.cfg, g.renderer); err != nil {
		return nil, err
	}
	return generated, nil
}
