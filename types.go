package aplos

import (
	"encoding/json"
	"fmt"
	"time"
)

// Time wraps the standard library's time.Time and supports the format returned
// by the Aplos API for time fields.
type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to unmarshal JSON field as a string: %w", err)
	}

	tmp, err := time.Parse("2006-01-02T15:04:05.999-0700", s)
	if err != nil {
		return fmt.Errorf("failed to parse time: %w", err)
	}
	*t = Time{tmp}

	return err
}

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to unmarshal JSON field as a string: %w", err)
	}

	tmp, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}
	y, m, day := tmp.Date()
	*d = Date{
		Year:  y,
		Month: m,
		Day:   day,
	}

	return err
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}
