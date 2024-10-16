package sqlite

import (
	"fmt"
	"time"
)

var ErrDirtyDate = fmt.Errorf("not a clean date")

type MonthRecord struct {
	ID    int
	Year  int
	Month int
}

func (sdb SqliteDb) CreateMonth(t time.Time) {
	// Make sure any prior date arithmetic, used a clean date
	isClean := t.Day() == 1 &&
		t.Hour() == 0 &&
		t.Minute() == 0 &&
		t.Second() == 0 &&
		t.Nanosecond() == 0

	if !isClean {
		panic(ErrDirtyDate)
	}
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(MONTHS, t.Year(), t.Month()),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryMonths(qm QueryMap) ([]MonthRecord, error) {
	rows := sdb.query(MONTHS, qm)
	var records []MonthRecord

	for rows.Next() {
		var record MonthRecord
		if err := rows.Scan(
			&record.ID,
			&record.Year,
			&record.Month,
		); err != nil {
			panic(err)
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []MonthRecord{}, fmt.Errorf("empty month table")
	}

	return records, nil
}
