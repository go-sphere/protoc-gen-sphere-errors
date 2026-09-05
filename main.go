package main

import (
	"flag"
	"fmt"

	"github.com/go-sphere/protoc-gen-sphere-errors/generate/errors"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.0.1"

var (
	showVersion  = flag.Bool("version", false, "print the version and exit")
	newErrorFunc = flag.String("new_errors_func", errors.DefaultNewErrorFunc, "new error func, must be func(status, code int32, message string, err error) error")
	templateFile = flag.String("template_file", "", "template file, if not set, use default template")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-sphere-errors %s\n", version)
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(run)
}

func run(plugin *protogen.Plugin) error {
	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	cfg, err := extractConfig()
	if err != nil {
		return err
	}
	generator, err := errors.NewGenerator(cfg)
	if err != nil {
		return err
	}
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		if _, err := generator.GenerateFile(plugin, file); err != nil {
			return err
		}
	}
	return nil
}

func extractConfig() (*errors.Config, error) {
	ident, err := errors.ParseGoIdent(*newErrorFunc)
	if err != nil {
		return nil, err
	}
	cfg := errors.DefaultConfig()
	cfg.TemplateFile = *templateFile
	cfg.NewErrorFunc = ident
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
