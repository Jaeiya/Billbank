
import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/billbank/lib/cmd"
	"github.com/jaeiya/billbank/lib/logger"
)

// # How a command works
//
// Each command has a command tree and it has a very specific hierarchy that
// lays out how commands are processed. There are branches and leaves. A
// leaf is a single node in a command branch and a branch is a string of
// nodes pulled one at a time from each slice within the tree.
//
// Example Tree:
//
//	[][]string{{"set"}, {"bill", "stat"}, {"amount", "name"}}
//
// Leaves:
//
//	"set", "bill", "stat", "amount", and "name"
//
// Branches
//
//	"set"
//	"set bill"
//	"set stat"
//	"set bill amount"
//	"set bill name"
//	"set stat amount"
//	"set stat name"
//
// Each branch is capable of being executed as a command, but not all
// branches need to be supported. Any unsupported branches that are
// executed, will result in a status msg indicating that it's
// unsupported.
func NewSkeletonCmd() cmd.Command {
	//
	// Anything that needs to be initialized should be here
	// so that it can be passed to your model.
	//
	m := skeletonModel{
		BaseModel: cmd.NewBaseModel[skeletonModel]([][]string{
			{"t", "test"}, // Aliases
			{"this", "that", "other", "nil", "nilview"}, // Branches
		}),
	}

	//
	// Will panic if you forget to add branches
	//
	m.AddBranch([]cmd.Branch[skeletonModel]{
		// If you want to use aliases, you'll need to add
		// a separate branch with the same funcs. We have
		// two branches below that reference the "this"
		// command.
		{String: "t this", Fn: loadThis, ViewFn: thisView},
		{String: "test this", Fn: loadThis, ViewFn: thisView},

		{String: "t that", Fn: loadThat, ViewFn: thatView},

		// Nil values are also valid
		{String: "t nil", Fn: nil, ViewFn: nil},
		{
			String: "t nilview",
			Fn:     func(sm skeletonModel) skeletonModel { return sm },
			ViewFn: nil,
		},
	}...)

	return cmd.New(
		cmd.Config{
			Model:               m,
			InputValidationFunc: func(arg string) error { return nil },
			KeyValidationFunc:   func(key rune) bool { return false },
			// You cannot mix arg & non-arg commands
			HasArg: false,
		},
	)
}

type skeletonModel struct {
	*cmd.BaseModel[skeletonModel]
	thisCounter int
	thatCounter int
}

func (m skeletonModel) Update(msg tea.Msg) (cmd.Model, tea.Cmd) {
	var teaCmd tea.Cmd
	var teaCmds []tea.Cmd

	m, teaCmd = m.BaseModel.Update(m, msg)
	teaCmds = append(teaCmds, teaCmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Executes for every branch
		if msg.String() == "ctrl+h" {
			logger.Log(logger.Info, "Skeleton", "hello from new command")
		}

		// Only watch certain keys on certain branches
		if m.IsActiveBranch("t this") || m.IsActiveBranch("test this") {
			if msg.String() == "ctrl+k" {
				m.thisCounter += 1
			}
		}
	}

	return m, tea.Batch(teaCmds...)
}

func (m skeletonModel) View() string {
	return m.BaseModel.View(m)
}

func loadThis(m skeletonModel) skeletonModel {
	return m
}

func thisView(m skeletonModel) string {
	return fmt.Sprintf(
		"Hit ctrl+k to increment the counter: %d",
		m.thisCounter,
	)
}

func loadThat(m skeletonModel) skeletonModel {
	m.thatCounter += 1
	return m
}

func thatView(m skeletonModel) string {
	return fmt.Sprintf("Execute the command again to increment the counter: %d", m.thatCounter)
}
