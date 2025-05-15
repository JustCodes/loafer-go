package loafergo

import (
	"log"
	"os"
)

// A Logger is a minimalistic interface for the loafer to log messages to. Should
// be used to provide custom logging writers for the loafer to use.
type Logger interface {
	Log(args ...any)
}

// A LoggerFunc is a convenience type to convert a function taking a variadic
// list of arguments and wrap it so the Logger interface can be used.
//
// Example:
//
//	loafergo.NewManager(context.Background(), loafergo.Config{Logger: loafergo.LoggerFunc(func(args ...interface{}) {
//	    fmt.Fprintln(os.Stdout, args...)
//	})})
type LoggerFunc func(...interface{})

// Log calls the wrapped function with the arguments provided
func (f LoggerFunc) Log(args ...interface{}) {
	f(args...)
}

// newDefaultLogger returns a Logger which will write log messages to stdout
//
//	and use the same formatting runes as the stdlib log.Logger
func newDefaultLogger() Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// A defaultLogger provides a minimalistic logger satisfying the Logger interface.
type defaultLogger struct {
	logger *log.Logger
}

// Log logs the parameters to the stdlib logger. See log.Println.
func (l defaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

// NoOpLogger is a logger that does nothing.
type NoOpLogger struct{}

// Log implements Logger but does nothing.
func (NoOpLogger) Log(args ...any) {}
