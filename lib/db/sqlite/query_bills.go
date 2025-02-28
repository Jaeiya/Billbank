package sqlite

import (
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
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

func (sdb SqliteDb) CreateNewBill(cfg BillsConfig) error {
	if _, err := sdb.handle.Exec(
		sdb.ToInsertIntoStr(BILLS, cfg.Name, cfg.Amount.GetStoredValue(), cfg.DueDay, cfg.Period),
	); err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBills(qm QueryMap) ([]BillRecord, error) {
	rows, err := sdb.query(BILLS, qm)
	if err != nil {
		return []BillRecord{}, err
	}

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
			return []BillRecord{}, err
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BillRecord{}, ErrBillsNotFound
	}

	return records, nil
}

func (sdb SqliteDb) CreateBillHistory(cfg BillHistoryConfig) error {
	paidAmount := utils.TryDeref(cfg.PaidAmount)
	if paidAmount != nil {
		paidAmount = cfg.PaidAmount.GetStoredValue()
	}

	if _, err := sdb.handle.Exec(
		sdb.ToInsertIntoStr(
			BILL_HISTORY,
			cfg.BillID,
			cfg.MonthID,
			cfg.Amount.GetStoredValue(),
			paidAmount,
			utils.TryDeref(cfg.PaidDate),
			cfg.DueDay,
			utils.TryDeref(cfg.Notes),
		),
	); err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRecord, error) {
	rows, err := sdb.query(BILL_HISTORY, qm)
	if err != nil {
		return []BillHistoryRecord{}, err
	}

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
			return []BillHistoryRecord{}, err
		}

		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)

		if paidAmount != nil {
			pa := lib.NewCurrencyFromStore(*paidAmount, sdb.currencyCode)
			record.PaidAmount = &pa
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, ErrBillHistoryNotFound
	}

	return records, nil
}
