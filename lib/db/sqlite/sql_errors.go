package sqlite

import (
	"errors"
	"fmt"
)

var (
	ErrForeignKey          = fmt.Errorf("foreign key failed validation")
	ErrDueDayInvalid       = fmt.Errorf("failed to validate due_day constraint")
	ErrTransferTypeInvalid = fmt.Errorf("failed to validate transfer_type constraint")
	ErrAmountInvalid       = fmt.Errorf("failed to validate amount constraint")
	ErrMonthInvalid        = fmt.Errorf("failed to validate month constraint")
	ErrUniqueName          = fmt.Errorf("failed unique 'name' constraint requirement")

	ErrMonthNotFound             = errors.New(setNotFound("months"))
	ErrTransfersNotFound         = errors.New(setNotFound("transfers"))
	ErrBillsNotFound             = errors.New(setNotFound("bills"))
	ErrBillHistoryNotFound       = errors.New(setNotFound("bill history"))
	ErrCreditCardsNotFound       = errors.New(setNotFound("credit cards"))
	ErrCreditCardHistoryNotFound = errors.New(setNotFound("credit card history"))
	ErrIncomeNotFound            = errors.New(setNotFound("income"))
	ErrIncomeHistoryNotFound     = errors.New(setNotFound("income history"))
	ErrAffixedIncomeNotFound     = errors.New(setNotFound("affixed income"))
)

func setNotFound(s string) string {
	return fmt.Sprintf("could not find any %s matching that query", s)
}
