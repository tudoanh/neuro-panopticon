package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"neuropanopticon/backend/config"
)

// Setup initializes the global slog logger. It writes structured JSON logs to a
// file in the config directory and text logs to stderr for development.
// The returned function closes the log file and should be called on shutdown.
func Setup() (logger *slog.Logger, close func()) {
	logPath := filepath.Join(config.ConfigDir(), "neuropanopticon.log")

	var writers []io.Writer
	writers = append(writers, os.Stderr)

	close = func() {} // no-op default

	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		writers = append(writers, f)
		close = func() { f.Close() }
	}

	w := io.MultiWriter(writers...)
	logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)
	return logger, close
}
