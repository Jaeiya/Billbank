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
		decrypt  bool
	}

	table := []MockTable{
		{
			should:   "save without account number or notes",
			actual:   []BankAccountConfig{{Name: "test"}},
			expected: []BankRecord{{ID: 1, Name: "test"}},
			decrypt:  false,
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
			decrypt: false,
		},
		{
			should: "save account number and notes",
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.GetPointer("test"),
					AccountNumber: lib.GetPointer("282841"),
					Notes:         lib.GetPointer("some notes"),
				},
			},
			expected: []BankRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: lib.GetPointer("282841"),
					Notes:         lib.GetPointer("some notes"),
				},
			},
			decrypt: true,
		},
		{
			should: "save just account number",
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.GetPointer("test"),
					AccountNumber: lib.GetPointer("1337420"),
				},
			},
			expected: []BankRecord{
				{
					ID:            1,
					Name:          "test",
					AccountNumber: lib.GetPointer("1337420"),
				},
			},
			decrypt: true,
		},
		{
			should: "save just notes",
			actual: []BankAccountConfig{
				{
					Name:     "test",
					Password: lib.GetPointer("test"),
					Notes:    lib.GetPointer("some notes"),
				},
			},
			expected: []BankRecord{
				{
					ID:    1,
					Name:  "test",
					Notes: lib.GetPointer("some notes"),
				},
			},
			decrypt: true,
		},
		{
			should: "get encoded versions of protected fields",
			actual: []BankAccountConfig{
				{
					Name:          "test",
					Password:      lib.GetPointer("test"),
					AccountNumber: lib.GetPointer("1337420"),
					Notes:         lib.GetPointer("sevenCh"),
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

			var res []BankRecord
			var err error
			if mock.decrypt {
				res, err = db.QueryDecryptedBankAccount(
					mock.expected[0].ID,
					*mock.actual[0].Password,
				)
			} else {
				res, err = db.QueryBankAccounts(QueryMap{})
			}
			r.NoError(err)
			r.Len(res, len(mock.expected))

			if !mock.decrypt {
				for i, m := range mock.expected {
					a.Equal(m.ID, res[i].ID)
					a.Equal(m.Name, res[i].Name)
					if res[i].AccountNumber != nil {
						a.True(isProbablyBase64(*res[i].AccountNumber))
					}
					if res[i].Notes != nil {
						a.True(isProbablyBase64(*res[i].Notes))
					}
				}
				return
			}
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
