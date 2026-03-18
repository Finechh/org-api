package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	error *log.Logger
}

func New() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO  ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR ", log.LstdFlags),
	}
}

func (l *Logger) Infof(format string, args ...any) {
	l.info.Output(2, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.error.Output(2, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.error.Fatalf(format, args...)
}
