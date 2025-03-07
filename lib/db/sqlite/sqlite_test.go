package sqlite

import (
	"time"

	"github.com/jaeiya/billbank/lib"
)

func createDate(year int, month time.Month, day int) lib.Date {
	d, _ := lib.NewDate(year, month, day)
	return d
}
