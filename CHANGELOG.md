# Changelog

## Unreleased

## v0.0.5 (2026-11-19)

### Added

- Support non-PascalCase top-level enum names (using `enum.GoIdent.GoName`) and
  message-nested error enums (walking `file.Messages` recursively while
  skipping map entries), using `ErrorInfo.CaseIdent()` in generated `switch`
  cases.

### Fixed

- Skip generating `.errors.pb.go` when no error enum in the file has any
  non-zero value (`number != 0`), eliminating the duplicate `hasErrorEnums`
  scan.
- Validate Go identifiers and import paths in `ParseGoIdent` and
  `Config.Validate`.
- Fix line 1 `gotype` hint and remove the unreachable fallback `else` branch in
  the embedded template.

## v0.0.4 (2026-10-08)

### Changed

- **BREAKING**: default `new_errors_func` from `github.com/go-sphere/httpx;NewError`
  to `github.com/go-sphere/errors/sphere/errors;NewError` (requiring
  `github.com/go-sphere/errors v0.0.3`+). Set
  `new_errors_func=github.com/go-sphere/httpx;NewError` explicitly to retain
  the previous output.
- Record plugin version in the generated file header (`Config.PluginVersion`).
