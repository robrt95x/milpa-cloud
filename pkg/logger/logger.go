package logger

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type hclogAdapter struct {
	log hclog.Logger
}

func (h *hclogAdapter) Debug(msg string, args ...interface{}) {
	h.log.Debug(msg, args...)
}

func (h *hclogAdapter) Info(msg string, args ...interface{}) {
	h.log.Info(msg, args...)
}

func (h *hclogAdapter) Warn(msg string, args ...interface{}) {
	h.log.Warn(msg, args...)
}

func (h *hclogAdapter) Error(msg string, args ...interface{}) {
	h.log.Error(msg, args...)
}

func New(level string) Logger {
	lvl := hclog.LevelFromString(level)
	if lvl == hclog.NoLevel {
		lvl = hclog.Info
	}

	return &hclogAdapter{
		log: hclog.New(&hclog.LoggerOptions{
			Name:   "milpa",
			Level:  lvl,
			Output: os.Stdout,
		}),
	}
}

func NewHCLog(name string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:  name,
		Level: hclog.Info,
	})
}

// Convenience function for quick logging
func Debug(msg string, args ...interface{}) {
	hclog.Default().Debug(fmt.Sprintf(msg, args...))
}

func Info(msg string, args ...interface{}) {
	hclog.Default().Info(fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	hclog.Default().Warn(fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...interface{}) {
	hclog.Default().Error(fmt.Sprintf(msg, args...))
}
