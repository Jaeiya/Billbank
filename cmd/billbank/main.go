package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib"
	"github.com/jaeiya/billbank/lib/commander"
	"github.com/jaeiya/billbank/lib/utils"
	"github.com/jaeiya/billbank/lib/utils/logger"
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

	logger.CreateLog(logger.Debug)
	logger.Log(logger.Info, "Starting BillBank")

	vp := lib.ViewPort{}
	h := utils.NewCmdHistory()

	vp.Commander = commander.NewCmdInput(
		h,
		commander.WithCommands(
			commander.NewDebugCmd(h),
		), // commands.NewTmpCmd()),
	)

	p1 := tea.NewProgram(
		vp,
		tea.WithAltScreen(),
	)

	if _, err := p1.Run(); err != nil {
		panic(err)
	}
}
