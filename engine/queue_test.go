// Copyright (C) 2026 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package engine

import "testing"

func TestQueuedRunnerFindValid(t *testing.T) {
	r := queuedRunner{}
	for _, args := range [][]string{
		{"file", "pack", "--format", "7z"},
		{"checksum", "create"},
		{"metadata", "query", "hash"},
	} {
		if err := r.Find(args); err != nil {
			t.Errorf("expected command %v to be resolvable, got %v", args, err)
		}
	}
}

func TestQueuedRunnerFindUnknown(t *testing.T) {
	r := queuedRunner{}
	if err := r.Find([]string{"bogus"}); err == nil {
		t.Fatal("expected unknown command to be rejected")
	}
}

func TestQueuedRunnerFindServeBlocked(t *testing.T) {
	r := queuedRunner{}
	if err := r.Find([]string{"serve"}); err == nil {
		t.Fatal("expected serve command to be rejected")
	}
}
