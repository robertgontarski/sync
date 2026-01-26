package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Level string

const (
	INFO  Level = "INFO"
	ERROR Level = "ERROR"
)

type Logger struct {
	out io.Writer
}

func New() *Logger {
	return &Logger{
		out: os.Stdout,
	}
}

func NewWithWriter(w io.Writer) *Logger {
	return &Logger{
		out: w,
	}
}

func (l *Logger) log(level Level, format string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.out, "[%s] %s: %s\n", level, timestamp, message)
}

func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}
