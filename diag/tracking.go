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

import (
	"fmt"
	"time"

	"github.com/tforce-io/tf-golib/random/securerng"
	"github.com/tforce-io/tf-golib/stdx/mathxt"
)

// FlowTracker tracks execution flow.
type FlowTracker struct {
	fnName    string
	instance  string
	createdAt time.Time
	notifier  Notifier
}

// Return a new FlowTracker instance and emit a Info event with Start message.
func NewFlowTracker(fnName string, notifier Notifier) *FlowTracker {
	instance := fmt.Sprintf("%s-%s", fnName, securerng.Hex(8))
	f := &FlowTracker{
		fnName:    fnName,
		instance:  instance,
		createdAt: time.Now().UTC(),
		notifier:  notifier,
	}
	if f.notifier == nil {
		f.notifier = &defaultNotifier{}
	}
	f.notifier.OnInfo(f.instance, "Started.")
	return f
}

// Emit an Error event.
func (f *FlowTracker) Error(err error, msg string) {
	f.notifier.OnError(f.instance, err, msg)
}

// Emit a Warn event.
func (f *FlowTracker) Warn(msg string) {
	f.notifier.OnWarn(f.instance, msg)
}

// Emit an Info event.
func (f *FlowTracker) Info(msg string) {
	f.notifier.OnInfo(f.instance, msg)
}

// Emit a Debug event.
func (f *FlowTracker) Debug(msg string) {
	f.notifier.OnDebug(f.instance, msg)
}

// Emit a Info event with Finished message.
func (f *FlowTracker) Done() {
	f.notifier.OnInfo(f.instance, "Finished.")
}

// ProgressTracker tracks long process.
type ProgressTracker struct {
	fnName    string
	instance  string
	createdAt time.Time
	total     uint64
	notifier  Notifier
}

// Return a new ProgressTracker instance and emit a Start event.
func NewProgressTracker(fnName string, notifier Notifier) *ProgressTracker {
	instance := fmt.Sprintf("%s-%s", fnName, securerng.Hex(8))
	p := &ProgressTracker{
		fnName:    fnName,
		instance:  instance,
		createdAt: time.Now().UTC(),
		notifier:  notifier,
	}
	if p.notifier == nil {
		p.notifier = &defaultNotifier{}
	}
	p.notifier.OnStart(p.instance, p.createdAt)
	return p
}

// Set total unit to process.
func (p *ProgressTracker) Total(v int64) {
	p.total = uint64(mathxt.AbsInt64(v))
}

// Emit a Info event.
func (p *ProgressTracker) Status(msg string) {
	p.notifier.OnInfo(p.instance, msg)
}

// Emit a Progress event.
func (p *ProgressTracker) Progress(v int64) {
	p.notifier.OnProgress(p.instance, uint64(v), p.total)
}

// Emit a Finish event.
func (p *ProgressTracker) Done() {
	duration := time.Since(p.createdAt)
	p.notifier.OnFinish(p.instance, duration)
}
