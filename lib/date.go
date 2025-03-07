package lib

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Date struct {
	time time.Time
}

func NewDate(year int, month time.Month, day int) (Date, error) {
	if month < 1 || month > 12 {
		return Date{}, fmt.Errorf("invalid month")
	}

	d := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	if d.Year() != year || d.Month() != month || d.Day() != day {
		return Date{}, fmt.Errorf("invalid date; date parsed into an unexpected value")
	}

	return Date{time: d}, nil
}

func (d Date) String() string {
	return d.time.Format(time.RFC3339)
}

func (d Date) GetTime() time.Time {
	return d.time
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *Date) Scan(data any) error {
	var err error
	if v, isStr := data.(string); isStr {
		d.time, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("failed date scan; %s", err)
		}
		return nil
	}
	return fmt.Errorf("failed date scan; bad data type: %T", data)
}
