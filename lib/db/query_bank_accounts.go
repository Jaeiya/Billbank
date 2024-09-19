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

type BankRecord struct {
	ID            int
	Name          string
	AccountNumber *string
	Notes         *string
}

func (dbr BankRecord) String() string {
	return fmt.Sprintf(
		"id: %d\nname: %s\nacc: %s\nnotes: %s",
		dbr.ID,
		dbr.Name,
		*dbr.AccountNumber,
		*dbr.Notes,
	)
}

type BankHistoryRecord struct {
	ID            int
	BankAccountID int
	MonthID       int
	Balance       lib.Currency
}

func (bhr BankHistoryRecord) String() string {
	return fmt.Sprintf(
		"id: %d\naccID: %d\nmonthID: %d\nbalance: %s",
		bhr.ID,
		bhr.BankAccountID,
		bhr.MonthID,
		bhr.Balance,
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

func (sdb SqliteDb) QueryBankAccounts(qm QueryMap) ([]BankRecord, error) {
	rows := sdb.query(BANK_ACCOUNTS, qm)
	var records []BankRecord

	for rows.Next() {
		var record BankRecord
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.AccountNumber,
			&record.Notes,
		); err != nil {
			panic(err)
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankRecord{}, fmt.Errorf("no bank accounts")
	}

	return records, nil
}

func (sdb SqliteDb) QueryDecryptedBankAccount(
	accountID int,
	password string,
) ([]BankRecord, error) {
	rows := sdb.query(BANK_ACCOUNTS, QueryMap{WHERE_ID: accountID})
	var encAcctNum, encNotes *string
	var records []BankRecord

	for rows.Next() {
		var record BankRecord
		var err error

		if err = rows.Scan(
			&record.ID,
			&record.Name,
			&encAcctNum,
			&encNotes,
		); err != nil {
			panic(err)
		}

		if record.AccountNumber, err = lib.DecryptNonNil(encAcctNum, password); err != nil {
			panic(err)
		}

		if record.Notes, err = lib.DecryptNonNil(encNotes, password); err != nil {
			panic(err)
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankRecord{}, fmt.Errorf("no query results")
	}

	return records, nil
}

func (sdb SqliteDb) CreateBankAccountHistory(config BankHistoryConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNT_HISTORY,
			config.BankAccountID,
			config.MonthID,
			config.Balance.GetStoredValue(),
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryBankAccountHistory(qm QueryMap) ([]BankHistoryRecord, error) {
	rows := sdb.query(BANK_ACCOUNT_HISTORY, qm)
	var balance int
	var records []BankHistoryRecord

	for rows.Next() {
		var record BankHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.BankAccountID,
			&record.MonthID,
			&balance,
		); err != nil {
			panic(err)
		}

		record.Balance = lib.NewCurrencyFromStore(balance, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankHistoryRecord{}, fmt.Errorf("no query results found")
	}

	return records, nil
}
