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
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

// BubbleteaNotifier implements diag.Notifier interface for routing message to Bubbletea program.
type BubbleteaNotifier struct {
	i *BubbleteaNotifierInternal
}

type BubbleteaNotifierInternal struct {
	mu sync.Mutex

	action       string
	item         string
	itemPercent  float64
	totalPercent float64
	errors       []string

	itemFinished uint64
	itemTotal    uint64

	serverMode   bool
	taskFinished uint64
	taskTotal    uint64

	teaProgram *tea.Program
}

// Return a new BubbleteaNotifier instance.
func NewBubbleteaNotifier() *BubbleteaNotifier {
	return &BubbleteaNotifier{
		i: &BubbleteaNotifierInternal{},
	}
}

// Set total number of item for overall progress processing.
func (n *BubbleteaNotifier) SetTotal(total uint64) {
	i := n.i
	i.mu.Lock()
	i.itemTotal = total
	i.mu.Unlock()
}

// Set the overall progress to finished tasks out of total tasks; from then on, items set by SetTotal no longer move it.
func (n *BubbleteaNotifier) SetTasks(finished, total uint64) {
	i := n.i
	i.mu.Lock()
	i.serverMode = true
	i.taskFinished = finished
	i.taskTotal = total
	i.totalPercent = 0
	if total > 0 {
		i.totalPercent = min(float64(finished)/float64(total), 1)
	}
	n.dispatch()
	i.mu.Unlock()
}

// Handle Start event.
func (n BubbleteaNotifier) OnStart(pid string, _ time.Time) {
	i := n.i
	i.mu.Lock()
	i.action = phaseFromPID(pid)
	i.item = ""
	i.itemPercent = 0
	i.mu.Unlock()
	i.mu.Lock()
	n.dispatch()
	i.mu.Unlock()
}

// Handle Error event.
func (n BubbleteaNotifier) OnError(pid string, err error, msg string) {
	i := n.i
	i.mu.Lock()
	var line string
	if err != nil {
		line = fmt.Sprintf("[ERR] %s: %s: %v", phaseFromPID(pid), msg, err)
	} else {
		line = fmt.Sprintf("[ERR] %s: %s", phaseFromPID(pid), msg)
	}
	i.appendError(line)
	n.dispatch()
	i.mu.Unlock()
}

// Handle Warn event.
func (n BubbleteaNotifier) OnWarn(pid, msg string) {
	i := n.i
	i.mu.Lock()
	i.appendError(fmt.Sprintf("[WARN] %s: %s", phaseFromPID(pid), msg))
	n.dispatch()
	i.mu.Unlock()
}

// Handle Info event.
func (n BubbleteaNotifier) OnInfo(pid, msg string) {
	i := n.i
	i.mu.Lock()
	if msg == "Started." || msg == "Finished." {
		i.action = phaseFromPID(pid)
	} else {
		i.item = msg
	}
	n.dispatch()
	i.mu.Unlock()
}

// Handle Debug event.
func (n BubbleteaNotifier) OnDebug(_, _ string) {}

// Handle Progress event.
func (n BubbleteaNotifier) OnProgress(pid string, cur, total uint64) {
	i := n.i
	i.mu.Lock()
	if total > 0 {
		i.itemPercent = float64(cur) / float64(total)
	} else {
		i.itemPercent = -1
	}
	if i.itemTotal > 0 && !i.serverMode {
		totalPercent := (float64(i.itemFinished) + i.itemPercent) / float64(i.itemTotal)
		if totalPercent > 1 {
			totalPercent = 1
		}
		i.totalPercent = totalPercent
	}
	n.dispatch()
	i.mu.Unlock()
}

// Handle Finish event.
func (n BubbleteaNotifier) OnFinish(_ string, _ time.Duration) {
	i := n.i
	i.mu.Lock()
	i.action = ""
	i.item = ""
	i.itemPercent = 1.0
	i.itemFinished++
	if !i.serverMode {
		if i.itemTotal > 0 {
			totalPercent := float64(i.itemFinished) / float64(i.itemTotal)
			if totalPercent > 1 {
				totalPercent = 1
			}
			i.totalPercent = totalPercent
		} else {
			i.totalPercent = 1
		}
	}
	n.dispatch()
	i.mu.Unlock()
}

func phaseFromPID(pid string) string {
	if idx := strings.LastIndex(pid, "-"); idx > 0 {
		return pid[:idx]
	}
	return pid
}

// Dispatch the message to paired Bubbletea program.
func (bn *BubbleteaNotifier) dispatch() {
	i := bn.i
	if i.teaProgram == nil {
		return
	}
	i.teaProgram.Send(i.statusMsg())
}

// Append an error line, keeping only the last lines the view can show so a long-running server doesn't accumulate them forever.
func (i *BubbleteaNotifierInternal) appendError(line string) {
	i.errors = append(i.errors, line)
	if len(i.errors) > maxErrorLines {
		i.errors = i.errors[len(i.errors)-maxErrorLines:]
	}
}

// Return whether this notifier is paired with a Bubbletea program.
func (n *BubbleteaNotifier) hasProgram() bool {
	i := n.i
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.teaProgram != nil
}

// Pair this notifier with a Bubbletea program.
func (n *BubbleteaNotifier) setProgram(p *tea.Program) {
	i := n.i
	i.mu.Lock()
	i.teaProgram = p
	i.mu.Unlock()
}

// Return the ProcessStatusMsg describing the current state.
func (i *BubbleteaNotifierInternal) statusMsg() ProcessStatusMsg {
	errsCopy := make([]string, len(i.errors))
	copy(errsCopy, i.errors)
	msg := ProcessStatusMsg{
		action:        i.action,
		item:          i.item,
		itemPercent:   i.itemPercent,
		totalInfinite: i.itemTotal <= 0 && !i.serverMode,
		totalPercent:  i.totalPercent,
		errors:        errsCopy,
	}
	if i.serverMode {
		msg.totalLabel = fmt.Sprintf("%d/%d tasks", i.taskFinished, i.taskTotal)
	}
	return msg
}
