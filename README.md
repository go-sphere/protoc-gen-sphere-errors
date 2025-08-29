# protoc-gen-sphere-errors

`protoc-gen-sphere-errors` is a protoc plugin that generates error handling code from `.proto` files. It is designed to inspect enum definitions within your protobuf files and automatically generate corresponding error handling code based on the sphere errors framework. This plugin creates Go code that provides structured error handling with HTTP status codes, error codes, and customizable messages.

This code is inspired by [protoc-gen-go-errors](https://github.com/go-kratos/kratos/tree/main/cmd/protoc-gen-go-errors) but is specifically designed for the go-sphere framework.

## Features

- Generates error structs with HTTP status codes
- Supports custom error messages and reasons
- Provides `Join` and `JoinWithMessage` methods for error composition
- Integrates with the sphere error handling framework
- Supports default status codes for enum types
- Individual error value customization through options

## Installation

To install `protoc-gen-sphere-errors`, use the following command:

```bash
go install github.com/go-sphere/protoc-gen-sphere-errors@latest
```

## Prerequisites

You need to have the sphere errors proto definitions in your project. Add the following dependency to your `buf.yaml`:

```yaml
deps:
  - buf.build/go-sphere/errors
```

## Usage with Buf

To use `protoc-gen-sphere-errors` with `buf`, you can configure it in your `buf.gen.yaml` file. Here is an example configuration:

```yaml
version: v2
managed:
  enabled: true
  disable:
    - file_option: go_package_prefix
      module: buf.build/go-sphere/errors
  override:
    - file_option: go_package_prefix
      value: github.com/go-sphere/sphere-layout/api
plugins:
  - local: protoc-gen-sphere-errors
    out: api
    opt: paths=source_relative
```

## Proto Definition Example

Here's how to define error enums in your `.proto` files:

```protobuf
syntax = "proto3";

package dash.v1;

import "sphere/errors/errors.proto";

enum AdminError {
  option (sphere.errors.default_status) = 500;
  ADMIN_ERROR_UNSPECIFIED = 0;
  ADMIN_ERROR_CANNOT_DELETE_SELF = 1001 [(sphere.errors.options) = {
    status: 400
    message: "不能删除当前登录的管理员账号"
  }];
  ADMIN_ERROR_INVALID_CREDENTIALS = 1002 [(sphere.errors.options) = {
    status: 401
    message: "用户名或密码错误"
  }];
}
```

## Generated Code

The plugin generates Go code with the following methods for each error enum:

- `Error() string` - Returns a string representation of the error
- `GetCode() int32` - Returns the error code (enum value)
- `GetStatus() int32` - Returns the HTTP status code
- `GetMessage() string` - Returns the custom error message
- `Join(errs ...error) error` - Wraps the error with additional errors
- `JoinWithMessage(msg string, errs ...error) error` - Wraps with custom message

Example generated code:

```go
func (e AdminError) Error() string {
    switch e {
    case AdminError_ADMIN_ERROR_UNSPECIFIED:
        return "AdminError:ADMIN_ERROR_UNSPECIFIED"
    case AdminError_ADMIN_ERROR_CANNOT_DELETE_SELF:
        return "AdminError:ADMIN_ERROR_CANNOT_DELETE_SELF"
    default:
        return "AdminError:UNKNOWN_ERROR"
    }
}

func (e AdminError) GetStatus() int32 {
    switch e {
    case AdminError_ADMIN_ERROR_UNSPECIFIED:
        return 500
    case AdminError_ADMIN_ERROR_CANNOT_DELETE_SELF:
        return 400
    default:
        return 500
    }
}

func (e AdminError) Join(errs ...error) error {
    allErrs := append(errs, e)
    msg := e.GetMessage()
    if msg == "" {
        msg = e.Error()
    }
    return statuserr.NewError(
        e.GetStatus(),
        e.GetCode(),
        msg,
        errors.Join(allErrs...),
    )
}
```

## Usage in Code

```go
// Direct error return
return nil, dashv1.AdminError_ADMIN_ERROR_CANNOT_DELETE_SELF

// With additional context
return nil, dashv1.AdminError_ADMIN_ERROR_CANNOT_DELETE_SELF.Join(originalErr)

// With custom message
return nil, dashv1.AdminError_ADMIN_ERROR_CANNOT_DELETE_SELF.JoinWithMessage("Additional context", originalErr)
```

## Options

### Enum Options

- `sphere.errors.default_status`: Sets the default HTTP status code for all values in the enum

### Enum Value Options

- `sphere.errors.options`: Configures individual error values
  - `status`: HTTP status code (overrides default)
  - `reason`: Custom reason string (optional)
  - `message`: Human-readable error message

