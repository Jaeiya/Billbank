package commander

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/utils/logger"
)

func NewSkeletonCmd() Command {
	m := skeletonModel{
		BaseCommand: NewBaseCommand[skeletonModel]([][]string{
			// Aliases
			{"t", "test"},
			// Branches
			{"this", "that", "other"},
		}),
	}

	m.AddBranch([]BranchEntry[skeletonModel]{
		{
			"t this",
			func(sm *skeletonModel) tea.Cmd { return sm.loadThis() },
			func(sm skeletonModel) string {
				return fmt.Sprintf(
					"This Counter goes up every time you enter the command: %d",
					sm.thisCounter,
				)
			},
		},
		{
			// FIXME  This shouldn't be possible. Do not allow identical branches.
			"t this",
			func(sm *skeletonModel) tea.Cmd { return sm.loadThat() },
			func(sm skeletonModel) string {
				return fmt.Sprintf("This counter goes up every 300ms: %d", sm.thatCounter)
			},
		},
		// You can also set a nil view
		{
			"t other",
			func(sm *skeletonModel) tea.Cmd { return func() tea.Msg { return "" } },
			nil,
		},
	}...)

	return NewCommand(
		CommandConfig{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			HasArg:              false,
		},
	)
}

type skeletonModel struct {
	BaseCommand[skeletonModel]
	thisCounter int
	thatCounter int
}

func (m skeletonModel) Update(msg tea.Msg) (CommandModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	model, cmd := m.BaseCommand.Update(&m, msg)
	m = *model
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+h" {
			logger.Log(logger.Info, "hello from new command")
		}
	}

	return model, tea.Batch(cmds...)
}

func (m skeletonModel) View() string {
	return m.BaseCommand.View(m)
}

func (m skeletonModel) SetStatus(status CommandStatus) (CommandModel, tea.Cmd) {
	return m, m.BaseCommand.SetStatus(status)
}

func (m *skeletonModel) loadThis() tea.Cmd {
	m.thisCounter += 1
	return nil
}

func (m *skeletonModel) loadThat() tea.Cmd {
	m.thatCounter += 1
	return m.poll(time.Millisecond * 300)
}
