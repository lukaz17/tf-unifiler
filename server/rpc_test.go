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

package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeDispatcher implements Dispatcher for handler tests.
type fakeDispatcher struct {
	enqueueErr error
	jobs       map[string]JobView
	cancelled  []string
}

func (f *fakeDispatcher) Enqueue(args []string) (string, error) {
	if f.enqueueErr != nil {
		return "", f.enqueueErr
	}
	id := "job-1"
	f.jobs[id] = JobView{
		ID:          id,
		Args:        args,
		Status:      "queued",
		SubmittedAt: time.Unix(0, 0).UTC(),
	}
	return id, nil
}

func (f *fakeDispatcher) Job(id string) (JobView, bool) {
	view, ok := f.jobs[id]
	return view, ok
}

func (f *fakeDispatcher) Jobs() []JobView {
	views := []JobView{}
	for _, view := range f.jobs {
		views = append(views, view)
	}
	return views
}

func (f *fakeDispatcher) Cancel(id string) bool {
	if _, ok := f.jobs[id]; !ok {
		return false
	}
	if strings.Contains(id, "running") {
		return false
	}
	f.cancelled = append(f.cancelled, id)
	job := f.jobs[id]
	job.Status = "cancelled"
	f.jobs[id] = job
	return true
}

func TestEnqueueJob(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{}}
	s := New(":0", WithDispatcher(d))

	body := `{"args": ["file", "pack", "--format", "7z"]}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d with body %s", http.StatusAccepted, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"queued"`) {
		t.Errorf("expected queued status in response, got %s", w.Body.String())
	}
}

func TestEnqueueJobBadRequest(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{}}
	s := New(":0", WithDispatcher(d))

	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{invalid`))
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"args": []}`))
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for empty args, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEnqueueJobValidation(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{}, enqueueErr: errors.New("unknown command")}
	s := New(":0", WithDispatcher(d))

	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"args": ["bogus"]}`))
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d with body %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestListJobs(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{
		"job-1": {ID: "job-1", Status: "done"},
	}}
	s := New(":0", WithDispatcher(d))

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Body.String(), `"job-1"`) {
		t.Errorf("expected job in listing, got %s", w.Body.String())
	}
}

func TestGetJob(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{
		"job-1": {ID: "job-1", Status: "running"},
	}}
	s := New(":0", WithDispatcher(d))

	req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/jobs/missing", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCancelJob(t *testing.T) {
	d := &fakeDispatcher{jobs: map[string]JobView{
		"job-1":       {ID: "job-1", Status: "queued"},
		"job-running": {ID: "job-running", Status: "running"},
	}}
	s := New(":0", WithDispatcher(d))

	req := httptest.NewRequest(http.MethodDelete, "/jobs/job-1", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/jobs/job-running", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/jobs/missing", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
