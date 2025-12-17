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

// Return clip duration based on number of captures, interval and adjustment settings
func CaptureNumberToDuration(num, intv, xrf, limit uint64) uint64 {
	adj := uint64(0)
	if num > limit {
		adj = (num - limit) * intv / xrf
	}
	return num * (intv + adj)
}

// Return the interval based on clip duration and adjustment settings
func DurationToInterval(time, intv, xrf, limit uint64) uint64 {
	upperNumber := time / intv
	lowerNumber := time / (intv + upperNumber*xrf)

	finalTime := time
	finalNumber := (upperNumber + lowerNumber) / 2
	for lowerNumber <= upperNumber {
		midNumber := (upperNumber + lowerNumber) / 2
		timeFromNum := CaptureNumberToDuration(midNumber, intv, xrf, limit)
		if timeFromNum > time {
			upperNumber = midNumber - 1
			continue
		}
		finalTime = timeFromNum
		finalNumber = midNumber
		lowerNumber = midNumber + 1
	}

	return finalTime / finalNumber
}

// Return number of captures and interval based on clip duration and adjustment settings
func DurationToCaptureNumberAndInterval(time, intv, xrf, limit uint64) (uint64, uint64) {
	adjIntv := DurationToInterval(time, intv, xrf, limit)
	adjNum := time / adjIntv
	return adjNum, adjIntv
}
