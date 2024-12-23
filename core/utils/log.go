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

// Debugf logs a debug message with formatted content.
func (l *ServiceLogger) Debugf(format string, args ...interface{}) {
	log.Debugf(l.getFormat(format), args...)
}

// Fatalf logs a fatal message with formatted content and exits the program.
func (l *ServiceLogger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(l.getFormat(format), args...)
}

func (l *ServiceLogger) getFormat(format string) string {
	return fmt.Sprintf("[%s] %s", l.prefix, format)
}
