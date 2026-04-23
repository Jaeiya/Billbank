package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jaeiya/billbank/assets"
	"github.com/jaeiya/billbank/internal"
	_ "modernc.org/sqlite"
)

type SqliteDb struct {
	handle       *sql.DB
	currencyCode internal.CurrencyCode
}

func NewSqliteDb(filePath string, cc internal.CurrencyCode) (*SqliteDb, error) {
	_, err := os.ReadDir(filepath.Dir(filePath))
	if err != nil {
		return nil, fmt.Errorf("cannot load database: %w", err)
	}

	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA user_version = 1;")
	if err != nil {
		return nil, err
	}

	// Creates the physical db file
	_, err = db.Exec(assets.SQL.CreateDatabase)
	if err != nil {
		return nil, err
	}

	return &SqliteDb{db, cc}, nil
}

func (sdb SqliteDb) Close() error {
	return sdb.handle.Close()
}

func (sdb SqliteDb) Query(s string) (*sql.Rows, error) {
	return sdb.handle.Query(s)
}

func (sdb SqliteDb) insertInto(t Table, args ...any) (sql.Result, error) {
	columns, exists := tableData[t]
	if !exists {
		return nil, ErrUnsupportedTable
	}

	if len(columns) != len(args) {
		return nil, ErrMismatchColsValues
	}

	return sdb.handle.Exec(toInsertStr(string(t), columns), args...)
}

func (sdb SqliteDb) queryAll(t Table) (*sql.Rows, error) {
	rows, err := sdb.handle.Query(fmt.Sprintf("SELECT * FROM %s", t))
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (sdb SqliteDb) query(t Table, qm QueryMap) (*sql.Rows, error) {
	var fm FieldMap
	var err error

	whereIDOrMonthID := WHERE_ID | WHERE_MONTH_ID
	switch t {
	case MONTHS:
		fm, err = buildFieldMap(WHERE_ID|WHERE_MONTH|WHERE_YEAR, qm)

	case BANK_ACCOUNTS, INCOME, BILLS:
		fm, err = buildFieldMap(WHERE_ID, qm)

	case BANK_ACCOUNT_HISTORY:
		fm, err = buildFieldMap(whereIDOrMonthID|WHERE_BANK_ACCOUNT_ID, qm)

	case BANK_TRANSFERS:
		fm, err = buildFieldMap(whereIDOrMonthID|WHERE_BANK_ACCOUNT_ID, qm)

	case CREDIT_CARDS:
		fm, err = buildFieldMap(WHERE_ID|WHERE_NAME, qm)

	case CREDIT_CARD_HISTORY:
		fm, err = buildFieldMap(whereIDOrMonthID|WHERE_CREDIT_CARD_ID, qm)

	case INCOME_HISTORY:
		fm, err = buildFieldMap(whereIDOrMonthID|WHERE_INCOME_ID, qm)

	case INCOME_AFFIXES:
		fm, err = buildFieldMap(WHERE_ID|WHERE_INCOME_ID, qm)

	case BILLS_HISTORY:
		fm, err = buildFieldMap(whereIDOrMonthID|WHERE_BILL_ID, qm)

	default:
		err = fmt.Errorf("unsupported table: %s", t)
	}

	if err != nil {
		return nil, err
	}

	queryStr, err := buildQueryStr(t, fm)
	if err != nil {
		return nil, err
	}

	rows, err := sdb.handle.Query(queryStr)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// insertMultiInto inserts multiple records at once using a
// resolve function that returns the necessary order of
// table values to be inserted.
func insertMultiInto[T any](
	db SqliteDb,
	t Table,
	records []T,
	resolve func(r T) []any,
) (sql.Result, error) {
	resolvedRecords := make([][]any, len(records))
	for i, r := range records {
		resolvedRecords[i] = resolve(r)
	}

	columns, exists := tableData[t]
	if !exists {
		return nil, ErrUnsupportedTable
	}

	for _, rr := range resolvedRecords {
		if len(rr) != len(columns) {
			return nil, ErrMismatchColsValues
		}
	}

	return db.handle.Exec(
		toInsertMultiStr(string(t), columns, len(records)),
		slices.Concat(resolvedRecords...)...,
	)
}

func toInsertStr(tName string, cols []string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "INSERT INTO %s (%s) VALUES (", tName, strings.Join(cols, ","))

	for i := range len(cols) {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
	}
	sb.WriteString(");")

	return sb.String()
}

func toInsertMultiStr(tName string, cols []string, count int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "INSERT INTO %s (%s) VALUES ", tName, strings.Join(cols, ","))

	for i := range count {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for ii := range len(cols) {
			if ii > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
		}
		sb.WriteString(")")
	}
	return sb.String() + ";"
}

func buildFieldMap(allowedFields WhereFlag, qm QueryMap) (FieldMap, error) {
	fm := FieldMap{}
	for ff, fieldValue := range qm {
		field, fieldExists := WhereFieldMap[ff]
		if !fieldExists {
			return FieldMap{}, ErrUnsupportedFieldMap
		}
		if allowedFields&ff == 0 {
			return FieldMap{}, fmt.Errorf("field not allowed: %v", WhereFieldMap[ff])
		}
		fm[field] = fieldValue
	}

	return fm, nil
}

func buildQueryStr(t Table, fm FieldMap) (string, error) {
	td, ok := tableData[t]
	if !ok {
		return "", fmt.Errorf("table does not exist [%s]", t)
	}

	var conditions []string
	if len(fm) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", t), nil
	}

	for field, val := range fm {
		// id's are not part of the table data because they are created
		// automatically by SQL.
		if field != "id" {
			if !slices.Contains(td, field) {
				return "", fmt.Errorf(
					"[%s] is an unsupported field for the table [%s]",
					field,
					t,
				)
			}
		}

		switch realVal := val.(type) {
		case string, Period:
			conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", field, realVal))
		case int, int64, int32:
			conditions = append(conditions, fmt.Sprintf("%s=%v", field, realVal))
		case internal.Currency:
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, realVal.GetStoredValue()))
		default:
			return "", fmt.Errorf("unsupported type [%T]", val)
		}
	}

	return fmt.Sprintf("SELECT * FROM %s WHERE %s", t, strings.Join(conditions, " AND ")), nil
}

func getExecError(err error) error {
	if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		return ErrForeignKey
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: due_day") {
		return ErrDueDayInvalid
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: type") {
		return ErrTransferTypeInvalid
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: amount") {
		return ErrAmountInvalid
	}
	if strings.Contains(err.Error(), "CHECK constraint failed: month") {
		return ErrMonthInvalid
	}
	if strings.Contains(err.Error(), "UNIQUE constraint failed") &&
		strings.Contains(err.Error(), ".name (") {
		return ErrUniqueName
	}

	return err
}
