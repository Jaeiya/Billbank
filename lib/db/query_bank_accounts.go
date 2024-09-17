package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type BankAccountConfig struct {
	Name          string
	Password      *string
	AccountNumber *string
	Notes         *string
}

type BankInfoRow struct {
	ID   int
	Name string
}

type BankHistoryRow struct {
	ID            int
	BankAccountID int
	MonthID       int
	Balance       lib.Currency
}

func (bhr BankHistoryRow) String() string {
	return fmt.Sprintf(
		"id: %d\naccID: %d\nmonthID: %d\nbalance: %s",
		bhr.ID,
		bhr.BankAccountID,
		bhr.MonthID,
		bhr.Balance,
	)
}

type DecryptedBankRow struct {
	ID            int
	Name          string
	AccountNumber *string
	Notes         *string
}

func (dbr DecryptedBankRow) String() string {
	return fmt.Sprintf(
		"id: %d\nname: %s\nacc: %s\nnotes: %s",
		dbr.ID,
		dbr.Name,
		*dbr.AccountNumber,
		*dbr.Notes,
	)
}

type BankHistoryConfig struct {
	MonthID       int
	BankAccountID int
	Balance       lib.Currency
}

func (sdb SqliteDb) CreateBankAccount(config BankAccountConfig) {
	if config.Password == nil || (config.AccountNumber == nil && config.Notes == nil) {
		if _, err := sdb.handle.Exec(
			sdb.InsertInto(BANK_ACCOUNTS, config.Name, nil, nil),
		); err != nil {
			panic(err)
		}
	}

	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNTS,
			config.Name,
			lib.EncryptNonNil(config.AccountNumber, config.Password),
			lib.EncryptNonNil(config.Notes, config.Password),
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryAllBankAccounts() ([]BankInfoRow, error) {
	queryStr := "SELECT id, name FROM bank_accounts"
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var bankInfoRows []BankInfoRow

	for rows.Next() {
		var row BankInfoRow
		if err = rows.Scan(&row.ID, &row.Name); err != nil {
			panic(err)
		}
		bankInfoRows = append(bankInfoRows, row)
	}

	if len(bankInfoRows) == 0 {
		return []BankInfoRow{}, fmt.Errorf("no bank accounts")
	}

	return bankInfoRows, nil
}

func (sdb SqliteDb) QueryDecryptedBankAccount(
	accountID int,
	password string,
) ([]DecryptedBankRow, error) {
	rows := sdb.query(BANK_ACCOUNTS, QueryMap{WHERE_ID: accountID})
	var encAcctNum, encNotes *string
	var decryptedBankRows []DecryptedBankRow

	for rows.Next() {
		var row DecryptedBankRow
		err := rows.Scan(&row.ID, &row.Name, &encAcctNum, &encNotes)
		if err != nil {
			panic(err)
		}

		row.AccountNumber, err = lib.DecryptNonNil(encAcctNum, password)
		if err != nil {
			panic(err)
		}

		row.Notes, err = lib.DecryptNonNil(encNotes, password)
		if err != nil {
			panic(err)
		}

		decryptedBankRows = append(decryptedBankRows, row)
	}

	if len(decryptedBankRows) == 0 {
		return []DecryptedBankRow{}, fmt.Errorf("no query results")
	}

	return decryptedBankRows, nil
}

func (sdb SqliteDb) CreateBankAccountHistory(config BankHistoryConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNT_HISTORY,
			config.BankAccountID,
			config.MonthID,
			config.Balance.ToInt(),
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryBankAccountHistory(qm QueryMap) ([]BankHistoryRow, error) {
	rows := sdb.query(BANK_ACCOUNT_HISTORY, qm)
	var balance int
	var bankHistoryRows []BankHistoryRow

	for rows.Next() {
		var row BankHistoryRow
		if err := rows.Scan(&row.ID, &row.BankAccountID, &row.MonthID, &balance); err != nil {
			panic(err)
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(balance)
		row.Balance = c
		bankHistoryRows = append(bankHistoryRows, row)
	}

	if len(bankHistoryRows) == 0 {
		return []BankHistoryRow{}, fmt.Errorf("no query results found")
	}

	return bankHistoryRows, nil
}
