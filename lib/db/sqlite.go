package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"reflect"
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

func NewSqliteDb(name string, cc lib.CurrencyCode) *SqliteDb {
	db, err := sql.Open("sqlite", name+".db")
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
		case string, Period:
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

func buildFieldMap(allowedFields FieldFlag, qm QueryMap) FieldMap {
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
	td := tableData[t]
	for k := range fm {
		if k == "id" {
			continue
		}
		for ii, d := range td {
			if d == k {
				break
			}
			if ii+1 == len(d) {
				panic("id name does not exist on provided table")
			}
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE ", t)

	sb := strings.Builder{}
	sb.WriteString(query)

	mapPos := 0
	mapLen := len(fm)
	for k, v := range fm {
		switch realVal := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("%s=%s", k, realVal))
		case int, int64, int32:
			sb.WriteString(fmt.Sprintf("%s=%v", k, realVal))
		case lib.Currency:
			sb.WriteString(fmt.Sprintf("%s=%d", k, realVal.ToInt()))
		default:
			panic(fmt.Sprintf("unsupported type: %v", reflect.TypeOf(v)))
		}
		if mapPos+1 != mapLen {
			sb.WriteString(" AND ")
		}
		mapPos++
	}

	return sb.String()
}
