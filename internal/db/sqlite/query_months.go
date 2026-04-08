package sqlite

import (
	"fmt"
	"time"

	"github.com/jaeiya/billbank/internal"
)

var ErrCreatePastTime = fmt.Errorf("invalid year or month; cannot create past months/years")

type MonthRecord struct {
	ID   int
	Date internal.Date
}

func (sdb SqliteDb) CreateMonth(year int, month time.Month) (int64, error) {
	now := time.Now()
	// Maybe in the future this might be a bad idea, but
	// for now, there's no reason to create months
	// or years that have already passed.
	if year < now.Year() || month < now.Month() {
		return 0, ErrCreatePastTime
	}

	d, _ := internal.NewDate(year, month, 1)

	res, err := sdb.insertInto(MONTHS, d)
	if err != nil {
		return 0, getExecError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
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
			&record.Date,
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
