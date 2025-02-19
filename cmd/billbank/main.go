package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/cmdmodel"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/utils"
)

// filePath := filepath.Join(wd, "hello.db")
// defer func() {
// 	err := os.Remove(filePath)
// 	if err != nil {
// 		panic(err)

// }()
// db := sqlite.NewSqliteDb(filePath, lib.USD)
// defer db.Close()

func main() {

	logger.SetLogLevel(logger.Info)
	logger.Log(logger.Info, "Starting BillBank")

	vp := lib.ViewPort{}
	h := utils.NewCmdHistory()

	vp.CommandInput = cmdmodel.NewInputModel(
		h,
		cmdmodel.With(
			commands.NewDebugCmd(h),
			commands.NewBillsCmd(),
		),
	)

	p1 := tea.NewProgram(
		vp,
		tea.WithAltScreen(),
	)

	if _, err := p1.Run(); err != nil {
		panic(err)
	}

	logger.Log(logger.Info, "exiting billbank")
	err := logger.CloseLog()
	if err != nil {
		panic(err)
	}
}
