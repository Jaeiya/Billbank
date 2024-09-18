package db

import (
	"fmt"
	"time"
)

type MonthInfo struct {
	ID    int
	Year  int
	Month int
}

func (sdb SqliteDb) CreateMonth(t time.Time) error {
	// Make sure any prior date arithmetic, used a clean date
	isClean := t.Day() == 1 &&
		t.Hour() == 0 &&
		t.Minute() == 0 &&
		t.Second() == 0 &&
		t.Nanosecond() == 0

	if !isClean {
		return fmt.Errorf("not a clean date")
	}
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(MONTHS, t.Year(), t.Month()),
	); err != nil {
		return err
	}
	return nil
}

func (sdb SqliteDb) QueryMonths(qm QueryMap) ([]MonthInfo, error) {
	rows := sdb.query(MONTHS, qm)
	var monthRows []MonthInfo

	for rows.Next() {
		var row MonthInfo
		if err := rows.Scan(&row.ID, &row.Year, &row.Month); err != nil {
			panic(err)
		}
		monthRows = append(monthRows, row)
	}

	if len(monthRows) == 0 {
		return []MonthInfo{}, fmt.Errorf("empty month table")
	}

	return monthRows, nil
}
