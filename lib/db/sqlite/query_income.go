package sqlite

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type IncomeRecord struct {
	ID     int
	Name   string
	Amount lib.Currency
	Period Period
}

type IncomeHistoryRecord struct {
	ID       int
	IncomeID int
	MonthID  int
	Amount   lib.Currency
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

func (sdb SqliteDb) CreateIncome(name string, amount lib.Currency, p Period) int64 {
	res, err := sdb.handle.Exec(sdb.InsertInto(INCOME, name, amount.GetStoredValue(), p))
	if err != nil {
		panic(err)
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

func (sdb SqliteDb) CreateIncomeHistory(incomeID int, monthID int, amount lib.Currency) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_HISTORY, incomeID, monthID, amount.GetStoredValue()),
	); err != nil {
		panic(err)
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
