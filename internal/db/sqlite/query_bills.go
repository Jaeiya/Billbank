package sqlite

import (
	"github.com/jaeiya/billbank/internal"
)

type BillRecord struct {
	ID      int
	TypeID  int
	Name    string
	Amount  internal.Currency
	DueDate internal.Date
	Status  string
	Period  Period
}

func (db SqliteDb) CreateNewBills(records []BillRecord) error {
	_, err := insertMultiInto(db, BILLS, records, func(r BillRecord) []any {
		return []any{r.TypeID, r.Name, r.Amount, r.DueDate, r.Status, r.Period}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryBills(qm QueryMap) ([]BillRecord, error) {
	rows, err := db.query(BILLS, qm)
	if err != nil {
		return []BillRecord{}, err
	}

	// var amount int
	var records []BillRecord
	for rows.Next() {
		var record BillRecord
		if err := rows.Scan(
			&record.ID,
			&record.TypeID,
			&record.Name,
			&record.Amount,
			&record.DueDate,
			&record.Status,
			&record.Period,
		); err != nil {
			return []BillRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BillRecord{}, ErrBillsNotFound
	}

	return records, nil
}

func (db SqliteDb) QueryBillsCount() (int, error) {
	var count int
	if err := db.handle.QueryRow("SELECT COUNT(*) FROM " + string(BILLS)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

type BillHistoryRecord struct {
	ID         int
	MonthID    int
	TypeID     int
	Name       string
	Amount     internal.Currency
	DueDate    internal.Date
	PaidAmount *internal.Currency
	PaidDate   *internal.Date
	PaidHow    *string
	ClearedDay *int
	Notes      *string
}

func (db SqliteDb) CreateBillHistory(records []BillHistoryRecord) error {
	_, err := insertMultiInto(db, BILLS_HISTORY, records, func(r BillHistoryRecord) []any {
		return []any{
			r.MonthID,
			r.TypeID,
			r.Name,
			r.Amount,
			r.DueDate,
			r.PaidAmount,
			r.PaidDate,
			r.PaidHow,
			r.ClearedDay,
			r.Notes,
		}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRecord, error) {
	rows, err := db.query(BILLS_HISTORY, qm)
	if err != nil {
		return []BillHistoryRecord{}, err
	}

	var records []BillHistoryRecord

	for rows.Next() {
		var record BillHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.MonthID,
			&record.TypeID,
			&record.Name,
			&record.Amount,
			&record.DueDate,
			&record.PaidAmount,
			&record.PaidDate,
			&record.PaidHow,
			&record.ClearedDay,
			&record.Notes,
		); err != nil {
			return []BillHistoryRecord{}, err
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, ErrBillHistoryNotFound
	}

	return records, nil
}

func (db SqliteDb) CreateBillTypes(names []string) error {
	_, err := insertMultiInto(db, BILL_TYPES, names, func(name string) []any {
		return []any{name}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryBillTypes() (types []string, err error) {
	rows, err := db.queryAll(BILL_TYPES)
	if err != nil {
		return []string{}, err
	}
	var name *string
	var id *int
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			return []string{}, err
		}
		types = append(types, *name)
	}
	return types, nil
}

type MonthlyBill struct {
	// Is ignored when creating
	ID       int
	BillID   int
	IsActive bool
}

func (db SqliteDb) CreateMonthlyBills(bills []MonthlyBill) error {
	_, err := insertMultiInto(db, BILLS_MONTHLY, bills, func(b MonthlyBill) []any {
		return []any{b.BillID, b.IsActive}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (db SqliteDb) QueryMonthlyBills() ([]MonthlyBill, error) {
	rows, err := db.queryAll(BILLS_MONTHLY)
	if err != nil {
		return []MonthlyBill{}, err
	}

	var id, billID int
	var isActive bool
	monthlyBills := make([]MonthlyBill, 0, 20)

	for rows.Next() {
		if err := rows.Scan(&id, &billID, &isActive); err != nil {
			return []MonthlyBill{}, err
		}
		monthlyBills = append(
			monthlyBills,
			MonthlyBill{id, billID, isActive},
		)
	}

	return monthlyBills, nil
}
