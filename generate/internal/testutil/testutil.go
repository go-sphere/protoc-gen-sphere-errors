// Package testutil provides shared helpers for building real protogen plugins
// from precompiled FileDescriptorSet (.pb) files in tests.
package testutil

import (
	"os"
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

// LoadDescriptorSet reads and unmarshals a FileDescriptorSet produced by
// `protoc --descriptor_set_out=... --include_imports`.
func LoadDescriptorSet(t *testing.T, path string) *descriptorpb.FileDescriptorSet {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read descriptor set %q: %v", path, err)
	}
	var set descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &set); err != nil {
		t.Fatalf("unmarshal descriptor set %q: %v", path, err)
	}
	if len(set.File) == 0 {
		t.Fatalf("descriptor set %q contains no files", path)
	}
	return &set
}

// MustCreatePlugin builds a real *protogen.Plugin from a descriptor set. The set
// must include every dependency (compile with --include_imports); fileToGenerate
// is the proto path (relative to the proto_path) that should be generated.
//
// The compiler version is fixed so generated headers stay deterministic.
func MustCreatePlugin(t *testing.T, set *descriptorpb.FileDescriptorSet, fileToGenerate string) *protogen.Plugin {
	t.Helper()
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{fileToGenerate},
		ProtoFile:      set.File,
		CompilerVersion: &pluginpb.Version{
			Major: proto.Int32(5),
			Minor: proto.Int32(29),
			Patch: proto.Int32(0),
		},
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("create plugin for %q: %v", fileToGenerate, err)
	}
	return plugin
}

// PluginFromPB is a convenience wrapper combining LoadDescriptorSet and
// MustCreatePlugin.
func PluginFromPB(t *testing.T, pbPath, fileToGenerate string) *protogen.Plugin {
	t.Helper()
	return MustCreatePlugin(t, LoadDescriptorSet(t, pbPath), fileToGenerate)
}

// FileToGenerate returns the single file marked for generation in the plugin.
func FileToGenerate(t *testing.T, plugin *protogen.Plugin) *protogen.File {
	t.Helper()
	for _, f := range plugin.Files {
		if f.Generate {
			return f
		}
	}
	t.Fatal("no file marked for generation")
	return nil
}
