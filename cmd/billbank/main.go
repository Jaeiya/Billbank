package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/ui"
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

	utils.CreateLog(utils.Debug)
	utils.Log(utils.Info, "Starting BillBank")

	vp := ui.ViewPort{}
	h := utils.NewCmdHistory()

	vp.Commander = ui.NewCmdInput(h, ui.WithCommands(commands.NewDebugCmd(h)))

	p1 := tea.NewProgram(
		vp,
		tea.WithAltScreen(),
	)

	if _, err := p1.Run(); err != nil {
		panic(err)
	}
}
