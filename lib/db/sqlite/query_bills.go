package sqlite

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

type BillRecord struct {
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

type BillHistoryRecord struct {
	ID int
	BillHistoryConfig
}

func (sdb SqliteDb) CreateNewBill(cfg BillsConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(BILLS, cfg.Name, cfg.Amount.GetStoredValue(), cfg.DueDay, cfg.Period),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryBills(qm QueryMap) ([]BillRecord, error) {
	rows := sdb.query(BILLS, qm)
	var amount int
	var records []BillRecord

	for rows.Next() {
		var record BillRecord
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&amount,
			&record.DueDay,
			&record.Period,
		); err != nil {
			panic(err)
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BillRecord{}, fmt.Errorf("no query results")
	}

	return records, nil
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
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRecord, error) {
	rows := sdb.query(BILL_HISTORY, qm)

	var amount int
	var paidAmount *int
	var records []BillHistoryRecord

	for rows.Next() {
		var record BillHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.BillID,
			&record.MonthID,
			&amount,
			&paidAmount,
			&record.PaidDate,
			&record.DueDay,
			&record.Notes,
		); err != nil {
			panic(err)
		}

		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)

		if paidAmount != nil {
			pa := lib.NewCurrencyFromStore(*paidAmount, sdb.currencyCode)
			record.PaidAmount = &pa
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no results from query")
	}

	return records, nil
}
