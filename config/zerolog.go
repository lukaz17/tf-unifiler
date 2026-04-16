// Copyright (C) 2024 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/filesys"
)

// Entrypoint for creating a ZeroLog logger instance.
func InitZerolog(configDir string, useFS bool) (zerolog.Logger, *os.File, error) {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.DateTime},
		},
		Level: zerolog.TraceLevel,
	}

	logFile, err := InitLogFile(useFS, configDir)
	if logFile == nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		return consoleLogger, nil, err
	}

	fileWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: logFile,
		},
		Level: zerolog.TraceLevel,
	}
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return logger, logFile, nil
}

// Create and return log file handle only if useFS is true.
func InitLogFile(useFS bool, workdingDir string) (*os.File, error) {
	if !useFS {
		return nil, nil
	}
	logDir := path.Join(opx.Ternary(workdingDir == "", ".", workdingDir), "logs")
	if !filesys.IsExist(logDir) {
		err := filesys.CreateDirectoryRecursive(logDir)
		if err != nil {
			return nil, err
		}
	}
	logFileName := fmt.Sprintf("unifiler-%s.log", time.Now().UTC().Format("20060102"))
	logFilePath := filepath.Join(logDir, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}
