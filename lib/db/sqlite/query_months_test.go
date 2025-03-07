package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jaeiya/billbank/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMonth(t *testing.T) {
	t.Parallel()
	type MockDate struct {
		year  int
		month time.Month
		day   int
	}
	type MockTable struct {
		should        string
		actual        MockDate
		expected      MonthRecord
		expectedError error
	}

	createDate := func(year int, month time.Month, day int) lib.Date {
		d, _ := lib.NewDate(year, month, day)
		return d
	}
	now := time.Now()

	table := []MockTable{
		{
			should: "create a month record",
			actual: MockDate{
				year:  now.Year(),
				month: now.Month(),
				day:   1,
			},
			expected: MonthRecord{
				ID:   1,
				Date: createDate(now.Year(), now.Month(), 1),
			},
		},
		{
			should: "should error creating prior year",
			actual: MockDate{
				year:  now.Year() - 1,
				month: now.Month(),
				day:   1,
			},
			expected:      MonthRecord{},
			expectedError: ErrCreatePastTime,
		},
		{
			should: "should error creating prior month",
			actual: MockDate{
				year:  now.Year(),
				month: now.Month() - 1,
				day:   1,
			},
			expected:      MonthRecord{},
			expectedError: ErrCreatePastTime,
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
				_, err := db.CreateMonth(mock.actual.year, mock.actual.month)
				a.ErrorIs(err, mock.expectedError, "expected to get correct error")
				return
			}
			db.CreateMonth(mock.actual.year, mock.actual.month)

			res, err := db.QueryMonths(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res[0])
		})
	}
}
