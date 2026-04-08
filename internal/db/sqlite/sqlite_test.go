package sqlite

import (
	"time"

	"github.com/jaeiya/billbank/internal"
)

func createDate(year int, month time.Month, day int) internal.Date {
	d, _ := internal.NewDate(year, month, day)
	return d
}
