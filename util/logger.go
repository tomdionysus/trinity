package util

import(
  "fmt"
  "strings"
  "time"
)

type Logger struct {
	LogLevel uint
}

func NewLogger(logLevel string) *Logger {
  logger := &Logger{
    LogLevel: parseLogLevel(strings.ToLower(strings.Trim(logLevel," "))),
  }
  if logger.LogLevel == 0 {
    logger.Warn("Logger", "Cannot parse log level '%s', assuming debug", logLevel)
  }
  return logger
}

func (me *Logger) Fatal(component string, message string, args... interface{}) {
  printLog(component,"FATAL",message, args...)
}

func (me *Logger) Error(component string, message string, args... interface{}) {
  printLog(component,"ERROR",message, args...)
}

func (me *Logger) Warn(component string, message string, args... interface{}) {
  if me.LogLevel > 3 { return }
  printLog(component,"WARN ",message, args...)
}

func (me *Logger) Info(component string, message string, args... interface{}) {
  if me.LogLevel > 2 { return }
  printLog(component,"INFO ",message, args...)
}

func (me *Logger) Debug(component string, message string, args... interface{}) {
  if me.LogLevel > 1 { return }
  printLog(component,"DEBUG",message, args...)
}

func printLog(component string, level string, message string, args... interface{}) {
  fmt.Printf("%s [%s] %s: ",getTimeUTCString(), level, component)
  fmt.Printf(message, args...)
  fmt.Print("\n")
}

func getTimeUTCString() string {
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