package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBills(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should        string
		actual        []BillRecord
		expected      []BillRecord
		expectedError error
	}

	now := time.Now()

	table := []MockTable{
		{
			should: "add bills to the bills table",
			actual: []BillRecord{
				{
					Name:    "t1",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Period:  MONTHLY,
				},
				{
					Name:    "t2",
					TypeID:  1,
					Amount:  internal.NewCurrency("39.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 27),
					Period:  MONTHLY,
				},
				{
					Name:    "t3",
					TypeID:  1,
					Amount:  internal.NewCurrency("10.45", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 11),
					Period:  MONTHLY,
				},
				{
					Name:    "t4",
					TypeID:  1,
					Amount:  internal.NewCurrency("2.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 8),
					Period:  MONTHLY,
				},
			},
			expected: []BillRecord{
				{
					ID:      1,
					Name:    "t1",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Period:  MONTHLY,
				},
				{
					ID:      2,
					Name:    "t2",
					TypeID:  1,
					Amount:  internal.NewCurrency("39.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 27),
					Period:  MONTHLY,
				},
				{
					ID:      3,
					Name:    "t3",
					TypeID:  1,
					Amount:  internal.NewCurrency("10.45", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 11),
					Period:  MONTHLY,
				},
				{
					ID:      4,
					Name:    "t4",
					TypeID:  1,
					Amount:  internal.NewCurrency("2.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 8),
					Period:  MONTHLY,
				},
			},
		},
		{
			should: "panic on foreign key type_id constraint",
			actual: []BillRecord{
				{
					Name:    "t1",
					TypeID:  2,
					Amount:  internal.NewCurrency("123.4", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 7),
					Period:  MONTHLY,
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

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected to create bill types")

			if mock.expectedError != nil {
				err = db.CreateNewBills(mock.actual)
				r.ErrorIs(err, mock.expectedError, "expected specific error")
				return
			}

			err = db.CreateNewBills(mock.actual)
			r.NoError(err, "expected bill to be created properly")

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

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBillTypes([]string{"test"})
		r.NoError(err, "expected to create bill types")

		err = db.CreateNewBills([]BillRecord{
			{
				Name:    "name",
				TypeID:  1,
				Amount:  internal.NewCurrency("13.37", internal.USD),
				DueDate: createDate(now.Year(), now.Month(), 10),
				Period:  MONTHLY,
			},
		})
		r.NoError(err, "expected to successfully create test bill")

		err = db.CreateNewBills([]BillRecord{
			{
				Name:    "name",
				TypeID:  1,
				Amount:  internal.NewCurrency("133.7", internal.USD),
				DueDate: createDate(now.Year(), now.Month(), 7),
				Period:  MONTHLY,
			},
		})
		a.ErrorIs(err, ErrUniqueName, "expected error when creating duplicate bill name")
	})
}

func TestCreateBillHistory(t *testing.T) {
	t.Parallel()
	type Mock struct {
		should        string
		bills         []BillRecord
		actual        []BillHistoryRecord
		expected      []BillHistoryRecord
		expectedError error
	}

	table := []Mock{
		{
			should: "add bill history entries to table",
			bills: []BillRecord{
				{
					Name:    "b1",
					TypeID:  1,
					Amount:  internal.NewCurrency("13.37", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
			},
			actual: []BillHistoryRecord{
				{
					MonthID:    1,
					TypeID:     1,
					Name:       "b1",
					Amount:     internal.NewCurrency("13.37", internal.USD),
					DueDate:    createDate(time.Now().Year(), time.Now().Month(), 3),
					PaidAmount: new(internal.NewCurrency("5", internal.USD)),
					PaidDate:   new(createDate(time.Now().Year(), time.Now().Month(), 3)),
					PaidHow:    new("online"),
					ClearedDay: new(3),
					Notes:      new("these are some notes"),
				},
			},
			expected: []BillHistoryRecord{
				{
					ID:         1,
					MonthID:    1,
					TypeID:     1,
					Name:       "b1",
					Amount:     internal.NewCurrency("13.37", internal.USD),
					DueDate:    createDate(time.Now().Year(), time.Now().Month(), 3),
					PaidAmount: new(internal.NewCurrency("5", internal.USD)),
					PaidDate:   new(createDate(time.Now().Year(), time.Now().Month(), 3)),
					PaidHow:    new("online"),
					ClearedDay: new(3),
					Notes:      new("these are some notes"),
				},
			},
		},
		{
			should: "allow nullable fields to be nil",
			bills: []BillRecord{
				{
					Name:    "b3",
					TypeID:  1,
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
			},
			actual: []BillHistoryRecord{
				{
					MonthID: 1,
					TypeID:  1,
					Name:    "b3",
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
				},
			},
			expected: []BillHistoryRecord{
				{
					ID:         1,
					MonthID:    1,
					TypeID:     1,
					Name:       "b3",
					Amount:     internal.NewCurrency("1337", internal.USD),
					DueDate:    createDate(time.Now().Year(), time.Now().Month(), 3),
					PaidAmount: nil,
					PaidDate:   nil,
					PaidHow:    nil,
					ClearedDay: nil,
					Notes:      nil,
				},
			},
		},
		{
			should: "error with foreign key month id violation",
			bills: []BillRecord{
				{
					Name:    "b4",
					TypeID:  1,
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
			},
			actual: []BillHistoryRecord{
				{
					MonthID: 2, // should not exist
					TypeID:  1,
					Name:    "b4",
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
				},
			},
			expectedError: ErrForeignKey,
		},
		{
			should: "error with foreign key type id violation",
			bills: []BillRecord{
				{
					Name:    "b4",
					TypeID:  1,
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
			},
			actual: []BillHistoryRecord{
				{
					MonthID: 1,
					TypeID:  2, // should not exist
					Name:    "b4",
					Amount:  internal.NewCurrency("1337", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
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
			r.NoError(err, "expected month to be created successfully")

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected bill types to be created")

			err = db.CreateNewBills(mock.bills)
			r.NoError(err)

			if mock.expectedError != nil {
				err = db.CreateBillHistory(mock.actual)
				r.ErrorIs(err, ErrForeignKey)
				return
			}

			err = db.CreateBillHistory(mock.actual)
			r.NoError(err, "expected to create bill history")

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
		bills         []BillRecord
		actual        []MonthlyBill
		expected      []MonthlyBill
		expectedError error
	}

	table := []Mock{
		{
			should: "create monthly bill entries",
			bills: []BillRecord{
				{
					TypeID:  1,
					Name:    "t1",
					Amount:  internal.NewCurrency("123", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
				{
					TypeID:  1,
					Name:    "t2",
					Amount:  internal.NewCurrency("123", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
				{
					TypeID:  1,
					Name:    "t3",
					Amount:  internal.NewCurrency("123", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
				},
				{
					TypeID:  1,
					Name:    "t4",
					Amount:  internal.NewCurrency("123", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
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
			bills: []BillRecord{
				{
					TypeID:  1,
					Name:    "hello",
					Amount:  internal.NewCurrency("1.12", internal.USD),
					DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
					Period:  MONTHLY,
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

			db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
			r.NoError(err)
			defer db.Close()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())
			r.NoError(err, "expected month to be created successfully")

			err = db.CreateBillTypes([]string{"test"})
			r.NoError(err, "expected bill types to be created")

			err = db.CreateNewBills(mock.bills)
			r.NoError(err)

			if mock.expectedError != nil {
				err = db.CreateMonthlyBills(mock.actual)
				r.ErrorIs(err, ErrForeignKey)
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

		db, err := NewSqliteDb(filepath.Join(dir, "mock.db"), internal.USD)
		r.NoError(err)
		defer db.Close()

		err = db.CreateBillTypes([]string{"test"})
		r.NoError(err, "expected to create bill types")

		err = db.CreateNewBills([]BillRecord{
			{
				Name:    "t1",
				TypeID:  1,
				Amount:  internal.NewCurrency("13.37", internal.USD),
				DueDate: createDate(time.Now().Year(), time.Now().Month(), 3),
				Period:  MONTHLY,
			},
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
