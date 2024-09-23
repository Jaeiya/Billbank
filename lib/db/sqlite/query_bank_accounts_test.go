package sqlite

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBankAccount(t *testing.T) {
	type MockTable struct {
		should   string
		actual   []BankAccountConfig
		expected []BankRecord
		password *string
	}

	table := []MockTable{
		{
			should:   "save without account number or notes",
			actual:   []BankAccountConfig{{Name: "test"}},
			expected: []BankRecord{{ID: 1, Name: "test"}},
			password: nil,
		},
		{
			should: "save a bunch of records",
			actual: []BankAccountConfig{
				{Name: "test"},
				{Name: "test1"},
				{Name: "test2"},
				{Name: "test3"},
				{Name: "test4"},
				{Name: "test5"},
			},
			expected: []BankRecord{
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
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.NewPointer("test"),
					AccountNumber: lib.NewPointer("282841"),
					Notes:         lib.NewPointer("some notes"),
				},
			},
			expected: []BankRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: lib.NewPointer("282841"),
					Notes:         lib.NewPointer("some notes"),
				},
			},
			password: lib.NewPointer("test"),
		},
		{
			should: "just save account number",
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.NewPointer("test"),
					AccountNumber: lib.NewPointer("1337420"),
				},
			},
			expected: []BankRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: lib.NewPointer("1337420"),
				},
			},
			password: lib.NewPointer("test"),
		},
		{
			should: "just save notes",
			actual: []BankAccountConfig{
				{
					Name:     "test",
					Password: lib.NewPointer("test"),
					Notes:    lib.NewPointer("some notes"),
				},
			},
			expected: []BankRecord{
				{
					ID:    1,
					Name:  "test",
					Notes: lib.NewPointer("some notes"),
				},
			},
			password: lib.NewPointer("test"),
		},
		{
			should: "get encoded versions of protected fields",
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.NewPointer("test"),
					AccountNumber: lib.NewPointer("1337420"),
					Notes:         lib.NewPointer("sevenCh"),
				},
			},
			expected: []BankRecord{
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

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			for _, acct := range mock.actual {
				db.CreateBankAccount(acct)
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
		dir := t.TempDir()

		db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		defer db.Close()

		a.PanicsWithValue(lib.ErrEncryptWithoutPassword, func() {
			db.CreateBankAccount(BankAccountConfig{
				Name:          "Test",
				AccountNumber: lib.NewPointer("1823842"),
			})
		})
	})
}

func TestBankAccountHistory(t *testing.T) {
	type MockTable struct {
		should      string
		accounts    []BankAccountConfig
		actual      []BankHistoryConfig
		expected    []BankHistoryRecord
		expectError error
	}

	table := []MockTable{
		{
			should: "create bank account history",
			accounts: []BankAccountConfig{
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
			accounts: []BankAccountConfig{
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
			accounts: []BankAccountConfig{
				{Name: "TestBank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 2, BankAccountID: 1, Balance: lib.NewCurrency("133.7", lib.USD)},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "fail account constraint",
			accounts: []BankAccountConfig{
				{Name: "TestBank"},
			},

			actual: []BankHistoryConfig{
				{MonthID: 1, BankAccountID: 2, Balance: lib.NewCurrency("133.7", lib.USD)},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "default balance to 0",
			accounts: []BankAccountConfig{
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

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			err := db.CreateMonth(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local))
			r.NoError(err)

			for _, acct := range mock.accounts {
				db.CreateBankAccount(acct)
			}

			for _, history := range mock.actual {
				if mock.expectError != nil {
					a.PanicsWithValue(ErrForeignKey, func() {
						db.CreateBankAccountHistory(history)
					})
					return
				}
				db.CreateBankAccountHistory(history)
			}

			res, err := db.QueryBankAccountHistory(QueryMap{})
			r.NoError(err)
			a.Equal(mock.expected, res)
		})
	}
}

func TestBankTransfers(t *testing.T) {
	type MockTable struct {
		should      string
		accounts    []BankAccountConfig
		history     []BankHistoryConfig
		actual      []TransferConfig
		expected    []TransferRecord
		expectError error
	}

	table := []MockTable{
		{
			should:   "record a transfer to a specific bank history",
			accounts: []BankAccountConfig{{Name: "Test"}},
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
					ToWhom:       lib.NewPointer("johnny"),
					FromWhom:     lib.NewPointer("bank of america"),
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
						ToWhom:       lib.NewPointer("johnny"),
						FromWhom:     lib.NewPointer("bank of america"),
					},
				},
			},
		},
		{
			should:   "allow nullable fields to be nil",
			accounts: []BankAccountConfig{{Name: "Test"}},
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
			accounts: []BankAccountConfig{{Name: "Test"}},
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
			accounts: []BankAccountConfig{{Name: "Test"}},
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
			accounts: []BankAccountConfig{{Name: "Test"}},
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

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			db.CreateMonth(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local))

			for _, acct := range mock.accounts {
				db.CreateBankAccount(acct)
			}

			for _, history := range mock.history {
				db.CreateBankAccountHistory(history)
			}

			for _, transfer := range mock.actual {
				if mock.expectError != nil {
					a.PanicsWithValue(mock.expectError, func() {
						db.CreateTransfer(transfer)
					})
				} else {
					db.CreateTransfer(transfer)
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
