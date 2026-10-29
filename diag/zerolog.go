// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package diag

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/tui"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName      = "unifiler.log"
	logFileMaxSize   = 10 // megabytes per file before rotation.
	logFileMaxBackup = 3  // maximum rotated files to retain.
	logFileMaxAge    = 28 // maximum days to retain old log files.
)

// syncLevelWriter serializes writes to the wrapped level writer so a shared base logger stays data-race free when used from multiple goroutines.
type syncLevelWriter struct {
	mu sync.Mutex
	w  zerolog.LevelWriter
}

func (s *syncLevelWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

func (s *syncLevelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.WriteLevel(level, p)
}

// Return a new thread-safe log writer instance (console plus rotating file when useFS is true).
func NewLogWriter(cfg *config.RootConfig, useFS bool) (zerolog.LevelWriter, *lumberjack.Logger, error) {
	zerolog.SetGlobalLevel(resolveLevel(cfg))

	consoleWriter := ConsoleLevelWriter(os.Stdout, tui.IsTTY())

	logFile, err := initLogFile(useFS, cfg.ConfigDir)
	if logFile == nil {
		return &syncLevelWriter{w: consoleWriter}, nil, err
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, FileLevelWriter(logFile))
	return &syncLevelWriter{w: multiWriter}, logFile, nil
}

// Build a logger from a pre-composed writer.
func NewLoggerFromWriter(w zerolog.LevelWriter) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Logger()
}

// Wrap a human-readable console writer around out as a zerolog level writer.
func ConsoleLevelWriter(out io.Writer, color bool) zerolog.LevelWriter {
	return &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{Out: out, NoColor: !color, TimeFormat: time.DateTime},
		},
		Level: zerolog.TraceLevel,
	}
}

// Wrap a raw writer as a zerolog level writer.
func FileLevelWriter(w io.Writer) zerolog.LevelWriter {
	return &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: w,
		},
		Level: zerolog.TraceLevel,
	}
}

// Create and return rotating log file writer only if useFS is true.
func initLogFile(useFS bool, workingDir string) (*lumberjack.Logger, error) {
	if !useFS {
		return nil, nil
	}
	logDir := workingDir
	if logDir == "" {
		logDir = "."
	}
	logDir = filepath.Join(logDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	return &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFileName),
		MaxSize:    logFileMaxSize,
		MaxBackups: logFileMaxBackup,
		MaxAge:     logFileMaxAge,
	}, nil
}

// Parse configured log level, fallback to info when invalid or missing.
func resolveLevel(cfg *config.RootConfig) zerolog.Level {
	if cfg == nil || cfg.Logger == nil {
		return zerolog.InfoLevel
	}
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return level
}
