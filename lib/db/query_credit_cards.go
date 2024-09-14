package db

import (
	"fmt"

	"github.com/jaeiya/billbank/lib"
)

type CreditCardConfig struct {
	Name           string
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
	CardNumber     *string
	LastFourDigits string
	Notes          *string
}

func (cr CreditCardRow) String() string {
	return fmt.Sprintf(
		"\nid: %d\nname: %s\nnum: %v\nlastFour: %v\nnotes: %v",
		cr.ID,
		cr.Name,
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
		chr.PaidDate,
		chr.DueDay,
	)
}

func (sdb SqliteDb) CreateCreditCard(config CreditCardConfig) {
	if config.Password == nil {
		_, err := sdb.handle.Exec(
			sdb.InsertInto(CREDIT_CARDS, config.Name, nil, config.LastFourDigits, nil),
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
		encCardNum     *string
		lastFourDigits string
		encNotes       *string
		cards          []CreditCardRow
	)

	for rows.Next() {
		err = rows.Scan(&id, &name, &encCardNum, &lastFourDigits, &encNotes)
		cards = append(cards, CreditCardRow{
			id, name, encCardNum, lastFourDigits, encNotes,
		})
	}

	if len(cards) == 0 {
		return []CreditCardRow{}, fmt.Errorf("no results found")
	}

	return cards, nil
}

func (sdb SqliteDb) CreateCreditCardHistory(config CreditCardHistoryConfig) {
	_, err := sdb.handle.Exec(
		sdb.InsertInto(
			CREDIT_CARD_HISTORY,
			config.CreditCardID,
			config.MonthID,
			config.Balance.ToInt(),
			nil,
			nil,
			nil,
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
