package sqlite

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type IncomeConfig struct {
	Name   string
	Amount lib.Currency
	Period Period
}

type IncomeRecord struct {
	ID int
	IncomeConfig
}

type IncomeHistoryConfig struct {
	IncomeID int
	MonthID  int
	Amount   lib.Currency
}

type IncomeHistoryRecord struct {
	ID int
	IncomeHistoryConfig
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
	Amount          lib.Currency
}

func (sdb SqliteDb) CreateIncome(config IncomeConfig) int64 {
	res, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME, config.Name, config.Amount.GetStoredValue(), config.Period),
	)
	if err != nil {
		panicOnExecErr(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	return id
}

func (sdb SqliteDb) SetIncome(id int, amount lib.Currency) {
	_, err := sdb.handle.Exec(
		fmt.Sprintf("UPDATE income SET amount=%d WHERE id=%d", amount.GetStoredValue(), id),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryIncome(qm QueryMap) ([]IncomeRecord, error) {
	rows := sdb.query(INCOME, qm)
	var amount int
	var records []IncomeRecord

	for rows.Next() {
		var record IncomeRecord
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&amount,
			&record.Period,
		); err != nil {
			panic(err)
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []IncomeRecord{}, fmt.Errorf("table is empty")
	}

	return records, nil
}

func (sdb SqliteDb) CreateIncomeHistory(config IncomeHistoryConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_HISTORY, config.IncomeID, config.MonthID, config.Amount.GetStoredValue()),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryIncomeHistory(qm QueryMap) ([]IncomeHistoryRecord, error) {
	rows := sdb.query(INCOME_HISTORY, qm)
	var amount int
	var records []IncomeHistoryRecord

	for rows.Next() {
		var record IncomeHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.IncomeID,
			&record.MonthID,
			&amount,
		); err != nil {
			panic(err)
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []IncomeHistoryRecord{}, fmt.Errorf("query returned no results")
	}

	return records, nil
}

/*
AffixIncome tracks an appended amount to an existing income. This could
be a bonus or overtime amount.
*/
func (sdb SqliteDb) AffixIncome(historyID int, name string, amount lib.Currency) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_AFFIXES, historyID, name, amount.GetStoredValue()),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryAffixIncome(qm QueryMap) ([]AffixIncomeRecord, error) {
	rows := sdb.query(INCOME_AFFIXES, qm)
	var amount int
	var records []AffixIncomeRecord

	for rows.Next() {
		var record AffixIncomeRecord
		if err := rows.Scan(
			&record.ID,
			&record.IncomeHistoryID,
			&record.Name,
			&amount,
		); err != nil {
			panic(err)
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []AffixIncomeRecord{}, fmt.Errorf("no results from query")
	}

	return records, nil
}
