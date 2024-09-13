package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type IncomeHistory struct {
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

func (sdb SqliteDb) QueryAllIncome() ([]IncomeInfo, error) {
	if sdb.getRowCount(INCOME) == 0 {
		return []IncomeInfo{}, fmt.Errorf("income table is empty")
	}

	rows, err := sdb.handle.Query("SELECT * FROM income")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id     int
		name   string
		amount int
		period string
		info   []IncomeInfo
	)

	for rows.Next() {
		err = rows.Scan(&id, &name, &amount, &period)
		if err != nil {
			return []IncomeInfo{}, err
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		info = append(info, IncomeInfo{id, name, c.LoadAmount(amount), Period(period)})
	}

	return info, nil
}

func (sdb SqliteDb) QueryIncome(incomeID int) (IncomeInfo, error) {
	queryStr := createQueryStr(INCOME, WhereMap{"id": incomeID})
	row := sdb.handle.QueryRow(queryStr)

	var (
		id     int
		name   string
		amount int
		period Period
	)

	err := row.Scan(&id, &name, &amount, &period)
	if err != nil {
		return IncomeInfo{}, err
	}
	c := lib.NewCurrency("", sdb.currencyCode)
	c.LoadAmount(amount)
	return IncomeInfo{id, name, c, period}, nil
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

func (sdb SqliteDb) QueryIncomeHistory(qw QueryWhereMap) (IncomeHistory, error) {
	wMap := WhereMap{}
	for k, v := range qw {
		switch k {

		case BY_ID:
			wMap["id"] = v

		case BY_INCOME_ID:
			wMap["income_id"] = v

		case BY_MONTH_ID:
			wMap["month_id"] = v

		default:
			panic("unsupported 'where' filter")

		}
	}

	queryStr := createQueryStr(INCOME_HISTORY, wMap)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		return IncomeHistory{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return IncomeHistory{}, fmt.Errorf("could not find any data related to query")
	}

	var (
		id       int
		incomeID int
		monthID  int
		amount   int
	)

	err = rows.Scan(&id, &incomeID, &monthID, &amount)
	if err != nil {
		return IncomeHistory{}, err
	}
	c := lib.NewCurrency("", sdb.currencyCode)
	c.LoadAmount(amount)
	return IncomeHistory{id, incomeID, monthID, c}, nil
}
