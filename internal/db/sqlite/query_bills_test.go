package sqlite

import (
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
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t2",
					TypeID:  1,
					Amount:  internal.NewCurrency("39.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 27),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t3",
					TypeID:  1,
					Amount:  internal.NewCurrency("10.45", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 11),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t4",
					TypeID:  1,
					Amount:  internal.NewCurrency("2.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 8),
					Status:  Pending,
					Period:  Monthly,
				},
			},
			expected: []BillRecord{
				{
					ID:      1,
					Name:    "t1",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					ID:      2,
					Name:    "t2",
					TypeID:  1,
					Amount:  internal.NewCurrency("39.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 27),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					ID:      3,
					Name:    "t3",
					TypeID:  1,
					Amount:  internal.NewCurrency("10.45", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 11),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					ID:      4,
					Name:    "t4",
					TypeID:  1,
					Amount:  internal.NewCurrency("2.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 8),
					Status:  Pending,
					Period:  Monthly,
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
					Status:  Pending,
					Period:  Monthly,
				},
			},
			expectedError: ErrForeignKey,
		},
		{
			should: "set bill to inactive",
			actual: []BillRecord{
				{
					Name:     "t1",
					TypeID:   1,
					Amount:   internal.NewCurrency("123.4", internal.USD),
					DueDate:  createDate(now.Year(), now.Month(), 7),
					Status:   Pending,
					IsActive: false,
					Period:   Monthly,
				},
			},
			expected: []BillRecord{
				{
					ID:       1,
					Name:     "t1",
					TypeID:   1,
					Amount:   internal.NewCurrency("123.4", internal.USD),
					DueDate:  createDate(now.Year(), now.Month(), 7),
					Status:   Pending,
					IsActive: false,
					Period:   Monthly,
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
			defer func() {
				r.NoError(db.Close())
			}()

			err = db.CreateBillTypes([]BillTypeRecord{{ID: 1, Name: "test"}})
			r.NoError(err, "expected to create bill types")

			if mock.expectedError != nil {
				err = db.AddNewBills(mock.actual)
				r.ErrorIs(err, mock.expectedError, "expected specific error")
				return
			}

			err = db.AddNewBills(mock.actual)
			r.NoError(err, "expected bill to be created properly")

			bills, err := db.QueryBills(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, bills)
		})
	}

	t.Run("should panic on unique constraint violation", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		r := require.New(t)

		db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
		r.NoError(err)
		defer func() {
			r.NoError(db.Close())
		}()

		err = db.CreateBillTypes([]BillTypeRecord{{ID: 1, Name: "test"}})
		r.NoError(err, "expected to create bill types")

		err = db.AddNewBills([]BillRecord{
			{
				Name:    "name",
				TypeID:  1,
				Amount:  internal.NewCurrency("13.37", internal.USD),
				DueDate: createDate(now.Year(), now.Month(), 10),
				Status:  Pending,
				Period:  Monthly,
			},
		})
		r.NoError(err, "expected to successfully create test bill")

		err = db.AddNewBills([]BillRecord{
			{
				Name:    "name",
				TypeID:  1,
				Amount:  internal.NewCurrency("133.7", internal.USD),
				DueDate: createDate(now.Year(), now.Month(), 7),
				Status:  Pending,
				Period:  Monthly,
			},
		})
		a.ErrorIs(err, ErrUniqueName, "expected error when creating duplicate bill name")
	})
}

func TestQueryBillsCount(t *testing.T) {
	t.Parallel()
	type MockTable struct {
		should   string
		actual   []BillRecord
		expected int
	}

	now := time.Now()

	table := []MockTable{
		{
			should:   "count 0 records in bills table",
			actual:   []BillRecord{},
			expected: 0,
		},
		{
			should: "count 1 record in bills table",
			actual: []BillRecord{
				{
					Name:    "t1",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
			},
			expected: 1,
		},
		{
			should: "count 5 records in bills table",
			actual: []BillRecord{
				{
					Name:    "t1",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t2",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t3",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t4",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
				{
					Name:    "t5",
					TypeID:  1,
					Amount:  internal.NewCurrency("19.99", internal.USD),
					DueDate: createDate(now.Year(), now.Month(), 3),
					Status:  Pending,
					Period:  Monthly,
				},
			},
			expected: 5,
		},
	}

	r := require.New(t)
	a := assert.New(t)

	for _, mock := range table {
		t.Run("should "+mock.should, func(t *testing.T) {
			db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
			r.NoError(err)
			defer func() {
				r.NoError(db.Close())
			}()

			err = db.CreateBillTypes([]BillTypeRecord{{ID: 1, Name: "test"}})
			r.NoError(err, "expected to create bill types")

			if len(mock.actual) > 0 {
				err = db.AddNewBills(mock.actual)
				r.NoError(err, "expected to create bills")
			}

			count, err := db.QueryBillsCount()
			r.NoError(err, "should successfully count bills")

			a.Equal(mock.expected, count)
		})
	}
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
					Status:  Pending,
					Period:  Monthly,
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
					Status:  Pending,
					Period:  Monthly,
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
					Status:  Pending,
					Period:  Monthly,
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
					Status:  Pending,
					Period:  Monthly,
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
			a := assert.New(t)
			r := require.New(t)

			db, err := NewSqliteDb("", internal.USD, WithMemoryDB())
			r.NoError(err)
			defer func() {
				r.NoError(db.Close())
			}()

			now := time.Now()
			_, err = db.CreateMonth(now.Year(), now.Month())
			r.NoError(err, "expected month to be created successfully")

			err = db.CreateBillTypes([]BillTypeRecord{{ID: 1, Name: "test"}})
			r.NoError(err, "expected bill types to be created")

			err = db.AddNewBills(mock.bills)
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
