package sqlite

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

type BankHistoryRecord struct {
	ID            int
	BankAccountID int
	MonthID       int
	Balance       lib.Currency
}

type BankHistoryConfig struct {
	MonthID       int
	BankAccountID int
	Balance       lib.Currency
}

type TransferConfig struct {
	BankAccountID int
	MonthID       int
	Name          string
	Amount        lib.Currency
	Date          string
	TransferType  TransferType

	ToWhom   *string
	FromWhom *string
}

type TransferRecord struct {
	TransferConfig
	ID int
}

func (sdb SqliteDb) CreateBankAccount(config BankAccountConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNTS,
			config.Name,
			lib.EncryptNonNil(config.AccountNumber, config.Password),
			lib.EncryptNonNil(config.Notes, config.Password),
		),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryBankAccounts(qm QueryMap, password *string) ([]BankRecord, error) {
	rows := sdb.query(BANK_ACCOUNTS, qm)
	var records []BankRecord

	for rows.Next() {
		var record BankRecord
		var err error

		if err = rows.Scan(
			&record.ID,
			&record.Name,
			&record.AccountNumber,
			&record.Notes,
		); err != nil {
			panic(err)
		}

		if password != nil && record.AccountNumber != nil {
			if record.AccountNumber, err = lib.DecryptNonNil(record.AccountNumber, *password); err != nil {
				panic(err)
			}
		}

		if password != nil && record.Notes != nil {
			if record.Notes, err = lib.DecryptNonNil(record.Notes, *password); err != nil {
				panic(err)
			}
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankRecord{}, fmt.Errorf("no bank accounts")
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
		panicOnExecErr(err)
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

func (sdb SqliteDb) CreateTransfer(td TransferConfig) {
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			TRANSFERS,
			td.BankAccountID,
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
			&record.BankAccountID,
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
