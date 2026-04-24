package sqlite

import (
	"fmt"
	"strings"

	"github.com/jaeiya/billbank/internal"
	"github.com/jaeiya/billbank/internal/utils"
)

type (
	CCField    string
	CCFieldMap map[CCField]any
)

const (
	CC_BALANCE     = CCField("balance")
	CC_LIMIT       = CCField("credit_limit")
	CC_PAID_AMOUNT = CCField("paid_amount")
	CC_PAID_DAY    = CCField("paid_day")
	CC_DUE_DAY     = CCField("due_day")
)

type CreditCardRecord struct {
	ID             int
	Name           string
	DueDay         int
	CreditLimit    *internal.Currency
	CardNumber     *string
	LastFourDigits string
	Notes          *string
}

type CreditCardHistoryRecord struct {
	ID           int
	CreditCardID int
	MonthID      int
	Balance      internal.Currency
	CreditLimit  *internal.Currency
	PaidAmount   *internal.Currency
	PaidDay      *int
	DueDay       int
	ClearedDay   *int
}

func (sdb SqliteDb) CreateCreditCard(config CreditCardRecord, pass *string) error {
	encCardNum, err := internal.EncryptNonNil(config.CardNumber, pass)
	if err != nil {
		return err
	}

	encNotes, err := internal.EncryptNonNil(config.Notes, pass)
	if err != nil {
		return err
	}

	_, err = sdb.insertInto(
		CREDIT_CARDS,
		config.Name,
		config.DueDay,
		config.CreditLimit,
		encCardNum,
		config.LastFourDigits,
		encNotes,
	)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryCreditCards(
	qm QueryMap,
	password *string,
) ([]CreditCardRecord, error) {
	rows, err := sdb.query(CREDIT_CARDS, qm)
	if err != nil {
		return []CreditCardRecord{}, err
	}

	var records []CreditCardRecord
	for rows.Next() {
		var record CreditCardRecord
		var err error

		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.DueDay,
			&record.CreditLimit,
			&record.CardNumber,
			&record.LastFourDigits,
			&record.Notes,
		); err != nil {
			return []CreditCardRecord{}, err
		}

		if password != nil && record.CardNumber != nil {
			if record.CardNumber, err = internal.DecryptNonNil(record.CardNumber, *password); err != nil {
				return []CreditCardRecord{}, err
			}
		}

		if password != nil && record.Notes != nil {
			if record.Notes, err = internal.DecryptNonNil(record.Notes, *password); err != nil {
				return []CreditCardRecord{}, err
			}
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []CreditCardRecord{}, ErrCreditCardsNotFound
	}

	return records, nil
}

func (sdb SqliteDb) CreateCreditCardHistory(r CreditCardHistoryRecord) error {
	_, err := sdb.insertInto(
		CREDIT_CARD_HISTORY,
		r.CreditCardID,
		r.MonthID,
		r.Balance.GetStoredValue(),
		r.DueDay,
		r.CreditLimit,
		r.PaidDay,
		r.PaidAmount,
		r.ClearedDay,
	)
	if err != nil {
		return getExecError(err)
	}
	return nil
}

func (sdb SqliteDb) QueryCreditCardHistory(qm QueryMap) ([]CreditCardHistoryRecord, error) {
	rows, err := sdb.query(CREDIT_CARD_HISTORY, qm)
	if err != nil {
		return []CreditCardHistoryRecord{}, err
	}

	var records []CreditCardHistoryRecord

	for rows.Next() {
		var record CreditCardHistoryRecord

		if err := rows.Scan(
			&record.ID,
			&record.CreditCardID,
			&record.MonthID,
			&record.Balance,
			&record.DueDay,
			&record.CreditLimit,
			&record.PaidDay,
			&record.PaidAmount,
			&record.ClearedDay,
		); err != nil {
			return []CreditCardHistoryRecord{}, err
		}

		records = append(records, record)
	}

	if len(records) == 0 {
		return []CreditCardHistoryRecord{}, ErrCreditCardHistoryNotFound
	}

	return records, nil
}

func (sdb SqliteDb) SetCreditCardHistory(historyID int, fieldMap CCFieldMap) error {
	conditions := make([]string, 0, len(fieldMap))
	for field, value := range fieldMap {
		switch field {

		case CC_BALANCE, CC_LIMIT, CC_PAID_AMOUNT:
			c, err := internal.ToCurrency(value)
			if err != nil {
				return fmt.Errorf("%s should be of type: internal.Currency", field)
			}
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, c.GetStoredValue()))

		case CC_DUE_DAY, CC_PAID_DAY:
			if !utils.IsInt(value) {
				return fmt.Errorf("%s should of of type: int", field)
			}
			conditions = append(conditions, fmt.Sprintf("%s=%d", field, value))

		default:
			return fmt.Errorf("unsupported credit card history field: %s", field)
		}
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = %d",
		CREDIT_CARD_HISTORY,
		strings.Join(conditions, ","),
		historyID,
	)

	if _, err := sdb.handle.Exec(query); err != nil {
		return err
	}
	return nil
}
