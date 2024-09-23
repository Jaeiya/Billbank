package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBills(t *testing.T) {
	type MockTable struct {
		should        string
		actual        []BillsConfig
		expected      []BillRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "add bills to the bills table",
			actual: []BillsConfig{
				{
					Name:   "t1",
					Amount: lib.NewCurrency("19.99", lib.USD),
					DueDay: 5,
					Period: MONTHLY,
				},
				{
					Name:   "t2",
					Amount: lib.NewCurrency("39.99", lib.USD),
					DueDay: 27,
					Period: MONTHLY,
				},
				{
					Name:   "t3",
					Amount: lib.NewCurrency("10.45", lib.USD),
					DueDay: 11,
					Period: MONTHLY,
				},
				{
					Name:   "t4",
					Amount: lib.NewCurrency("2.99", lib.USD),
					DueDay: 8,
					Period: MONTHLY,
				},
			},
			expected: []BillRecord{
				{
					ID: 1,
					BillsConfig: BillsConfig{
						Name:   "t1",
						Amount: lib.NewCurrency("19.99", lib.USD),
						DueDay: 5,
						Period: MONTHLY,
					},
				},
				{
					ID: 2,
					BillsConfig: BillsConfig{
						Name:   "t2",
						Amount: lib.NewCurrency("39.99", lib.USD),
						DueDay: 27,
						Period: MONTHLY,
					},
				},
				{
					ID: 3,
					BillsConfig: BillsConfig{
						Name:   "t3",
						Amount: lib.NewCurrency("10.45", lib.USD),
						DueDay: 11,
						Period: MONTHLY,
					},
				},
				{
					ID: 4,
					BillsConfig: BillsConfig{
						Name:   "t4",
						Amount: lib.NewCurrency("2.99", lib.USD),
						DueDay: 8,
						Period: MONTHLY,
					},
				},
			},
		},
		{
			should: "panic on violated due_day constraint",
			actual: []BillsConfig{
				{
					Name:   "t1",
					Amount: lib.NewCurrency("133.7", lib.USD),
					DueDay: 32,
				},
				{
					Name:   "t2",
					Amount: lib.NewCurrency("133.7", lib.USD),
					DueDay: 0,
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
				for _, bill := range mock.actual {
					a.PanicsWithValue(ErrDueDayInvalid, func() {
						db.CreateNewBill(bill)
					})
				}
				return
			}

			for _, bill := range mock.actual {
				db.CreateNewBill(bill)
			}

			bills, err := db.QueryBills(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, bills)
		})
	}

	t.Run("should panic on unique constraint violation", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		a := assert.New(t)

		db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		defer db.Close()

		db.CreateNewBill(BillsConfig{
			Name:   "name",
			Amount: lib.NewCurrency("13.37", lib.USD),
			DueDay: 3,
			Period: MONTHLY,
		})

		a.PanicsWithValue(ErrUniqueName, func() {
			db.CreateNewBill(BillsConfig{
				Name:   "name",
				Amount: lib.NewCurrency("133.7", lib.USD),
				DueDay: 7,
				Period: MONTHLY,
			})
		})
	})
}

func TestCreateBillHistory(t *testing.T) {
	type Mock struct {
		should        string
		bills         []BillsConfig
		actual        []BillHistoryConfig
		expected      []BillHistoryRecord
		expectedError error
	}

	table := []Mock{
		{
			should: "add bill history entries to table",
			bills: []BillsConfig{
				{
					Name:   "b1",
					Amount: lib.NewCurrency("13.37", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					BillID:     1,
					MonthID:    1,
					Amount:     lib.NewCurrency("13.37", lib.USD),
					DueDay:     3,
					PaidAmount: lib.NewPointer(lib.NewCurrency("5", lib.USD)),
				},
			},
			expected: []BillHistoryRecord{
				{
					ID: 1,
					BillHistoryConfig: BillHistoryConfig{
						BillID:     1,
						MonthID:    1,
						Amount:     lib.NewCurrency("13.37", lib.USD),
						DueDay:     3,
						PaidAmount: lib.NewPointer(lib.NewCurrency("5", lib.USD)),
					},
				},
			},
		},
		{
			should: "default to a zero paid amount",
			bills: []BillsConfig{
				{
					Name:   "b1",
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					BillID:  1,
					MonthID: 1,
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
				},
			},
			expected: []BillHistoryRecord{
				{
					ID: 1,
					BillHistoryConfig: BillHistoryConfig{
						BillID:     1,
						MonthID:    1,
						Amount:     lib.NewCurrency("1337", lib.USD),
						DueDay:     3,
						PaidAmount: lib.NewPointer(lib.NewCurrency("0", lib.USD)),
					},
				},
			},
		},
		{
			should: "allow nullable fields to be nil",
			bills: []BillsConfig{
				{
					Name:   "b1",
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					BillID:  1,
					MonthID: 1,
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
				},
			},
			expected: []BillHistoryRecord{
				{
					ID: 1,
					BillHistoryConfig: BillHistoryConfig{
						BillID:     1,
						MonthID:    1,
						Amount:     lib.NewCurrency("1337", lib.USD),
						DueDay:     3,
						PaidAmount: lib.NewPointer(lib.NewCurrency("0", lib.USD)),
						PaidDate:   nil,
						Notes:      nil,
					},
				},
			},
		},
		{
			should: "panic with foreign key violations",
			bills: []BillsConfig{
				{
					Name:   "b1",
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					BillID:  2,
					MonthID: 1,
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
				},
				{
					BillID:  1,
					MonthID: 2,
					Amount:  lib.NewCurrency("13.37", lib.USD),
					DueDay:  7,
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

			db := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
			defer db.Close()

			db.CreateMonth(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local))

			for _, b := range mock.bills {
				db.CreateNewBill(b)
			}

			if mock.expectedError != nil {
				for _, history := range mock.actual {
					if mock.expectedError != nil {
						a.PanicsWithValue(ErrForeignKey, func() {
							db.CreateBillHistory(history)
						})
					}
				}
				return
			}

			for _, history := range mock.actual {
				db.CreateBillHistory(history)
			}

			res, err := db.QueryBillHistory(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res)
		})
	}
}
