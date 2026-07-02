package sqlite

import (
	"fmt"

	"github.com/jaeiya/billbank/internal"
)

type IncomeRecord struct {
	ID     int
	Name   string
	Amount internal.Currency
	Period Period
}

type IncomeHistoryRecord struct {
	ID       int
	IncomeID int
	MonthID  int
	Amount   internal.Currency
}

func (ih IncomeHistoryRecord) String() string {
	s := fmt.Sprintf(
		"id: %d\nincomeID: %d\nmonthID: %d\namount: %s",
		ih.ID,
		ih.IncomeID,
		ih.MonthID,
		ih.Amount.String(),
	)
	return s
}

type AffixIncomeRecord struct {
	ID              int
	IncomeHistoryID int
	Name            string
	Amount          internal.Currency
}

func (db SqliteDb) CreateIncome(config IncomeRecord) (int64, error) {
	res, err := db.insertInto(
		INCOME,
		config.Name,
		config.Amount,
		config.Period,
	)
	if err != nil {
		return 0, getExecError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db SqliteDb) SetIncome(id int, amount internal.Currency) error {
	_, err := db.handle.Exec(
		fmt.Sprintf("UPDATE income SET amount=%d WHERE id=%d", amount.GetStoredValue(), id),
	)
	if err != nil {
		return err
	}
	return nil
}

func (db SqliteDb) QueryIncome(qm QueryMap) ([]IncomeRecord, error) {
	rows, err := db.query(INCOME, qm)
	if err != nil {
		return []IncomeRecord{}, err
	}

	var records []IncomeRecord
	for rows.Next() {
		var record IncomeRecord
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.Amount,
			&record.Period,
		); err != nil {
			return []IncomeRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []IncomeRecord{}, ErrIncomeNotFound
	}

	return records, nil
}

func (db SqliteDb) CreateIncomeHistory(config IncomeHistoryRecord) error {
	_, err := db.insertInto(
		INCOME_HISTORY,
		config.IncomeID,
		config.MonthID,
		config.Amount,
	)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryIncomeHistory(qm QueryMap) ([]IncomeHistoryRecord, error) {
	rows, err := db.query(INCOME_HISTORY, qm)
	if err != nil {
		return []IncomeHistoryRecord{}, err
	}

	var records []IncomeHistoryRecord
	for rows.Next() {
		var record IncomeHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.IncomeID,
			&record.MonthID,
			&record.Amount,
		); err != nil {
			return []IncomeHistoryRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []IncomeHistoryRecord{}, ErrIncomeHistoryNotFound
	}

	return records, nil
}

/*
AffixIncome tracks an appended amount to an existing income. This could
be a bonus or overtime amount.
*/
func (db SqliteDb) AffixIncome(historyID int, name string, amount internal.Currency) error {
	_, err := db.insertInto(INCOME_AFFIXES, historyID, name, amount.GetStoredValue())
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryAffixIncome(qm QueryMap) ([]AffixIncomeRecord, error) {
	rows, err := db.query(INCOME_AFFIXES, qm)
	if err != nil {
		return []AffixIncomeRecord{}, err
	}

	var records []AffixIncomeRecord
	for rows.Next() {
		var record AffixIncomeRecord
		if err := rows.Scan(
			&record.ID,
			&record.IncomeHistoryID,
			&record.Name,
			&record.Amount,
		); err != nil {
			return []AffixIncomeRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []AffixIncomeRecord{}, ErrAffixedIncomeNotFound
	}

	return records, nil
}
