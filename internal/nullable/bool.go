package nullable

import (
	"database/sql/driver"
	"encoding/json"
)

// Bool SQL type that can retrieve NULL value
type Bool struct {
	RealValue bool
	IsValid   bool
}

func FromBool(value bool) Bool {
	return Bool{
		RealValue: value,
		IsValid:   true,
	}
}

// NewBool creates a new nullable boolean
func NewBool(value *bool) Bool {
	if value == nil {
		return Bool{
			RealValue: false,
			IsValid:   false,
		}
	}
	return Bool{
		RealValue: *value,
		IsValid:   true,
	}
}

// Get either nil or boolean
func (n Bool) Get() *bool {
	if !n.IsValid {
		return nil
	}
	return &n.RealValue
}

// Set either nil or boolean
func (n *Bool) Set(value *bool) {
	n.IsValid = (value != nil)
	if n.IsValid {
		n.RealValue = *value
	} else {
		n.RealValue = false
	}
}

// MarshalJSON converts current value to JSON
func (n Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Get())
}

// UnmarshalJSON writes JSON to this type
func (n *Bool) UnmarshalJSON(data []byte) error {
	dataString := string(data)
	if len(dataString) == 0 || dataString == "null" {
		n.IsValid = false
		n.RealValue = false
		return nil
	}

	var parsed bool
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	n.IsValid = true
	n.RealValue = parsed
	return nil
}

// Scan implements scanner interface
func (n *Bool) Scan(value interface{}) error {
	if value == nil {
		n.RealValue, n.IsValid = false, false
		return nil
	}
	n.IsValid = true
	return convertAssign(&n.RealValue, value)
}

// Value implements the driver Valuer interface.
func (n Bool) Value() (driver.Value, error) {
	if !n.IsValid {
		return nil, nil
	}
	return n.RealValue, nil
}
