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
	t.Parallel()
	type MockTable struct {
		should        string
		actual        []CreditCardRecord
		expected      []CreditCardRecord
		password      *string
		expectedError error
	}

	table := []MockTable{
		{
			should: "create credit card records",
			actual: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     new("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          new("some notes"),
				},
			},
			expected: []CreditCardRecord{
				{
					ID:             1,
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     new("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          new("some notes"),
				},
			},
			password: new("password"),
		},
		{
			should: "nullable values should be nil",
			actual: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     new("2382 3812 4582 5822"),
					LastFourDigits: "5822",
				},
				{
					Name:           "test2",
					DueDay:         7,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "0023",
					Notes:          new("some notes"),
				},
				{
					Name:           "test3",
					DueDay:         8,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
				},
			},
			expected: []CreditCardRecord{
				{
					ID:             1,
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					CardNumber:     new("2382 3812 4582 5822"),
					LastFourDigits: "5822",
					Notes:          nil,
				},
				{
					ID:             2,
					Name:           "test2",
					DueDay:         7,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "0023",
					Notes:          new("some notes"),
					CardNumber:     nil,
				},
				{
					ID:             3,
					Name:           "test3",
					DueDay:         8,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
					CardNumber:     nil,
					Notes:          nil,
				},
			},
			password: new("password"),
		},
		{
			should: "panic on due day constraint violation",
			actual: []CreditCardRecord{
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

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			if mock.expectedError != nil {
				for _, cardConfig := range mock.actual {
					err = db.CreateCreditCard(cardConfig, mock.password)
					a.ErrorIs(err, ErrDueDayInvalid)
				}
				return
			}

			for _, cardConfig := range mock.actual {
				err = db.CreateCreditCard(cardConfig, mock.password)
				r.NoError(err)
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
		r := require.New(t)

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateCreditCard(CreditCardRecord{
			Name:           "test",
			DueDay:         5,
			LastFourDigits: "1234",
		}, nil)
		r.NoError(err)

		err = db.CreateCreditCard(CreditCardRecord{
			Name:           "test",
			DueDay:         5,
			LastFourDigits: "1234",
		}, nil)
		a.ErrorIs(err, ErrUniqueName)
	})
}

func TestCreateCreditCardHistory(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should        string
		cards         []CreditCardRecord
		actual        []CreditCardHistoryRecord
		expected      []CreditCardHistoryRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "create credit card history records",
			cards: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryRecord{
				{
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
					CreditLimit:  new(lib.NewCurrency("5000", lib.USD)),
					PaidDay:      new(3),
					PaidAmount:   new(lib.NewCurrency("32.32", lib.USD)),
					ClearedDay:   new(7),
				},
			},
			expected: []CreditCardHistoryRecord{
				{
					ID:           1,
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
					CreditLimit:  new(lib.NewCurrency("5000", lib.USD)),
					PaidDay:      new(3),
					PaidAmount:   new(lib.NewCurrency("32.32", lib.USD)),
					ClearedDay:   new(7),
				},
			},
		},
		{
			should: "set nullable fields to nil",
			cards: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					CreditLimit:    new(lib.NewCurrency("5000", lib.USD)),
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryRecord{
				{
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
				},
			},
			expected: []CreditCardHistoryRecord{
				{
					ID:           1,
					CreditCardID: 1,
					MonthID:      1,
					Balance:      lib.NewCurrency("500", lib.USD),
					DueDay:       5,
					CreditLimit:  nil,
					PaidDay:      nil,
					PaidAmount:   nil,
					ClearedDay:   nil,
				},
			},
		},
		{
			should: "panic on foreign key violations",
			cards: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryRecord{
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
			cards: []CreditCardRecord{
				{
					Name:           "test",
					DueDay:         5,
					LastFourDigits: "1234",
				},
			},
			actual: []CreditCardHistoryRecord{
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

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())
			r.NoError(err)

			for _, cardConfig := range mock.cards {
				err = db.CreateCreditCard(cardConfig, nil)
				r.NoError(err)
			}

			if mock.expectedError != nil {
				for _, histConfig := range mock.actual {
					err = db.CreateCreditCardHistory(histConfig)
					a.ErrorIs(err, mock.expectedError)
				}
				return
			}

			for _, histConfig := range mock.actual {
				err = db.CreateCreditCardHistory(histConfig)
				r.NoError(err)
			}

			res, err := db.QueryCreditCardHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}
}

func TestSetCreditCardHistory(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should              string
		actual              CCFieldMap
		expected            CreditCardHistoryRecord
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
			expected: CreditCardHistoryRecord{
				ID:           1,
				CreditCardID: 1,
				MonthID:      1,
				Balance:      lib.NewCurrency("500", lib.USD),
				CreditLimit:  new(lib.NewCurrency("1234.56", lib.USD)),
				DueDay:       10,
				PaidDay:      new(20),
				PaidAmount:   new(lib.NewCurrency("250", lib.USD)),
			},
		},
		{
			should: "error on unsupported field",
			actual: CCFieldMap{
				"invalidField": nil,
			},
			expectedErrContains: new("unsupported credit card history field"),
		},
		{
			should: "error on invalid balance field type",
			actual: CCFieldMap{
				CC_BALANCE: 8008,
			},
			expectedErrContains: new("type: lib.Currency"),
		},
		{
			should: "error on invalid credit limit field type",
			actual: CCFieldMap{
				CC_LIMIT: 8008,
			},
			expectedErrContains: new("type: lib.Currency"),
		},
		{
			should: "error on invalid paid amount field type",
			actual: CCFieldMap{
				CC_PAID_AMOUNT: 8008,
			},
			expectedErrContains: new("type: lib.Currency"),
		},
		{
			should: "error on invalid due day field type",
			actual: CCFieldMap{
				CC_DUE_DAY: "invalidType",
			},
			expectedErrContains: new("type: int"),
		},
		{
			should: "error on invalid paid day field type",
			actual: CCFieldMap{
				CC_PAID_DAY: "invalidType",
			},
			expectedErrContains: new("type: int"),
		},
	}

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			a := assert.New(t)
			r := require.New(t)

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			r.NoError(err)
			defer db.Close()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())
			r.NoError(err)

			err = db.CreateCreditCard(CreditCardRecord{
				Name:           "test",
				DueDay:         1,
				LastFourDigits: "1234",
			}, nil)
			r.NoError(err)

			err = db.CreateCreditCardHistory(CreditCardHistoryRecord{
				CreditCardID: 1,
				MonthID:      1,
				Balance:      lib.NewCurrency("0", lib.USD),
				DueDay:       1,
			})
			r.NoError(err)

			if mock.expectedErrContains != nil {
				err := db.SetCreditCardHistory(1, mock.actual)
				r.Error(err)
				a.Contains(err.Error(), *mock.expectedErrContains)
				return
			}

			err = db.SetCreditCardHistory(1, mock.actual)
			r.NoError(err)

			res, err := db.QueryCreditCardHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res[0])
		})
	}
}
