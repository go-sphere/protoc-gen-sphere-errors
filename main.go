package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/go-sphere/protoc-gen-sphere-errors/generate/errors"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	defaultErrorsPackage = "github.com/go-sphere/httpx"
)

var (
	showVersion   = flag.Bool("version", false, "print the version and exit")
	newErrorsFunc = flag.String("new_errors_func", defaultErrorsPackage+";NewError", "new errors func, must be func(status, code int32, message string, err error) error")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-sphere-errors %v\n", "0.0.1")
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			errPkg := strings.Split(*newErrorsFunc, ";")
			if len(errPkg) != 2 {
				return fmt.Errorf("invalid new_errors_func format, expected 'path;ident'")
			}
			_, gErr := errors.GenerateFile(gen, f, &errors.Config{
				NewErrorsFunc: protogen.GoIdent{
					GoName:       errPkg[1],
					GoImportPath: protogen.GoImportPath(errPkg[0]),
				},
			})
			if gErr != nil {
				return gErr
			}
		}
		return nil
	})
}
