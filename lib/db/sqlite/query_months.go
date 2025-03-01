package sqlite

import (
	"fmt"
	"time"
)

var ErrDirtyDate = fmt.Errorf("found dirty date; use NewMonth() to create the date")

type Month time.Time

func NewMonth(year int, m time.Month) Month {
	return Month(time.Date(year, m, 1, 0, 0, 0, 0, time.Local))
}

type MonthRecord struct {
	ID    int
	Year  int
	Month int
}

func (sdb SqliteDb) CreateMonth(m Month) error {
	t := time.Time(m)

	// New months should just contain a modified year & month
	isClean := t.Day() == 1 &&
		t.Hour() == 0 &&
		t.Minute() == 0 &&
		t.Second() == 0 &&
		t.Nanosecond() == 0

	if !isClean {
		return ErrDirtyDate
	}
	if _, err := sdb.handle.Exec(sdb.ToInsertIntoStr(MONTHS, t.Year(), t.Month())); err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryMonths(qm QueryMap) ([]MonthRecord, error) {
	rows, err := sdb.query(MONTHS, qm)
	if err != nil {
		return []MonthRecord{}, err
	}

	var records []MonthRecord
	for rows.Next() {
		var record MonthRecord
		if err := rows.Scan(
			&record.ID,
			&record.Year,
			&record.Month,
		); err != nil {
			return []MonthRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []MonthRecord{}, ErrMonthNotFound
	}

	return records, nil
}
