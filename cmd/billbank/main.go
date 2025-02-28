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
		commands.NewBillsCmd(),
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
