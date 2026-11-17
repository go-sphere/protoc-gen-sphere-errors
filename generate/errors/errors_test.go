package errors

import (
	"go/parser"
	"go/token"
	"strings"
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
		{name: "stdlib single segment", input: "errors;NewError"},
		{name: "empty", input: "", wantErr: true},
		{name: "missing separator", input: "github.com/example/errors/NewError", wantErr: true},
		{name: "multiple separators", input: "github.com/example/errors;NewError;Other", wantErr: true},
		{name: "empty import path", input: ";NewError", wantErr: true},
		{name: "empty identifier", input: "github.com/example/errors;", wantErr: true},
		{name: "identifier starts with digit", input: "github.com/example/errors;123Error", wantErr: true},
		{name: "identifier with space", input: "github.com/example/errors;New Error", wantErr: true},
		{name: "identifier with dot", input: "github.com/example/errors;pkg.NewError", wantErr: true},
		{name: "keyword identifier", input: "github.com/example/errors;func", wantErr: true},
		{name: "call expression identifier", input: "github.com/example/errors;NewError()", wantErr: true},
		{name: "leading slash import path", input: "/github.com/example/errors;NewError", wantErr: true},
		{name: "trailing slash import path", input: "github.com/example/errors/;NewError", wantErr: true},
		{name: "double slash import path", input: "github.com//example/errors;NewError", wantErr: true},
		{name: "relative dot segment", input: "./errors;NewError", wantErr: true},
		{name: "parent dot segment", input: "github.com/../errors;NewError", wantErr: true},
		{name: "space in import path", input: "github.com/example/bad pkg;NewError", wantErr: true},
		{name: "version suffix in import path", input: "github.com/example/errors@v1;NewError", wantErr: true},
		{name: "quote in import path", input: `github.com/example/"errors";NewError`, wantErr: true},
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

// TestDefaultNewErrorFuncValue pins the default constructor. Changing it is a
// breaking output change (guidelines §11): it must come with a release note and
// golden compat cases for the previous default.
func TestDefaultNewErrorFuncValue(t *testing.T) {
	const want = "github.com/go-sphere/errors/sphere/errors;NewError"
	if DefaultNewErrorFunc != want {
		t.Fatalf("DefaultNewErrorFunc = %q, want %q", DefaultNewErrorFunc, want)
	}
}

func TestConfigValidate(t *testing.T) {
	if err := (*Config)(nil).Validate(); err == nil {
		t.Fatal("nil Config.Validate() error = nil")
	}
	if err := (&Config{}).Validate(); err == nil {
		t.Fatal("empty Config.Validate() error = nil")
	}
	if err := (&Config{NewErrorFunc: protogen.GoIdent{GoImportPath: "bad path", GoName: "NewError"}}).Validate(); err == nil {
		t.Fatal("invalid import path Config.Validate() error = nil")
	}
	if err := (&Config{NewErrorFunc: protogen.GoIdent{GoImportPath: "example.com/errors", GoName: "123Bad"}}).Validate(); err == nil {
		t.Fatal("invalid identifier Config.Validate() error = nil")
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

func TestGenerateFile_ZeroOnlyErrorEnum(t *testing.T) {
	enumOpts := &descriptorpb.EnumOptions{}
	proto.SetExtension(enumOpts, sphereerrors.E_DefaultStatus, int32(400))

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("zero_only.proto"),
		Package: proto.String("tests.zero"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/zero")},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name:    proto.String("ZeroError"),
				Options: enumOpts,
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("ZERO_ERROR_UNSPECIFIED"), Number: proto.Int32(0)},
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
		t.Error("expected nil for zero-only error enum, got non-nil")
	}
}

func TestGenerateFile_SnakeCaseTopLevelEnum(t *testing.T) {
	enumOpts := &descriptorpb.EnumOptions{}
	proto.SetExtension(enumOpts, sphereerrors.E_DefaultStatus, int32(400))

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("snake.proto"),
		Package: proto.String("tests.snake"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/snake")},
		EnumType: []*descriptorpb.EnumDescriptorProto{
			{
				Name:    proto.String("user_error"),
				Options: enumOpts,
				Value: []*descriptorpb.EnumValueDescriptorProto{
					{Name: proto.String("USER_ERROR_UNSPECIFIED"), Number: proto.Int32(0)},
					{Name: proto.String("USER_ERROR_INVALID"), Number: proto.Int32(1)},
				},
			},
		},
	}
	plugin := mustPluginFromFD(t, fd)
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file for snake_case enum, got nil")
	}
	raw, err := genFile.Content()
	if err != nil {
		t.Fatalf("genFile.Content: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "snake.errors.pb.go", raw, parser.ParseComments); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n%s", err, raw)
	}
	content := string(raw)
	for _, want := range []string{
		"func (e UserError) Error() string",
		"case UserError_USER_ERROR_INVALID:",
		`return "UserError:USER_ERROR_INVALID"`,
		`return "UserError:UNKNOWN_ERROR"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("generated content missing %q:\n%s", want, content)
		}
	}
}

func TestGenerateFile_NestedMessageErrorEnum(t *testing.T) {
	nestedOpts := &descriptorpb.EnumOptions{}
	proto.SetExtension(nestedOpts, sphereerrors.E_DefaultStatus, int32(404))

	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("nested.proto"),
		Package: proto.String("tests.nested"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/nested")},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("outer"),
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("inner"),
						EnumType: []*descriptorpb.EnumDescriptorProto{
							{
								Name:    proto.String("nested_error"),
								Options: nestedOpts,
								Value: []*descriptorpb.EnumValueDescriptorProto{
									{Name: proto.String("NESTED_ERROR_UNSPECIFIED"), Number: proto.Int32(0)},
									{Name: proto.String("NESTED_ERROR_NOT_FOUND"), Number: proto.Int32(1)},
								},
							},
						},
					},
					{
						Name:    proto.String("LabelsEntry"),
						Options: &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)},
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:     proto.String("key"),
								Number:   proto.Int32(1),
								JsonName: proto.String("key"),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
								Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
							},
							{
								Name:     proto.String("value"),
								Number:   proto.Int32(2),
								JsonName: proto.String("value"),
								Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
								Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
							},
						},
					},
				},
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:     proto.String("labels"),
						Number:   proto.Int32(1),
						JsonName: proto.String("labels"),
						Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
						TypeName: proto.String(".tests.nested.outer.LabelsEntry"),
						Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
					},
				},
			},
		},
	}
	plugin := mustPluginFromFD(t, fd)
	genFile, err := GenerateFile(plugin, testutil.FileToGenerate(t, plugin), testConfig)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}
	if genFile == nil {
		t.Fatal("expected generated file for nested enum, got nil")
	}
	raw, err := genFile.Content()
	if err != nil {
		t.Fatalf("genFile.Content: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "nested.errors.pb.go", raw, parser.ParseComments); err != nil {
		t.Fatalf("generated code is not valid Go: %v\n%s", err, raw)
	}
	content := string(raw)
	for _, want := range []string{
		"func (e OuterInnerNestedError) Error() string",
		"case OuterInner_NESTED_ERROR_NOT_FOUND:",
		`return "OuterInnerNestedError:NESTED_ERROR_NOT_FOUND"`,
		`return "OuterInnerNestedError:UNKNOWN_ERROR"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("generated content missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "LabelsEntry") {
		t.Errorf("map entry message should be skipped:\n%s", content)
	}

	t.Run("pascalcase nested enum", func(t *testing.T) {
		pascalOpts := &descriptorpb.EnumOptions{}
		proto.SetExtension(pascalOpts, sphereerrors.E_DefaultStatus, int32(400))

		pfd := &descriptorpb.FileDescriptorProto{
			Name:    proto.String("pascal_nested.proto"),
			Package: proto.String("tests.pascalnested"),
			Syntax:  proto.String("proto3"),
			Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/example/pascalnested")},
			MessageType: []*descriptorpb.DescriptorProto{
				{
					Name: proto.String("OuterMsg"),
					EnumType: []*descriptorpb.EnumDescriptorProto{
						{
							Name:    proto.String("InnerError"),
							Options: pascalOpts,
							Value: []*descriptorpb.EnumValueDescriptorProto{
								{Name: proto.String("INNER_ERROR_UNSPECIFIED"), Number: proto.Int32(0)},
								{Name: proto.String("INNER_ERROR_BAD"), Number: proto.Int32(1)},
							},
						},
					},
				},
			},
		}
		pPlugin := mustPluginFromFD(t, pfd)
		pGenFile, err := GenerateFile(pPlugin, testutil.FileToGenerate(t, pPlugin), testConfig)
		if err != nil {
			t.Fatalf("GenerateFile failed: %v", err)
		}
		if pGenFile == nil {
			t.Fatal("expected generated file, got nil")
		}
		pRaw, err := pGenFile.Content()
		if err != nil {
			t.Fatalf("genFile.Content: %v", err)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), "pascal_nested.errors.pb.go", pRaw, parser.ParseComments); err != nil {
			t.Fatalf("generated code is not valid Go: %v\n%s", err, pRaw)
		}
		pContent := string(pRaw)
		for _, want := range []string{
			"func (e OuterMsg_InnerError) Error() string",
			"case OuterMsg_INNER_ERROR_BAD:",
			`return "OuterMsg_InnerError:INNER_ERROR_BAD"`,
			`return "OuterMsg_InnerError:UNKNOWN_ERROR"`,
		} {
			if !strings.Contains(pContent, want) {
				t.Errorf("generated content missing %q:\n%s", want, pContent)
			}
		}
	})
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
