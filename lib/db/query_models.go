package db

import "github.com/jaeiya/billbank/lib"

type Period string

const (
	YEARLY   = Period("yearly")
	MONTHLY  = Period("monthly")
	WEEKLY   = Period("weekly")
	BIWEEKLY = Period("biweekly")
)

type Where int

const (
	BY_ID = Where(iota)
	BY_MONTH_ID
	BY_ACCOUNT_ID
)

type QueryWhere map[Where]int

type WhereMap map[string]any

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
		"bank_account_id",
		"month_id",
		"name",
		"amount",
		"date",
		"type",
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

type MonthInfo struct {
	Id   int
	Date string
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
