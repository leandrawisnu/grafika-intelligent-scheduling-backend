package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafika-scheduling/backend/pkg/timezone"
)

// APIDate accepts JSON "YYYY-MM-DD" (frontend) or RFC3339 (legacy).
type APIDate struct {
	time.Time
}

func (d APIDate) calendarDateInJakarta() time.Time {
	if d.Time.IsZero() {
		return d.Time
	}
	t := d.Time.In(timezone.Loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timezone.Loc)
}

func (d APIDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.calendarDateInJakarta().Format("2006-01-02"))
}

func (d *APIDate) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(strings.Trim(string(data), `"`))
	if raw == "" || raw == "null" {
		d.Time = time.Time{}
		return nil
	}
	if len(raw) >= 10 && raw[4] == '-' && raw[7] == '-' {
		t, err := time.ParseInLocation("2006-01-02", raw[:10], timezone.Loc)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", raw, err)
	}
	j := t.In(timezone.Loc)
	d.Time = time.Date(j.Year(), j.Month(), j.Day(), 0, 0, 0, 0, timezone.Loc)
	return nil
}

func (d APIDate) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.calendarDateInJakarta().Format("2006-01-02"), nil
}

func (d *APIDate) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		j := v.In(timezone.Loc)
		d.Time = time.Date(j.Year(), j.Month(), j.Day(), 0, 0, 0, 0, timezone.Loc)
		return nil
	case []byte:
		return d.parseDateString(string(v))
	case string:
		return d.parseDateString(v)
	default:
		return fmt.Errorf("cannot scan %T into APIDate", value)
	}
}

func (d *APIDate) parseDateString(s string) error {
	s = strings.TrimSpace(s)
	if len(s) >= 10 {
		s = s[:10]
	}
	t, err := time.ParseInLocation("2006-01-02", s, timezone.Loc)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}
