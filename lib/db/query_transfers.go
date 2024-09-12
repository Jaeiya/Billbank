package db

import "github.com/jaeiya/billbank/lib"

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

func (sdb SqliteDb) CreateTransfer(td TransferConfig) error {
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
		return err
	}

	return nil
}

func (sdb SqliteDb) QueryTransfer(qwm QueryWhere) TransferData {
	wMap := WhereMap{}
	for k, v := range qwm {
		switch k {

		case BY_ID:
			wMap["id"] = v

		case BY_MONTH_ID:
			wMap["month_id"] = v

		case BY_ACCOUNT_ID:
			wMap["bank_account_id"] = v

		}
	}
	return sdb.queryTransfers(wMap)
}

func (sdb SqliteDb) queryTransfers(wm WhereMap) TransferData {
	queryStr := createQueryStr(TRANSFERS, wm)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if !rows.Next() {
		panic("transfers table is empty")
	}

	var (
		id        int
		accountID int
		monthID   int
		name      string
		amount    int
		date      string
		tt        TransferType
		toWhom    *string
		fromWhom  *string
	)

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

	d := TransferData{
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
	}
	return d
}
