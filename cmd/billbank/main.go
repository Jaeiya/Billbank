package main

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/cmdmodel"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/db/sqlite"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/utils"
)

func main() {
	err := logger.SetLogLevel(logger.Debug)
	if err != nil {
		panic(err)
	}
	logger.Log(logger.Info, "starting billbank")

	filePath := filepath.Join(utils.GetWorkingDir(), "billbank.db")
	// TODO  Do not delete database in production
	defer func() {
		err := os.Remove(filePath)
		if err != nil {
			panic(err)
		}
	}()

	db, err := sqlite.NewSqliteDb(filePath, lib.USD)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	logger.Log(logger.Info, "loaded database")

	vp := lib.ViewPort{}
	h := utils.NewCmdHistory()

	vp.CommandInput = cmdmodel.NewInputModel(
		h,
		"bills",
		commands.NewBillsCmd(db),
		commands.NewDebugCmd(h),
	)

	logger.Log(logger.Info, "finished loading commands")

	p1 := tea.NewProgram(
		vp,
		tea.WithAltScreen(),
	)

	if _, err := p1.Run(); err != nil {
		panic(err)
	}

	logger.Log(logger.Info, "exiting billbank")
	err = logger.CloseLog()
	if err != nil {
		panic(err)
	}
}

// db, err := sqlite.NewSqliteDb("billbank.db", lib.USD)
// if err != nil {
// 	panic(err)
// }

// defer func() {
// 	err := db.Close()
// 	if err != nil {
// 		panic(err)
// 	}
// }()

// _, err = db.CreateMonth(2025, time.March)
// if err != nil {
// 	panic(err)
// }

// err = db.CreateBillTypes([]string{"test type"})
// if err != nil {
// 	panic(err)
// }

// // IMPORTANT: Due dates for bills should be set according to period
// // IMPORTANT: [monthly]: we set month to January and the day is specified by user
// // IMPORTANT:  [yearly]: user should set month and day
// dd, err := lib.NewDate(2025, time.February, 3)
// if err != nil {
// 	panic(err)
// }

// err = db.CreateNewBills([]sqlite.BillRecord{
// 	{
// 		TypeID:  1,
// 		Name:    "test bill 1",
// 		Amount:  lib.NewCurrency("32.41", lib.USD),
// 		DueDate: dd,
// 		Period:  sqlite.MONTHLY,
// 	},
// 	{
// 		TypeID:  1,
// 		Name:    "test bill 2",
// 		Amount:  lib.NewCurrency("11.33", lib.USD),
// 		DueDate: dd,
// 		Period:  sqlite.MONTHLY,
// 	},
// 	{
// 		TypeID:  1,
// 		Name:    "test bill 3",
// 		Amount:  lib.NewCurrency("782.23", lib.USD),
// 		DueDate: dd,
// 		Period:  sqlite.MONTHLY,
// 	},
// })
// if err != nil {
// 	panic(err)
// }

// err = db.CreateMonthlyBills([]sqlite.MonthlyBill{
// 	{
// 		BillID:   2,
// 		IsActive: true,
// 	},
// 	{
// 		BillID:   1,
// 		IsActive: true,
// 	},
// 	{
// 		BillID:   3,
// 		IsActive: true,
// 	},
// })

// rows, err := db.Query(sqlite.MonthlyBillsSql)
// if err != nil {
// 	panic(err)
// }

// var id int
// var isActive bool
// var amount lib.Currency
// var dueDate lib.Date
// var period sqlite.Period
// var billType string

// for rows.Next() {
// 	err := rows.Scan(&id, &isActive, &amount, &dueDate, &period, &billType)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(id, isActive, amount, dueDate, period, billType)
// }
