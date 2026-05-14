package cmdcore

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/jaeiya/billbank/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCommandModel struct {
	ModelBase
	someInt int
}

func (m mockCommandModel) Update(_ tea.Msg) (CommandModel, tea.Cmd) { return m, nil }

type mockCommandModel2 struct{ ModelBase }

func (m mockCommandModel2) Update(_ tea.Msg) (CommandModel, tea.Cmd) { return m, nil }

func defaultView(mockCommandModel) tea.View { return tea.NewView("") }

func TestCommand(t *testing.T) {
	t.Parallel()

	t.Run("executes command successfully", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, NoArg]{
			RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
				model.someInt = 1
				return model, nil
			},
			ViewFunc: defaultView,
		})

		m, err := cmd.Run(mockCommandModel{})
		require.NoError(t, err, "should run command without error")

		v, isType := m.(mockCommandModel)
		if !isType {
			assert.Fail(t, "RunFunc should be executed")
		}
		assert.Equal(t, v.someInt, 1)
	})

	t.Run("fails when wrong model passed to run func", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, NoArg]{
			RunFunc: func(model mockCommandModel, _ *NoArg) (CommandModel, error) {
				return model, nil
			},
			ViewFunc: defaultView,
		})

		_, err := cmd.Run(0)
		require.ErrorIs(t, err, ErrModelTypeMismatch)
	})

	t.Run("fails when wrong model returned from run func", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, NoArg]{
			RunFunc: func(_ mockCommandModel, _ *NoArg) (CommandModel, error) {
				return mockCommandModel2{}, nil
			},
			ViewFunc: defaultView,
		})

		_, err := cmd.Run(mockCommandModel{})
		require.ErrorIs(t, err, ErrModelTypeReturnMismatch)
	})

	t.Run("fails when wrong model passed to view", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, NoArg]{
			RunFunc: func(_ mockCommandModel, _ *NoArg) (CommandModel, error) {
				return mockCommandModel2{}, nil
			},
			ViewFunc: defaultView,
		})

		_, err := cmd.View(mockCommandModel2{})
		require.ErrorIs(t, err, ErrModelTypeMismatch)
	})

	t.Run("resolves argument with method", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, int]{
			ArgType: ArgRequired,
			RunFunc: func(model mockCommandModel, arg *int) (CommandModel, error) {
				if arg != nil {
					model.someInt = *arg
				}
				return model, nil
			},
			ViewFunc:  defaultView,
			ParseFunc: func(arg string) (int, error) { return utils.ParseInt(arg) },
		})

		err := cmd.ResolveArg("5")
		require.NoError(t, err)
		mm := mockCommandModel{}
		v, err := cmd.Run(mm)
		require.NoError(t, err)

		m, ok := v.(mockCommandModel)
		assert.True(t, ok, true)
		assert.Equal(t, m.someInt, 5)
	})

	t.Run("clears argument on each run", func(t *testing.T) {
		cmd := NewCommand(CommandOptions[mockCommandModel, int]{
			ArgType: ArgRequired,
			RunFunc: func(model mockCommandModel, arg *int) (CommandModel, error) {
				if arg != nil {
					model.someInt = *arg
				} else {
					model.someInt = 0
				}
				return model, nil
			},
			ViewFunc:  defaultView,
			ParseFunc: func(arg string) (int, error) { return utils.ParseInt(arg) },
		})

		err := cmd.ResolveArg("10")
		require.NoError(t, err)

		// First run
		m1, err := cmd.Run(mockCommandModel{})
		require.NoError(t, err)

		mm1, ok := m1.(mockCommandModel)
		require.True(t, ok)
		assert.Equal(t, 10, mm1.someInt, "arg should be passed through")

		// Second run
		m2, err := cmd.Run(mockCommandModel{})
		require.NoError(t, err)

		mm2, ok := m2.(mockCommandModel)
		require.True(t, ok)
		assert.Equal(t, 0, mm2.someInt, "arg should be cleared of value")
	})
}
