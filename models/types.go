package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONStringArray is a slice of strings that scans from/to JSON.
type JSONStringArray []string

// Scan implements the sql.Scanner interface.
func (a *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("type assertion to []byte or string failed")
	}

	return json.Unmarshal(b, a)
}

// Value implements the driver.Valuer interface.
func (a JSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}
