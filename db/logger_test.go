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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestConnectLogsQueryErrorsToGivenLogger(t *testing.T) {
	var buf bytes.Buffer
	ctx, err := Connect(filepath.Join(t.TempDir(), "unifiler.db"), zerolog.New(&buf))
	if err != nil {
		t.Fatalf("cannot open database: %v", err)
	}
	t.Cleanup(func() { ctx.Disconnect() })

	ctx.db.Exec("SELECT * FROM missing_table")

	if !strings.Contains(buf.String(), "missing_table") {
		t.Errorf("expected the failed query in the log, got %q", buf.String())
	}
}

func TestZerologAdapterTrace(t *testing.T) {
	boom := errors.New("boom")
	fast := time.Now()
	slow := time.Now().Add(-time.Minute)
	tests := []struct {
		name      string
		level     logger.LogLevel
		begin     time.Time
		err       error
		wantLevel string // empty when nothing must be logged.
	}{
		{"failed query is logged as error", logger.Warn, fast, boom, "error"},
		{"record not found is not logged", logger.Warn, fast, gorm.ErrRecordNotFound, ""},
		{"slow query is logged as warning", logger.Warn, slow, nil, "warn"},
		{"fast query is not logged", logger.Warn, fast, nil, ""},
		{"silent mode drops failed query", logger.Silent, fast, boom, ""},
		{"error mode drops slow query", logger.Error, slow, nil, ""},
		{"info mode logs fast query", logger.Info, fast, nil, "info"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := newZerologAdapter(zerolog.New(&buf)).LogMode(tt.level)

			l.Trace(context.Background(), tt.begin, func() (string, int64) { return "SELECT 1", 1 }, tt.err)

			if tt.wantLevel == "" {
				if buf.Len() != 0 {
					t.Errorf("expected no log, got %q", buf.String())
				}
				return
			}
			var got map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("decoding log line %q: %v", buf.String(), err)
			}
			if got["level"] != tt.wantLevel {
				t.Errorf("expected level %q, got %v", tt.wantLevel, got["level"])
			}
			if got["sql"] != "SELECT 1" {
				t.Errorf("expected sql %q, got %v", "SELECT 1", got["sql"])
			}
		})
	}
}
