package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/commands"
	"github.com/jaeiya/billbank/lib/ui"
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

	vp := ui.ViewPort{}

	vp.Commander = ui.NewCmdInput(
		ui.WithCommands(commands.NewTestCmd()),
	)

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
