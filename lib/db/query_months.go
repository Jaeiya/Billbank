package db

import (
	"fmt"
	"strings"
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

func (sdb SqliteDb) QueryAllMonths() ([]MonthInfo, error) {
	rows, err := sdb.handle.Query(fmt.Sprintf("SELECT * FROM %s", MONTHS))
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id        int
		year      int
		month     int
		monthRows []MonthInfo
	)

	for rows.Next() {
		if err = rows.Scan(&id, &year, &month); err != nil {
			panic(err)
		}
		monthRows = append(monthRows, MonthInfo{id, year, month})
	}

	if len(monthRows) == 0 {
		return []MonthInfo{}, fmt.Errorf("empty month table")
	}

	return monthRows, nil
}

func (sdb SqliteDb) QueryMonth(qm QueryMap) (MonthInfo, error) {
	fm := buildFieldMap(WHERE_MONTH_ID|WHERE_MONTH|WHERE_YEAR, qm)
	row := sdb.handle.QueryRow(buildQueryStr(MONTHS, fm))

	var (
		id    int
		year  int
		month int
	)

	if err := row.Scan(&id, &year, &month); err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return MonthInfo{}, err
		}
		panic(err)
	}

	return MonthInfo{id, year, month}, nil
}
