package db

type Period string

const (
	YEARLY   = Period("yearly")
	MONTHLY  = Period("monthly")
	WEEKLY   = Period("weekly")
	BIWEEKLY = Period("biweekly")
)

type FieldFlag int

const (
	BY_ID = FieldFlag(1 << iota)
	BY_MONTH_ID
	BY_BANK_ACCOUNT_ID
	BY_INCOME_ID
	BY_INCOME_HISTORY_ID
	BY_CREDIT_CARD_ID
)

var WhereFieldMap = map[FieldFlag]string{
	BY_ID:                "id",
	BY_MONTH_ID:          "month_id",
	BY_BANK_ACCOUNT_ID:   "bank_account_id",
	BY_INCOME_ID:         "income_id",
	BY_INCOME_HISTORY_ID: "income_history_id",
	BY_CREDIT_CARD_ID:    "credit_card_id",
}

type QueryMap map[FieldFlag]int

type FieldMap map[string]int

type Table string

const (
	MONTHS               = Table("months")
	INCOME               = Table("income")
	INCOME_HISTORY       = Table("income_history")
	INCOME_AFFIXES       = Table("income_affixes")
	BANK_ACCOUNTS        = Table("bank_accounts")
	BANK_ACCOUNT_HISTORY = Table("bank_account_history")
	TRANSFERS            = Table("transfers")
	CREDIT_CARDS         = Table("credit_cards")
	CREDIT_CARD_HISTORY  = Table("credit_card_history")
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
	CREDIT_CARDS: {"name", "due_day", "card_number", "last_four_digits", "notes"},
	CREDIT_CARD_HISTORY: {
		"credit_card_id",
		"month_id",
		"balance",
		"credit_limit",
		"paid_amount",
		"paid_date",
		"due_day",
		"period",
	},
}

type MonthInfo struct {
	Id   int
	Date string
}
