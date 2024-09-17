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

type TransferData struct {
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

func (sdb SqliteDb) QueryTransfers(qm QueryMap) ([]TransferData, error) {
	rows := sdb.query(TRANSFERS, qm)
	var amount int
	var transferRows []TransferData

	for rows.Next() {
		var row TransferData
		if err := rows.Scan(
			&row.ID,
			&row.AccountID,
			&row.MonthID,
			&row.Name,
			&amount,
			&row.Date,
			&row.TransferType,
			&row.ToWhom,
			&row.FromWhom,
		); err != nil {
			panic(err)
		}
		row.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		transferRows = append(transferRows, row)
	}

	if len(transferRows) == 0 {
		return transferRows, fmt.Errorf("no data found")
	}

	return transferRows, nil
}
