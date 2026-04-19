package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus to provide a clean logging interface
type Logger struct {
	log *logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return &Logger{log: log}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log.Info(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error) {
	l.log.WithError(err).Error(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log.Warn(msg)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log.Debug(msg)
}

// WithField adds a field to the log entry
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.log.WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.log.WithFields(logrus.Fields(fields))
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	l.log.SetLevel(lvl)
	return nil
}

// GetLogger returns the underlying logrus logger
func (l *Logger) GetLogger() *logrus.Logger {
	return l.log
}
