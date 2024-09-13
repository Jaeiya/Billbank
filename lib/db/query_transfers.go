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
	tryDereference := func(data *string) any /* nil|string */ {
		if data == nil {
			return nil
		}
		return *data
	}
	_, err := sdb.handle.Exec(
		sdb.InsertInto(
			TRANSFERS,
			td.AccountID,
			td.MonthID,
			td.Name,
			td.Amount.ToInt(),
			td.Date,
			td.TransferType,
			tryDereference(td.ToWhom),
			tryDereference(td.FromWhom),
		),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryTransfers(qm QueryMap) ([]TransferData, error) {
	fieldMap := buildFieldMap(BY_ID|BY_MONTH_ID|BY_BANK_ACCOUNT_ID, qm)
	queryStr := buildQueryStr(TRANSFERS, fieldMap)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var (
		id           int
		accountID    int
		monthID      int
		name         string
		amount       int
		date         string
		tt           TransferType
		toWhom       *string
		fromWhom     *string
		transferRows []TransferData
	)

	for rows.Next() {
		err = rows.Scan(
			&id,
			&accountID,
			&monthID,
			&name,
			&amount,
			&date,
			&tt,
			&toWhom,
			&fromWhom,
		)
		if err != nil {
			panic(err)
		}
		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(amount)

		transferRows = append(transferRows, TransferData{
			ID: id,
			TransferConfig: TransferConfig{
				AccountID:    accountID,
				MonthID:      monthID,
				Name:         name,
				Amount:       c,
				Date:         date,
				TransferType: tt,
				ToWhom:       toWhom,
				FromWhom:     fromWhom,
			},
		})

	}

	if len(transferRows) == 0 {
		return transferRows, fmt.Errorf("no data found")
	}

	return transferRows, nil
}
