package errors

import (
	"strings"
	"testing"

	sphereerrors "github.com/go-sphere/errors/sphere/errors"
	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/testutil"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var testConfig = &Config{
	NewErrorsFunc: protogen.GoIdent{
		GoName:       "NewError",
		GoImportPath: "github.com/go-sphere/httpx",
	},
}

// --- Pure function unit tests (no protogen involved) ---

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

// --- Layer 2: functional tests loading precompiled .pb files ---

func TestGenerateFile_WithErrorEnums(t *testing.T) {
	plugin := testutil.PluginFromPB(t, "testdata/pb/basic_errors.pb", "basic_errors.proto")
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file, got nil")
	}
	content := mustContent(t, genFile)

	for _, want := range []string{
		"func (e UserError) Error() string",
		"func (e UserError) GetCode() int32",
		"func (e UserError) GetStatus() int32",
		"func (e UserError) GetMessage() string",
		"func (e UserError) Join(errs ...error) error",
		"func (e UserError) JoinWithMessage(msg string, errs ...error) error",
		"func (e OrderError) Error() string",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("generated content missing: %q", want)
		}
	}
	// The value without options falls back to the generated reason.
	if !strings.Contains(content, "UserError:USER_ERROR_DEFAULTED") {
		t.Error("expected generated reason for value without options")
	}
}

func TestGenerateFile_SkipsNormalEnums(t *testing.T) {
	plugin := testutil.PluginFromPB(t, "testdata/pb/mixed_enums.pb", "mixed_enums.proto")
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file for mixed enums, got nil")
	}
	content := mustContent(t, genFile)

	if !strings.Contains(content, "AuthError") {
		t.Error("expected AuthError in generated content")
	}
	if strings.Contains(content, "Color") {
		t.Error("normal enum Color should not appear in generated content")
	}
}

func TestGenerateFile_NoErrorAnnotations(t *testing.T) {
	plugin := testutil.PluginFromPB(t, "testdata/pb/no_errors.pb", "no_errors.proto")
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile != nil {
		t.Error("expected nil for file without error annotations, got non-nil")
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

func mustContent(t *testing.T, g *protogen.GeneratedFile) string {
	t.Helper()
	b, err := g.Content()
	if err != nil {
		t.Fatalf("GeneratedFile.Content() failed: %v", err)
	}
	return string(b)
}
