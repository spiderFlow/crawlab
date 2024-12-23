package utils

import (
	"fmt"
	"github.com/apex/log"
)

// ServiceLogger represents a logger with a specific service prefix.
type ServiceLogger struct {
	prefix string
}

// NewServiceLogger creates a new logger with the given service name as a prefix.
func NewServiceLogger(serviceName string) *ServiceLogger {
	return &ServiceLogger{
		prefix: serviceName,
	}
}

// Debug logs a debug message.
func (l *ServiceLogger) Debug(message string) {
	log.Debug(l.getFormat(message))
}

// Info logs an informational message.
func (l *ServiceLogger) Info(message string) {
	log.Info(l.getFormat(message))
}

// Warn logs a warning message.
func (l *ServiceLogger) Warn(message string) {
	log.Warn(l.getFormat(message))
}

// Error logs an error message.
func (l *ServiceLogger) Error(message string) {
	log.Error(l.getFormat(message))
}

// Fatal logs a fatal message.
func (l *ServiceLogger) Fatal(message string) {
	log.Fatal(l.getFormat(message))
}

// Debugf logs a debug message with formatted content.
func (l *ServiceLogger) Debugf(format string, args ...interface{}) {
	log.Debugf(l.getFormat(format), args...)
}

// Infof logs an informational message with formatted content.
func (l *ServiceLogger) Infof(format string, args ...interface{}) {
	log.Infof(l.getFormat(format), args...)
}

// Warnf logs a warning message with formatted content.
func (l *ServiceLogger) Warnf(format string, args ...interface{}) {
	log.Warnf(l.getFormat(format), args...)
}

// Errorf logs an error message with formatted content.
func (l *ServiceLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(l.getFormat(format), args...)
}

// Fatalf logs an error message with formatted content.
func (l *ServiceLogger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(l.getFormat(format), args...)
}

func (l *ServiceLogger) getFormat(format string) string {
	return fmt.Sprintf("[%s] %s", l.prefix, format)
}
