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
	tea "charm.land/bubbletea/v2"
)

// TeaProgramHandle provides caller a way to stop the Bubbletea program.
type TeaProgramHandle struct {
	program  *tea.Program
	notifier *BubbleteaNotifier
	done     chan error
}

// Stop running Bubbletea program.
func (p *TeaProgramHandle) Stop() {
	p.notifier.setProgram(nil)
	p.program.Quit()
	<-p.done
}

// Run CompressOptions and show it on the terminal.
func RunCompressOptions(defaults *CompressOptionsValue) (*CompressOptionsValue, error) {
	m := NewCompressOptions().
		WithSelected(defaults)

	return m.Run()
}

// Run ProcessStatus and show it on the terminal.
func RunProcessStatus(notifier *BubbleteaNotifier, onInterrupt func()) *TeaProgramHandle {
	m := NewProcessStatus()
	return m.Run(notifier, onInterrupt)
}
