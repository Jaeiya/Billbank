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

type DecryptedBankInfo struct {
	ID            int
	Name          string
	AccountNumber *string
	Notes         *string
}

type BankAccountHistoryConfig struct {
	MonthID       int
	BankAccountID int
	Balance       lib.Currency
}

func (sdb SqliteDb) CreateBankAccount(config BankAccountConfig) {
	if config.Password == nil || (config.AccountNumber == nil && config.Notes == nil) {
		sdb.handle.Exec(sdb.InsertInto(BANK_ACCOUNTS, config.Name, nil, nil))
	}

	encrypt := func(data, password *string) any /* nil|string */ {
		if data == nil {
			return nil
		}
		return lib.EncryptData(*data, *password)
	}

	_, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNTS,
			config.Name,
			encrypt(config.AccountNumber, config.Password),
			encrypt(config.Notes, config.Password),
		),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryAllBankAccounts() ([]BankInfoRow, error) {
	queryStr := "SELECT id, name FROM bank_accounts"
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var (
		id           int
		name         string
		bankInfoRows []BankInfoRow
	)

	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			panic(err)
		}
		bankInfoRows = append(bankInfoRows, BankInfoRow{id, name})
	}

	if len(bankInfoRows) == 0 {
		return []BankInfoRow{}, fmt.Errorf("no bank accounts")
	}

	return bankInfoRows, nil
}

func (sdb SqliteDb) QueryDecryptedBankAccount(
	accountID int,
	password string,
) (DecryptedBankInfo, error) {
	queryStr := buildQueryStr(BANK_ACCOUNTS, FieldMap{"id": accountID})
	row := sdb.handle.QueryRow(queryStr)

	var (
		id               int
		name             string
		encryptedAcctNum *string
		encryptedNotes   *string
		decryptedAcctNum string
		decryptedNotes   string
	)

	err := row.Scan(&id, &name, &encryptedAcctNum, &encryptedNotes)
	if err != nil {
		panic(err)
	}

	if encryptedAcctNum != nil {
		decryptedAcctNum, err = lib.DecryptData(*encryptedAcctNum, password)
		if err != nil {
			return DecryptedBankInfo{}, err
		}
	}

	if encryptedNotes != nil {
		decryptedNotes, err = lib.DecryptData(*encryptedNotes, password)
		if err != nil {
			return DecryptedBankInfo{}, err
		}

	}

	return DecryptedBankInfo{id, name, &decryptedAcctNum, &decryptedNotes}, nil
}

func (sdb SqliteDb) CreateBankAccountHistory(config BankAccountHistoryConfig) {
	_, err := sdb.handle.Exec(
		sdb.InsertInto(
			BANK_ACCOUNT_HISTORY,
			config.BankAccountID,
			config.MonthID,
			config.Balance.ToInt(),
		),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryBankAccountHistory(qm QueryMap) ([]BankHistoryRow, error) {
	fm := buildFieldMap(BY_ID|BY_BANK_ACCOUNT_ID|BY_MONTH_ID, qm)
	queryStr := buildQueryStr(BANK_ACCOUNT_HISTORY, fm)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var (
		id              int
		bankAcctID      int
		monthID         int
		balance         int
		bankHistoryRows []BankHistoryRow
	)

	for rows.Next() {
		err := rows.Scan(&id, &bankAcctID, &monthID, &balance)
		if err != nil {
			panic(err)
		}

		c := lib.NewCurrency("", sdb.currencyCode)
		c.LoadAmount(balance)
		bankHistoryRows = append(bankHistoryRows, BankHistoryRow{id, bankAcctID, monthID, c})
	}

	if len(bankHistoryRows) == 0 {
		return []BankHistoryRow{}, fmt.Errorf("no query results found")
	}

	return bankHistoryRows, nil
}
