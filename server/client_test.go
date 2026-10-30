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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tforceaio/tf-unifiler/task"
)

func TestClientStatus(t *testing.T) {
	c := newTestClient(t, StatusInfo{Service: "tf-unifiler", Version: "9.9.9"}, &fakeTasks{
		statsFn: func() (int, int) { return 1, 3 },
	})

	st, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Service != "tf-unifiler" {
		t.Errorf("expected service %q, got %q", "tf-unifiler", st.Service)
	}
	if st.Version != "9.9.9" {
		t.Errorf("expected version echoed, got %q", st.Version)
	}
	if st.PendingTasks != 1 || st.TotalTasks != 3 {
		t.Errorf("expected counts 1/3, got %d/%d", st.PendingTasks, st.TotalTasks)
	}
}

func TestClientStatusUnreachable(t *testing.T) {
	// A server that is created then immediately closed is not listening.
	ts := httptest.NewServer(New(":0", "", StatusInfo{}, &fakeTasks{}).Handler())
	url := ts.URL
	ts.Close()

	if _, err := NewClient(url, "").Status(context.Background()); err == nil {
		t.Fatal("expected error contacting a closed server, got nil")
	}
}

func TestClientStatusForeignServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service":"something-else","status":"ok"}`))
	}))
	t.Cleanup(ts.Close)

	st, err := NewClient(ts.URL, "").Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Service == "tf-unifiler" {
		t.Errorf("foreign server should not report service %q", "tf-unifiler")
	}
}

func TestClientEnqueueAccepted(t *testing.T) {
	c := newTestClient(t, StatusInfo{}, &fakeTasks{
		enqueueFn: func(args []string) (task.Task, error) {
			return task.Task{ID: "t1", Args: args, Status: task.StatusQueued}, nil
		},
	})

	got, err := c.Enqueue(context.Background(), []string{"file", "hash"})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if got.ID != "t1" || got.Status != task.StatusQueued {
		t.Errorf("unexpected task returned: %+v", got)
	}
}

func TestClientEnqueueQueueFull(t *testing.T) {
	c := newTestClient(t, StatusInfo{}, &fakeTasks{
		enqueueFn: func([]string) (task.Task, error) { return task.Task{}, task.ErrQueueFull },
	})

	if _, err := c.Enqueue(context.Background(), []string{"x"}); err == nil {
		t.Fatal("expected error on a full queue, got nil")
	}
}

func TestClientStatusSendsAPIKey(t *testing.T) {
	url := newKeyedTestServer(t, "s3cret")

	st, err := NewClient(url, "s3cret").Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Service != "tf-unifiler" {
		t.Errorf("expected service %q, got %q", "tf-unifiler", st.Service)
	}
}

func TestClientStatusWrongAPIKey(t *testing.T) {
	url := newKeyedTestServer(t, "s3cret")

	_, err := NewClient(url, "wrong").Status(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestClientEnqueueMissingAPIKey(t *testing.T) {
	url := newKeyedTestServer(t, "s3cret")

	_, err := NewClient(url, "").Enqueue(context.Background(), []string{"x"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

// Stand up a real server handler guarded by apiKey over httptest and return its URL.
func newKeyedTestServer(t *testing.T, apiKey string) string {
	t.Helper()
	ts := httptest.NewServer(New(":0", apiKey, StatusInfo{Service: "tf-unifiler"}, &fakeTasks{
		statsFn: func() (int, int) { return 0, 0 },
		enqueueFn: func(args []string) (task.Task, error) {
			return task.Task{ID: "t1", Args: args, Status: task.StatusQueued}, nil
		},
	}).Handler())
	t.Cleanup(ts.Close)
	return ts.URL
}

// Stand up a real server handler over httptest and return a client pointed at it.
func newTestClient(t *testing.T, info StatusInfo, tasks TaskService) *Client {
	t.Helper()
	ts := httptest.NewServer(New(":0", "", info, tasks).Handler())
	t.Cleanup(ts.Close)
	return NewClient(ts.URL, "")
}
