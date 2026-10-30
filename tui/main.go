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
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// TeaProgramHandle provides caller a way to stop the Bubbletea program.
// A handle without a program shows nothing, and stopping it is a no-op.
type TeaProgramHandle struct {
	program  *tea.Program
	notifier *BubbleteaNotifier
	exited   chan struct{}
	done     chan error
}

// Stop running Bubbletea program.
func (p *TeaProgramHandle) Stop() {
	if p.program == nil {
		return
	}
	p.notifier.setProgram(nil)
	p.program.Quit()
	<-p.done
}

// Return a writer that prints each write as its own line above the program while it runs, and passes writes to fallback once it has exited.
func (p *TeaProgramHandle) Writer(fallback io.Writer) io.Writer {
	if p.program == nil {
		return fallback
	}
	return &printAboveWriter{
		program:  p.program,
		exited:   p.exited,
		fallback: fallback,
	}
}

// Run CompressOptions and show it on the terminal.
func RunCompressOptions(defaults *CompressOptionsValue) (*CompressOptionsValue, error) {
	m := NewCompressOptions().
		WithSelected(defaults)

	return m.Run()
}

// Run ProcessStatus and show it on the terminal.
// When notifier already feeds a running program, return a handle without a program, so progress lands in that program instead of a second one fighting over the terminal.
func RunProcessStatus(notifier *BubbleteaNotifier, onInterrupt func()) *TeaProgramHandle {
	if notifier.hasProgram() {
		return &TeaProgramHandle{}
	}
	m := NewProcessStatus()
	return m.Run(notifier, onInterrupt)
}

// printAboveWriter prints writes above a Bubbletea program, falling back to another writer once the program has exited.
type printAboveWriter struct {
	program  *tea.Program
	exited   <-chan struct{}
	fallback io.Writer
}

// Implement io.Writer: print p above the program without its trailing newline, as the printed line gets its own.
func (w *printAboveWriter) Write(p []byte) (int, error) {
	select {
	case <-w.exited:
		return w.fallback.Write(p)
	default:
	}
	// Send, unlike Program.Println, returns once the program has exited instead of blocking forever.
	// Sending the print message directly keeps lines in write order; returning it as a Cmd from Update runs it on its own goroutine.
	w.program.Send(tea.Println(strings.TrimSuffix(string(p), "\n"))())
	return len(p), nil
}
