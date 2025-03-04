package sqlite

import (
	"fmt"
	"strings"

	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
)

type BillsConfig struct {
	TypeID int
	Name   string
	Amount lib.Currency
	DueDay int
	Period Period
}

type BillRecord struct {
	ID int
	BillsConfig
}

type BillHistoryConfig struct {
	MonthID      int
	TypeID       int
	Name         string
	Amount       lib.Currency
	DueDay       int
	PaidAmount   *lib.Currency
	PaidDay      *int
	PaidHow      *string
	ClearedOnDay *int
	Notes        *string
}

type BillHistoryRecord struct {
	ID int
	BillHistoryConfig
}

func (sdb SqliteDb) CreateNewBill(cfg BillsConfig) error {
	insStr, err := sdb.ToInsertIntoStr(
		BILLS,
		cfg.TypeID,
		cfg.Name,
		cfg.Amount.GetStoredValue(),
		cfg.DueDay,
		cfg.Period,
	)
	if err != nil {
		return err
	}

	if _, err := sdb.handle.Exec(insStr); err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBills(qm QueryMap) ([]BillRecord, error) {
	rows, err := sdb.query(BILLS, qm)
	if err != nil {
		return []BillRecord{}, err
	}

	var amount int
	var records []BillRecord
	for rows.Next() {
		var record BillRecord
		if err := rows.Scan(
			&record.ID,
			&record.TypeID,
			&record.Name,
			&amount,
			&record.DueDay,
			&record.Period,
		); err != nil {
			return []BillRecord{}, err
		}
		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)
		records = append(records, record)
	}

	if len(records) == 0 {
		return []BillRecord{}, ErrBillsNotFound
	}

	return records, nil
}

func (sdb SqliteDb) CreateBillHistory(cfg BillHistoryConfig) error {
	paidAmount := utils.TryDeref(cfg.PaidAmount)
	if paidAmount != nil {
		paidAmount = cfg.PaidAmount.GetStoredValue()
	}

	insStr, err := sdb.ToInsertIntoStr(
		BILLS_HISTORY,
		cfg.MonthID,
		cfg.TypeID,
		cfg.Name,
		cfg.Amount.GetStoredValue(),
		cfg.DueDay,
		paidAmount,
		utils.TryDeref(cfg.PaidDay),
		utils.TryDeref(cfg.PaidHow),
		utils.TryDeref(cfg.ClearedOnDay),
		utils.TryDeref(cfg.Notes),
	)
	if err != nil {
		return err
	}

	if _, err := sdb.handle.Exec(insStr); err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryBillHistory(qm QueryMap) ([]BillHistoryRecord, error) {
	rows, err := sdb.query(BILLS_HISTORY, qm)
	if err != nil {
		return []BillHistoryRecord{}, err
	}

	var amount int
	var paidAmount *int
	var records []BillHistoryRecord

	for rows.Next() {
		var record BillHistoryRecord
		if err := rows.Scan(
			&record.ID,
			&record.MonthID,
			&record.TypeID,
			&record.Name,
			&amount,
			&record.DueDay,
			&paidAmount,
			&record.PaidDay,
			&record.PaidHow,
			&record.ClearedOnDay,
			&record.Notes,
		); err != nil {
			return []BillHistoryRecord{}, err
		}

		record.Amount = lib.NewCurrencyFromStore(amount, sdb.currencyCode)

		if paidAmount != nil {
			pa := lib.NewCurrencyFromStore(*paidAmount, sdb.currencyCode)
			record.PaidAmount = &pa
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, ErrBillHistoryNotFound
	}

	return records, nil
}

func (sdb SqliteDb) CreateBillTypes(names []string) error {
	var sb strings.Builder
	sb.WriteString("INSERT INTO bill_types (name) VALUES ")
	for _, n := range names {
		sb.WriteString(fmt.Sprintf("('%s'),", n))
	}
	insStr := sb.String()
	_, err := sdb.handle.Exec(insStr[:len(insStr)-1] + ";")
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
	var sb strings.Builder
	sb.WriteString("INSERT INTO bills_monthly (bill_id, is_active) VALUES ")
	for _, cfg := range bills {
		isActiveStr := "FALSE"
		if cfg.IsActive {
			isActiveStr = "TRUE"
		}
		sb.WriteString(fmt.Sprintf("(%d, '%s'),", cfg.BillID, isActiveStr))
	}
	insStr := sb.String()
	_, err := sdb.handle.Exec(insStr[:len(insStr)-1] + ";")
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
