package interfaces

// Logger interface for reporting informational and warning messages.
type Logger interface {
	// Debugf logs a formatted debugging message.
	Debugf(format string, args ...interface{})

	// Infof logs a formatted informational message.
	Infof(format string, args ...interface{})

	// Warnf logs a formatted warning message.
	Warnf(format string, args ...interface{})

	// Errorf logs a formatted error message.
	Errorf(format string, args ...interface{})

	// Fatalf logs a formatted fatal message.
	Fatalf(format string, args ...interface{})
}
