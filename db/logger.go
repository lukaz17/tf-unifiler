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

package db

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Queries running longer than this are logged as slow.
const slowQueryThreshold = time.Second

// zerologAdapter is a gorm logger writing to a zerolog logger, so database logs follow the caller's sinks.
type zerologAdapter struct {
	log   zerolog.Logger
	level logger.LogLevel
}

// Return a gorm logger that writes failed and slow queries to log.
func newZerologAdapter(log zerolog.Logger) logger.Interface {
	return &zerologAdapter{log: log, level: logger.Warn}
}

// Return a copy of this logger that logs at level.
func (l *zerologAdapter) LogMode(level logger.LogLevel) logger.Interface {
	c := *l
	c.level = level
	return &c
}

// Log an informational message from gorm.
func (l *zerologAdapter) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Info {
		l.log.Info().Msgf(msg, args...)
	}
}

// Log a warning message from gorm.
func (l *zerologAdapter) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Warn {
		l.log.Warn().Msgf(msg, args...)
	}
}

// Log an error message from gorm.
func (l *zerologAdapter) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Error {
		l.log.Error().Msgf(msg, args...)
	}
}

// Log a finished query if it failed or ran slower than slowQueryThreshold, or every query in info mode; record not found is not a failure.
func (l *zerologAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.level >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.log.Err(err).Str("sql", sql).Int64("rows", rows).Dur("elapsed", elapsed).Msg("Database query failed.")
	case elapsed > slowQueryThreshold && l.level >= logger.Warn:
		sql, rows := fc()
		l.log.Warn().Str("sql", sql).Int64("rows", rows).Dur("elapsed", elapsed).Msg("Database query is slow.")
	case l.level >= logger.Info:
		sql, rows := fc()
		l.log.Info().Str("sql", sql).Int64("rows", rows).Dur("elapsed", elapsed).Msg("Database query executed.")
	}
}
