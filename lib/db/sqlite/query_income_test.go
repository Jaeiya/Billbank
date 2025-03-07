package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIncome(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should        string
		actual        []IncomeRecord
		expected      []IncomeRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "add an income record to the table",
			actual: []IncomeRecord{
				{
					Name:   "test",
					Amount: lib.NewCurrency("200", lib.USD),
					Period: MONTHLY,
				},
			},
			expected: []IncomeRecord{
				{
					ID:     1,
					Name:   "test",
					Amount: lib.NewCurrency("200", lib.USD),
					Period: MONTHLY,
				},
			},
		},
		{
			should: "panic on amount constraint violation",
			actual: []IncomeRecord{
				{Name: "test", Amount: lib.NewCurrency("-5", lib.USD)},
			},
			expectedError: ErrAmountInvalid,
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
				for _, iConfig := range mock.actual {
					_, err = db.CreateIncome(iConfig)
					a.ErrorIs(err, mock.expectedError)
				}
				return
			}

			for _, iConfig := range mock.actual {
				_, err = db.CreateIncome(iConfig)
				r.NoError(err)
			}

			res, err := db.QueryIncome(QueryMap{})
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

		_, err = db.CreateIncome(IncomeRecord{
			Name:   "name",
			Amount: lib.NewCurrency("13.37", lib.USD),
			Period: MONTHLY,
		})
		r.NoError(err)

		_, err = db.CreateIncome(IncomeRecord{
			Name:   "name",
			Amount: lib.NewCurrency("133.7", lib.USD),
			Period: MONTHLY,
		})
		a.ErrorIs(err, ErrUniqueName)
	})
}

func TestCreateIncomeHistory(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should        string
		incomes       []IncomeRecord
		actual        []IncomeHistoryRecord
		expected      []IncomeHistoryRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "create an income history record",
			incomes: []IncomeRecord{
				{
					Name:   "test",
					Amount: lib.NewCurrency("250", lib.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 1,
					MonthID:  1,
					Amount:   lib.NewCurrency("250", lib.USD),
				},
			},
			expected: []IncomeHistoryRecord{
				{
					ID:       1,
					IncomeID: 1,
					MonthID:  1,
					Amount:   lib.NewCurrency("250", lib.USD),
				},
			},
		},
		{
			should: "panic on amount constraint violation",
			incomes: []IncomeRecord{
				{
					Name:   "test",
					Amount: lib.NewCurrency("250", lib.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 1,
					MonthID:  1,
					Amount:   lib.NewCurrency("-5", lib.USD),
				},
			},
			expectedError: ErrAmountInvalid,
		},
		{
			should: "panic on foreign key constraint violation",
			incomes: []IncomeRecord{
				{
					Name:   "test",
					Amount: lib.NewCurrency("250", lib.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 2,
					MonthID:  1,
					Amount:   lib.NewCurrency("1000", lib.USD),
				},
				{
					IncomeID: 1,
					MonthID:  2,
					Amount:   lib.NewCurrency("500", lib.USD),
				},
			},
			expectedError: ErrForeignKey,
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

			for _, iConfig := range mock.incomes {
				_, err = db.CreateIncome(iConfig)
				r.NoError(err)
			}

			if mock.expectedError != nil {
				for _, ihConfig := range mock.actual {
					err = db.CreateIncomeHistory(ihConfig)
					a.ErrorIs(err, mock.expectedError)
				}
				return
			}

			for _, ihConfig := range mock.actual {
				err = db.CreateIncomeHistory(ihConfig)
				r.NoError(err)
			}

			res, err := db.QueryIncomeHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}
}
