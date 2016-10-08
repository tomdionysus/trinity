package util

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	LogLevel uint

	mutex *sync.Mutex
}

func NewLogger(logLevel string) *Logger {
	logger := &Logger{
		LogLevel: parseLogLevel(strings.ToLower(strings.Trim(logLevel, " "))),
		mutex:    &sync.Mutex{},
	}
	if logger.LogLevel == 0 {
		logger.Warn("Logger", "Cannot parse log level '%s', assuming debug", logLevel)
	}
	return logger
}

func (log *Logger) Raw(component string, message string, args ...interface{}) {
	log.printLog(component, "-----", message, args...)
}

func (log *Logger) Fatal(component string, message string, args ...interface{}) {
	log.printLog(component, "FATAL", message, args...)
}

func (log *Logger) Error(component string, message string, args ...interface{}) {
	log.printLog(component, "ERROR", message, args...)
}

func (log *Logger) Warn(component string, message string, args ...interface{}) {
	if log.LogLevel > 3 {
		return
	}
	log.printLog(component, "WARN ", message, args...)
}

func (log *Logger) Info(component string, message string, args ...interface{}) {
	if log.LogLevel > 2 {
		return
	}
	log.printLog(component, "INFO ", message, args...)
}

func (log *Logger) Debug(component string, message string, args ...interface{}) {
	if log.LogLevel > 1 {
		return
	}
	log.printLog(component, "DEBUG", message, args...)
}

func (log *Logger) printLog(component string, level string, message string, args ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	fmt.Printf("%s [%s] %s: ", log.getTimeUTCString(), level, component)
	fmt.Printf(message, args...)
	fmt.Print("\n")
}

func (log *Logger) getTimeUTCString() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseLogLevel(logLevel string) uint {
	switch logLevel {
	case "fatal":
		return 5
	case "error":
		return 4
	case "warn":
		return 3
	case "info":
		return 2
	case "debug":
		return 1
	}
	return 0
}
