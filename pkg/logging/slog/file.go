package slog

import (
	"fmt"
	"os"

	"log/slog"

	"aidicti.top/pkg/utils"
)

func NewLocalFileHandler(filename string, opts *slog.HandlerOptions) (*slog.TextHandler, func() error) {
	path := fmt.Sprintf(filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	utils.Assert(err == nil, fmt.Sprintf("open log file fail, \"name\":\"%s\"", filename))

	return slog.NewTextHandler(file, opts), func() error { return file.Close() }
}
