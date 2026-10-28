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

import "time"

// Interface for handling events emitted by FlowTracker and ProgressTracker.
type Notifier interface {
	// Indicate a long running process has started.
	OnStart(pid string, time time.Time)

	// Indicate an error has occurred.
	OnError(pid string, err error, msg string)

	// Indicate an warning has occurred.
	OnWarn(pid, msg string)

	// Indicate an informational message.
	OnInfo(pid, msg string)

	// Indicate an internal message.
	OnDebug(pid, msg string)

	// Indicate a progress has advanced.
	OnProgress(pid string, cur, total uint64)

	// Indicate a long running process has finished.
	OnFinish(pid string, duration time.Duration)
}

type defaultNotifier struct{}

func (d *defaultNotifier) OnStart(pid string, time time.Time) {}

func (d *defaultNotifier) OnError(pid string, err error, msg string) {}

func (d *defaultNotifier) OnWarn(pid, msg string) {}

func (d *defaultNotifier) OnInfo(pid, msg string) {}

func (d *defaultNotifier) OnDebug(pid, msg string) {}

func (d *defaultNotifier) OnProgress(pid string, cur, total uint64) {}

func (d *defaultNotifier) OnFinish(pid string, duration time.Duration) {}
