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

package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler/task"
)

func TestEnqueueAccepted(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		enqueueFn: func(args []string) (task.Task, error) {
			return task.Task{ID: "t1", Args: args, Status: task.StatusQueued}, nil
		},
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":["checksum","create"]}`)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d (%s)", http.StatusAccepted, w.Code, w.Body.String())
	}
}

func TestEnqueueEmptyArgsBadRequest(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) { return task.Task{}, task.ErrEmptyArgs },
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":[]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEnqueueInvalidArgsBadRequest(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) {
			return task.Task{}, fmt.Errorf("%w: %w", task.ErrInvalidArgs, errors.New("unknown flag: --bogus"))
		},
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":["file","hash","--bogus"]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "unknown flag: --bogus") {
		t.Errorf("expected the rejection reason in the body, got %s", w.Body.String())
	}
}

func TestEnqueueQueueFull(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) { return task.Task{}, task.ErrQueueFull },
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":["x"]}`)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestEnqueueInvalidBody(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{})
	w := do(s, http.MethodPost, "/tasks", `not json`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		getFn: func(string) (task.Task, bool) { return task.Task{}, false },
	})
	w := do(s, http.MethodGet, "/tasks/missing", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetFound(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		getFn: func(id string) (task.Task, bool) { return task.Task{ID: id, Status: task.StatusRunning}, true },
	})
	w := do(s, http.MethodGet, "/tasks/t1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestList(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		listFn: func() []task.Task { return []task.Task{{ID: "a"}, {ID: "b"}} },
	})
	w := do(s, http.MethodGet, "/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCancelNotFound(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		cancelFn: func(string) error { return task.ErrNotFound },
	})
	w := do(s, http.MethodPost, "/tasks/missing/cancel", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCancelFound(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		cancelFn: func(string) error { return nil },
		getFn:    func(id string) (task.Task, bool) { return task.Task{ID: id, Status: task.StatusCanceled}, true },
	})
	w := do(s, http.MethodPost, "/tasks/t1/cancel", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestClear(t *testing.T) {
	s := New(":0", StatusInfo{}, &fakeTasks{
		clearFn: func() int { return 3 },
	})
	w := do(s, http.MethodDelete, "/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
