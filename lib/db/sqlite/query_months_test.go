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
	type MockTable struct {
		should        string
		actual        Month
		expected      MonthRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "create a month record",
			actual: NewMonth(2024, time.January),
			expected: MonthRecord{
				ID:    1,
				Year:  2024,
				Month: 1,
			},
		},
		{
			should:        "panic on dirty date",
			actual:        Month(time.Now()),
			expectedError: ErrDirtyDate,
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
				_, err := db.CreateMonth(mock.actual)
				a.ErrorIs(err, mock.expectedError, "expected to get correct error")
				return
			}
			db.CreateMonth(mock.actual)

			res, err := db.QueryMonths(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res[0])
		})
	}
}
