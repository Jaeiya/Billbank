package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/cmd"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/logger"
	"github.com/jaeiya/billbank/lib/utils"
)

func main() {
	// filePath := filepath.Join(wd, "hello.db")
	// defer func() {
	// 	err := os.Remove(filePath)
	// 	if err != nil {
	// 		panic(err)
	//
	// }()
	// db := sqlite.NewSqliteDb(filePath, lib.USD)
	// defer db.Close()

	logger.SetLogLevel(logger.Hot)
	logger.Log(logger.Info, "Main", "Starting BillBank")

	vp := lib.ViewPort{}
	h := utils.NewCmdHistory()

	vp.CommandInput = cmd.NewInput(
		h,
		cmd.With(
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
}
