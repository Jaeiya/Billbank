package sqlite

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
	BILLS                = Table("bills")
	BILL_HISTORY         = Table("bill_history")
)

type TableFields = map[Table][]string

var tableData = TableFields{
	MONTHS:               {"year", "month"},
	INCOME:               {"name", "amount", "period"},
	INCOME_HISTORY:       {"income_id", "month_id", "amount"},
	INCOME_AFFIXES:       {"history_id", "name", "amount"},
	BANK_ACCOUNTS:        {"name", "account_number", "notes"},
	BANK_ACCOUNT_HISTORY: {"account_id", "month_id", "balance"},
	TRANSFERS: {
		"account_id",
		"month_id",
		"name",
		"amount",
		"date",
		"type",
		"to_whom",
		"from_whom",
	},
	CREDIT_CARDS: {
		"name",
		"due_day",
		"credit_limit",
		"card_number",
		"last_four_digits",
		"notes",
	},
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
	BILLS: {
		"name",
		"amount",
		"due_day",
		"period",
	},
	BILL_HISTORY: {
		"bill_id",
		"month_id",
		"amount",
		"paid_amount",
		"paid_date",
		"due_day",
		"notes",
	},
}

type (
	QueryMap  map[WhereFlag]any
	FieldMap  map[string]any
	WhereFlag int
)

const (
	WHERE_ID = WhereFlag(1 << iota)
	WHERE_NAME
	WHERE_AMOUNT
	WHERE_BALANCE
	WHERE_YEAR
	WHERE_MONTH
	WHERE_MONTH_ID
	WHERE_BANK_ACCOUNT_ID
	WHERE_INCOME_ID
	WHERE_INCOME_HISTORY_ID
	WHERE_CREDIT_CARD_ID
	WHERE_BILL_ID
)

var WhereFieldMap = map[WhereFlag]string{
	WHERE_ID:                "id",
	WHERE_NAME:              "name",
	WHERE_AMOUNT:            "amount",
	WHERE_BALANCE:           "balance",
	WHERE_YEAR:              "year",
	WHERE_MONTH:             "month",
	WHERE_MONTH_ID:          "month_id",
	WHERE_BANK_ACCOUNT_ID:   "bank_account_id",
	WHERE_INCOME_ID:         "income_id",
	WHERE_INCOME_HISTORY_ID: "income_history_id",
	WHERE_CREDIT_CARD_ID:    "credit_card_id",
	WHERE_BILL_ID:           "bill_id",
}

type Period string

const (
	YEARLY   = Period("yearly")
	MONTHLY  = Period("monthly")
	WEEKLY   = Period("weekly")
	BIWEEKLY = Period("biweekly")
)
