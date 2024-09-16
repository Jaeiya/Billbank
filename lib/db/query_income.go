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

func (sdb SqliteDb) QueryAllIncome() ([]IncomeRow, error) {
	rows, err := sdb.handle.Query("SELECT * FROM income")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id        int
		name      string
		amount    int
		period    string
		incomeRow []IncomeRow
	)

	for rows.Next() {
		if err = rows.Scan(&id, &name, &amount, &period); err != nil {
			panic(err)
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		incomeRow = append(incomeRow, IncomeRow{id, name, c.LoadAmount(amount), Period(period)})
	}

	if len(incomeRow) == 0 {
		return []IncomeRow{}, fmt.Errorf("table is empty")
	}

	return incomeRow, nil
}

func (sdb SqliteDb) QueryIncome(incomeID int) IncomeRow {
	queryStr := buildQueryStr(INCOME, FieldMap{"id": incomeID})
	row := sdb.handle.QueryRow(queryStr)

	var (
		id     int
		name   string
		amount int
		period Period
	)

	if err := row.Scan(&id, &name, &amount, &period); err != nil {
		panic(err)
	}
	c := lib.NewCurrency("", sdb.currencyCode)
	c.LoadAmount(amount)
	return IncomeRow{id, name, c, period}
}

func (sdb SqliteDb) CreateIncomeHistory(incomeID int, monthID int) {
	incomeRow := sdb.QueryIncome(incomeID)
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(INCOME_HISTORY, incomeID, monthID, incomeRow.Amount.ToInt()),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryIncomeHistory(qw QueryMap) ([]IncomeHistoryRow, error) {
	fieldMap := buildFieldMap(BY_ID|BY_INCOME_ID|BY_MONTH_ID, qw)
	queryStr := buildQueryStr(INCOME_HISTORY, fieldMap)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id                int
		incomeID          int
		monthID           int
		amount            int
		incomeHistoryRows []IncomeHistoryRow
	)

	for rows.Next() {
		if err = rows.Scan(&id, &incomeID, &monthID, &amount); err != nil {
			panic(err)
		}
		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(amount)

		incomeHistoryRows = append(
			incomeHistoryRows,
			IncomeHistoryRow{id, incomeID, monthID, c},
		)
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
	fieldMap := buildFieldMap(BY_ID|BY_INCOME_ID, qm)
	queryStr := buildQueryStr(INCOME_AFFIXES, fieldMap)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var (
		id           int
		incomeHistID int
		name         string
		amount       int
		affixRows    []AffixIncomeRow
	)

	for rows.Next() {
		if err = rows.Scan(&id, &incomeHistID, &name, &amount); err != nil {
			panic(err)
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(amount)

		affixRows = append(affixRows, AffixIncomeRow{
			id, incomeHistID, name, c,
		})
	}

	if len(affixRows) == 0 {
		return []AffixIncomeRow{}, fmt.Errorf("no results from query")
	}

	return affixRows, nil
}
