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

import "testing"

func TestDialHost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"empty host is reached over IPv4 loopback", "", "127.0.0.1"},
		{"IPv4 wildcard is reached over IPv4 loopback", "0.0.0.0", "127.0.0.1"},
		{"IPv6 wildcard is reached over IPv6 loopback", "::", "::1"},
		{"loopback IP is kept", "127.0.0.1", "127.0.0.1"},
		{"LAN IP is kept", "192.168.1.5", "192.168.1.5"},
		{"IPv6 address is kept", "fd00::5", "fd00::5"},
		{"DNS name is kept", "nas.lan", "nas.lan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DialHost(tt.host); got != tt.want {
				t.Errorf("DialHost(%q): expected %q, got %q", tt.host, tt.want, got)
			}
		})
	}
}

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"localhost is loopback", "localhost", true},
		{"localhost is matched case-insensitively", "LocalHost", true},
		{"IPv4 loopback is loopback", "127.0.0.1", true},
		{"whole IPv4 loopback block is loopback", "127.1.2.3", true},
		{"IPv6 loopback is loopback", "::1", true},
		{"empty host listens on all interfaces", "", false},
		{"IPv4 wildcard listens on all interfaces", "0.0.0.0", false},
		{"IPv6 wildcard listens on all interfaces", "::", false},
		{"LAN IP is reachable from the network", "192.168.1.5", false},
		{"DNS name is treated as reachable from the network", "nas.lan", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLoopback(tt.host); got != tt.want {
				t.Errorf("IsLoopback(%q): expected %t, got %t", tt.host, tt.want, got)
			}
		})
	}
}
