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

type BankAccountRecord struct {
	ID            int
	Name          string
	AccountNumber *string
	Notes         *string
}

type BankHistoryRecord struct {
	ID            int
	MonthID       int
	BankAccountID int
	Balance       lib.Currency
}

type TransferRecord struct {
	ID           int
	HistoryID    int
	MonthID      int
	Name         string
	Amount       lib.Currency
	DueDay       int
	TransferType TransferType

	ToWhom   *string
	FromWhom *string
}

func (sdb SqliteDb) CreateBankAccount(config BankAccountRecord, password *string) error {
	encAccountNum, err := lib.EncryptNonNil(config.AccountNumber, password)
	if err != nil {
		return err
	}

	encNotes, err := lib.EncryptNonNil(config.Notes, password)
	if err != nil {
		return err
	}

	_, err = sdb.insertInto(BANK_ACCOUNTS, config.Name, encAccountNum, encNotes)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBankAccounts(qm QueryMap, pass *string) ([]BankAccountRecord, error) {
	rows, err := sdb.query(BANK_ACCOUNTS, qm)
	if err != nil {
		return []BankAccountRecord{}, err
	}

	var records []BankAccountRecord
	for rows.Next() {
		var record BankAccountRecord
		var err error

		if err = rows.Scan(
			&record.ID,
			&record.Name,
			&record.AccountNumber,
			&record.Notes,
		); err != nil {
			return []BankAccountRecord{}, err
		}

		if pass != nil && record.AccountNumber != nil {
			if record.AccountNumber, err = lib.DecryptNonNil(record.AccountNumber, *pass); err != nil {
				return []BankAccountRecord{}, err
			}
		}

		if pass != nil && record.Notes != nil {
			if record.Notes, err = lib.DecryptNonNil(record.Notes, *pass); err != nil {
				return []BankAccountRecord{}, err
			}
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankAccountRecord{}, fmt.Errorf("no bank accounts")
	}

	return records, nil
}

func (sdb SqliteDb) CreateBankAccountHistory(config BankHistoryRecord) error {
	_, err := sdb.insertInto(
		BANK_ACCOUNT_HISTORY,
		config.BankAccountID,
		config.MonthID,
		config.Balance,
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

	var records []BankHistoryRecord
	for rows.Next() {
		var record BankHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.BankAccountID,
			&record.MonthID,
			&record.Balance,
		); err != nil {
			return []BankHistoryRecord{}, err
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []BankHistoryRecord{}, fmt.Errorf("no query results found")
	}

	return records, nil
}

func (sdb SqliteDb) CreateTransfer(tr TransferRecord) error {
	_, err := sdb.insertInto(
		TRANSFERS,
		tr.HistoryID,
		tr.MonthID,
		tr.Name,
		tr.Amount,
		tr.DueDay,
		tr.TransferType,
		utils.TryDeref(tr.ToWhom),
		utils.TryDeref(tr.FromWhom),
	)
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

	var records []TransferRecord
	for rows.Next() {
		var record TransferRecord
		if err := rows.Scan(
			&record.ID,
			&record.HistoryID,
			&record.MonthID,
			&record.Name,
			&record.Amount,
			&record.DueDay,
			&record.TransferType,
			&record.ToWhom,
			&record.FromWhom,
		); err != nil {
			return []TransferRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return records, ErrTransfersNotFound
	}

	return records, nil
}
