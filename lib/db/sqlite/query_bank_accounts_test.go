package sqlite

import (
	"encoding/base64"
	"path/filepath"
	"regexp"
	"testing"

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
			should: "save just account number",
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
			should: "save just notes",
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
