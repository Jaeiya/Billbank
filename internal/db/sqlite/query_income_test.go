package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/internal"
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
					Amount: internal.NewCurrency("200", internal.USD),
					Period: MONTHLY,
				},
			},
			expected: []IncomeRecord{
				{
					ID:     1,
					Name:   "test",
					Amount: internal.NewCurrency("200", internal.USD),
					Period: MONTHLY,
				},
			},
		},
		{
			should: "panic on amount constraint violation",
			actual: []IncomeRecord{
				{Name: "test", Amount: internal.NewCurrency("-5", internal.USD)},
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

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
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

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
		r.NoError(err)
		defer db.Close()

		_, err = db.CreateIncome(IncomeRecord{
			Name:   "name",
			Amount: internal.NewCurrency("13.37", internal.USD),
			Period: MONTHLY,
		})
		r.NoError(err)

		_, err = db.CreateIncome(IncomeRecord{
			Name:   "name",
			Amount: internal.NewCurrency("133.7", internal.USD),
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
					Amount: internal.NewCurrency("250", internal.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 1,
					MonthID:  1,
					Amount:   internal.NewCurrency("250", internal.USD),
				},
			},
			expected: []IncomeHistoryRecord{
				{
					ID:       1,
					IncomeID: 1,
					MonthID:  1,
					Amount:   internal.NewCurrency("250", internal.USD),
				},
			},
		},
		{
			should: "panic on amount constraint violation",
			incomes: []IncomeRecord{
				{
					Name:   "test",
					Amount: internal.NewCurrency("250", internal.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 1,
					MonthID:  1,
					Amount:   internal.NewCurrency("-5", internal.USD),
				},
			},
			expectedError: ErrAmountInvalid,
		},
		{
			should: "panic on foreign key constraint violation",
			incomes: []IncomeRecord{
				{
					Name:   "test",
					Amount: internal.NewCurrency("250", internal.USD),
					Period: BIWEEKLY,
				},
			},
			actual: []IncomeHistoryRecord{
				{
					IncomeID: 2,
					MonthID:  1,
					Amount:   internal.NewCurrency("1000", internal.USD),
				},
				{
					IncomeID: 1,
					MonthID:  2,
					Amount:   internal.NewCurrency("500", internal.USD),
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

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
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
