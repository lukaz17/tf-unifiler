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

package app

import "testing"

func TestCheckListenAuth(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		apiKey  string
		wantErr bool
	}{
		{"loopback without key is allowed", "127.0.0.1", "", false},
		{"localhost without key is allowed", "localhost", "", false},
		{"all interfaces without key is refused", "0.0.0.0", "", true},
		{"empty host without key is refused", "", "", true},
		{"LAN IP without key is refused", "192.168.1.5", "", true},
		{"LAN IP with key is allowed", "192.168.1.5", "s3cret", false},
		{"all interfaces with key is allowed", "0.0.0.0", "s3cret", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkListenAuth(tt.host, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkListenAuth(%q, key set: %t): expected error %t, got %v", tt.host, tt.apiKey != "", tt.wantErr, err)
			}
		})
	}
}
