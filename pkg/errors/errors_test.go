package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestNewError(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	err := NewError(ErrCodeInvalidConfig, "test error", wrappedErr)
	if err == nil {
		t.Error("NewError should not return nil")
	}

	if err.Code != ErrCodeInvalidConfig {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeInvalidConfig)
	}

	if err.Message != "test error" {
		t.Errorf("Error message = %q, want %q", err.Message, "test error")
	}

	if err.Err != wrappedErr {
		t.Errorf("Wrapped error = %v, want %v", err.Err, wrappedErr)
	}
}

func TestApplicationErrorError(t *testing.T) {
	err := NewError(ErrCodeInvalidConfig, "test error", nil)
	errorString := err.Error()

	if !strings.Contains(errorString, "[1]") {
		t.Errorf("Error string = %q, want to contain '[1]'", errorString)
	}

	if !strings.Contains(errorString, "test error") {
		t.Errorf("Error string = %q, want to contain 'test error'", errorString)
	}
}

func TestApplicationErrorErrorWithWrappedError(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	err := NewError(ErrCodeDatabaseConnection, "connection failed", wrappedErr)
	errorString := err.Error()

	if !strings.Contains(errorString, "connection failed") {
		t.Errorf("Error string = %q, want to contain 'connection failed'", errorString)
	}

	if !strings.Contains(errorString, "wrapped error") {
		t.Errorf("Error string = %q, want to contain 'wrapped error'", errorString)
	}
}

func TestApplicationErrorUnwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	err := &ApplicationError{
		Code:    ErrCodeDatabaseConnection,
		Message: "connection failed",
		Err:     wrappedErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestNewInvalidConfigError(t *testing.T) {
	wrappedErr := errors.New("config parse error")
	err := NewInvalidConfigError("config is invalid", wrappedErr)

	if err.Code != ErrCodeInvalidConfig {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeInvalidConfig)
	}

	if !strings.Contains(err.Message, "config is invalid") {
		t.Errorf("Error message = %q, want to contain 'config is invalid'", err.Message)
	}

	if err.Err != wrappedErr {
		t.Errorf("Wrapped error = %v, want %v", err.Err, wrappedErr)
	}
}

func TestNewDatabaseConnectionError(t *testing.T) {
	wrappedErr := errors.New("connection timeout")
	err := NewDatabaseConnectionError("failed to connect", wrappedErr)

	if err.Code != ErrCodeDatabaseConnection {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeDatabaseConnection)
	}

	if !strings.Contains(err.Message, "failed to connect") {
		t.Errorf("Error message = %q, want to contain 'failed to connect'", err.Message)
	}

	if err.Err != wrappedErr {
		t.Errorf("Wrapped error = %v, want %v", err.Err, wrappedErr)
	}
}

func TestNewDatabaseNotFoundError(t *testing.T) {
	err := NewDatabaseNotFoundError("testdb")

	if err.Code != ErrCodeDatabaseNotFound {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeDatabaseNotFound)
	}

	if !strings.Contains(err.Message, "testdb") {
		t.Errorf("Error message = %q, want to contain 'testdb'", err.Message)
	}

	if !strings.Contains(err.Message, "not found") {
		t.Errorf("Error message = %q, want to contain 'not found'", err.Message)
	}
}

func TestNewDatabaseExistsError(t *testing.T) {
	err := NewDatabaseExistsError("existing_db")

	if err.Code != ErrCodeDatabaseExists {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeDatabaseExists)
	}

	if !strings.Contains(err.Message, "existing_db") {
		t.Errorf("Error message = %q, want to contain 'existing_db'", err.Message)
	}

	if !strings.Contains(err.Message, "already exists") {
		t.Errorf("Error message = %q, want to contain 'already exists'", err.Message)
	}
}

func TestNewInvalidDBNameError(t *testing.T) {
	err := NewInvalidDBNameError("123-invalid")

	if err.Code != ErrCodeInvalidDBName {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeInvalidDBName)
	}

	if !strings.Contains(err.Message, "123-invalid") {
		t.Errorf("Error message = %q, want to contain '123-invalid'", err.Message)
	}

	if !strings.Contains(err.Message, "invalid database name") {
		t.Errorf("Error message = %q, want to contain 'invalid database name'", err.Message)
	}
}

func TestNewProcessFailedError(t *testing.T) {
	wrappedErr := errors.New("command failed")
	err := NewProcessFailedError("odoo-bin failed", wrappedErr)

	if err.Code != ErrCodeProcessFailed {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeProcessFailed)
	}

	if !strings.Contains(err.Message, "odoo-bin failed") {
		t.Errorf("Error message = %q, want to contain 'odoo-bin failed'", err.Message)
	}

	if err.Err != wrappedErr {
		t.Errorf("Wrapped error = %v, want %v", err.Err, wrappedErr)
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("operation timed out")

	if err.Code != ErrCodeTimeout {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodeTimeout)
	}

	if !strings.Contains(err.Message, "operation timed out") {
		t.Errorf("Error message = %q, want to contain 'operation timed out'", err.Message)
	}

	if err.Err != nil {
		t.Error("Wrapped error should be nil for timeout errors")
	}
}

func TestNewPermissionError(t *testing.T) {
	wrappedErr := errors.New("permission denied")
	err := NewPermissionError("access denied to /etc/config", wrappedErr)

	if err.Code != ErrCodePermission {
		t.Errorf("Error code = %d, want %d", err.Code, ErrCodePermission)
	}

	if !strings.Contains(err.Message, "access denied to /etc/config") {
		t.Errorf("Error message = %q, want to contain 'access denied to /etc/config'", err.Message)
	}

	if err.Err != wrappedErr {
		t.Errorf("Wrapped error = %v, want %v", err.Err, wrappedErr)
	}
}

func TestErrorCodeValues(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		want ErrorCode
	}{
		{"Unknown", ErrCodeUnknown, 0},
		{"InvalidConfig", ErrCodeInvalidConfig, 1},
		{"DatabaseConnection", ErrCodeDatabaseConnection, 2},
		{"DatabaseNotFound", ErrCodeDatabaseNotFound, 3},
		{"DatabaseExists", ErrCodeDatabaseExists, 4},
		{"InvalidDBName", ErrCodeInvalidDBName, 5},
		{"ProcessFailed", ErrCodeProcessFailed, 6},
		{"Timeout", ErrCodeTimeout, 7},
		{"Permission", ErrCodePermission, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("%s code = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}
