package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/ui"
	"github.com/jaeiya/billbank/lib/utils"
)

func main() {
	// wd, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }

	// filePath := filepath.Join(wd, "hello.db")
	// defer func() {
	// 	err := os.Remove(filePath)
	// 	if err != nil {
	// 		panic(err)
	//
	// }()
	// db := sqlite.NewSqliteDb(filePath, lib.USD)
	// defer db.Close()
	// ti := components.NewCommanderInput()
	utils.CreateLog()
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

	// p := tea.NewProgram(
	// 	ui.NewCmdInput(
	// 		ui.WithCommands(commands.NewTestCmd()),
	// 	),
	// 	tea.WithAltScreen(),
	// )
	// if _, err := p.Run(); err != nil {
	// 	panic(err)
	// }

	// cmd := commands.NewTestCmd()
	// res := cmd.ParseCommand("view bills amount")
	// fmt.Printf("%v\n", res)
}
