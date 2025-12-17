// Copyright (C) 2025 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeToShotNumberAndInterval(t *testing.T) {
	tests := []struct {
		name    string
		time    uint64
		intv    uint64
		xrf     uint64
		limit   uint64
		adjNum  uint64
		adjIntv uint64
	}{
		{"15/225/10", 70000, 15000, 225, 10, 4, 15000},
		{"15/225/10", 150000, 15000, 225, 10, 10, 15000},
		{"15/225/10", 165000, 15000, 225, 10, 11, 15000},
		{"15/225/10", 170000, 15000, 225, 10, 11, 15066},
		{"15/225/10", 6565000, 15000, 225, 10, 224, 29266},
		{"30/300/90", 1801000, 30000, 300, 90, 60, 30000},
		{"30/300/90", 2700000, 30000, 300, 90, 90, 30000},
		{"30/300/90", 2730000, 30000, 300, 90, 91, 30000},
		{"30/300/90", 2740000, 30000, 300, 90, 91, 30100},
		{"30/300/90", 10801000, 30000, 300, 90, 240, 45000},
		{"50/250/216", 9000000, 50000, 250, 216, 180, 50000},
		{"50/250/216", 10800000, 50000, 250, 216, 216, 50000},
		{"50/250/216", 10890000, 50000, 250, 216, 217, 50000},
		{"50/250/216", 10896000, 50000, 250, 216, 217, 50200},
		{"50/250/216", 28370000, 50000, 250, 216, 360, 78800},
		{"60/360/180", 9000000, 60000, 360, 180, 150, 60000},
		{"60/360/180", 10800000, 60000, 360, 180, 180, 60000},
		{"60/360/180", 10890000, 60000, 360, 180, 181, 60000},
		{"60/360/180", 10896000, 60000, 360, 180, 181, 60166},
		{"60/360/180", 32401000, 60000, 360, 180, 360, 90000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjIntv := DurationToInterval(tt.time, tt.intv, tt.xrf, tt.limit)
			assert.Equal(t, tt.adjIntv, adjIntv)
			adjNum, adjIntv := DurationToCaptureNumberAndInterval(tt.time, tt.intv, tt.xrf, tt.limit)
			assert.Equal(t, tt.adjNum, adjNum)
			assert.Equal(t, tt.adjIntv, adjIntv)
		})
	}
}
