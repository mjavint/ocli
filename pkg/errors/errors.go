package errors

import (
	"fmt"
)

// ErrorCode represents different types of errors
type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeInvalidConfig
	ErrCodeDatabaseConnection
	ErrCodeDatabaseNotFound
	ErrCodeDatabaseExists
	ErrCodeInvalidDBName
	ErrCodeProcessFailed
	ErrCodeTimeout
	ErrCodePermission
)

// ApplicationError represents a custom application error
type ApplicationError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface
func (e *ApplicationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *ApplicationError) Unwrap() error {
	return e.Err
}

// NewError creates a new ApplicationError
func NewError(code ErrorCode, message string, err error) *ApplicationError {
	return &ApplicationError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error constructors
func NewInvalidConfigError(msg string, err error) *ApplicationError {
	return NewError(ErrCodeInvalidConfig, msg, err)
}

func NewDatabaseConnectionError(msg string, err error) *ApplicationError {
	return NewError(ErrCodeDatabaseConnection, msg, err)
}

func NewDatabaseNotFoundError(dbname string) *ApplicationError {
	return NewError(ErrCodeDatabaseNotFound, fmt.Sprintf("database '%s' not found", dbname), nil)
}

func NewDatabaseExistsError(dbname string) *ApplicationError {
	return NewError(ErrCodeDatabaseExists, fmt.Sprintf("database '%s' already exists", dbname), nil)
}

func NewInvalidDBNameError(dbname string) *ApplicationError {
	return NewError(ErrCodeInvalidDBName, fmt.Sprintf("invalid database name: '%s'", dbname), nil)
}

func NewProcessFailedError(msg string, err error) *ApplicationError {
	return NewError(ErrCodeProcessFailed, msg, err)
}

func NewTimeoutError(msg string) *ApplicationError {
	return NewError(ErrCodeTimeout, msg, nil)
}

func NewPermissionError(msg string, err error) *ApplicationError {
	return NewError(ErrCodePermission, msg, err)
}
