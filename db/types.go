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

package db

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)

// Bytes32 is a 32-byte value that can be stored as a Text column in SQLite.
type Bytes32 [32]byte

func (h Bytes32) Value() (driver.Value, error) {
	return hex.EncodeToString(h[:]), nil
}

func (h *Bytes32) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		b, err := hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("Bytes32: invalid hex string: %w", err)
		}
		if len(b) != 32 {
			return fmt.Errorf("Bytes32: expected 32 bytes, got %d", len(b))
		}
		copy(h[:], b)
		return nil
	case []byte:
		if len(v) == 32 {
			copy(h[:], v)
			return nil
		}
		return h.Scan(string(v))
	default:
		return fmt.Errorf("Bytes32: unable to scan type %T", value)
	}
}
