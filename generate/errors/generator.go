package errors

import (
	"github.com/go-sphere/errors/sphere/errors"
	"github.com/go-sphere/protoc-gen-sphere-errors/generate/internal/template"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// generateFileContent renders the error-helper methods for every error enum in
// file and writes them to g.
func generateFileContent(file *protogen.File, g *protogen.GeneratedFile, config *Config) error {
	newErrorsFunc := g.QualifiedGoIdent(config.NewErrorsFunc)
	errorsJoinFunc := g.QualifiedGoIdent(errorsPackage.Ident("Join"))
	for _, enum := range file.Enums {
		ew := buildErrorWrapper(enum, newErrorsFunc, errorsJoinFunc)
		if ew == nil {
			continue
		}
		content, err := ew.Execute()
		if err != nil {
			return err
		}
		g.P(content)
		g.P("\n\n")
	}
	return nil
}

// buildErrorWrapper builds a template.ErrorWrapper from an enum. It returns nil
// when the enum is not an error enum (the default_status option is missing) or
// when it has no values. newErrorsFunc and errorsJoinFunc must be the already
// qualified Go identifiers used by the generated code.
func buildErrorWrapper(enum *protogen.Enum, newErrorsFunc, errorsJoinFunc string) *template.ErrorWrapper {
	if !proto.HasExtension(enum.Desc.Options(), errors.E_DefaultStatus) {
		return nil
	}
	defaultStatus, _ := proto.GetExtension(enum.Desc.Options(), errors.E_DefaultStatus).(int32)
	ew := &template.ErrorWrapper{
		Name:           string(enum.Desc.Name()),
		NewErrorsFunc:  newErrorsFunc,
		ErrorsJoinFunc: errorsJoinFunc,
	}
	// When allow_alias is enabled an enum can declare several names for the
	// same number, each a distinct EnumValueDescriptor. Emitting a switch case
	// per descriptor would produce duplicate `case X_A:`/`case X_B:` entries
	// that share a value, which the Go compiler rejects. Keep only the first
	// declared name for each number.
	seen := make(map[int32]struct{}, len(enum.Values))
	for _, v := range enum.Values {
		number := int32(v.Desc.Number())
		// The zero value (typically <ENUM>_UNSPECIFIED) is the Go zero value of
		// the type and semantically means "no error". Emitting a case for it
		// would turn the zero value into a defined error (Error() non-empty,
		// GetStatus() = default_status), so a plain `var e SomeError` returned
		// as error would look like a real 500-ish failure. Skip number 0 so the
		// zero value falls through to the switch default (unknown / 500) instead
		// of being a mapped error.
		if number == 0 {
			continue
		}
		if _, ok := seen[number]; ok {
			continue
		}
		seen[number] = struct{}{}
		info := resolveErrorInfo(
			string(enum.Desc.Name()),
			string(v.Desc.Name()),
			number,
			enumValueOptions(v),
			defaultStatus,
		)
		ew.Errors = append(ew.Errors, info)
	}
	if len(ew.Errors) == 0 {
		return nil
	}
	return ew
}

// resolveErrorInfo computes the final template.ErrorInfo for an enum value,
// falling back to the enum's default status and a generated reason when the
// value omits them. It is pure and independent of protogen.
func resolveErrorInfo(enumName, valueName string, code int32, opt *errors.Error, defaultStatus int32) *template.ErrorInfo {
	status := opt.GetStatus()
	if status == 0 {
		status = defaultStatus
	}
	reason := opt.GetReason()
	if reason == "" {
		reason = enumName + ":" + valueName
	}
	return &template.ErrorInfo{
		Name:    enumName,
		Value:   valueName,
		Status:  status,
		Code:    code,
		Reason:  reason,
		Message: opt.GetMessage(),
	}
}

// enumValueOptions returns the errors.Error options attached to an enum value,
// or an empty value when none are set.
func enumValueOptions(v *protogen.EnumValue) *errors.Error {
	if proto.HasExtension(v.Desc.Options(), errors.E_Options) {
		if opt, ok := proto.GetExtension(v.Desc.Options(), errors.E_Options).(*errors.Error); ok && opt != nil {
			return opt
		}
	}
	return &errors.Error{}
}

// hasErrorEnums reports whether enums contains at least one error enum (an enum
// carrying the default_status option and at least one value).
func hasErrorEnums(enums []*protogen.Enum) bool {
	for _, v := range enums {
		if proto.HasExtension(v.Desc.Options(), errors.E_DefaultStatus) && len(v.Values) > 0 {
			return true
		}
	}
	return false
}
