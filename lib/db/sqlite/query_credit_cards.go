package sqlite

import (
	"fmt"
	"strings"

	"github.com/jaeiya/billbank/lib"
)

type (
	CCField    string
	CCFieldMap map[CCField]any
)

const (
	CC_BALANCE     = CCField("balance")
	CC_LIMIT       = CCField("credit_limit")
	CC_PAID_AMOUNT = CCField("paid_amount")
	CC_PAID_DAY    = CCField("paid_day")
	CC_DUE_DAY     = CCField("due_day")
)

type CreditCardConfig struct {
	Name           string
	DueDay         int
	CreditLimit    *lib.Currency
	CardNumber     *string
	LastFourDigits string
	Notes          *string
	Password       *string
}

type CreditCardHistoryConfig struct {
	CreditCardID int
	MonthID      int
	Balance      lib.Currency
	CreditLimit  *lib.Currency
	DueDay       int
}

type CreditCardRecord struct {
	ID             int
	Name           string
	DueDay         int
	CreditLimit    *lib.Currency
	CardNumber     *string
	LastFourDigits string
	Notes          *string
}

func (cr CreditCardRecord) String() string {
	return fmt.Sprintf(
		"\nid: %d\nname: %s\ndueDay: %d\nlimit: %s\nnum: %v\nlastFour: %v\nnotes: %v",
		cr.ID,
		cr.Name,
		cr.DueDay,
		cr.CreditLimit,
		lib.TryDeref(cr.CardNumber),
		cr.LastFourDigits,
		lib.TryDeref(cr.Notes),
	)
}

type CardHistoryRecord struct {
	ID           int
	CreditCardID int
	MonthID      int
	Balance      lib.Currency
	CreditLimit  *lib.Currency
	PaidAmount   lib.Currency
	PaidDay      *int
	DueDay       int
	Period       Period
}

func (chr CardHistoryRecord) String() string {
	return fmt.Sprintf(
		"\nid: %d\nccid: %d\nmid: %d\nbal: %s\nclimit: %v\npaidA: %s\npaidD: %v\ndueday: %d",
		chr.ID,
		chr.CreditCardID,
		chr.MonthID,
		chr.Balance.String(),
		chr.CreditLimit,
		chr.PaidAmount.String(),
		lib.TryDeref(chr.PaidDay),
		chr.DueDay,
	)
}

func (sdb SqliteDb) CreateCreditCard(config CreditCardConfig) {
	creditLimit := lib.TryDeref(config.CreditLimit)
	if creditLimit != nil {
		creditLimit = config.CreditLimit.GetStoredValue()
	}

	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			CREDIT_CARDS,
			config.Name,
			config.DueDay,
			creditLimit,
			lib.EncryptNonNil(config.CardNumber, config.Password),
			config.LastFourDigits,
			lib.EncryptNonNil(config.Notes, config.Password),
		),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryCreditCards(
	qm QueryMap,
	password *string,
) ([]CreditCardRecord, error) {
	rows := sdb.query(CREDIT_CARDS, qm)
	var creditLimit *int
	var records []CreditCardRecord

	for rows.Next() {
		var record CreditCardRecord
		var err error

		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.DueDay,
			&creditLimit,
			&record.CardNumber,
			&record.LastFourDigits,
			&record.Notes,
		); err != nil {
			panic(err)
		}

		if creditLimit != nil {
			c := lib.NewCurrencyFromStore(*creditLimit, sdb.currencyCode)
			record.CreditLimit = &c
		}

		if password != nil && record.CardNumber != nil {
			if record.CardNumber, err = lib.DecryptNonNil(record.CardNumber, *password); err != nil {
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
		return []CreditCardRecord{}, fmt.Errorf("no results found")
	}

	return records, nil
}

func (sdb SqliteDb) CreateCreditCardHistory(config CreditCardHistoryConfig) {
	creditLimit := lib.TryDeref(config.CreditLimit)
	if creditLimit != nil {
		creditLimit = config.CreditLimit.GetStoredValue()
	}
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			CREDIT_CARD_HISTORY,
			config.CreditCardID,
			config.MonthID,
			config.Balance.GetStoredValue(),
			creditLimit,
			nil, // paid amount -- defaults to 0
			nil, // paid date
			config.DueDay,
			MONTHLY,
		),
	); err != nil {
		panicOnExecErr(err)
	}
}

func (sdb SqliteDb) QueryCreditCardHistory(qm QueryMap) ([]CardHistoryRecord, error) {
	rows := sdb.query(CREDIT_CARD_HISTORY, qm)

	var (
		balance     int
		creditLimit *int
		paidAmount  int
		records     []CardHistoryRecord
	)

	for rows.Next() {
		var record CardHistoryRecord

		if err := rows.Scan(
			&record.ID,
			&record.CreditCardID,
			&record.MonthID,
			&balance,
			&creditLimit,
			&paidAmount,
			&record.PaidDay,
			&record.DueDay,
			&record.Period,
		); err != nil {
			panic(err)
		}

		if creditLimit != nil {
			c := lib.NewCurrencyFromStore(*creditLimit, sdb.currencyCode)
			record.CreditLimit = &c
		}

		record.Balance = lib.NewCurrencyFromStore(balance, sdb.currencyCode)
		record.PaidAmount = lib.NewCurrencyFromStore(paidAmount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []CardHistoryRecord{}, fmt.Errorf("no query results")
	}

	return records, nil
}

func (sdb SqliteDb) SetCreditCardHistory(historyID int, fieldMap CCFieldMap) error {
	conditions := make([]string, 0, len(fieldMap))
	for field, value := range fieldMap {
		switch field {

		case CC_BALANCE, CC_LIMIT, CC_PAID_AMOUNT:
			c, err := lib.ToCurrency(value)
			if err != nil {
				return fmt.Errorf("%s should be of type: lib.Currency", field)
			}
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, c.GetStoredValue()))

		case CC_DUE_DAY, CC_PAID_DAY:
			if !lib.IsInt(value) {
				return fmt.Errorf("%s should of of type: int", field)
			}
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, value))

		default:
			return fmt.Errorf("unsupported credit card history field: %s", field)
		}
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = %d",
		CREDIT_CARD_HISTORY,
		strings.Join(conditions, ","),
		historyID,
	)

	if _, err := sdb.handle.Exec(query); err != nil {
		panic(err)
	}
	return nil
}
