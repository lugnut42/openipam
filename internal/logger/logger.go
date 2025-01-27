package logger

import "log"

var debugMode bool

// SetDebugMode enables or disables debug logging
func SetDebugMode(enabled bool) {
	debugMode = enabled
}

// Debug logs a debug message if debug mode is enabled
func Debug(format string, v ...interface{}) {
	if debugMode {
		log.Printf("DEBUG: "+format, v...)
	}
}
