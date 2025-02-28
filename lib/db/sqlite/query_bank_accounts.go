package sqlite

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
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
	HistoryID    int
	MonthID      int
	Name         string
	Amount       lib.Currency
	DueDay       int
	TransferType TransferType

	ToWhom   *string
	FromWhom *string
}

type TransferRecord struct {
	TransferConfig
	ID int
}

func (sdb SqliteDb) CreateBankAccount(config BankAccountConfig) error {
	encAccountNum, err := lib.EncryptNonNil(config.AccountNumber, config.Password)
	if err != nil {
		return err
	}

	encNotes, err := lib.EncryptNonNil(config.Notes, config.Password)
	if err != nil {
		return err
	}

	_, err = sdb.handle.Exec(
		sdb.ToInsertIntoStr(BANK_ACCOUNTS, config.Name, encAccountNum, encNotes),
	)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBankAccounts(qm QueryMap, password *string) ([]BankRecord, error) {
	rows, err := sdb.query(BANK_ACCOUNTS, qm)
	if err != nil {
		return []BankRecord{}, err
	}

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
			return []BankRecord{}, err
		}

		if password != nil && record.AccountNumber != nil {
			if record.AccountNumber, err = lib.DecryptNonNil(record.AccountNumber, *password); err != nil {
				return []BankRecord{}, err
			}
		}

		if password != nil && record.Notes != nil {
			if record.Notes, err = lib.DecryptNonNil(record.Notes, *password); err != nil {
				return []BankRecord{}, err
			}
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankRecord{}, fmt.Errorf("no bank accounts")
	}

	return records, nil
}

func (sdb SqliteDb) CreateBankAccountHistory(config BankHistoryConfig) error {
	_, err := sdb.handle.Exec(
		sdb.ToInsertIntoStr(
			BANK_ACCOUNT_HISTORY,
			config.BankAccountID,
			config.MonthID,
			config.Balance.GetStoredValue(),
		),
	)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBankAccountHistory(qm QueryMap) ([]BankHistoryRecord, error) {
	rows, err := sdb.query(BANK_ACCOUNT_HISTORY, qm)
	if err != nil {
		return []BankHistoryRecord{}, err
	}

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
			return []BankHistoryRecord{}, err
		}

		record.Balance = lib.NewCurrencyFromStore(balance, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankHistoryRecord{}, fmt.Errorf("no query results found")
	}

	return records, nil
}

func (sdb SqliteDb) CreateTransfer(td TransferConfig) error {
	_, err := sdb.handle.Exec(sdb.ToInsertIntoStr(
		TRANSFERS,
		td.HistoryID,
		td.MonthID,
		td.Name,
		td.Amount.GetStoredValue(),
		td.DueDay,
		td.TransferType,
		utils.TryDeref(td.ToWhom),
		utils.TryDeref(td.FromWhom),
	))
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryTransfers(qm QueryMap) ([]TransferRecord, error) {
	rows, err := sdb.query(TRANSFERS, qm)
	if err != nil {
		return []TransferRecord{}, err
	}

	var amount int
	var records []TransferRecord
	for rows.Next() {
		var record TransferRecord
		if err := rows.Scan(
			&record.ID,
			&record.HistoryID,
			&record.MonthID,
			&record.Name,
			&amount,
			&record.DueDay,
			&record.TransferType,
			&record.ToWhom,
			&record.FromWhom,
		); err != nil {
			return []TransferRecord{}, err
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return records, ErrTransfersNotFound
	}

	return records, nil
}
