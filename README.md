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

package shared.v1;

import "sphere/errors/errors.proto";

enum TestError {
  option (sphere.errors.default_status) = 500;
  TEST_ERROR_UNSPECIFIED = 0;
  TEST_ERROR_INVALID_FIELD_TEST1 = 1000 [(sphere.errors.options) = {
    status: 400
    reason: "INVALID_ARGUMENT"
    message: "Invalid field_test1 value"
  }];
  TEST_ERROR_INVALID_PATH_TEST2 = 1001 [(sphere.errors.options) = {
    status: 400
    message: "Invalid path_test2 parameter"
  }];
  TEST_ERROR_UNAUTHORIZED = 1002 [(sphere.errors.options) = {
    status: 401
    reason: "UNAUTHORIZED"
    message: "Authentication required"
  }];
  TEST_ERROR_FORBIDDEN = 1003 [(sphere.errors.options) = {
    status: 403
    reason: "FORBIDDEN"
    message: "Permission denied"
  }];
}

enum UserError {
  option (sphere.errors.default_status) = 500;
  USER_ERROR_UNSPECIFIED = 0;
  USER_ERROR_NOT_FOUND = 2001 [(sphere.errors.options) = {
    status: 404
    message: "User not found"
  }];
  USER_ERROR_EMAIL_EXISTS = 2002 [(sphere.errors.options) = {
    status: 409
    message: "Email already exists"
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

Example generated code for the `TestError` enum:

```go
// Error implements the error interface
func (e TestError) Error() string {
    switch e {
    case TestError_TEST_ERROR_UNSPECIFIED:
        return "TestError_TEST_ERROR_UNSPECIFIED"
    case TestError_TEST_ERROR_INVALID_FIELD_TEST1:
        return "INVALID_ARGUMENT"  // Uses reason when specified
    case TestError_TEST_ERROR_INVALID_PATH_TEST2:
        return "TestError_TEST_ERROR_INVALID_PATH_TEST2"
    case TestError_TEST_ERROR_UNAUTHORIZED:
        return "UNAUTHORIZED"  // Uses reason when specified
    case TestError_TEST_ERROR_FORBIDDEN:
        return "FORBIDDEN"  // Uses reason when specified
    default:
        return "TestError:UNKNOWN_ERROR"
    }
}

// GetCode returns the error code (enum value)
func (e TestError) GetCode() int32 {
    switch e {
    case TestError_TEST_ERROR_UNSPECIFIED:
        return 0
    case TestError_TEST_ERROR_INVALID_FIELD_TEST1:
        return 1000
    case TestError_TEST_ERROR_INVALID_PATH_TEST2:
        return 1001
    case TestError_TEST_ERROR_UNAUTHORIZED:
        return 1002
    case TestError_TEST_ERROR_FORBIDDEN:
        return 1003
    default:
        return 0
    }
}

// GetStatus returns the HTTP status code
func (e TestError) GetStatus() int32 {
    switch e {
    case TestError_TEST_ERROR_UNSPECIFIED:
        return 500  // Uses default_status
    case TestError_TEST_ERROR_INVALID_FIELD_TEST1:
        return 400
    case TestError_TEST_ERROR_INVALID_PATH_TEST2:
        return 400
    case TestError_TEST_ERROR_UNAUTHORIZED:
        return 401
    case TestError_TEST_ERROR_FORBIDDEN:
        return 403
    default:
        return 500  // Uses default_status
    }
}

// GetMessage returns the custom error message
func (e TestError) GetMessage() string {
    switch e {
    case TestError_TEST_ERROR_INVALID_FIELD_TEST1:
        return "Invalid field_test1 value"
    case TestError_TEST_ERROR_INVALID_PATH_TEST2:
        return "Invalid path_test2 parameter"
    case TestError_TEST_ERROR_UNAUTHORIZED:
        return "Authentication required"
    case TestError_TEST_ERROR_FORBIDDEN:
        return "Permission denied"
    default:
        return ""
    }
}

// Join wraps the error with additional errors
func (e TestError) Join(errs ...error) error {
    allErrs := append([]error{e}, errs...)
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

// JoinWithMessage wraps the error with a custom message and additional errors
func (e TestError) JoinWithMessage(msg string, errs ...error) error {
    allErrs := append([]error{e}, errs...)
    return statuserr.NewError(
        e.GetStatus(),
        e.GetCode(),
        msg,
        errors.Join(allErrs...),
    )
}
```

## Usage in Code

### Direct Error Returns

```go
func (s *service) ValidateField(field string) error {
    if field == "" {
        return sharedv1.TestError_TEST_ERROR_INVALID_FIELD_TEST1
    }
    return nil
}
```

### Error Handling in HTTP Handlers

```go
func (s *service) RunTest(ctx context.Context, req *sharedv1.RunTestRequest) (*sharedv1.RunTestResponse, error) {
    if req.FieldTest1 == "" {
        return nil, sharedv1.TestError_TEST_ERROR_INVALID_FIELD_TEST1
    }
    
    if req.PathTest2 <= 0 {
        return nil, sharedv1.TestError_TEST_ERROR_INVALID_PATH_TEST2
    }
    
    // Business logic here...
    
    return &sharedv1.RunTestResponse{
        FieldTest1: req.FieldTest1,
        PathTest1:  req.PathTest1,
    }, nil
}
```

### Error Wrapping with Context

```go
func (s *service) ProcessUser(userID int64) error {
    user, err := s.userRepo.GetUser(userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return sharedv1.UserError_USER_ERROR_NOT_FOUND
        }
        // Wrap with additional context
        return sharedv1.TestError_TEST_ERROR_INVALID_FIELD_TEST1.Join(err)
    }
    
    // Process user...
    return nil
}
```

### Custom Error Messages

```go
func (s *service) CreateUser(email string) error {
    if s.userExists(email) {
        return sharedv1.UserError_USER_ERROR_EMAIL_EXISTS.JoinWithMessage(
            fmt.Sprintf("User with email %s already exists", email),
        )
    }
    
    // Create user...
    return nil
}
```

### Error Handling in Middleware

```go
func ErrorHandlingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            // Check if it's a sphere error with status info
            if statusErr, ok := err.(interface {
                GetStatus() int32
                GetCode() int32
                GetMessage() string
            }); ok {
                c.JSON(int(statusErr.GetStatus()), gin.H{
                    "error": gin.H{
                        "code":    statusErr.GetCode(),
                        "message": statusErr.GetMessage(),
                    },
                })
                return
            }
            
            // Default error handling
            c.JSON(500, gin.H{
                "error": gin.H{
                    "code":    -1,
                    "message": "Internal server error",
                },
            })
        }
    }
}
```

## Features

- **HTTP Status Code Integration**: Each error automatically provides the correct HTTP status code
- **Custom Error Messages**: Support for human-readable error messages in multiple languages
- **Error Reasons**: Machine-readable reason codes for programmatic error handling
- **Error Composition**: `Join` and `JoinWithMessage` methods for error wrapping and context
- **Default Status Codes**: Enum-level default status codes with per-value overrides
- **Framework Integration**: Seamless integration with sphere error handling framework
- **Type Safety**: Generated errors implement Go's error interface with additional methods

## Error Handling Best Practices

### 1. Use Meaningful Error Codes

```protobuf
enum UserError {
  option (sphere.errors.default_status) = 500;
  
  USER_ERROR_UNSPECIFIED = 0;
  USER_ERROR_NOT_FOUND = 1001;        // Clear what the error is
  USER_ERROR_INVALID_EMAIL = 1002;    // Specific validation error
  USER_ERROR_DUPLICATE_EMAIL = 1003;  // Specific conflict error
}
```

### 2. Group Related Errors

```protobuf
// Authentication errors (1000-1099)
enum AuthError {
  option (sphere.errors.default_status) = 401;
  AUTH_ERROR_UNSPECIFIED = 0;
  AUTH_ERROR_INVALID_TOKEN = 1001;
  AUTH_ERROR_TOKEN_EXPIRED = 1002;
  AUTH_ERROR_INSUFFICIENT_PERMISSIONS = 1003;
}

// User management errors (2000-2099)
enum UserError {
  option (sphere.errors.default_status) = 400;
  USER_ERROR_UNSPECIFIED = 0;
  USER_ERROR_NOT_FOUND = 2001;
  USER_ERROR_INVALID_INPUT = 2002;
}
```

### 3. Provide Clear Messages

```protobuf
enum ValidationError {
  option (sphere.errors.default_status) = 400;
  
  VALIDATION_ERROR_UNSPECIFIED = 0;
  VALIDATION_ERROR_REQUIRED_FIELD = 1001 [(sphere.errors.options) = {
    status: 400
    reason: "REQUIRED_FIELD_MISSING"
    message: "Required field is missing"
  }];
  VALIDATION_ERROR_INVALID_FORMAT = 1002 [(sphere.errors.options) = {
    status: 400
    reason: "INVALID_FORMAT"
    message: "Field format is invalid"
  }];
}
```

### 4. Use Appropriate HTTP Status Codes

- `400`: Bad Request - Client input validation errors
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Permission denied
- `404`: Not Found - Resource doesn't exist
- `409`: Conflict - Resource conflict (e.g., duplicate email)
- `422`: Unprocessable Entity - Semantic validation errors
- `429`: Too Many Requests - Rate limiting
- `500`: Internal Server Error - Server-side errors
- `502`: Bad Gateway - External service errors
- `503`: Service Unavailable - Service temporarily down

## Integration with Other Sphere Components

The error plugin works seamlessly with other sphere components:

- **protoc-gen-sphere**: HTTP handlers automatically handle sphere errors and return appropriate status codes
- **sphere/server/ginx**: Response wrapper functions understand sphere errors
- **sphere/core/errors**: Base error handling framework
- **protovalidate**: Validation errors can be wrapped with sphere errors for consistent error responses

## Options

### Enum Options

- `sphere.errors.default_status`: Sets the default HTTP status code for all values in the enum

### Enum Value Options

- `sphere.errors.options`: Configures individual error values
  - `status`: HTTP status code (overrides default)
  - `reason`: Custom reason string (optional)
  - `message`: Human-readable error message

