package assets

import _ "embed"

//go:embed sql/init_db.sqlite
var slqInitStr string

//go:embed sql/get_monthly_bills.sqlite
var billsQueryStr string

// References raw SQLite strings for various operations
// that are meant to be used with Exec()
type SQLite struct {
	CreateDatabase string
	QueryBills     string
}

var SQL SQLite = SQLite{
	CreateDatabase: slqInitStr,
	QueryBills:     billsQueryStr,
}
