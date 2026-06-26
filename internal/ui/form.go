package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jaeiya/billbank/internal/utils"
)

const (
	blinkSpeed = 400
	inputSize  = 18
)

type FormInputType uint8

const (
	AnyInput FormInputType = iota

	AlphaInput    // Alphabet, Space, and Backspace keys only; no validation
	AlphaNumInput // Alphabet, Number, Space, and Backspace keys only; no validation
	IntegerInput  // Number and Backspace keys only; validates as positive int64
	FloatInput    // Number, Decimal, and Backspace keys only; validates as positive float64
	PriceInput    // FloatInput keys only; validation enforces 2 decimal places
)

var (
	formActiveColor = BrightYellow
	formBorderColor = lipgloss.Color("#505072")
)

var formStyles = struct {
	itemTitle   lipgloss.Style
	inputBorder lipgloss.Style
	textInput   textinput.Styles
}{
	itemTitle: Style.
		Bold(true).
		Width(inputSize + 1).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(formBorderColor),

	inputBorder: Style.
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(formBorderColor),

	textInput: func() textinput.Styles {
		s := textinput.DefaultStyles(false)
		s.Cursor.Color = BrightGreen
		s.Cursor.BlinkSpeed = time.Millisecond * blinkSpeed
		s.Focused.Text = s.Focused.Text.Foreground(BrightBlue)
		s.Blurred.Text = s.Blurred.Text.Foreground(Green)
		s.Focused.Prompt = s.Focused.Prompt.Foreground(BrightGreen)
		s.Focused.Suggestion = s.Focused.Suggestion.Foreground(Gray)
		return s
	}(),
}

type FormInput struct {
	input       textinput.Model
	validator   func(s string) error
	title       string
	description string
	prompt      string
	inputType   FormInputType
	isOptional  bool
}

func NewFormInput() FormInput {
	fi := FormInput{}
	fi.prompt = "> "
	fi.input = NewDefaultInput(inputSize - 3)
	fi.input.SetStyles(formStyles.textInput)
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

// Validator validates the form input with the provided validator
// func.
//
// 🟡 Overrides default validation activated by InputType()
func (fi FormInput) Validator(v func(s string) error) FormInput {
	fi.validator = v
	return fi
}

// Optional sets the field to optional. By default, all input
// fields are required.
func (fi FormInput) Optional() FormInput {
	fi.isOptional = true
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
	titleStyle := formStyles.itemTitle
	borderStyle := formStyles.inputBorder

	fi.input.Prompt = ""

	borderUtil := utils.GetBorderStyle(utils.SingleLine)
	focusedLine := string(borderUtil.RightT)

	if isFocused {
		titleStyle = titleStyle.BorderForeground(formActiveColor).Foreground(formActiveColor)
		borderStyle = borderStyle.BorderForeground(formActiveColor)
		fi.input.SetWidth(inputSize - 3)
		fi.input.Prompt = fi.prompt
		focusedLine = strings.Repeat(string(borderUtil.Horizontal), inputSize-1) +
			Style.Foreground(formActiveColor).Render(string(borderUtil.Horizontal)+
				string(borderUtil.RightT),
			)
	}

	return Style.Width(inputSize + 5).Align(lipgloss.Left).Render(
		JoinVertical(
			lipgloss.Left,
			titleStyle.Render(fi.title),
			Style.Width(inputSize).Render(fi.input.View())+borderStyle.Render(""),
			Style.Foreground(formBorderColor).
				Width(inputSize+1).
				Align(lipgloss.Right).
				Render(focusedLine),
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
	buttons struct {
		save   Button
		cancel Button
	}
	tabPos int
	isInit bool
	err    error
}

func NewForm(header string, inputs ...FormInput) Form {
	f := Form{
		header:  header,
		isInit:  true,
		entries: inputs,
	}

	saveButton := NewButton("Save")
	cancelButton := NewButton("Cancel")

	s := saveButton.Styles()
	s.Text.Focused = s.Text.Focused.Foreground(BrightGreen)
	s.Text.Blurred = s.Text.Blurred.Foreground(Gray)
	s.Selected.Foreground(White)
	saveButton.SetStyle(s)

	s.Text.Focused = s.Text.Focused.Foreground(BrightRed)
	cancelButton.SetStyle(s)

	f.buttons.save = saveButton
	f.buttons.cancel = cancelButton

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
		case "ctrl+j":
			if f.hasFocusedButtons() {
				return f, nil
			}
			f.entries[f.tabPos].input.Blur()
			f.tabPos = len(f.entries) - 1
			f.buttons.save.Focus()
			return f, nil

		case "enter":
			if f.tabPos == len(f.entries)-1 {
				if f.hasFocusedButtons() {
					return f, nil
				}
				f.entries[f.tabPos].input.Blur()
				f.buttons.save.Focus()
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

		case "h", "left":
			if f.buttons.cancel.Focused() {
				f.buttons.cancel.Blur()
				f.buttons.save.Focus()
			}

		case "l", "right":
			if f.buttons.save.Focused() {
				f.buttons.save.Blur()
				f.buttons.cancel.Focus()
			}

		case "tab":
			// Toggle buttons back and forth
			if f.hasFocusedButtons() {
				if f.buttons.save.Focused() {
					f.buttons.save.Blur()
					f.buttons.cancel.Focus()
				} else {
					f.buttons.cancel.Blur()
					f.buttons.save.Focus()
				}
				return f, nil
			}

			// Manually update focused input
			f.entries[f.tabPos].input, cmd = f.entries[f.tabPos].input.Update(msg)
			cmds = append(cmds, cmd)

			f.err = f.entries[f.tabPos].validate()
			if f.err != nil { // do not tab on error
				return f, nil
			}

			if f.tabPos+1 == len(f.entries) {
				f.buttons.save.Focus()
				f.entries[f.tabPos].input.Blur()
				return f, cmd
			}

			f.entries[f.tabPos].input.Blur()
			f.tabPos++
			cmds = append(cmds, f.entries[f.tabPos].input.Focus())
			return f, tea.Batch(cmds...)

		case "backspace":
			if f.hasFocusedButtons() {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				return f, f.entries[f.tabPos].input.Focus()
			}

		case "shift+tab":
			if f.hasFocusedButtons() {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				return f, f.entries[f.tabPos].input.Focus()
			}

			if f.tabPos == 0 {
				return f, nil
			}

			f.err = f.entries[f.tabPos].validate()
			if f.err != nil { // do not tab on error
				return f, nil
			}

			f.entries[f.tabPos].input.Blur()
			f.tabPos--
			f.isInit = true
			return f, f.entries[f.tabPos].input.Focus()
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

	for i, entry := range f.entries {
		if i != 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(entry.view())
	}

	formStatus := f.formStatus("Ok", true)
	if f.err != nil {
		formStatus = f.formStatus(f.err.Error(), false)
	}

	cancelView := f.buttons.cancel.View() + " "
	if f.buttons.cancel.Focused() {
		cancelView = strings.TrimRight(cancelView, " ")
	}

	buttons := f.buttons.save.View() + "   " + cancelView

	border := Style.Border(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		BorderForeground(lipgloss.Color("#505072"))
	form := border.Render(
		JoinVertical(
			lipgloss.Right,
			JoinHorizontal(
				lipgloss.Left,
				Style.Width(inputSize+2).Render(sb.String()),
				formStatus,
			),
			buttons,
		),
	)
	header := Style.Foreground(BrightMagenta).
		Width(lipgloss.Width(form)).
		Align(lipgloss.Center).
		Render(f.header)

	return JoinVertical(lipgloss.Left, header, form)
}

func (f Form) hasFocusedButtons() bool {
	return f.buttons.save.Focused() || f.buttons.cancel.Focused()
}

func (f Form) formStatus(status string, isGood bool) string {
	const formSize = 35

	statusWrapper := Style.
		Width(formSize).
		MarginTop(0).
		Foreground(White).
		MarginLeft(1).
		Padding(0, 1, 0)

	entry := f.entries[f.tabPos]
	title := entry.title
	if entry.isOptional {
		title = JoinHorizontal(
			lipgloss.Left,
			title,
			Style.Foreground(Magenta).Render(" (Optional)"),
		)
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
		// Focused input
		if i == f.tabPos {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				if !f.isValidKey(f.entries[i].inputType, msg) {
					return cmd
				}
			}

			f.entries[i].input, cmd = f.entries[i].input.Update(msg)
			break
		}
	}
	return cmd
}

// isValidKey returns true if the specified input type allows
// the pressed key.
func (f Form) isValidKey(t FormInputType, keyMsg tea.KeyPressMsg) bool {
	key := keyMsg.Code
	keyStr := keyMsg.String()

	// Special keys that should always be allowed
	allowedKeyMatches := [4]bool{
		keyStr == "enter",
		keyStr == "tab",
		keyStr == "backspace",
		keyMsg.Mod.Contains(tea.ModAlt),
	}

	if slices.Contains(allowedKeyMatches[:], true) {
		return true
	}

	switch t {
	case AnyInput:
		return true

	case AlphaInput:
		return utils.IsAlpha(key) || key == ' '

	case AlphaNumInput:
		return utils.IsAlphaNum(key) || key == ' '

	case IntegerInput:
		return utils.IsNumber(key)

	case FloatInput, PriceInput:
		return key == '.' || utils.IsNumber(key)
	}

	return false
}
