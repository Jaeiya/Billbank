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
	_, err := sdb.handle.Exec(sdb.InsertInto(MONTHS, t.Year(), t.Month()))
	if err != nil {
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
		err = rows.Scan(&id, &year, &month)
		if err != nil {
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
	fm := buildFieldMap(BY_MONTH_ID|BY_MONTH|BY_YEAR, qm)
	row := sdb.handle.QueryRow(buildQueryStr(MONTHS, fm))

	var (
		id    int
		year  int
		month int
	)

	err := row.Scan(&id, &year, &month)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return MonthInfo{}, err
		}
		panic(err)
	}

	return MonthInfo{id, year, month}, nil
}
