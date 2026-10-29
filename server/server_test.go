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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tforceaio/tf-unifiler/task"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// fakeTasks is a controllable TaskService stub for handler tests.
type fakeTasks struct {
	enqueueFn func([]string) (task.Task, error)
	getFn     func(string) (task.Task, bool)
	listFn    func() []task.Task
	cancelFn  func(string) error
	clearFn   func() int
}

func (f *fakeTasks) Enqueue(args []string) (task.Task, error) { return f.enqueueFn(args) }
func (f *fakeTasks) Get(id string) (task.Task, bool)          { return f.getFn(id) }
func (f *fakeTasks) List() []task.Task                        { return f.listFn() }
func (f *fakeTasks) Cancel(id string) error                   { return f.cancelFn(id) }
func (f *fakeTasks) Clear() int                               { return f.clearFn() }

func do(s *Server, method, path, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

func TestHealthEndpoint(t *testing.T) {
	s := New(":0", &fakeTasks{})
	w := do(s, http.MethodGet, "/health", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body %q, got %q", "ok", w.Body.String())
	}
}

func TestUnknownRoute(t *testing.T) {
	s := New(":0", &fakeTasks{})
	w := do(s, http.MethodGet, "/unknown", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestEnqueueAccepted(t *testing.T) {
	s := New(":0", &fakeTasks{
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
	s := New(":0", &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) { return task.Task{}, task.ErrEmptyArgs },
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":[]}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEnqueueQueueFull(t *testing.T) {
	s := New(":0", &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) { return task.Task{}, task.ErrQueueFull },
	})
	w := do(s, http.MethodPost, "/tasks", `{"args":["x"]}`)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestEnqueueInvalidBody(t *testing.T) {
	s := New(":0", &fakeTasks{})
	w := do(s, http.MethodPost, "/tasks", `not json`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	s := New(":0", &fakeTasks{
		getFn: func(string) (task.Task, bool) { return task.Task{}, false },
	})
	w := do(s, http.MethodGet, "/tasks/missing", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetFound(t *testing.T) {
	s := New(":0", &fakeTasks{
		getFn: func(id string) (task.Task, bool) { return task.Task{ID: id, Status: task.StatusRunning}, true },
	})
	w := do(s, http.MethodGet, "/tasks/t1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestList(t *testing.T) {
	s := New(":0", &fakeTasks{
		listFn: func() []task.Task { return []task.Task{{ID: "a"}, {ID: "b"}} },
	})
	w := do(s, http.MethodGet, "/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCancelNotFound(t *testing.T) {
	s := New(":0", &fakeTasks{
		cancelFn: func(string) error { return task.ErrNotFound },
	})
	w := do(s, http.MethodPost, "/tasks/missing/cancel", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCancelFound(t *testing.T) {
	s := New(":0", &fakeTasks{
		cancelFn: func(string) error { return nil },
		getFn:    func(id string) (task.Task, bool) { return task.Task{ID: id, Status: task.StatusCanceled}, true },
	})
	w := do(s, http.MethodPost, "/tasks/t1/cancel", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}

func TestClear(t *testing.T) {
	s := New(":0", &fakeTasks{
		clearFn: func() int { return 3 },
	})
	w := do(s, http.MethodDelete, "/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
}
