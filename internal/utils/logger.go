package utils

import (
	"log"
	"os"
)

type Logger struct {
	verbose    bool
	infoLogger *log.Logger
	errLogger  *log.Logger
}

func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose:    verbose,
		infoLogger: log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errLogger:  log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if l.verbose {
		l.infoLogger.Printf(format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.errLogger.Printf(format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.infoLogger.Printf("WARNING: "+format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		l.infoLogger.Printf("DEBUG: "+format, args...)
	}
}
