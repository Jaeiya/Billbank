package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCreditCards(t *testing.T) {
	type MockTable struct {
		should        string
		actual        []CreditCardConfig
		expected      []CreditCardRecord
		password      *string
		expectedError error
	}

	table := []MockTable{
		{
			should: "create credit card records",
			actual: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     lib.NewPointer("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          lib.NewPointer("some notes"),
					Password:       lib.NewPointer("password"),
				},
			},
			expected: []CreditCardRecord{
				{
					ID:             1,
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     lib.NewPointer("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          lib.NewPointer("some notes"),
				},
			},
			password: lib.NewPointer("password"),
		},
		{
			should: "nullable values should be nil",
			actual: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     lib.NewPointer("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Password:       lib.NewPointer("password"),
				},
				{
					Name:           "test2",
					DueDay:         7,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "0023",
					Notes:          lib.NewPointer("some notes"),
					Password:       lib.NewPointer("password"),
				},
				{
					Name:           "test3",
					DueDay:         8,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
					Password:       lib.NewPointer("password"),
				},
			},
			expected: []CreditCardRecord{
				{
					ID:             1,
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     lib.NewPointer("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          nil,
				},
				{
					ID:             2,
					Name:           "test2",
					DueDay:         7,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "0023",
					Notes:          lib.NewPointer("some notes"),
					CardNumber:     nil,
				},
				{
					ID:             3,
					Name:           "test3",
					DueDay:         8,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
					CardNumber:     nil,
					Notes:          nil,
				},
			},
			password: lib.NewPointer("password"),
		},
		{
			should: "panic on due day constraint violation",
			actual: []CreditCardConfig{
				{
					Name:   "test",
					DueDay: 0,
				},
				{
					Name:   "test2",
					DueDay: 32,
				},
			},
			expectedError: ErrDueDayInvalid,
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			a := assert.New(t)
			r := require.New(t)

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			if mock.expectedError != nil {
				for _, cardConfig := range mock.actual {
					a.PanicsWithValue(ErrDueDayInvalid, func() {
						db.CreateCreditCard(cardConfig)
					})
				}
				return
			}

			for _, cardConfig := range mock.actual {
				db.CreateCreditCard(cardConfig)
			}

			res, err := db.QueryCreditCards(QueryMap{}, mock.password)
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}

	t.Run("should panic on unique name constraint violation", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		a := assert.New(t)

		db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		defer db.Close()

		db.CreateCreditCard(CreditCardConfig{
			Name:           "test",
			DueDay:         5,
			LastFourDigits: "1234",
		})

		a.PanicsWithValue(ErrUniqueName, func() {
			db.CreateCreditCard(CreditCardConfig{
				Name:           "test",
				DueDay:         5,
				LastFourDigits: "1234",
			})
		})
	})
}

func TestCreateCreditCardHistory(t *testing.T) {
	type MockTable struct {
		should        string
		cards         []CreditCardConfig
		actual        []CreditCardHistoryConfig
		expected      []CardHistoryRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "create credit card history records",
			cards: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryConfig{
				{
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					CreditLimit:  lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					DueDay:       5,
				},
			},

			expected: []CardHistoryRecord{
				{
					ID:           1,
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					CreditLimit:  lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					DueDay:       5,
					PaidAmount:   lib.NewCurrency("0", lib.USD),
					Period:       MONTHLY,
				},
			},
		},
		{
			should: "set nullable fields to nil",
			cards: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    lib.NewPointer(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryConfig{
				{
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
				},
			},
			expected: []CardHistoryRecord{
				{
					ID:           1,
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
					CreditLimit:  nil,
					PaidDay:      nil,
					PaidAmount:   lib.NewCurrency("0", lib.USD),
					Period:       MONTHLY,
				},
			},
		},
		{
			should: "panic on foreign key violations",
			cards: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryConfig{
				{
					CreditCardID: 2,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
				},
				{
					CreditCardID: 1,
					MonthID:      2,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
				},
			},
			expectedError: ErrForeignKey,
		},
		{
			should: "panic on due day constraint violation",
			cards: []CreditCardConfig{
				{
					Name:           "test",
					DueDay:         5,
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryConfig{
				{
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       0,
				},
				{
					CreditCardID: 1,
					MonthID:      2,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       32,
				},
			},
			expectedError: ErrDueDayInvalid,
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			a := assert.New(t)
			r := require.New(t)

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			db.CreateMonth(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local))

			for _, cardConfig := range mock.cards {
				db.CreateCreditCard(cardConfig)
			}

			if mock.expectedError != nil {
				for _, histConfig := range mock.actual {
					a.PanicsWithValue(mock.expectedError, func() {
						db.CreateCreditCardHistory(histConfig)
					})
				}
				return
			}

			for _, histConfig := range mock.actual {
				db.CreateCreditCardHistory(histConfig)
			}

			res, err := db.QueryCreditCardHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}
}

func TestSetCreditCardHistory(t *testing.T) {
	type MockTable struct {
		should              string
		actual              CCFieldMap
		expected            CardHistoryRecord
		expectedErrContains *string
	}

	table := []MockTable{
		{
			should: "set all available fields",
			actual: CCFieldMap{
				CC_BALANCE:     lib.NewCurrency("500", lib.USD),
				CC_LIMIT:       lib.NewCurrency("1234.56", lib.USD),
				CC_DUE_DAY:     10,
				CC_PAID_DAY:    20,
				CC_PAID_AMOUNT: lib.NewCurrency("250", lib.USD),
			},
			expected: CardHistoryRecord{
				ID:           1,
				CreditCardID: 1,
				MonthID:      1,
				Balance:      lib.NewCurrency("500", lib.USD),
				CreditLimit:  lib.NewPointer(lib.NewCurrency("1234.56", lib.USD)),
				DueDay:       10,
				PaidDay:      lib.NewPointer(20),
				PaidAmount:   lib.NewCurrency("250", lib.USD),
				Period:       MONTHLY,
			},
		},
		{
			should: "error on unsupported field",
			actual: CCFieldMap{
				"invalidField": nil,
			},
			expectedErrContains: lib.NewPointer("unsupported credit card history field"),
		},
		{
			should: "error on invalid balance field type",
			actual: CCFieldMap{
				CC_BALANCE: 8008,
			},
			expectedErrContains: lib.NewPointer("type: lib.Currency"),
		},
		{
			should: "error on invalid credit limit field type",
			actual: CCFieldMap{
				CC_LIMIT: 8008,
			},
			expectedErrContains: lib.NewPointer("type: lib.Currency"),
		},
		{
			should: "error on invalid paid amount field type",
			actual: CCFieldMap{
				CC_PAID_AMOUNT: 8008,
			},
			expectedErrContains: lib.NewPointer("type: lib.Currency"),
		},
		{
			should: "error on invalid due day field type",
			actual: CCFieldMap{
				CC_DUE_DAY: "invalidType",
			},
			expectedErrContains: lib.NewPointer("type: int"),
		},
		{
			should: "error on invalid paid day field type",
			actual: CCFieldMap{
				CC_PAID_DAY: "invalidType",
			},
			expectedErrContains: lib.NewPointer("type: int"),
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			a := assert.New(t)
			r := require.New(t)

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			db.CreateMonth(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local))

			db.CreateCreditCard(CreditCardConfig{
				Name:           "test",
				DueDay:         1,
				LastFourDigits: "1234",
			})

			db.CreateCreditCardHistory(CreditCardHistoryConfig{
				CreditCardID: 1,
				MonthID:      1,
				Balance:      lib.NewCurrency("0", lib.USD),
				DueDay:       1,
			})

			if mock.expectedErrContains != nil {
				err := db.SetCreditCardHistory(1, mock.actual)
				r.Error(err)
				a.Contains(err.Error(), *mock.expectedErrContains)
				return
			}

			db.SetCreditCardHistory(1, mock.actual)

			res, err := db.QueryCreditCardHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res[0])
		})
	}
}
