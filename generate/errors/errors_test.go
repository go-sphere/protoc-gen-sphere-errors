package errors

import (
	"testing"

	sphereerrors "github.com/go-sphere/errors/sphere/errors"
	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/testutil"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var testConfig = DefaultConfig()

// --- Pure function unit tests (no protogen involved) ---

func TestParseGoIdent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: "github.com/example/errors;NewError"},
		{name: "empty", input: "", wantErr: true},
		{name: "missing separator", input: "github.com/example/errors/NewError", wantErr: true},
		{name: "multiple separators", input: "github.com/example/errors;NewError;Other", wantErr: true},
		{name: "empty import path", input: ";NewError", wantErr: true},
		{name: "empty identifier", input: "github.com/example/errors;", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGoIdent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseGoIdent(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig().Validate() error = %v", err)
	}
	if got := string(cfg.NewErrorFunc.GoImportPath) + ";" + cfg.NewErrorFunc.GoName; got != DefaultNewErrorFunc {
		t.Errorf("default NewErrorFunc = %q, want %q", got, DefaultNewErrorFunc)
	}
}

func TestConfigValidate(t *testing.T) {
	if err := (*Config)(nil).Validate(); err == nil {
		t.Fatal("nil Config.Validate() error = nil")
	}
	if err := (&Config{}).Validate(); err == nil {
		t.Fatal("empty Config.Validate() error = nil")
	}
}

func TestResolveErrorInfo(t *testing.T) {
	tests := []struct {
		name          string
		opt           *sphereerrors.Error
		defaultStatus int32
		wantStatus    int32
		wantReason    string
		wantMessage   string
	}{
		{
			name:          "empty options use defaults",
			opt:           &sphereerrors.Error{},
			defaultStatus: 400,
			wantStatus:    400,
			wantReason:    "UserError:USER_NOT_FOUND",
			wantMessage:   "",
		},
		{
			name:          "explicit zero status falls back to default",
			opt:           &sphereerrors.Error{Status: 0, Reason: "kept"},
			defaultStatus: 500,
			wantStatus:    500,
			wantReason:    "kept",
			wantMessage:   "",
		},
		{
			name:          "explicit status is kept",
			opt:           &sphereerrors.Error{Status: 404, Reason: "nope", Message: "missing"},
			defaultStatus: 400,
			wantStatus:    404,
			wantReason:    "nope",
			wantMessage:   "missing",
		},
		{
			name:          "empty reason is generated from names",
			opt:           &sphereerrors.Error{Status: 403},
			defaultStatus: 400,
			wantStatus:    403,
			wantReason:    "UserError:USER_NOT_FOUND",
			wantMessage:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveErrorInfo("UserError", "USER_NOT_FOUND", 2, tt.opt, tt.defaultStatus)
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %d, want %d", got.Status, tt.wantStatus)
			}
			if got.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, tt.wantReason)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantMessage)
			}
			if got.Code != 2 {
				t.Errorf("Code = %d, want 2", got.Code)
			}
			if got.Name != "UserError" || got.Value != "USER_NOT_FOUND" {
				t.Errorf("Name/Value = %q/%q", got.Name, got.Value)
			}
		})
	}
}

func TestFormatProtocVersion(t *testing.T) {
	tests := []struct {
		name string
		v    *pluginpb.Version
		want string
	}{
		{name: "nil", v: nil, want: "(unknown)"},
		{
			name: "no suffix",
			v:    &pluginpb.Version{Major: proto.Int32(5), Minor: proto.Int32(29), Patch: proto.Int32(3)},
			want: "v5.29.3",
		},
		{
			name: "with suffix",
			v:    &pluginpb.Version{Major: proto.Int32(4), Minor: proto.Int32(25), Patch: proto.Int32(0), Suffix: proto.String("rc1")},
			want: "v4.25.0-rc1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatProtocVersion(tt.v); got != tt.want {
				t.Errorf("formatProtocVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Layer 1: hand-written descriptors (skip logic, no extensions, no .pb) ---

func TestGenerateFile_NoEnums(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("empty.proto"),
		Package: proto.String("tests.empty"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/empty")},
	}
	plugin := mustPluginFromFD(t, fd)
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for file with no enums, got non-nil")
	}
}

func TestGenerateFile_OnlyNormalEnum(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("normal.proto"),
		Package: proto.String("tests.normal"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/normal")},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name: proto.String("Color"),
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("COLOR_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: proto.String("COLOR_RED"), Number: proto.Int32(1)},
				},
			},
		},
	}
	plugin := mustPluginFromFD(t, fd)
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for file with only a normal enum, got non-nil")
	}
}

func TestNewGeneratorSnapshotsConfig(t *testing.T) {
	cfg := DefaultConfig()
	generator, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	cfg.NewErrorFunc.GoName = "Changed"
	if generator.cfg.NewErrorFunc.GoName != "NewError" {
		t.Fatalf("Generator NewErrorFunc = %q after caller mutation", generator.cfg.NewErrorFunc.GoName)
	}
	if _, err := NewGenerator(nil); err == nil {
		t.Fatal("NewGenerator(nil) error = nil")
	}
}

// --- helpers ---

func mustPluginFromFD(t *testing.T, fd *descriptorpb.FileDescriptorProto) *protogen.Plugin {
	t.Helper()
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{fd.GetName()},
		ProtoFile:      []*descriptorpb.FileDescriptorProto{fd},
	}
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}
	return plugin
}
