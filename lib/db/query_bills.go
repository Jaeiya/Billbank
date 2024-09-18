package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type BillsConfig struct {
	Name   string
	Amount lib.Currency
	DueDay int
	Period Period
}

type BillRow struct {
	ID int
	BillsConfig
}

type BillHistoryConfig struct {
	BillID     int
	MonthID    int
	Amount     lib.Currency
	DueDay     int
	PaidAmount *lib.Currency
	PaidDate   *string
	Notes      *string
}

type BillHistoryRow struct {
	ID int
	BillHistoryConfig
}

func (sdb SqliteDb) CreateNewBill(cfg BillsConfig) {
	_, err := sdb.handle.Exec(
		sdb.InsertInto(BILLS, cfg.Name, cfg.Amount.GetStoredValue(), cfg.DueDay, cfg.Period),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryBills(qm QueryMap) ([]BillRow, error) {
	rows := sdb.query(BILLS, qm)
	var amount int
	var serializedRows []BillRow

	for rows.Next() {
		var row BillRow
		err := rows.Scan(&row.ID, &row.Name, &amount, &row.DueDay, &row.Period)
		if err != nil {
			panic(err)
		}
		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		serializedRows = append(serializedRows, row)
	}

	if len(serializedRows) == 0 {
		return []BillRow{}, fmt.Errorf("no query results")
	}

	return serializedRows, nil
}

func (sdb SqliteDb) CreateBillHistory(cfg BillHistoryConfig) {
	paidAmount := lib.TryDeref(cfg.PaidAmount)
	if paidAmount != nil {
		paidAmount = cfg.PaidAmount.GetStoredValue()
	}

	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			BILL_HISTORY,
			cfg.BillID,
			cfg.MonthID,
			cfg.Amount.GetStoredValue(),
			paidAmount,
			lib.TryDeref(cfg.PaidDate),
			cfg.DueDay,
			lib.TryDeref(cfg.Notes),
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRow, error) {
	rows := sdb.query(BILL_HISTORY, qm)

	var amount int
	var paidAmount *int
	var serializedRows []BillHistoryRow

	for rows.Next() {
		var row BillHistoryRow
		err := rows.Scan(
			&row.ID,
			&row.BillID,
			&row.MonthID,
			&amount,
			&paidAmount,
			&row.PaidDate,
			&row.DueDay,
			&row.Notes,
		)
		if err != nil {
			panic(err)
		}

		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)

		if paidAmount != nil {
			pa := lib.NewCurrencyFromStore(*paidAmount, sdb.currencyCode)
			row.PaidAmount = &pa
		}

		serializedRows = append(serializedRows, row)
	}

	if len(serializedRows) == 0 {
		return nil, fmt.Errorf("no results from query")
	}

	return serializedRows, nil
}
