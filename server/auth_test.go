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
)

func TestAPIKeyGuard(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		want          int
	}{
		{"missing header is rejected", "", http.StatusUnauthorized},
		{"wrong key is rejected", "Bearer wrong", http.StatusUnauthorized},
		{"prefix of the key is rejected", "Bearer s3cre", http.StatusUnauthorized},
		{"key without a scheme is rejected", "s3cret", http.StatusUnauthorized},
		{"key under another scheme is rejected", "Basic s3cret", http.StatusUnauthorized},
		{"matching key is accepted", "Bearer s3cret", http.StatusOK},
		{"scheme is matched case-insensitively", "bearer s3cret", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doAuthorized("/status", tt.authorization)
			if w.Code != tt.want {
				t.Errorf("expected status %d, got %d", tt.want, w.Code)
			}
		})
	}
}

func TestAPIKeyRejectionChallengesForBearer(t *testing.T) {
	w := doAuthorized("/status", "Bearer wrong")

	if got := w.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Errorf("expected WWW-Authenticate %q, got %q", "Bearer", got)
	}
	if !strings.Contains(w.Body.String(), ErrUnauthorized.Error()) {
		t.Errorf("expected body to carry %q, got %s", ErrUnauthorized.Error(), w.Body.String())
	}
}

func TestAPIKeyGuardsUnknownRoutes(t *testing.T) {
	w := doAuthorized("/unknown", "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// Serve a GET of path on a server guarded by the key "s3cret", sending authorization as the Authorization header unless it is empty.
func doAuthorized(path, authorization string) *httptest.ResponseRecorder {
	s := New(":0", "s3cret", StatusInfo{}, &fakeTasks{
		statsFn: func() (int, int) { return 0, 0 },
	})
	r := httptest.NewRequest(http.MethodGet, path, nil)
	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
