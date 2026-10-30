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

func TestAPIKeyGuardsShutdown(t *testing.T) {
	s := New(":0", "s3cret", StatusInfo{}, &fakeTasks{})
	w := do(s, http.MethodPost, "/shutdown", "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
	if isClosed(s.ShutdownRequested()) {
		t.Error("an unauthorized request requested a shutdown")
	}
}

func TestCrossOriginGuard(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{"cross-site page is rejected", map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://attacker.example"}, http.StatusForbidden},
		{"page on another localhost port is rejected", map[string]string{"Sec-Fetch-Site": "same-site", "Origin": "http://localhost:3000"}, http.StatusForbidden},
		{"foreign origin without fetch metadata is rejected", map[string]string{"Origin": "https://attacker.example"}, http.StatusForbidden},
		{"same-origin request is accepted", map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusAccepted},
		{"matching origin without fetch metadata is accepted", map[string]string{"Origin": "http://" + loopbackHost}, http.StatusAccepted},
		{"non-browser client is accepted", nil, http.StatusAccepted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(":0", "", StatusInfo{}, &fakeTasks{})
			r := httptest.NewRequest(http.MethodPost, "/shutdown", nil)
			r.Host = loopbackHost
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}

			w := serve(s, r)

			if w.Code != tt.want {
				t.Errorf("expected status %d, got %d", tt.want, w.Code)
			}
			if want := tt.want == http.StatusAccepted; isClosed(s.ShutdownRequested()) != want {
				t.Errorf("expected shutdown requested to be %v", want)
			}
		})
	}
}

func TestCrossOriginGuardAppliesWithAPIKey(t *testing.T) {
	s := New(":0", "s3cret", StatusInfo{}, &fakeTasks{})
	r := httptest.NewRequest(http.MethodPost, "/shutdown", nil)
	r.Host = loopbackHost
	r.Header.Set("Authorization", "Bearer s3cret")
	r.Header.Set("Sec-Fetch-Site", "cross-site")

	w := serve(s, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if isClosed(s.ShutdownRequested()) {
		t.Error("a cross-origin request requested a shutdown")
	}
}

func TestLoopbackHostGuard(t *testing.T) {
	tests := []struct {
		name string
		host string
		want int
	}{
		{"IPv4 loopback is accepted", "127.0.0.1:8080", http.StatusOK},
		{"IPv6 loopback is accepted", "[::1]:8080", http.StatusOK},
		{"localhost on the default port is accepted", "localhost", http.StatusOK},
		{"rebound domain is rejected", "attacker.example:8080", http.StatusForbidden},
		{"domain starting with a loopback IP is rejected", "127.0.0.1.attacker.example:8080", http.StatusForbidden},
		{"LAN address is rejected", "192.168.1.10:8080", http.StatusForbidden},
		{"missing host is rejected", "", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(":0", "", StatusInfo{}, &fakeTasks{
				statsFn: func() (int, int) { return 0, 0 },
			})
			r := httptest.NewRequest(http.MethodGet, "/status", nil)
			r.Host = tt.host

			w := serve(s, r)

			if w.Code != tt.want {
				t.Errorf("expected status %d, got %d", tt.want, w.Code)
			}
		})
	}
}

func TestRebindingPageCannotShutDown(t *testing.T) {
	s := New(":0", "", StatusInfo{}, &fakeTasks{})
	r := httptest.NewRequest(http.MethodPost, "/shutdown", nil)
	r.Host = "attacker.example:8080"
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("Origin", "http://attacker.example:8080")

	w := serve(s, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if isClosed(s.ShutdownRequested()) {
		t.Error("a rebinding page requested a shutdown")
	}
}

func TestAPIKeyServerAcceptsLANHost(t *testing.T) {
	s := New(":0", "s3cret", StatusInfo{}, &fakeTasks{
		statsFn: func() (int, int) { return 0, 0 },
	})
	r := httptest.NewRequest(http.MethodGet, "/status", nil)
	r.Host = "192.168.1.10:8080"
	r.Header.Set("Authorization", "Bearer s3cret")

	w := serve(s, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
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
