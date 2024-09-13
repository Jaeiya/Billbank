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

func (sdb SqliteDb) CreateIncome(name string, amount lib.Currency, p Period) (int64, error) {
	res, err := sdb.handle.Exec(sdb.InsertInto(INCOME, name, amount.ToInt(), p))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sdb SqliteDb) SetIncome(id int, amount lib.Currency) error {
	_, err := sdb.handle.Exec(
		fmt.Sprintf("UPDATE income SET amount=%d WHERE id=%d", amount.ToInt(), id),
	)
	return err
}

/*
AffixIncome tracks an appended amount to an existing income. This could
be a bonus or overtime amount.

🟡The id is an income_history_id, not an income_id
*/
func (sdb SqliteDb) AffixIncome(id int, name string, amount lib.Currency) error {
	_, err := sdb.handle.Exec(sdb.InsertInto(INCOME_AFFIXES, id, amount.ToInt()))
	return err
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
		err = rows.Scan(&id, &name, &amount, &period)
		if err != nil {
			return []IncomeRow{}, err
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		incomeRow = append(incomeRow, IncomeRow{id, name, c.LoadAmount(amount), Period(period)})
	}

	if len(incomeRow) == 0 {
		return []IncomeRow{}, fmt.Errorf("income table is empty")
	}

	return incomeRow, nil
}

func (sdb SqliteDb) QueryIncome(incomeID int) (IncomeRow, error) {
	queryStr := createQueryStr(INCOME, FieldMap{"id": incomeID})
	row := sdb.handle.QueryRow(queryStr)

	var (
		id     int
		name   string
		amount int
		period Period
	)

	err := row.Scan(&id, &name, &amount, &period)
	if err != nil {
		return IncomeRow{}, err
	}
	c := lib.NewCurrency("", sdb.currencyCode)
	c.LoadAmount(amount)
	return IncomeRow{id, name, c, period}, nil
}

func (sdb SqliteDb) CreateIncomeHistory(incomeID int, monthID int) error {
	info, err := sdb.QueryIncome(incomeID)
	_, err = sdb.handle.Exec(
		sdb.InsertInto(INCOME_HISTORY, incomeID, monthID, info.Amount.ToInt()),
	)
	if err != nil {
		return err
	}
	return nil
}

func (sdb SqliteDb) QueryIncomeHistory(qw QueryMap) ([]IncomeHistoryRow, error) {
	fieldMap := buildFieldMap(BY_ID|BY_INCOME_ID|BY_MONTH_ID, qw)
	queryStr := createQueryStr(INCOME_HISTORY, fieldMap)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		return []IncomeHistoryRow{}, err
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
		err = rows.Scan(&id, &incomeID, &monthID, &amount)
		if err != nil {
			return []IncomeHistoryRow{}, err
		}
		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(amount)

		incomeHistoryRows = append(
			incomeHistoryRows,
			IncomeHistoryRow{id, incomeID, monthID, c},
		)
	}

	if len(incomeHistoryRows) == 0 {
		return []IncomeHistoryRow{}, fmt.Errorf("could not find any data related to query")
	}

	return incomeHistoryRows, nil
}
