package db

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
	CC_PAID_DATE   = CCField("paid_date")
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

type CreditCardRow struct {
	ID             int
	Name           string
	DueDay         int
	CreditLimit    *lib.Currency
	CardNumber     *string
	LastFourDigits string
	Notes          *string
}

func (cr CreditCardRow) String() string {
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

type CreditCardHistoryRow struct {
	ID           int
	CreditCardID int
	MonthID      int
	Balance      lib.Currency
	CreditLimit  *lib.Currency
	PaidAmount   lib.Currency
	PaidDate     *string
	DueDay       int
	Period       Period
}

func (chr CreditCardHistoryRow) String() string {
	return fmt.Sprintf(
		"\nid: %d\nccid: %d\nmid: %d\nbal: %s\nclimit: %v\npaidA: %s\npaidD: %v\ndueday: %d",
		chr.ID,
		chr.CreditCardID,
		chr.MonthID,
		chr.Balance.String(),
		chr.CreditLimit,
		chr.PaidAmount.String(),
		lib.TryDeref(chr.PaidDate),
		chr.DueDay,
	)
}

func (sdb SqliteDb) CreateCreditCard(config CreditCardConfig) {
	creditLimit := lib.TryDeref(config.CreditLimit)
	if creditLimit != nil {
		creditLimit = config.CreditLimit.ToInt()
	}

	if config.Password == nil {
		if _, err := sdb.handle.Exec(
			sdb.InsertInto(
				CREDIT_CARDS,
				config.Name,
				config.DueDay,
				creditLimit,
				nil,
				config.LastFourDigits,
				nil,
			),
		); err != nil {
			panic(err)
		}
		return
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
		panic(err)
	}
}

func (sdb SqliteDb) QueryCreditCards(qm QueryMap) ([]CreditCardRow, error) {
	rows := sdb.query(CREDIT_CARDS, qm)
	var creditLimit *int
	var cards []CreditCardRow

	for rows.Next() {
		var card CreditCardRow

		if err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.DueDay,
			&creditLimit,
			&card.CardNumber,
			&card.LastFourDigits,
			&card.Notes,
		); err != nil {
			panic(err)
		}

		if creditLimit != nil {
			c := lib.NewCurrency("", sdb.currencyCode)
			c.LoadAmount(*creditLimit)
			card.CreditLimit = &c
		}

		cards = append(cards, card)
	}

	if len(cards) == 0 {
		return []CreditCardRow{}, fmt.Errorf("no results found")
	}

	return cards, nil
}

func (sdb SqliteDb) QueryDecryptedCreditCard(
	qm QueryMap,
	password string,
) ([]CreditCardRow, error) {
	rows := sdb.query(CREDIT_CARDS, qm)
	var creditLimit *int
	var encCardNum, encNotes *string
	var cards []CreditCardRow

	for rows.Next() {
		var card CreditCardRow

		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.DueDay,
			&creditLimit,
			&encCardNum,
			&card.LastFourDigits,
			&encNotes,
		)
		if err != nil {
			panic(err)
		}

		card.CardNumber, err = lib.DecryptNonNil(encCardNum, password)
		if err != nil {
			panic(err)
		}

		card.Notes, err = lib.DecryptNonNil(encNotes, password)
		if err != nil {
			panic(err)
		}

		if creditLimit != nil {
			c := lib.NewCurrency("", sdb.currencyCode)
			c.LoadAmount(*creditLimit)
			card.CreditLimit = &c
		}

		cards = append(cards, card)
	}

	if len(cards) == 0 {
		return []CreditCardRow{}, fmt.Errorf("no results found")
	}

	return cards, nil
}

func (sdb SqliteDb) CreateCreditCardHistory(config CreditCardHistoryConfig) {
	creditLimit := lib.TryDeref(config.CreditLimit)
	if creditLimit != nil {
		creditLimit = config.CreditLimit.ToInt()
	}
	if _, err := sdb.handle.Exec(
		sdb.InsertInto(
			CREDIT_CARD_HISTORY,
			config.CreditCardID,
			config.MonthID,
			config.Balance.ToInt(),
			creditLimit,
			nil, // paid amount -- defaults to 0
			nil, // paid date
			config.DueDay,
			MONTHLY,
		),
	); err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryCreditCardHistory(qm QueryMap) ([]CreditCardHistoryRow, error) {
	rows := sdb.query(CREDIT_CARD_HISTORY, qm)

	var (
		balance     int
		creditLimit *int
		paidAmount  int
		ccRows      []CreditCardHistoryRow
	)

	for rows.Next() {
		var row CreditCardHistoryRow

		if err := rows.Scan(
			&row.ID,
			&row.CreditCardID,
			&row.MonthID,
			&balance,
			&creditLimit,
			&paidAmount,
			&row.PaidDate,
			&row.DueDay,
			&row.Period,
		); err != nil {
			panic(err)
		}

		realBalance := lib.NewCurrency("", sdb.currencyCode)
		realBalance.LoadAmount(balance)
		row.Balance = realBalance

		if creditLimit != nil {
			c := lib.NewCurrency("", sdb.currencyCode)
			c.LoadAmount(*creditLimit)
			row.CreditLimit = &c
		}

		realPaidAmount := lib.NewCurrency("", sdb.currencyCode)
		realPaidAmount.LoadAmount(paidAmount)
		row.PaidAmount = realPaidAmount

		ccRows = append(ccRows, row)
	}

	if len(ccRows) == 0 {
		return []CreditCardHistoryRow{}, fmt.Errorf("no query results")
	}

	return ccRows, nil
}

func (sdb SqliteDb) SetCreditCardHistory(historyID int, fieldMap CCFieldMap) {
	conditions := make([]string, 0, len(fieldMap))
	for field, value := range fieldMap {
		switch field {

		case CC_BALANCE, CC_LIMIT, CC_PAID_AMOUNT:
			c, err := lib.ToCurrency(value)
			if err != nil {
				panic(fmt.Sprintf("%s should be of type: lib.Currency", field))
			}
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, c.ToInt()))

		case CC_DUE_DAY:
			if !lib.IsInt(value) {
				panic(fmt.Sprintf("%s should of of type: int", field))
			}
			conditions = append(conditions, fmt.Sprintf("due_day=%d", value))

		case CC_PAID_DATE:
			if !lib.IsString(value) {
				panic(fmt.Sprintf("%s should be of type: string", field))
			}
			conditions = append(conditions, fmt.Sprintf("paid_date='%s'", value))

		default:
			panic(fmt.Sprintf("unsupported credit card history field: %s", field))
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
}
