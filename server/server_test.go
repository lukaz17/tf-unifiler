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
	"encoding/json"
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
	statsFn   func() (int, int)
}

func (f *fakeTasks) Enqueue(args []string) (task.Task, error) { return f.enqueueFn(args) }
func (f *fakeTasks) Get(id string) (task.Task, bool)          { return f.getFn(id) }
func (f *fakeTasks) List() []task.Task                        { return f.listFn() }
func (f *fakeTasks) Cancel(id string) error                   { return f.cancelFn(id) }
func (f *fakeTasks) Clear() int                               { return f.clearFn() }
func (f *fakeTasks) Stats() (int, int)                        { return f.statsFn() }

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

func TestStatusEndpoint(t *testing.T) {
	s := New(":0", "", StatusInfo{
		Service:    "tf-unifiler",
		Version:    "1.2.3",
		WorkingDir: "/wd",
		ConfigDir:  "/cfg",
		Executable: "/bin/unifiler",
		Portable:   true,
	}, &fakeTasks{
		statsFn: func() (int, int) { return 2, 5 },
	})
	w := do(s, http.MethodGet, "/status", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decoding status response: %v", err)
	}
	if got.Service != "tf-unifiler" {
		t.Errorf("expected service %q, got %q", "tf-unifiler", got.Service)
	}
	if got.Version != "1.2.3" || got.WorkingDir != "/wd" || got.ConfigDir != "/cfg" || got.Executable != "/bin/unifiler" || !got.Portable {
		t.Errorf("status identity fields not echoed back: %+v", got)
	}
	if got.PendingTasks != 2 || got.TotalTasks != 5 {
		t.Errorf("expected counts 2/5, got %d/%d", got.PendingTasks, got.TotalTasks)
	}
}

func TestUnknownRoute(t *testing.T) {
	s := New(":0", "", StatusInfo{}, &fakeTasks{})
	w := do(s, http.MethodGet, "/unknown", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
