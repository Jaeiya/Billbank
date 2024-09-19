package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type TransferType string

const (
	WITHDRAWAL = TransferType("withdrawal")
	DEPOSIT    = TransferType("deposit")
	MOVE       = TransferType("move")
)

type TransferConfig struct {
	AccountID    int
	MonthID      int
	Name         string
	Amount       lib.Currency
	Date         string
	TransferType TransferType

	ToWhom   *string
	FromWhom *string
}

type TransferRecord struct {
	TransferConfig
	ID int
}

func (sdb SqliteDb) CreateTransfer(td TransferConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			TRANSFERS,
			td.AccountID,
			td.MonthID,
			td.Name,
			td.Amount.GetStoredValue(),
			td.Date,
			td.TransferType,
			lib.TryDeref(td.ToWhom),
			lib.TryDeref(td.FromWhom),
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryTransfers(qm QueryMap) ([]TransferRecord, error) {
	rows := sdb.query(TRANSFERS, qm)
	var amount int
	var records []TransferRecord

	for rows.Next() {
		var record TransferRecord
		if err := rows.Scan(
			&record.ID,
			&record.AccountID,
			&record.MonthID,
			&record.Name,
			&amount,
			&record.Date,
			&record.TransferType,
			&record.ToWhom,
			&record.FromWhom,
		); err != nil {
			panic(err)
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return records, fmt.Errorf("no data found")
	}

	return records, nil
}
