package cmdcore

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jaeiya/billbank/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommand(t *testing.T) {
	t.Parallel()

	cmd1 := NewCommand(CommandOptions[mockCommandModel, NoArg]{
		Path: "",
		RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
			model.someInt = 1
			return model, nil
		},
		ViewFunc: defaultView,
	})
	cmd2 := NewCommand(CommandOptions[mockCommandModel, NoArg]{
		Path: "test",
		RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
			model.someInt = 1
			return model, nil
		},
		ViewFunc: defaultView,
	})
	cmd3 := NewCommand(CommandOptions[mockCommandModel, int]{
		Path:    "arg",
		ArgType: ArgRequired,
		RunFunc: func(model mockCommandModel, arg *int) (CommandModel, error) {
			model.someInt = 1
			return model, nil
		},
		ViewFunc: defaultView,
		ParseFunc: func(arg string) (int, error) {
			return utils.ParseInt(arg)
		},
	})

	handler := NewCmdHandler(
		"mock",
		[]string{"mock", "m"},
		[]Command[mockCommandModel]{cmd1, cmd2, cmd3},
		mockCommandModel{},
	)

	tests := []struct {
		should    string
		input     string
		wantArg   bool
		wantError error
		wantPath  string
	}{
		{"recognizes default command path", "mock", false, nil, "mock"},
		{"recognizes alias", "m test", false, nil, "m test"},
		{"errors on unrecognized command", "mo", false, ErrNotCommand, ""},
		{"errors on incomplete command", "mock te", false, ErrIncompleteCmd, ""},
		{"errors on missing argument", "mock arg", false, ErrMissingArg, ""},
		{"errors on empty command", "      ", false, ErrEmptyCommand, ""},
		{"accepts args", "mock arg 10", true, nil, "mock arg"},
		{
			"errors on failed arg parse",
			"mock arg asdf",
			true,
			fmt.Errorf("not a number"),
			"mock arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.should, func(t *testing.T) {
			got := handler.ParseCommand(tt.input)

			if tt.wantArg {
				if tt.wantError != nil {
					assert.ErrorContains(t, got.Error, tt.wantError.Error())
				} else {
					require.NoError(t, got.Error)
					assert.Equal(t, got.Path, tt.wantPath)
				}
				return
			}

			if tt.wantError != nil {
				require.ErrorIs(t, got.Error, tt.wantError)
			} else {
				require.NoError(t, got.Error)
				assert.Equal(t, got.Path, tt.wantPath)
			}
		})
	}
}

func TestCommandHandler(t *testing.T) {
	t.Parallel()

	cmd1 := NewCommand(CommandOptions[mockCommandModel, NoArg]{
		Path: "abc",
		RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
			return model, nil
		},
		ViewFunc: defaultView,
	})
	cmd2 := NewCommand(CommandOptions[mockCommandModel, NoArg]{
		Path: "def",
		RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
			return model, nil
		},
		ViewFunc: defaultView,
	})
	cmd3 := NewCommand(CommandOptions[mockCommandModel, NoArg]{
		Path: "ghi jkl",
		RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
			return model, nil
		},
		ViewFunc: defaultView,
	})

	t.Run("successfully creates a command handler", func(t *testing.T) {
		NewCmdHandler(
			"mock",
			[]string{"m"},
			[]Command[any]{cmd1, cmd2, cmd3},
			nil,
		)
	})

	t.Run("creates proper command mappings", func(t *testing.T) {
		aliases := []string{"mock", "m"}

		handler := NewCmdHandler("test", aliases, []Command[any]{cmd1, cmd2, cmd3}, nil)
		paths := handler.CommandPaths()

		// Ensure all combinations are created (3 cmds * 2 aliases)
		require.Len(t, paths, 6, "Should have 6 mapped command paths")

		expected := []string{
			"mock abc",
			"mock def",
			"mock ghi jkl",
			"m abc",
			"m def",
			"m ghi jkl",
		}
		assert.ElementsMatch(t, expected, paths)
	})
}

func TestPopulateSuggestions(t *testing.T) {
	cmdPaths := []string{"cmd status", "cmd start", "c status", "c start timer"}

	t.Run("matches partial input", func(t *testing.T) {
		suggestions := populateSuggestions(cmdPaths, strings.Fields("cmd st"))

		assert.Len(t, suggestions, 4)
		assert.Equal(
			t,
			suggestions,
			[]string{"cmd status", "cmd start", "c status", "c start"},
			"should contain all second tier paths",
		)

		suggestions2 := populateSuggestions(cmdPaths, strings.Fields("c start a"))

		assert.Len(t, suggestions2, 1)
		assert.Equal(
			t,
			suggestions2,
			[]string{"c start timer"},
			"should contain all third tier paths",
		)
	})

	t.Run("has default matches", func(t *testing.T) {
		suggestions := populateSuggestions(cmdPaths, []string{"other"})
		assert.Len(t, suggestions, 4)
		assert.Equal(t, suggestions, []string{"cmd", "cmd", "c", "c"})
	})
}
