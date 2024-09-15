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
	CardNumber     *string
	DueDay         int
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
	CardNumber     *string
	LastFourDigits string
	Notes          *string
}

func (cr CreditCardRow) String() string {
	return fmt.Sprintf(
		"\nid: %d\nname: %s\ndueDay: %d\nnum: %v\nlastFour: %v\nnotes: %v",
		cr.ID,
		cr.Name,
		cr.DueDay,
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
	if config.Password == nil {
		_, err := sdb.handle.Exec(
			sdb.InsertInto(
				CREDIT_CARDS,
				config.Name,
				config.DueDay,
				nil,
				config.LastFourDigits,
				nil,
			),
		)
		if err != nil {
			panic(err)
		}
		return
	}
	_, err := sdb.handle.Exec(
		sdb.InsertInto(
			CREDIT_CARDS,
			config.Name,
			config.DueDay,
			lib.EncryptNotNil(config.CardNumber, *config.Password),
			config.LastFourDigits,
			lib.EncryptNotNil(config.Notes, *config.Password),
		),
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryAllCreditCards() ([]CreditCardRow, error) {
	queryStr := fmt.Sprintf("SELECT * FROM %s", getTableName(CREDIT_CARDS))
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var (
		id             int
		name           string
		dueDay         int
		encCardNum     *string
		lastFourDigits string
		encNotes       *string
		cards          []CreditCardRow
	)

	for rows.Next() {
		err = rows.Scan(&id, &name, &dueDay, &encCardNum, &lastFourDigits, &encNotes)
		if err != nil {
			panic(err)
		}
		cards = append(cards, CreditCardRow{
			id, name, dueDay, encCardNum, lastFourDigits, encNotes,
		})
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
	_, err := sdb.handle.Exec(
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
	)
	if err != nil {
		panic(err)
	}
}

func (sdb SqliteDb) QueryCreditCardHistory() ([]CreditCardHistoryRow, error) {
	queryStr := fmt.Sprintf("SELECT * FROM %s", getTableName(CREDIT_CARD_HISTORY))
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}

	var (
		id          int
		ccID        int
		monthID     int
		balance     int
		creditLimit *int
		paidAmount  int
		paidDate    *string
		dueDay      int
		period      Period
		ccRows      []CreditCardHistoryRow
	)

	for rows.Next() {
		err = rows.Scan(
			&id,
			&ccID,
			&monthID,
			&balance,
			&creditLimit,
			&paidAmount,
			&paidDate,
			&dueDay,
			&period,
		)
		if err != nil {
			panic(err)
		}
		balanceCurrency := lib.NewCurrency("", sdb.currencyCode)
		balanceCurrency.LoadAmount(balance)

		var cLimitCurrency *lib.Currency
		if creditLimit != nil {
			c := lib.NewCurrency("", sdb.currencyCode)
			c.LoadAmount(*creditLimit)
			cLimitCurrency = &c
		}

		paidAmountCurrency := lib.NewCurrency("", sdb.currencyCode)
		paidAmountCurrency.LoadAmount(paidAmount)

		ccRows = append(ccRows, CreditCardHistoryRow{
			id,
			ccID,
			monthID,
			balanceCurrency,
			cLimitCurrency,
			paidAmountCurrency,
			paidDate,
			dueDay,
		})
	}

	if len(ccRows) == 0 {
		return []CreditCardHistoryRow{}, fmt.Errorf("no query results")
	}

	return ccRows, nil
}

func (sdb SqliteDb) SetCreditCardHistory(historyID int, fieldMap CCFieldMap) {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("UPDATE %s\nSET ", getTableName(CREDIT_CARD_HISTORY)))

	for field, value := range fieldMap {
		switch field {

		case CC_BALANCE, CC_LIMIT, CC_PAID_AMOUNT:
			c, err := lib.ToCurrency(value)
			if err != nil {
				panic(fmt.Sprintf("%s should be of type: lib.Currency", field))
			}
			sb.WriteString(fmt.Sprintf("%s = %d,", field, c.ToInt()))

		case CC_DUE_DAY:
			if !lib.IsInt(value) {
				panic(fmt.Sprintf("%s should of of type: int", field))
			}
			sb.WriteString(fmt.Sprintf("due_day = %d,", value))

		case CC_PAID_DATE:
			if _, isType := value.(string); !isType {
				panic(fmt.Sprintf("%s should be of type: string", field))
			}
			sb.WriteString(fmt.Sprintf("paid_date = '%s',", value))

		default:
			panic("unsupported credit card history field")
		}
	}

	query := fmt.Sprintf("%s\nWHERE id = %d;", sb.String()[:sb.Len()-1], historyID)
	_, err := sdb.handle.Exec(query)
	if err != nil {
		panic(err)
	}
}
