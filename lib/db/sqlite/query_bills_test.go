package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBills(t *testing.T) {
	t.Parallel()
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
					TypeID: 1,
					Amount: lib.NewCurrency("19.99", lib.USD),
					DueDay: 5,
					Period: MONTHLY,
				},
				{
					Name:   "t2",
					TypeID: 1,
					Amount: lib.NewCurrency("39.99", lib.USD),
					DueDay: 27,
					Period: MONTHLY,
				},
				{
					Name:   "t3",
					TypeID: 1,
					Amount: lib.NewCurrency("10.45", lib.USD),
					DueDay: 11,
					Period: MONTHLY,
				},
				{
					Name:   "t4",
					TypeID: 1,
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
						TypeID: 1,
						Amount: lib.NewCurrency("19.99", lib.USD),
						DueDay: 5,
						Period: MONTHLY,
					},
				},
				{
					ID: 2,
					BillsConfig: BillsConfig{
						Name:   "t2",
						TypeID: 1,
						Amount: lib.NewCurrency("39.99", lib.USD),
						DueDay: 27,
						Period: MONTHLY,
					},
				},
				{
					ID: 3,
					BillsConfig: BillsConfig{
						Name:   "t3",
						TypeID: 1,
						Amount: lib.NewCurrency("10.45", lib.USD),
						DueDay: 11,
						Period: MONTHLY,
					},
				},
				{
					ID: 4,
					BillsConfig: BillsConfig{
						Name:   "t4",
						TypeID: 1,
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
					TypeID: 1,
					Amount: lib.NewCurrency("133.7", lib.USD),
					DueDay: 32,
				},
				{
					Name:   "t2",
					TypeID: 1,
					Amount: lib.NewCurrency("133.7", lib.USD),
					DueDay: 0,
				},
			},
			expectedError: ErrDueDayInvalid,
		},
		{
			should: "panic on foreign key type_id constraint",
			actual: []BillsConfig{
				{
					Name:   "t1",
					TypeID: 2,
					Amount: lib.NewCurrency("123.4", lib.USD),
					DueDay: 12,
					Period: MONTHLY,
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

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected to create bill types")

			if mock.expectedError != nil {
				for _, bill := range mock.actual {
					err = db.CreateNewBill(bill)
					a.ErrorIs(err, mock.expectedError, "expected specific error")
				}
				return
			}

			for _, bill := range mock.actual {
				err = db.CreateNewBill(bill)
				r.NoError(err, "expected bill to be created properly")
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
		r := require.New(t)

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBillTypes([]string{"test"})
		r.NoError(err, "expected to create bill types")

		err = db.CreateNewBill(BillsConfig{
			Name:   "name",
			TypeID: 1,
			Amount: lib.NewCurrency("13.37", lib.USD),
			DueDay: 3,
			Period: MONTHLY,
		})
		r.NoError(err, "expected to successfully create test bill")

		err = db.CreateNewBill(BillsConfig{
			Name:   "name",
			TypeID: 1,
			Amount: lib.NewCurrency("133.7", lib.USD),
			DueDay: 7,
			Period: MONTHLY,
		})
		a.ErrorIs(err, ErrUniqueName, "expected error when creating duplicate bill name")
	})
}

func TestCreateBillHistory(t *testing.T) {
	t.Parallel()
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
					TypeID: 1,
					Amount: lib.NewCurrency("13.37", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					MonthID:    1,
					TypeID:     1,
					Name:       "b1",
					Amount:     lib.NewCurrency("13.37", lib.USD),
					DueDay:     3,
					PaidAmount: utils.NewPointer(lib.NewCurrency("5", lib.USD)),
				},
			},
			expected: []BillHistoryRecord{
				{
					ID: 1,
					BillHistoryConfig: BillHistoryConfig{
						MonthID:    1,
						TypeID:     1,
						Name:       "b1",
						Amount:     lib.NewCurrency("13.37", lib.USD),
						DueDay:     3,
						PaidAmount: utils.NewPointer(lib.NewCurrency("5", lib.USD)),
					},
				},
			},
		},
		{
			should: "allow nullable fields to be nil",
			bills: []BillsConfig{
				{
					Name:   "b3",
					TypeID: 1,
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					MonthID: 1,
					TypeID:  1,
					Name:    "b3",
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
				},
			},
			expected: []BillHistoryRecord{
				{
					ID: 1,
					BillHistoryConfig: BillHistoryConfig{
						MonthID:    1,
						TypeID:     1,
						Name:       "b3",
						Amount:     lib.NewCurrency("1337", lib.USD),
						DueDay:     3,
						PaidAmount: nil,
						PaidDay:    nil,
						Notes:      nil,
					},
				},
			},
		},
		{
			should: "error with foreign key month id violation",
			bills: []BillsConfig{
				{
					Name:   "b4",
					TypeID: 1,
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					MonthID: 2, // should not exist
					TypeID:  1,
					Name:    "b4",
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
				},
			},
			expectedError: ErrForeignKey,
		},
		{
			should: "error with foreign key type id violation",
			bills: []BillsConfig{
				{
					Name:   "b4",
					TypeID: 1,
					Amount: lib.NewCurrency("1337", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []BillHistoryConfig{
				{
					MonthID: 1,
					TypeID:  2, // should not exist
					Name:    "b4",
					Amount:  lib.NewCurrency("1337", lib.USD),
					DueDay:  3,
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

			_, err = db.CreateMonth(NewMonth(2024, time.January))
			r.NoError(err, "expected month to be created successfully")

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected bill types to be created")

			for _, b := range mock.bills {
				err = db.CreateNewBill(b)
				r.NoError(err)
			}

			if mock.expectedError != nil {
				for _, history := range mock.actual {
					if mock.expectedError != nil {
						err = db.CreateBillHistory(history)
						a.ErrorIs(err, ErrForeignKey)
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

func TestBillsMonthly(t *testing.T) {
	t.Parallel()
	type Mock struct {
		should        string
		bills         []BillsConfig
		actual        []MonthlyBill
		expected      []MonthlyBill
		expectedError error
	}

	table := []Mock{
		{
			should: "create monthly bill entries",
			bills: []BillsConfig{
				{
					TypeID: 1,
					Name:   "t1",
					Amount: lib.NewCurrency("123", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
				{
					TypeID: 1,
					Name:   "t2",
					Amount: lib.NewCurrency("123", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
				{
					TypeID: 1,
					Name:   "t3",
					Amount: lib.NewCurrency("123", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
				{
					TypeID: 1,
					Name:   "t4",
					Amount: lib.NewCurrency("123", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual: []MonthlyBill{
				{BillID: 2, IsActive: true},
				{BillID: 4, IsActive: false},
				{BillID: 1, IsActive: false},
				{BillID: 3, IsActive: true},
			},
			expected: []MonthlyBill{
				{ID: 1, BillID: 2, IsActive: true},
				{ID: 2, BillID: 4, IsActive: false},
				{ID: 3, BillID: 1, IsActive: false},
				{ID: 4, BillID: 3, IsActive: true},
			},
		},
		{
			should: "panic on foreign key bill_id violation",
			bills: []BillsConfig{
				{
					TypeID: 1,
					Name:   "hello",
					Amount: lib.NewCurrency("1.12", lib.USD),
					DueDay: 3,
					Period: MONTHLY,
				},
			},
			actual:        []MonthlyBill{{BillID: 2}},
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

			_, err = db.CreateMonth(NewMonth(2024, time.January))
			r.NoError(err, "expected month to be created successfully")

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected bill types to be created")

			for _, b := range mock.bills {
				err = db.CreateNewBill(b)
				r.NoError(err)
			}

			if mock.expectedError != nil {
				err = db.CreateMonthlyBills(mock.actual)
				a.ErrorIs(err, ErrForeignKey)
				return
			}

			err = db.CreateMonthlyBills(mock.actual)
			r.NoError(err, "expect monthly bills to be created successfully")

			rows, err := db.QueryMonthlyBills()
			r.NoError(err)
			a.Equal(mock.expected, rows)
		})
	}

	t.Run("should panic on unique constraint violation", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		a := assert.New(t)
		r := require.New(t)

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), lib.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBillTypes([]string{"test"})
		r.NoError(err, "expected to create bill types")

		err = db.CreateNewBill(BillsConfig{
			Name:   "t1",
			TypeID: 1,
			Amount: lib.NewCurrency("13.37", lib.USD),
			DueDay: 3,
			Period: MONTHLY,
		})
		r.NoError(err, "expected to successfully create test bill")

		err = db.CreateMonthlyBills([]MonthlyBill{
			{BillID: 1}, {BillID: 1},
		})
		a.ErrorContains(
			err,
			"UNIQUE constraint failed: bills_monthly.bill_id",
			"expected error when creating duplicate bill name",
		)
	})
}
