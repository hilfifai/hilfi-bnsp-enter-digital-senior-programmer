package helper

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DateFormat struct {
	time.Time
}

const DateLayout = "2006-01-02"

// --- JSON ---
func (d *DateFormat) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d DateFormat) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte(`null`), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", d.Time.Format(DateLayout))), nil
}

// --- SQL ---
func (d *DateFormat) Scan(value interface{}) error {
	if value == nil {
		*d = DateFormat{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := time.Parse(DateLayout, v)
		if err != nil {
			return err
		}
		d.Time = t
	case []byte:
		t, err := time.Parse(DateLayout, string(v))
		if err != nil {
			return err
		}
		d.Time = t
	default:
		return fmt.Errorf("cannot scan type %T into DateFormat", value)
	}
	return nil
}

func (d DateFormat) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time.Format(DateLayout), nil // biar DB nulis DATE sebagai string
}
