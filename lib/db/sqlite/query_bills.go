package sqlite

import (
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
)

type BillRecord struct {
	ID     int
	TypeID int
	Name   string
	Amount lib.Currency
	DueDay int
	Period Period
}

func (sdb SqliteDb) CreateNewBills(records []BillRecord) error {
	_, err := insertMultiInto(sdb, BILLS, records, func(r BillRecord) []any {
		return []any{r.TypeID, r.Name, r.Amount, r.DueDay, r.Period}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBills(qm QueryMap) ([]BillRecord, error) {
	rows, err := sdb.query(BILLS, qm)
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
			&record.DueDay,
			&record.Period,
		); err != nil {
			return []BillRecord{}, err
		}
		// record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BillRecord{}, ErrBillsNotFound
	}

	return records, nil
}

type BillHistoryRecord struct {
	ID         int
	MonthID    int
	TypeID     int
	Name       string
	Amount     lib.Currency
	DueDay     int
	PaidAmount *lib.Currency
	PaidDay    *int
	PaidHow    *string
	ClearedDay *int
	Notes      *string
}

func (sdb SqliteDb) CreateBillHistory(records []BillHistoryRecord) error {
	_, err := insertMultiInto(sdb, BILLS_HISTORY, records, func(r BillHistoryRecord) []any {
		return []any{
			r.MonthID,
			r.TypeID,
			r.Name,
			r.Amount,
			r.DueDay,
			r.PaidAmount,
			utils.TryDeref(r.PaidDay),
			utils.TryDeref(r.PaidHow),
			utils.TryDeref(r.ClearedDay),
			utils.TryDeref(r.Notes),
		}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRecord, error) {
	rows, err := sdb.query(BILLS_HISTORY, qm)
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
			&record.DueDay,
			&record.PaidAmount,
			&record.PaidDay,
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

func (sdb SqliteDb) CreateBillTypes(names []string) error {
	_, err := insertMultiInto(sdb, BILL_TYPES, names, func(name string) []any {
		return []any{name}
	})
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBillTypes() (types []string, err error) {
	rows, err := sdb.queryAll(BILL_TYPES)
	if err != nil {
		return []string{}, err
	}
	var name *string
	var id *int
	for rows.Next() {
		if err = rows.Scan(
			&id,
			&name,
		); err != nil {
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

func (sdb SqliteDb) CreateMonthlyBills(bills []MonthlyBill) error {
	insStr := toInsertMultiStr(string(BILLS_MONTHLY), tableData[BILLS_MONTHLY], len(bills))
	execArgs := make([]any, 0, len(bills)*2)
	for _, bill := range bills {
		execArgs = append(execArgs, bill.BillID, bill.IsActive)
	}
	_, err := sdb.handle.Exec(insStr, execArgs...)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryMonthlyBills() ([]MonthlyBill, error) {
	rows, err := sdb.queryAll(BILLS_MONTHLY)
	if err != nil {
		return []MonthlyBill{}, err
	}

	var id, billID int
	var isActive bool
	monthlyBills := make([]MonthlyBill, 0, 20)

	for rows.Next() {
		rows.Scan(&id, &billID, &isActive)
		monthlyBills = append(
			monthlyBills,
			MonthlyBill{id, billID, isActive},
		)
	}

	return monthlyBills, nil
}
