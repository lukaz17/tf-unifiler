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

package tui

import (
	"fmt"
	"slices"
	"testing"
)

func TestSetTasksReportsFinishedOverTotal(t *testing.T) {
	tests := []struct {
		name        string
		finished    uint64
		total       uint64
		wantPercent float64
		wantLabel   string
	}{
		{"idle server with no task yet", 0, 0, 0, "0/0 tasks"},
		{"some tasks finished", 3, 5, 0.6, "3/5 tasks"},
		{"every task finished", 5, 5, 1, "5/5 tasks"},
		{"more finished than total is capped", 6, 5, 1, "6/5 tasks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewBubbleteaNotifier()

			n.SetTasks(tt.finished, tt.total)

			msg := n.i.statusMsg()
			if msg.totalPercent != tt.wantPercent {
				t.Errorf("expected total percent %v, got %v", tt.wantPercent, msg.totalPercent)
			}
			if msg.totalLabel != tt.wantLabel {
				t.Errorf("expected total label %q, got %q", tt.wantLabel, msg.totalLabel)
			}
			if msg.totalInfinite {
				t.Error("expected a task bar, got the spinner of an unknown total")
			}
		})
	}
}

func TestTaskCountIgnoresItemsOfRunningTask(t *testing.T) {
	n := NewBubbleteaNotifier()
	n.SetTasks(1, 4)

	n.SetTotal(2)
	n.OnProgress("Hash File-1", 50, 100)
	n.OnFinish("Hash File-1", 0)
	n.OnProgress("Hash File-2", 25, 100)

	msg := n.i.statusMsg()
	if msg.totalPercent != 0.25 {
		t.Errorf("expected total percent 0.25 from 1/4 tasks, got %v", msg.totalPercent)
	}
	if msg.itemPercent != 0.25 {
		t.Errorf("expected item percent 0.25 from the running item, got %v", msg.itemPercent)
	}
}

func TestItemCountDrivesTotalWithoutTasks(t *testing.T) {
	n := NewBubbleteaNotifier()
	n.SetTotal(2)

	n.OnFinish("Hash File-1", 0)

	msg := n.i.statusMsg()
	if msg.totalPercent != 0.5 {
		t.Errorf("expected total percent 0.5 from 1/2 items, got %v", msg.totalPercent)
	}
	if msg.totalLabel != "" {
		t.Errorf("expected no total label, got %q", msg.totalLabel)
	}
}

func TestErrorsKeepOnlyLinesTheViewShows(t *testing.T) {
	n := NewBubbleteaNotifier()

	for i := 1; i <= maxErrorLines+2; i++ {
		n.OnWarn("Hash File-1", fmt.Sprintf("warning %d", i))
	}

	want := []string{
		"[WARN] Hash File: warning 3",
		"[WARN] Hash File: warning 4",
		"[WARN] Hash File: warning 5",
	}
	if got := n.i.statusMsg().errors; !slices.Equal(got, want) {
		t.Errorf("expected last %d warnings %v, got %v", maxErrorLines, want, got)
	}
}
