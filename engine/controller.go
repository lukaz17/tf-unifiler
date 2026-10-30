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

package engine

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/crypto/hasher"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/filesys/exec"
	"github.com/tforceaio/tf-unifiler/tui"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Controller is the entrypoint for working with application configurations and loggings.
type Controller struct {
	Config   *config.RootConfig
	Logger   zerolog.Logger
	Notifier diag.Notifier

	Interactive bool

	logFile *lumberjack.Logger
}

// Entrypoint for creating new instance of Controller.
// useFS will instruct this function to read configurations and create log file.
func NewController(useFS bool) *Controller {
	cfg, err := config.InitKoanf(useFS)
	writer, logFile, err2 := diag.NewLogWriter(cfg, useFS)
	logger := diag.NewLoggerFromWriter(writer)
	if err != nil {
		logger.Err(err).Msg("error initializing config")
	}
	if err2 != nil {
		logger.Err(err2).Msg("error initializing log file")
	}
	interactive := tui.IsTTY()
	var notifier diag.Notifier
	if interactive {
		notifier = tui.NewBubbleteaNotifier()
		hasher.SetNotify(notifier)
		filesys.SetNotify(notifier)
		exec.SetNotify(notifier)
	}
	return &Controller{
		Config:   cfg,
		Logger:   logger,
		Notifier: notifier,

		Interactive: interactive,

		logFile: logFile,
	}
}

// Execute additional clean up when terminate the app.
func (c *Controller) Close() {
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}
}

// Get a ZeroLog logger instance for command handler from root instance.
func (c *Controller) CommandLogger(module, command string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Str("command", command).Logger()
}

// Get a ZeroLog logger instance for module from root instance.
func (c *Controller) ModuleLogger(module string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Logger()
}

// Derive a task-scoped Controller for running a single task, it shares this Controller's rotating log file but tees log output into the per-task sink out for capture.
func (c *Controller) fork(out io.Writer) *Controller {
	sink := diag.ConsoleLevelWriter(out, false)
	var writer zerolog.LevelWriter = sink
	if c.logFile != nil {
		writer = zerolog.MultiLevelWriter(diag.FileLevelWriter(c.logFile), sink)
	}
	return &Controller{
		Config:   c.Config,
		Logger:   diag.NewLoggerFromWriter(writer),
		Notifier: c.Notifier,

		Interactive: false,

		logFile: nil,
	}
}
