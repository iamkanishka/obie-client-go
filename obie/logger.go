package obie

import (
	"fmt"
	"log"
)

// Logger is the pluggable logging interface consumed by the SDK.
// Callers may supply their own implementation (zerolog, zap, logrus, etc.).
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// nopLogger discards all log output.
type nopLogger struct{}

func (nopLogger) Debugf(_ string, _ ...any) {}
func (nopLogger) Infof(_ string, _ ...any)  {}
func (nopLogger) Warnf(_ string, _ ...any)  {}
func (nopLogger) Errorf(_ string, _ ...any) {}

// StdLogger wraps the standard library logger.
type StdLogger struct{ l *log.Logger }

// NewStdLogger returns a Logger backed by the standard library.
func NewStdLogger(l *log.Logger) Logger { return &StdLogger{l: l} }

func (s *StdLogger) Debugf(f string, a ...any) { s.l.Printf("[DEBUG] "+f, a...) }
func (s *StdLogger) Infof(f string, a ...any)  { s.l.Printf("[INFO]  "+f, a...) }
func (s *StdLogger) Warnf(f string, a ...any)  { s.l.Printf("[WARN]  "+f, a...) }
func (s *StdLogger) Errorf(f string, a ...any) { s.l.Printf("[ERROR] "+f, a...) }

// SlogLogger wraps the standard library log/slog logger (Go 1.25+).
// Usage:
//
//	import "log/slog"
//	obie.NewSlogLogger(slog.Default())
type SlogLogger struct {
	l interface {
		Debug(msg string, args ...any)
		Info(msg string, args ...any)
		Warn(msg string, args ...any)
		Error(msg string, args ...any)
	}
}

// NewSlogLogger returns a Logger backed by a *slog.Logger.
// Pass slog.Default() to use the default structured logger.
func NewSlogLogger(l interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}) Logger {
	return &SlogLogger{l: l}
}

func (s *SlogLogger) Debugf(f string, a ...any) { s.l.Debug(fmt.Sprintf(f, a...)) }
func (s *SlogLogger) Infof(f string, a ...any)  { s.l.Info(fmt.Sprintf(f, a...)) }
func (s *SlogLogger) Warnf(f string, a ...any)  { s.l.Warn(fmt.Sprintf(f, a...)) }
func (s *SlogLogger) Errorf(f string, a ...any) { s.l.Error(fmt.Sprintf(f, a...)) }
