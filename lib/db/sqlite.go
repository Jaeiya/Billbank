package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/jaeiya/billbank/lib"
	_ "modernc.org/sqlite"
)

type Period string

//go:embed sql/init_db.sqlite
var sqlCreateBank string

const (
	YEARLY   = Period("yearly")
	MONTHLY  = Period("monthly")
	WEEKLY   = Period("weekly")
	BIWEEKLY = Period("biweekly")
)

type Table int

const (
	MONTHS = Table(iota)
	INCOME
	INCOME_HISTORY
	INCOME_AFFIXES
	BANK_ACCOUNTS
	BANK_ACCOUNT_HISTORY
	TRANSFERS
	CREDIT_CARDS
	CREDIT_CARD_HISTORY
)

type TableData = map[Table][]string

var tableData = TableData{
	MONTHS:               {"date"},
	INCOME:               {"name", "amount", "period"},
	INCOME_HISTORY:       {"income_id", "month_id", "amount"},
	INCOME_AFFIXES:       {"income_history_id", "name", "amount"},
	BANK_ACCOUNTS:        {"name", "account_number", "notes"},
	BANK_ACCOUNT_HISTORY: {"bank_account_id", "month_id", "balance"},
	TRANSFERS: {
		"bank_account_history_id",
		"month_id",
		"name",
		"amount",
		"date",
		"kind",
		"to_whom",
		"from_whom",
	},
	CREDIT_CARDS: {"name", "card_number", "last_four_digits"},
	CREDIT_CARD_HISTORY: {
		"credit_card_id",
		"month_id",
		"balance",
		"credit_limit",
		"paid_amount",
		"paid_date",
		"due_date",
		"period",
	},
}

type BankInfo struct {
	Name          string
	AccountNumber string
	Notes         string
}

type IncomeInfo struct {
	Name   string
	Amount lib.Currency
	Period Period
}

type SqliteDb struct {
	handle       *sql.DB
	currencyCode lib.CurrencyCode
}

func NewSqliteDb(name string, cc lib.CurrencyCode) *SqliteDb {
	db, err := sql.Open("sqlite", name+".db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(sqlCreateBank)
	if err != nil {
		panic(err)
	}

	return &SqliteDb{db, cc}
}

func (sdb SqliteDb) CreateMonth(t time.Time) error {
	// Make sure any prior date arithmetic, used a clean date
	isClean := t.Day() == 1 &&
		t.Hour() == 0 &&
		t.Minute() == 0 &&
		t.Second() == 0 &&
		t.Nanosecond() == 0

	if !isClean {
		return fmt.Errorf("not a clean date")
	}
	_, err := sdb.handle.Exec(sdb.InsertInto(MONTHS, t.Format("2006-01")))
	if err != nil {
		return err
	}
	return nil
}

func (sdb SqliteDb) QueryMonth() {
	rows, err := sdb.handle.Query("SELECT * FROM months")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var date string
		err := rows.Scan(&id, &date)
		if err != nil {
			panic(err)
		}
		fmt.Println(id, date)
	}
}

func (sdb SqliteDb) CreateBankAccount(
	name string,
	password, accountNumber, notes *string,
) error {
	if password == nil || (accountNumber == nil && notes == nil) {
		sdb.handle.Exec(sdb.InsertInto(BANK_ACCOUNTS, name, nil, nil))
		return nil
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
			name,
			encrypt(accountNumber, password),
			encrypt(notes, password),
		),
	)
	return err
}

func (sdb SqliteDb) QueryBankAccount(accountId int, password *string) (BankInfo, error) {
	var query string

	if password == nil {
		query = "SELECT name FROM bank_accounts WHERE id=%d"
	} else {
		query = "SELECT * FROM bank_accounts WHERE id=%d"
	}

	rows, err := sdb.handle.Query(fmt.Sprintf(query, accountId))
	if err != nil {
		return BankInfo{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return BankInfo{}, fmt.Errorf("missing bank account")
	}

	if password == nil {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return BankInfo{}, err
		}
		return BankInfo{Name: name}, nil
	}

	var (
		id               string
		name             string
		encryptedAcctNum *string
		encryptedNotes   *string
		acctNum          string
		notes            string
	)

	if err = rows.Scan(&id, &name, &encryptedAcctNum, &encryptedNotes); err != nil {
		return BankInfo{}, err
	}

	if encryptedAcctNum != nil {
		acctNum, err = lib.DecryptData(*encryptedAcctNum, *password)
		if err != nil {
			return BankInfo{}, err
		}
	}

	if encryptedNotes != nil {
		notes, err = lib.DecryptData(*encryptedNotes, *password)
		if err != nil {
			return BankInfo{}, err
		}

	}
	return BankInfo{name, acctNum, notes}, nil
}

func (sdb SqliteDb) CreateIncome(name string, amount lib.Currency, p Period) (int64, error) {
	res, err := sdb.handle.Exec(sdb.InsertInto(INCOME, name, amount.ToInt(), p))
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sdb SqliteDb) SetIncome(id int, amount lib.Currency) error {
	_, err := sdb.handle.Exec(
		fmt.Sprintf("UPDATE income SET amount=%d WHERE id=%d", amount.ToInt(), id),
	)
	return err
}

/*
AffixIncome tracks an appended amount to an existing income. This could
be a bonus or overtime amount.

🟡The id is an income_history_id, not an income_id
*/
func (sdb SqliteDb) AffixIncome(id int, name string, amount lib.Currency) error {
	_, err := sdb.handle.Exec(sdb.InsertInto(INCOME_AFFIXES, id, amount.ToInt()))
	return err
}

func (sdb SqliteDb) QueryIncome() (IncomeInfo, error) {
	rows, err := sdb.handle.Query("SELECT * FROM income")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return IncomeInfo{}, fmt.Errorf("missing income entry")
	}

	var (
		id     int
		name   string
		amount int
		period string
	)

	err = rows.Scan(&id, &name, &amount, &period)
	if err != nil {
		return IncomeInfo{}, err
	}

	c := lib.NewCurrency("", sdb.currencyCode)

	return IncomeInfo{name, c.LoadAmount(amount), Period(period)}, nil
}

func (sdb SqliteDb) InsertInto(t Table, values ...any) string {
	columns, exists := tableData[t]
	if !exists {
		panic("unsupported table")
	}

	if len(columns) > len(values) || len(columns) < len(values) {
		panic(fmt.Sprintf("expected %d values, but got %d", len(columns), len(values)))
	}

	sbColumns := strings.Builder{}
	var realCols []string
	for i, col := range columns {
		if values[i] == nil {
			continue
		}
		realCols = append(realCols, col)
	}
	sbColumns.WriteString(
		fmt.Sprintf("INSERT INTO %s (%s)", getTableName(t), strings.Join(realCols, ",")),
	)

	sbColumns.WriteString(" VALUES (")
	valTemplate := "%v,"
	for _, val := range values {
		switch val.(type) {
		case string:
			valTemplate = "'%v',"
		case nil:
			continue
		}
		sbColumns.WriteString(fmt.Sprintf(valTemplate, val))
	}

	return sbColumns.String()[:len(sbColumns.String())-1] + ")"
}

func (sdb SqliteDb) Close() {
	_ = sdb.handle.Close()
}

func getTableName(table Table) string {
	switch table {
	case MONTHS:
		return "months"

	case BANK_ACCOUNTS:
		return "bank_accounts"
	case BANK_ACCOUNT_HISTORY:
		return "bank_account_history"
	case TRANSFERS:
		return "transfers"

	case INCOME:
		return "income"
	case INCOME_HISTORY:
		return "income_history"
	case INCOME_AFFIXES:
		return "income_affixes"

	case CREDIT_CARDS:
		return "credit_cards"
	case CREDIT_CARD_HISTORY:
		return "credit_card_history"
	}

	panic("unsupported table name")
}
