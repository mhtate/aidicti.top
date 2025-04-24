package logging

import (
	"fmt"
	"log/slog"
	"runtime"

	"aidicti.top/pkg/utils"
)

var loggerInstance *slog.Logger = nil

func Set(l *slog.Logger) {
	utils.Assert(l != nil, "*slog.Logger is nil")

	loggerInstance = l
}

func Logger() *slog.Logger {
	utils.Assert(loggerInstance != nil, "logger not set")
	return loggerInstance
}

func Warn(msg string, args ...any) {
	utils.Assert(loggerInstance != nil, "loggerInstance is nil")

	_, file, line, _ := runtime.Caller(1)

	args = append(args, "stack", slog.GroupValue(slog.String("file", fmt.Sprintf("%s:%d", file, line))))
	loggerInstance.Warn(msg, args...)
}

func Info(msg string, args ...any) {
	utils.Assert(loggerInstance != nil, "loggerInstance is nil")

	_, file, line, _ := runtime.Caller(1)

	args = append(args, "stack", slog.GroupValue(slog.String("file", fmt.Sprintf("%s:%d", file, line))))
	loggerInstance.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	utils.Assert(loggerInstance != nil, "loggerInstance is nil")

	_, file, line, _ := runtime.Caller(1)

	args = append(args, "stack", slog.GroupValue(slog.String("file", fmt.Sprintf("%s:%d", file, line))))
	loggerInstance.Debug(msg, args...)
}

type NullWriter struct{}

func (w NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
