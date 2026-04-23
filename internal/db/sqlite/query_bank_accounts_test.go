package sqlite

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/jaeiya/billbank/internal"
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
					AccountNumber: new("282841"),
					Notes:         new("some notes"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: new("282841"),
					Notes:         new("some notes"),
				},
			},
			password: new("test"),
		},
		{
			should: "just save account number",
			actual: []BankAccountRecord{
				{
					Name:          "test",
					AccountNumber: new("1337420"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: new("1337420"),
				},
			},
			password: new("test"),
		},
		{
			should: "just save notes",
			actual: []BankAccountRecord{
				{
					Name:  "test",
					Notes: new("some notes"),
				},
			},
			expected: []BankAccountRecord{
				{
					ID:    1,
					Name:  "test",
					Notes: new("some notes"),
				},
			},
			password: new("test"),
		},
		{
			should: "get encoded versions of protected fields",
			actual: []BankAccountRecord{
				{
					Name:          "test",
					AccountNumber: new("1337420"),
					Notes:         new("sevenCh"),
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

			db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
			r.NoError(err)
			defer db.Close()

			for _, acct := range mock.actual {
				err = db.CreateBankAccount(acct, new("test"))
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

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBankAccount(BankAccountRecord{
			Name:          "Test",
			AccountNumber: new("1823842"),
		}, nil)
		a.ErrorIs(err, internal.ErrEncryptWithoutPassword)
	})
}

func TestBankAccountHistory(t *testing.T) {
	type MockTable struct {
		should      string
		accounts    []BankAccountRecord
		actual      []BankHistoryRecord
		expected    []BankHistoryRecord
		expectError error
	}

	table := []MockTable{
		{
			should: "create bank account history",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryRecord{
				{
					MonthID:       1,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("133.7", internal.USD),
				},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("133.7", internal.USD),
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

			actual: []BankHistoryRecord{
				{
					MonthID:       1,
					BankAccountID: 3,
					Balance:       internal.NewCurrency("7242.31", internal.USD),
				},
				{
					MonthID:       1,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("13.37", internal.USD),
				},
				{
					MonthID:       1,
					BankAccountID: 4,
					Balance:       internal.NewCurrency("1337.69", internal.USD),
				},
				{
					MonthID:       1,
					BankAccountID: 2,
					Balance:       internal.NewCurrency("80.08", internal.USD),
				},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 3,
					Balance:       internal.NewCurrency("7242.31", internal.USD),
				},
				{
					ID:            2,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("13.37", internal.USD),
				},
				{
					ID:            3,
					MonthID:       1,
					BankAccountID: 4,
					Balance:       internal.NewCurrency("1337.69", internal.USD),
				},
				{
					ID:            4,
					MonthID:       1,
					BankAccountID: 2,
					Balance:       internal.NewCurrency("80.08", internal.USD),
				},
			},
		},
		{
			should: "fail month constraint",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryRecord{
				{
					MonthID:       2,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("133.7", internal.USD),
				},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "fail account constraint",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},

			actual: []BankHistoryRecord{
				{
					MonthID:       1,
					BankAccountID: 2,
					Balance:       internal.NewCurrency("133.7", internal.USD),
				},
			},
			expectError: ErrForeignKey,
		},
		{
			should: "default balance to 0",
			accounts: []BankAccountRecord{
				{Name: "TestBank"},
			},
			actual: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			expected: []BankHistoryRecord{
				{
					ID:            1,
					MonthID:       1,
					BankAccountID: 1,
					Balance:       internal.NewCurrency("0", internal.USD),
				},
			},
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)

			db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
			r.NoError(err)
			defer db.Close()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())

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
		history     []BankHistoryRecord
		actual      []TransferRecord
		expected    []TransferRecord
		expectError error
	}

	table := []MockTable{
		{
			should:   "record a transfer to a specific bank history",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferRecord{
				{
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       internal.NewCurrency("72.28", internal.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
					ToWhom:       new("johnny"),
					FromWhom:     new("bank of america"),
				},
			},
			expected: []TransferRecord{
				{
					ID:           1,
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       internal.NewCurrency("72.28", internal.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
					ToWhom:       new("johnny"),
					FromWhom:     new("bank of america"),
				},
			},
		},
		{
			should:   "allow nullable fields to be nil",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferRecord{
				{
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       internal.NewCurrency("72.28", internal.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
				},
			},
			expected: []TransferRecord{
				{
					ID:           1,
					HistoryID:    1,
					MonthID:      1,
					Name:         "test",
					Amount:       internal.NewCurrency("72.28", internal.USD),
					DueDay:       5,
					TransferType: DEPOSIT,
					ToWhom:       nil,
					FromWhom:     nil,
				},
			},
		},
		{
			should:   "panic on foreign key constraint violations",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferRecord{
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
			history: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferRecord{
				{DueDay: 32},
				{DueDay: 0},
			},
			expectError: ErrDueDayInvalid,
		},
		{
			should:   "panic on transfer type constraint violations",
			accounts: []BankAccountRecord{{Name: "Test"}},
			history: []BankHistoryRecord{
				{MonthID: 1, BankAccountID: 1},
			},
			actual: []TransferRecord{
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

			db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
			r.NoError(err)
			defer db.Close()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())
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
