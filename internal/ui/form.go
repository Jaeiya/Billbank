package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/utils"
)

const blinkSpeed = 400
const inputSize = 18

type FormInputType uint8

const (
	AnyInput FormInputType = iota

	AlphaInput    // Allow Alphabet, Space, and Backspace keys only; no validation limits
	AlphaNumInput // Allow Alphabet, Number, Space, and Backspace keys only; no validation limits
	IntegerInput  // Allow Number and Backspace keys only; validates as positive int64
	FloatInput    // Allow Number, Decimal, and Backspace keys only; validates as positive float64
	PriceInput    // Allows FloatInput keys; validation enforces 2 decimal places
)

var (
	formItemTitleStyle = Style.Bold(true).
				Width(inputSize + 1).
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Gray).
				Foreground(White)
	formItemBorderStyle = Style.
				BorderRight(true).
				Width(inputSize + 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Gray)
)

type FormInput struct {
	title       string
	description string
	prompt      string
	inputType   FormInputType
	input       textinput.Model
	validator   func(s string) error
	isOptional  bool
}

func NewFormInput() FormInput {
	fi := FormInput{}
	fi.prompt = "> "

	fi.input = NewDefaultInput(inputSize - 3)

	s := fi.input.Styles()
	s.Cursor.Color = BrightGreen
	s.Cursor.BlinkSpeed = time.Millisecond * blinkSpeed
	s.Focused.Text = s.Focused.Text.Foreground(BrightBlue)
	s.Blurred.Text = s.Blurred.Text.Foreground(Green)
	s.Focused.Prompt = s.Focused.Prompt.Foreground(BrightGreen)
	s.Focused.Suggestion = s.Focused.Suggestion.Foreground(Gray)
	fi.input.SetStyles(s)

	return fi
}

func (fi FormInput) Title(s string) FormInput {
	fi.title = s
	return fi
}

func (fi FormInput) Description(s string) FormInput {
	fi.description = s
	return fi
}

func (fi FormInput) InputType(t FormInputType) FormInput {
	fi.inputType = t
	if t == PriceInput {
		fi.prompt = "$ "
	}
	return fi
}

func (fi FormInput) Value(s string) FormInput {
	fi.input.SetValue(s)
	return fi
}

func (fi FormInput) Validate(v func(s string) error) FormInput {
	fi.validator = v
	return fi
}

func (fi FormInput) Optional(v bool) FormInput {
	fi.isOptional = v
	return fi
}

func (fi FormInput) Suggestions(sugs ...string) FormInput {
	fi.input.ShowSuggestions = true
	fi.input.SetSuggestions(sugs)
	return fi
}

func (fi FormInput) Prompt(p string) FormInput {
	fi.prompt = p
	return fi
}

func (fi FormInput) view() string {
	isFocused := fi.input.Focused()
	titleStyle := formItemTitleStyle
	activeBorderStyle := Style.BorderRight(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(Gray)
	fi.input.Prompt = ""

	if isFocused {
		titleStyle = titleStyle.BorderForeground(BrightYellow).Foreground(Yellow)
		activeBorderStyle = activeBorderStyle.BorderForeground(BrightYellow)
		fi.input.SetWidth(inputSize - 3)
		fi.input.Prompt = fi.prompt
	}

	return Style.Width(inputSize + 5).Align(lipgloss.Left).Render(
		JoinVertical(lipgloss.Left,
			titleStyle.Render(fi.title),
			Style.Width(inputSize).Render(fi.input.View())+activeBorderStyle.Render(""),
			formItemBorderStyle.Render(""),
		),
	)

}

func (fi FormInput) validate() error {
	inputStr := fi.input.Value()

	if !fi.isOptional && len(fi.input.Value()) == 0 {
		return fmt.Errorf("%s is a required field", fi.title)
	}

	if fi.validator != nil {
		return fi.validator(inputStr)
	}

	switch fi.inputType {
	case AnyInput:
		return nil

	case AlphaNumInput:
		if !utils.IsAlphaStr(inputStr) {
			return errors.New("invalid characters in string")
		}
		return nil

	case IntegerInput:
		if !utils.IsIntStr(inputStr) {
			return errors.New("invalid number")
		}
		return nil

	case FloatInput:
		if !utils.IsFloatStr(inputStr) {
			return errors.New("invalid float")
		}
		return nil

	case PriceInput:
		if utils.IsIntStr(inputStr) {
			return nil
		}
		if !utils.IsFloatStr(inputStr) {
			return errors.New("invalid price")
		}
		if len(inputStr) > 2 && inputStr[len(inputStr)-3] != '.' {
			return errors.New("too many or too few cent places")
		}
		return nil

	default:
		return errors.New("fatal::missing form input type")
	}

}

type Form struct {
	entries []FormInput
	values  []string
	header  string
	tabPos  int
	isInit  bool
	err     error
}

func NewForm(header string, inputs ...FormInput) Form {
	f := Form{}
	f.header = header
	f.isInit = true
	f.entries = inputs
	for i := range f.entries {
		if i == 0 {
			continue
		}
		f.entries[i].input.Blur()
	}
	f.values = make([]string, len(inputs))
	return f
}

func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "ctrl+j":
			if f.tabPos == len(f.entries)-1 {
				return f, nil
			}

			f.err = f.entries[f.tabPos].validate()
			if f.err != nil { // do not tab on error
				return f, nil
			}

			f.entries[f.tabPos].input.Blur()
			f.tabPos++
			f.entries[f.tabPos].input.Focus()
			f.isInit = true

		case "tab":
			// Manually update focused input
			f.entries[f.tabPos].input, cmd = f.entries[f.tabPos].input.Update(msg)

			f.err = f.entries[f.tabPos].validate()
			if f.err != nil { // do not tab on error
				return f, nil
			}
			return f, cmd

		case "shift+tab", "ctrl+k":
			if f.tabPos == 0 {
				return f, nil
			}

			f.err = f.entries[f.tabPos].validate()
			if f.err != nil { // do not tab on error
				return f, nil
			}

			f.entries[f.tabPos].input.Blur()
			f.tabPos--
			f.entries[f.tabPos].input.Focus()
			f.isInit = true
		}
	}

	cmds = append(cmds, f.updateInputs(msg))

	// Jump-start blinking virtual cursor
	if f.isInit {
		cmds = append(cmds, textinput.Blink)
		f.isInit = false
	}

	return f, tea.Batch(cmds...)
}

func (f Form) View() string {
	var sb strings.Builder

	for _, entry := range f.entries {
		sb.WriteString(entry.view())
		sb.WriteString("\n")
	}

	formStatus := f.formStatus("Ok", true)
	if f.err != nil {
		formStatus = f.formStatus(f.err.Error(), false)
	}

	header := Style.Foreground(BrightMagenta).Render(f.header)
	form := JoinHorizontal(lipgloss.Left, Style.Width(inputSize+2).PaddingTop(1).Render(sb.String()), formStatus)

	return JoinVertical(lipgloss.Center, header, form)
}

func (f Form) formStatus(status string, isGood bool) string {
	const formSize = 35

	statusWrapper := Style.
		Width(formSize).
		Foreground(White).
		MarginTop(1).
		MarginLeft(1).
		Padding(0, 1, 0)

	entry := f.entries[f.tabPos]
	title := entry.title
	if entry.isOptional {
		title = JoinHorizontal(lipgloss.Left, title, Style.Foreground(Magenta).Render(" (Optional)"))
	}

	formHead := Style.Foreground(Yellow).Align(lipgloss.Left).Render(title)

	statusChar := "\u2713 "
	statusFg := FgSuccessColor
	if !isGood {
		statusChar = "\u2717 "
		statusFg = FgErrColor
	}

	formBody := Style.Foreground(statusFg).Render(statusChar + status)

	if len(f.entries[f.tabPos].description) > 0 {
		return statusWrapper.Render(
			JoinVertical(
				lipgloss.Left,
				formHead,
				"",
				Style.Foreground(Blue).Render(f.entries[f.tabPos].description),
				"",
				formBody,
			),
		)
	}

	return statusWrapper.Render(
		JoinVertical(
			lipgloss.Left,
			formHead,
			"",
			formBody,
		),
	)
}

func (f *Form) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	for i := range f.entries {
		if i == f.tabPos {
			var key rune

			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				key = msg.Code
				switch msg.String() {
				// Do not restrict control keys
				case "enter", "tab", "backspace":
					f.entries[i].input, cmd = f.entries[i].input.Update(msg)
					return cmd
				}

			default:
				f.entries[i].input, cmd = f.entries[i].input.Update(msg)
				return cmd
			}

			switch f.entries[i].inputType {
			case AlphaInput:
				if utils.IsAlpha(key) || key == 32 {
					f.entries[i].input, cmd = f.entries[i].input.Update(msg)
				}

			case AlphaNumInput:
				// Only allows numbers, upper/lower case alphabet, and space
				if utils.IsAlphaNum(key) || key == 32 {
					f.entries[i].input, cmd = f.entries[i].input.Update(msg)
				}

			case IntegerInput:
				// Only allow numbers
				if utils.IsNumber(key) {
					f.entries[i].input, cmd = f.entries[i].input.Update(msg)
				}

			case FloatInput, PriceInput:
				// Only allow decimal and numbers
				if key == '.' || utils.IsNumber(key) {
					f.entries[i].input, cmd = f.entries[i].input.Update(msg)
				}

			default:
				f.entries[i].input, cmd = f.entries[i].input.Update(msg)
			}
		}
	}
	return cmd
}
