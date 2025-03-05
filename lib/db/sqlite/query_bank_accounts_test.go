package sqlite

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBankAccount(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should   string
		actual   []BankAccountRecord
		expected []BankAccountRecord
		password *string
	}

	table := []MockTable{
		{
			should:   "save without account number or notes",
			actual:   []BankAccountRecord{{Name: "test"}},
			expected: []BankAccountRecord{{ID: 1, Name: "test"}},
			password: nil,
		},
		{
			should: "save a bunch of records",
			actual: []BankAccountRecord{
				{Name: "test"},
				{Name: "test1"},
				{Name: "test2"},
				{Name: "test3"},
				{Name: "test4"},
				{Name: "test5"},
			},
			expected: []BankAccountRecord{
				{ID: 1, Name: "test"},
				{ID: 2, Name: "test1"},
				{ID: 3, Name: "test2"},
				{ID: 4, Name: "test3"},
				{ID: 5, Name: "test4"},
				{ID: 6, Name: "test5"},
			},
			password: nil,
		},
		{
			should: "save account number and notes",
			actual: []BankAccountRecord{
				{
					Name:          "test",
					AccountNumber: utils.NewPointer("282841"),
					Notes:         utils.NewPointer("some notes"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: utils.NewPointer("282841"),
					Notes:         utils.NewPointer("some notes"),
				},
			},
			password: utils.NewPointer("test"),
		},
		{
			should: "just save account number",
			actual: []BankAccountRecord{
				{
					Name:          "test",
					AccountNumber: utils.NewPointer("1337420"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: utils.NewPointer("1337420"),
				},
			},
			password: utils.NewPointer("test"),
		},
		{
			should: "just save notes",
			actual: []BankAccountRecord{
				{
					Name:  "test",
					Notes: utils.NewPointer("some notes"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:    1,
					Name:  "test",
					Notes: utils.NewPointer("some notes"),
				},
			},
			password: utils.NewPointer("test"),
		},
		{
			should: "get encoded versions of protected fields",
			actual: []BankAccountRecord{
				{
					Name:          "test",
					AccountNumber: utils.NewPointer("1337420"),
					Notes:         utils.NewPointer("sevenCh"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:   1,
					Name: "test",
				},
			},
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)
			dir := t.TempDir()

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			for _, acct := range mock.actual {
				err = db.CreateBankAccount(acct, utils.NewPointer("test"))
				r.NoError(err)
			}

			res, err := db.QueryBankAccounts(QueryMap{}, mock.password)
			r.NoError(err)
			r.Len(res, len(mock.expected))
			for i, r := range res {
				if mock.password == nil {
					a.Equal(r.ID, mock.expected[i].ID)
					a.Equal(r.Name, mock.expected[i].Name)
					if r.AccountNumber != nil {
						a.True(isProbablyBase64(*res[0].AccountNumber))
					}
					if r.Notes != nil {
						a.True(isProbablyBase64(*res[0].Notes))
					}
					return
				}
				a.Equal(mock.expected, res)
			}
		})
	}

	t.Run("should error when passing nil password & sensitive data", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		r := require.New(t)
		dir := t.TempDir()

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBankAccount(BankAccountRecord{
			Name:          "Test",
			AccountNumber: utils.NewPointer("1823842"),
		}, nil)
		a.ErrorIs(err, lib.ErrEncryptWithoutPassword)
	})
}

func TestBankAccountHistory(t *testing.T) {
	type MockTable struct {
		should      string
		accounts    []BankAccountRecord
		actual      []BankHistoryConfig
		expected    []BankHistoryRecord
		expectError error
	}

	table := []MockTable{
		{
			should: "create bank account history",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1, Balance: lib.NewCurrency("133.7", lib.USD)},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       lib.NewCurrency("133.7", lib.USD),
				},
			},
		},
		{
			should: "create multiple bank histories",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
				{Name: "DaddyBank"},
				{Name: "BigBank"},
				{Name: "1337Bank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 3, Balance: lib.NewCurrency("7242.31", lib.USD)},
				{MonthID: 1, BankAccountID: 1, Balance: lib.NewCurrency("13.37", lib.USD)},
				{MonthID: 1, BankAccountID: 4, Balance: lib.NewCurrency("1337.69", lib.USD)},
				{MonthID: 1, BankAccountID: 2, Balance: lib.NewCurrency("80.08", lib.USD)},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 3,
					Balance:       lib.NewCurrency("7242.31", lib.USD),
				},
				{
					ID:            2,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       lib.NewCurrency("13.37", lib.USD),
				},
				{
					ID:            3,
					MonthID:       1,
					BankAccountID: 4,
					Balance:       lib.NewCurrency("1337.69", lib.USD),
				},
				{
					ID:            4,
					MonthID:       1,
					BankAccountID: 2,
					Balance:       lib.NewCurrency("80.08", lib.USD),
				},
			},
		},
		{
			should: "fail month constraint",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 2, BankAccountID: 1, Balance: lib.NewCurrency("133.7", lib.USD)},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "fail account constraint",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 2, Balance: lib.NewCurrency("133.7", lib.USD)},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "default balance to 0",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},
			actual: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       lib.NewCurrency("0", lib.USD),
				},
			},
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)
			dir := t.TempDir()

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			db.CreateMonth(NewMonth(2024, time.January))

			for _, acct := range mock.accounts {
				err = db.CreateBankAccount(acct, nil)
				r.NoError(err)
			}

			for _, history := range mock.actual {
				if mock.expectError != nil {
					err = db.CreateBankAccountHistory(history)
					a.ErrorIs(err, ErrForeignKey)
					return
				}
				err = db.CreateBankAccountHistory(history)
				a.NoError(err)
			}

			res, err := db.QueryBankAccountHistory(QueryMap{})
			r.NoError(err)
			a.Equal(mock.expected, res)
		})
	}
}

func TestBankTransfers(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should      string
		accounts    []BankAccountRecord
		history     []BankHistoryConfig
		actual      []TransferConfig
		expected    []TransferRecord
		expectError error
	}

	table := []MockTable{
		{
			should:   "record a transfer to a specific bank history",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferConfig{
				{
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       lib.NewCurrency("72.28", lib.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
					ToWhom:       utils.NewPointer("johnny"),
					FromWhom:     utils.NewPointer("bank of america"),
				},
			},
			expected: []TransferRecord{
				{
					ID: 1,
					TransferConfig: TransferConfig{
						HistoryID:    1,
						MonthID:      1,
						Name:         "test",
						Amount:       lib.NewCurrency("72.28", lib.USD),
						DueDay:       5,
						TransferType: DEPOSIT,
						ToWhom:       utils.NewPointer("johnny"),
						FromWhom:     utils.NewPointer("bank of america"),
					},
				},
			},
		},
		{
			should:   "allow nullable fields to be nil",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferConfig{
				{
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       lib.NewCurrency("72.28", lib.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
				},
			},
			expected: []TransferRecord{
				{
					ID: 1,
					TransferConfig: TransferConfig{
						HistoryID:    1,
						MonthID:      1,
						Name:         "test",
						Amount:       lib.NewCurrency("72.28", lib.USD),
						DueDay:       5,
						TransferType: DEPOSIT,
						ToWhom:       nil,
						FromWhom:     nil,
					},
				},
			},
		},
		{
			should:   "panic on foreign key constraint violations",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferConfig{
				{
					HistoryID:    1,
					MonthID:      2,
					DueDay:       5,
					TransferType: DEPOSIT,
				},
				{
					HistoryID:    2,
					MonthID:      1,
					DueDay:       5,
					TransferType: DEPOSIT,
				},
			},
			expectError: ErrForeignKey,
		},
		{
			should:   "panic on due date constraint violations",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferConfig{
				{DueDay: 32},
				{DueDay: 0},
			},
			expectError: ErrDueDayInvalid,
		},
		{
			should:   "panic on transfer type constraint violations",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferConfig{
				{DueDay: 1, TransferType: "not a good type"},
				{DueDay: 1, TransferType: "withdrawals"},
			},
			expectError: ErrTransferTypeInvalid,
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := assert.New(t)
			dir := t.TempDir()

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			_, err = db.CreateMonth(NewMonth(2024, time.January))
			r.NoError(err)

			for _, acct := range mock.accounts {
				err = db.CreateBankAccount(acct, nil)
				r.NoError(err)
			}

			for _, history := range mock.history {
				err = db.CreateBankAccountHistory(history)
				r.NoError(err)
			}

			for _, transfer := range mock.actual {
				if mock.expectError != nil {
					err = db.CreateTransfer(transfer)
					a.ErrorIs(err, mock.expectError)
				} else {
					err = db.CreateTransfer(transfer)
					r.NoError(err)
				}
			}

			// Error tests have no expected values
			if mock.expectError != nil {
				return
			}

			res, err := db.QueryTransfers(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}
}

func isProbablyBase64(s string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)
	if !re.MatchString(s) {
		return false
	}

	// base64 strings are always in multiples of 4
	if len(s)%4 != 0 {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
