package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaeiya/billbank/lib"
	_ "modernc.org/sqlite"
)

type SqliteDb struct {
	handle       *sql.DB
	currencyCode lib.CurrencyCode
}

//go:embed sql/init_db.sqlite
var initBankSQL string

var (
	ErrForeignKey          = fmt.Errorf("foreign key failed validation")
	ErrDueDayInvalid       = fmt.Errorf("failed to validate due_day constraint")
	ErrTransferTypeInvalid = fmt.Errorf("failed to valid transfer_type constraint")
	ErrAmountInvalid       = fmt.Errorf("failed to validate amount constraint")
	ErrMonthInvalid        = fmt.Errorf("failed to validate month constraint")
	ErrUniqueName          = fmt.Errorf("failed unique 'name' constraint requirement")
)

func NewSqliteDb(filePath string, cc lib.CurrencyCode) *SqliteDb {
	_, err := os.ReadDir(filepath.Dir(filePath))
	if err != nil {
		panic(fmt.Errorf("cannot load database: %w", err))
	}

	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(initBankSQL)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("PRAGMA user_version = 1;")
	if err != nil {
		panic(err)
	}

	return &SqliteDb{db, cc}
}

func (sdb SqliteDb) InsertInto(t Table, values ...any) string {
	columns, exists := tableData[t]
	if !exists {
		panic("unsupported table")
	}

	if len(columns) != len(values) {
		panic(fmt.Sprintf("expected %d values, but got %d", len(columns), len(values)))
	}

	var realCols []string
	var realValues []string

	for i, col := range columns {
		if values[i] == nil {
			continue
		}
		realCols = append(realCols, col)

		switch v := values[i].(type) {
		case string, Period, TransferType:
			realValues = append(realValues, fmt.Sprintf("'%v'", v))
		case time.Month:
			realValues = append(realValues, fmt.Sprintf("%d", v))
		default:
			realValues = append(realValues, fmt.Sprintf("%v", v))
		}
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		t,
		strings.Join(realCols, ","),
		strings.Join(realValues, ","),
	)
}

func (sdb SqliteDb) Close() {
	_ = sdb.handle.Close()
}

func (sdb SqliteDb) query(t Table, qm QueryMap) *sql.Rows {
	var fm FieldMap
	whereIDOrMonthID := WHERE_ID | WHERE_MONTH_ID
	switch t {
	case MONTHS:
		fm = buildFieldMap(WHERE_ID|WHERE_MONTH|WHERE_YEAR, qm)

	case BANK_ACCOUNTS, INCOME, BILLS:
		fm = buildFieldMap(WHERE_ID, qm)

	case BANK_ACCOUNT_HISTORY:
		fm = buildFieldMap(whereIDOrMonthID|WHERE_BANK_ACCOUNT_ID, qm)

	case TRANSFERS:
		fm = buildFieldMap(whereIDOrMonthID|WHERE_BANK_ACCOUNT_ID, qm)

	case CREDIT_CARDS:
		fm = buildFieldMap(WHERE_ID|WHERE_NAME, qm)

	case CREDIT_CARD_HISTORY:
		fm = buildFieldMap(whereIDOrMonthID|WHERE_CREDIT_CARD_ID, qm)

	case INCOME_HISTORY:
		fm = buildFieldMap(whereIDOrMonthID|WHERE_INCOME_ID, qm)

	case INCOME_AFFIXES:
		fm = buildFieldMap(WHERE_ID|WHERE_INCOME_ID, qm)

	case BILL_HISTORY:
		fm = buildFieldMap(whereIDOrMonthID|WHERE_BILL_ID, qm)

	default:
		panic(fmt.Sprintf("unsupported table: %s", t))
	}

	queryStr := buildQueryStr(t, fm)
	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		panic(err)
	}
	return rows
}

func buildFieldMap(allowedFields WhereFlag, qm QueryMap) FieldMap {
	fm := FieldMap{}
	for ff, fieldValue := range qm {
		field, fieldExists := WhereFieldMap[ff]
		if !fieldExists {
			panic("unsupported field")
		}
		if allowedFields&ff == 0 {
			panic(fmt.Sprintf("field not allowed: %v", WhereFieldMap[ff]))
		}
		fm[field] = fieldValue
	}

	return fm
}

func buildQueryStr(t Table, fm FieldMap) string {
	td, ok := tableData[t]
	if !ok {
		panic("table does not exist")
	}

	var conditions []string
	if len(fm) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", t)
	}

	for field, val := range fm {
		// id's are not part of the table data because they are created
		// automatically by SQL.
		if field != "id" {
			if !lib.StrSliceContains(td, field) {
				panic(fmt.Sprintf("%s is an unsupported field for the table: %s", field, t))
			}
		}

		switch realVal := val.(type) {
		case string, Period:
			conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", field, realVal))
		case int, int64, int32:
			conditions = append(conditions, fmt.Sprintf("%s=%v", field, realVal))
		case lib.Currency:
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, realVal.GetStoredValue()))
		default:
			panic(fmt.Sprintf("unsupported type: %T", val))
		}
	}

	return fmt.Sprintf("SELECT * FROM %s WHERE %s", t, strings.Join(conditions, " AND "))
}

func panicOnExecErr(err error) {
	if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		panic(ErrForeignKey)
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: due_day") {
		panic(ErrDueDayInvalid)
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: transfer_type") {
		panic(ErrTransferTypeInvalid)
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: amount") {
		panic(ErrAmountInvalid)
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: month") {
		panic(ErrMonthInvalid)
	}
	if strings.Contains(err.Error(), "UNIQUE constraint failed") &&
		strings.Contains(err.Error(), ".name (") {
		panic(ErrUniqueName)
	}
	panic(err)
}
