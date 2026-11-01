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
	"bytes"
	"strings"
	"sync"
	"testing"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// idleModel renders nothing and only quits when told to.
type idleModel struct{}

func (idleModel) Init() tea.Cmd                         { return nil }
func (m idleModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (idleModel) View() tea.View                        { return tea.NewView("") }

// lockedBuffer is a bytes.Buffer safe to write from the program's goroutines.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// Run a program rendering to out in the background, detached from the terminal, and return its handle.
func runIdleProgram(t *testing.T, out *lockedBuffer) *TeaProgramHandle {
	t.Helper()
	p := tea.NewProgram(idleModel{}, tea.WithInput(nil), tea.WithOutput(out), tea.WithoutSignalHandler())
	n := NewBubbleteaNotifier()
	n.setProgram(p)
	exited := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		close(exited)
		done <- err
	}()
	return &TeaProgramHandle{program: p, notifier: n, exited: exited, done: done}
}

func TestRunProcessStatusReusesRunningProgram(t *testing.T) {
	n := NewBubbleteaNotifier()
	running := tea.NewProgram(idleModel{})
	n.setProgram(running)

	h := RunProcessStatus(n, nil)
	h.Stop()

	if h.program != nil {
		t.Error("expected a handle without program, got a second program on the terminal")
	}
	if n.i.teaProgram != running {
		t.Error("stopping the reusing handle unpaired the running program")
	}
}

func TestWriterPrintsAboveRunningProgram(t *testing.T) {
	out := &lockedBuffer{}
	fallback := &lockedBuffer{}
	h := runIdleProgram(t, out)

	w := h.Writer(fallback)
	if _, err := w.Write([]byte("Task started.\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	h.Stop()

	if !strings.Contains(out.String(), "Task started.") {
		t.Errorf("expected line printed above the program, got output %q", out.String())
	}
	if fallback.String() != "" {
		t.Errorf("expected nothing on fallback while the program runs, got %q", fallback.String())
	}
}

func TestWriterFallsBackAfterProgramExits(t *testing.T) {
	out := &lockedBuffer{}
	fallback := &lockedBuffer{}
	h := runIdleProgram(t, out)
	w := h.Writer(fallback)
	h.Stop()

	n, err := w.Write([]byte("Shutdown signal received.\n"))

	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len("Shutdown signal received.\n") {
		t.Errorf("expected %d bytes written, got %d", len("Shutdown signal received.\n"), n)
	}
	if fallback.String() != "Shutdown signal received.\n" {
		t.Errorf("expected line on fallback, got %q", fallback.String())
	}
}

func TestWriterWithoutProgramIsFallback(t *testing.T) {
	fallback := &lockedBuffer{}
	h := &TeaProgramHandle{}

	w := h.Writer(fallback)

	if w != fallback {
		t.Errorf("expected fallback writer itself, got %T", w)
	}
}

func TestViewLabelsTaskBar(t *testing.T) {
	m := ProcessStatus{
		totalPercent:  0.6,
		totalLabel:    "3/5 tasks",
		itemProgress:  progress.New(progress.WithWidth(barWidth)),
		itemSpinner:   spinner.New(),
		totalProgress: progress.New(progress.WithWidth(barWidth)),
	}

	lines := strings.Split(m.View().Content, "\n")

	if len(lines) < 4 || !strings.HasSuffix(lines[3], "  3/5 tasks") {
		t.Errorf("expected overall bar line ending with the task label, got %q", lines)
	}
	if strings.Contains(m.View().Content, "Please wait") {
		t.Error("expected a task bar, got the spinner of an unknown total")
	}
}
