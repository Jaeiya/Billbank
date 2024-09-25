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
	type MockTable struct {
		should        string
		actual        time.Time
		expected      MonthRecord
		expectedError error
	}

	table := []MockTable{
		{
			should: "create a month record",
			actual: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			expected: MonthRecord{
				ID:    1,
				Year:  2024,
				Month: 1,
			},
		},
		{
			should:        "panic on dirty date",
			actual:        time.Now(),
			expectedError: ErrDirtyDate,
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
				a.PanicsWithValue(mock.expectedError, func() {
					db.CreateMonth(mock.actual)
				})
				return
			}
			db.CreateMonth(mock.actual)

			res, err := db.QueryMonths(QueryMap{})
			r.NoError(err)

			a.Equal(mock.expected, res[0])
		})
	}
}
