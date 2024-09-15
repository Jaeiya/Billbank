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

func (sdb SqliteDb) CreateMonth(t time.Time) error {
	// Make sure any prior date arithmetic, used a clean date
	isClean := t.Day() == 1 &&
		t.Hour() == 0 &&
		t.Minute() == 0 &&
		t.Second() == 0 &&
		t.Nanosecond() == 0

	if !isClean {
		return fmt.Errorf("not a clean date")
	}
	_, err := sdb.handle.Exec(sdb.InsertInto(MONTHS, t.Format("2006-01")))
	if err != nil {
		return err
	}
	return nil
}

func (sdb SqliteDb) QueryMonths() (MonthInfo, error) {
	rows, err := sdb.handle.Query("SELECT * FROM months")
	if err != nil {
		return MonthInfo{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return MonthInfo{}, fmt.Errorf("empty month table")
	}

	var (
		id   int
		date string
	)

	err = rows.Scan(&id, &date)
	if err != nil {
		return MonthInfo{}, err
	}

	return MonthInfo{id, date}, nil
}

func (sdb SqliteDb) InsertInto(t Table, values ...any) string {
	columns, exists := tableData[t]
	if !exists {
		panic("unsupported table")
	}

	if len(columns) > len(values) || len(columns) < len(values) {
		panic(fmt.Sprintf("expected %d values, but got %d", len(columns), len(values)))
	}

	sbColumns := strings.Builder{}
	var realCols []string
	for i, col := range columns {
		if values[i] == nil {
			continue
		}
		realCols = append(realCols, col)
	}
	sbColumns.WriteString(
		fmt.Sprintf("INSERT INTO %s (%s)", t, strings.Join(realCols, ",")),
	)

	sbColumns.WriteString(" VALUES (")
	valTemplate := "%v,"
	for _, val := range values {
		switch val.(type) {
		case string, Period:
			valTemplate = "'%v',"
		case nil:
			continue
		}
		sbColumns.WriteString(fmt.Sprintf(valTemplate, val))
	}

	return sbColumns.String()[:len(sbColumns.String())-1] + ")"
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
