package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger() should not return nil")
	}

	if logger.log == nil {
		t.Error("NewLogger() should create a logrus logger")
	}
}

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	logger.Info("test info message")

	output := buf.String()
	if !strings.Contains(output, "test info message") {
		t.Errorf("Info() output = %q, want to contain 'test info message'", output)
	}
	if !strings.Contains(output, "level=info") {
		t.Errorf("Info() output = %q, want to contain 'level=info'", output)
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	testErr := errors.New("test error")
	logger.Error("test error message", testErr)

	output := buf.String()
	if !strings.Contains(output, "test error message") {
		t.Errorf("Error() output = %q, want to contain 'test error message'", output)
	}
	if !strings.Contains(output, "level=error") {
		t.Errorf("Error() output = %q, want to contain 'level=error'", output)
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("Error() output = %q, want to contain 'test error'", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	logger.Warn("test warn message")

	output := buf.String()
	if !strings.Contains(output, "test warn message") {
		t.Errorf("Warn() output = %q, want to contain 'test warn message'", output)
	}
	if !strings.Contains(output, "level=warning") {
		t.Errorf("Warn() output = %q, want to contain 'level=warning'", output)
	}
}

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetLevel(logrus.DebugLevel)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	logger.Debug("test debug message")

	output := buf.String()
	if !strings.Contains(output, "test debug message") {
		t.Errorf("Debug() output = %q, want to contain 'test debug message'", output)
	}
	if !strings.Contains(output, "level=debug") {
		t.Errorf("Debug() output = %q, want to contain 'level=debug'", output)
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	logger.WithField("key", "value").Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("WithField().Info() output = %q, want to contain 'test message'", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("WithField().Info() output = %q, want to contain 'key=value'", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.log.SetOutput(&buf)
	logger.log.SetFormatter(&logrus.TextFormatter{DisableColors: true, DisableTimestamp: true})

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	logger.WithFields(fields).Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("WithFields().Info() output = %q, want to contain 'test message'", output)
	}
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("WithFields().Info() output = %q, want to contain 'key1=value1'", output)
	}
	if !strings.Contains(output, "key2=42") {
		t.Errorf("WithFields().Info() output = %q, want to contain 'key2=42'", output)
	}
}

func TestSetLevel(t *testing.T) {
	logger := NewLogger()

	// Default should be info
	if logger.log.GetLevel() != logrus.InfoLevel {
		t.Errorf("Default level = %v, want %v", logger.log.GetLevel(), logrus.InfoLevel)
	}

	// Set to debug
	err := logger.SetLevel("debug")
	if err != nil {
		t.Errorf("SetLevel(debug) error = %v, want nil", err)
	}
	if logger.log.GetLevel() != logrus.DebugLevel {
		t.Errorf("After SetLevel(debug), level = %v, want %v", logger.log.GetLevel(), logrus.DebugLevel)
	}

	// Set to error
	err = logger.SetLevel("error")
	if err != nil {
		t.Errorf("SetLevel(error) error = %v, want nil", err)
	}
	if logger.log.GetLevel() != logrus.ErrorLevel {
		t.Errorf("After SetLevel(error), level = %v, want %v", logger.log.GetLevel(), logrus.ErrorLevel)
	}

	// Test invalid level
	err = logger.SetLevel("invalid")
	if err == nil {
		t.Error("SetLevel(invalid) should return error")
	}
}

func TestGetLogger(t *testing.T) {
	logger := NewLogger()
	underlyingLogger := logger.GetLogger()

	if underlyingLogger == nil {
		t.Error("GetLogger() should not return nil")
	}

	if underlyingLogger != logger.log {
		t.Error("GetLogger() should return the underlying logrus logger")
	}
}
