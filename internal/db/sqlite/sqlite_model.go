//nolint:goconst
package sqlite

type Table string

const (
	Months            = Table("months")
	Income            = Table("income")
	IncomeHistory     = Table("income_history")
	IncomeAffixes     = Table("income_affixes")
	BankAccts         = Table("bank_accounts")
	BankAcctHistory   = Table("bank_account_history")
	BankTranx         = Table("bank_transfers")
	CreditCards       = Table("credit_cards")
	CreditCardHistory = Table("credit_card_history")
	Bills             = Table("bills")
	BillsHistory      = Table("bills_history")
	BillTypes         = Table("bill_types")
)

type TableFields = map[Table][]string

var tableData = TableFields{
	Months:          {"date"},
	Income:          {"name", "amount", "period"},
	IncomeHistory:   {"income_id", "month_id", "amount"},
	IncomeAffixes:   {"history_id", "name", "amount"},
	BankAccts:       {"name", "acct_num", "notes"},
	BankAcctHistory: {"account_id", "month_id", "balance"},
	BankTranx: {
		"bank_history_id",
		"month_id",
		"name",
		"amount",
		"due_day",
		"type",
		"to_whom",
		"from_whom",
	},
	CreditCards: {
		"name",
		"due_day",
		"credit_limit",
		"card_number",
		"last_four_digits",
		"notes",
	},
	CreditCardHistory: {
		"card_id",
		"month_id",
		"balance",
		"due_day",
		"credit_limit",
		"paid_day",
		"paid_amount",
		"cleared_day",
	},
	Bills: {
		"type_id",
		"name",
		"amount",
		"due_date",
		"status",
		"period",
		"is_active",
	},
	BillsHistory: {
		"month_id",
		"type_id",
		"name",
		"amount",
		"due_date",
		"paid_amount",
		"paid_date",
		"paid_how",
		"cleared_day",
		"notes",
	},
	BillTypes: {
		"id",
		"name",
	},
}

type (
	QueryMap  map[WhereFlag]any
	FieldMap  map[string]any
	WhereFlag int
)

const (
	WhereID = WhereFlag(1 << iota)
	WhereName
	WhereAmount
	WhereBalance
	WhereYear
	WhereMonth
	WhereMonthID
	WhereBankAcctID
	WhereIncomeID
	WhereIncomeHistoryID
	WhereCreditCardID
	WhereBillID
)

var WhereFieldMap = map[WhereFlag]string{
	WhereID:              "id",
	WhereName:            "name",
	WhereAmount:          "amount",
	WhereBalance:         "balance",
	WhereYear:            "year",
	WhereMonth:           "month",
	WhereMonthID:         "month_id",
	WhereBankAcctID:      "bank_account_id",
	WhereIncomeID:        "income_id",
	WhereIncomeHistoryID: "income_history_id",
	WhereCreditCardID:    "credit_card_id",
	WhereBillID:          "bill_id",
}

type Period string

const (
	Yearly    = Period("yearly")
	BiYearly  = Period("bi-yearly")
	TriYearly = Period("tri-yearly")

	Monthly   = Period("monthly")
	BiMonthly = Period("bi-monthly")

	Weekly   = Period("weekly")
	BiWeekly = Period("bi-weekly")
)

var PeriodStrings = [...]string{
	string(Yearly),
	string(BiYearly),
	string(TriYearly),
	string(Monthly),
	string(BiMonthly),
	string(Weekly),
	string(BiWeekly),
}

// Bill types mapped directly to their database ID.
//
// 🔴 These ID's should NEVER change! If a type needs to be
// added, add the type name (keeping alphabetical order) and
// set its ID to the current largest id + 1.
var BillTypeMap = map[string]int{
	"credit card":   1,
	"donation":      2,
	"education":     3,
	"entertainment": 4,
	"family":        5,
	"food":          6,
	"health":        7,
	"housing":       8,
	"insurance":     9,
	"loan":          10,
	"maintenance":   11,
	"medical":       12,
	"subscription":  13,
	"tax":           14,
	"transit":       15,
	"utility":       16,
}
