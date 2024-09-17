package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type IncomeRow struct {
	ID     int
	Name   string
	Amount lib.Currency
	Period Period
}

type IncomeHistoryRow struct {
	ID       int
	IncomeID int
	MonthID  int
	Amount   lib.Currency
}

func (ih IncomeHistoryRow) String() string {
	s := fmt.Sprintf(
		"id: %d\nincomeID: %d\nmonthID: %d\namount: %s",
		ih.ID,
		ih.IncomeID,
		ih.MonthID,
		ih.Amount.String(),
	)
	return s
}

type AffixIncomeRow struct {
	ID              int
	IncomeHistoryID int
	Name            string
	Amount          lib.Currency
}

func (sdb SqliteDb) CreateIncome(name string, amount lib.Currency, p Period) int64 {
	res, err := sdb.handle.Exec(sdb.InsertInto(INCOME, name, amount.ToInt(), p))
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
		fmt.Sprintf("UPDATE income SET amount=%d WHERE id=%d", amount.ToInt(), id),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryIncome(qm QueryMap) ([]IncomeRow, error) {
	rows := sdb.query(INCOME, qm)
	var amount int
	var incomeRow []IncomeRow

	for rows.Next() {
		var row IncomeRow
		if err := rows.Scan(&row.ID, &row.Name, &amount, &row.Period); err != nil {
			panic(err)
		}
		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		incomeRow = append(incomeRow, row)
	}

	if len(incomeRow) == 0 {
		return []IncomeRow{}, fmt.Errorf("table is empty")
	}

	return incomeRow, nil
}

func (sdb SqliteDb) CreateIncomeHistory(incomeID int, monthID int, amount lib.Currency) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_HISTORY, incomeID, monthID, amount.ToInt()),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryIncomeHistory(qm QueryMap) ([]IncomeHistoryRow, error) {
	rows := sdb.query(INCOME_HISTORY, qm)
	var amount int
	var incomeHistoryRows []IncomeHistoryRow

	for rows.Next() {
		var row IncomeHistoryRow
		if err := rows.Scan(&row.ID, &row.IncomeID, &row.MonthID, &amount); err != nil {
			panic(err)
		}
		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		incomeHistoryRows = append(incomeHistoryRows, row)
	}

	if len(incomeHistoryRows) == 0 {
		return []IncomeHistoryRow{}, fmt.Errorf("query returned no results")
	}

	return incomeHistoryRows, nil
}

/*
AffixIncome tracks an appended amount to an existing income. This could
be a bonus or overtime amount.
*/
func (sdb SqliteDb) AffixIncome(historyID int, name string, amount lib.Currency) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_AFFIXES, historyID, name, amount.ToInt()),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryAffixIncome(qm QueryMap) ([]AffixIncomeRow, error) {
	rows := sdb.query(INCOME_AFFIXES, qm)
	var amount int
	var affixRows []AffixIncomeRow

	for rows.Next() {
		var row AffixIncomeRow
		if err := rows.Scan(&row.ID, &row.IncomeHistoryID, &row.Name, &amount); err != nil {
			panic(err)
		}
		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		affixRows = append(affixRows, row)
	}

	if len(affixRows) == 0 {
		return []AffixIncomeRow{}, fmt.Errorf("no results from query")
	}

	return affixRows, nil
}
