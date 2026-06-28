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
	inputWidth     = 18
	inputMargin    = 2              // Should always be >= 2
	inputCharLimit = inputWidth - 2 // Should always be < inputWidth
	blinkSpeed     = 400
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

type formFocusState uint8

const (
	formFocusInput formFocusState = iota
	formFocusSave
	formFocusCancel
)

var (
	formActiveColor = BrightYellow
	formBorderColor = lipgloss.Color("#505072")
)

type (
	FormSavedMsg  []string
	FormCancelMsg struct{}
)

var formStyles = struct {
	border      lipgloss.Style
	header      lipgloss.Style
	itemTitle   lipgloss.Style
	inputBorder lipgloss.Style
	textInput   textinput.Styles
	statusTitle lipgloss.Style
}{
	border: Style.
		PaddingLeft(1).
		PaddingRight(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(formBorderColor),

	header: Style.Foreground(BrightMagenta).Align(lipgloss.Center),

	itemTitle: Style.Bold(true),

	inputBorder: Style.
		Width(inputWidth + inputMargin).
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

	statusTitle: Style.Foreground(Yellow).Align(lipgloss.Left),
}

type FormInput struct {
	input       textinput.Model
	validator   func(s string, defaultValidator func() error) error
	title       string
	description string
	prompt      string
	inputType   FormInputType
	isOptional  bool
}

func NewFormInput() FormInput {
	fi := FormInput{}
	fi.prompt = "> "
	fi.input = NewDefaultInput(inputCharLimit)
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

// Validator overwrites the default InputType validator with
// the one specified. Not all InputType's have a validator.
//
// 🔵 The default validator is passed as the defaultV arg.
func (fi FormInput) Validator(v func(s string, defaultV func() error) error) FormInput {
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

	fi.input.Prompt = ""

	if isFocused {
		titleStyle = titleStyle.Foreground(formActiveColor)
		fi.input.SetWidth(inputWidth)
		fi.input.Prompt = fi.prompt
	}

	return JoinVertical(
		lipgloss.Left,
		titleStyle.Render(fi.title),
		fi.input.View(),
	)
}

func (fi FormInput) validate() error {
	inputStr := fi.input.Value()

	if !fi.isOptional && len(fi.input.Value()) == 0 {
		return fmt.Errorf("%s is a required field", fi.title)
	}

	var defaultValidator func() error

	switch fi.inputType {
	case AnyInput, AlphaInput:
		defaultValidator = func() error {
			return nil
		}

	case AlphaNumInput:
		defaultValidator = func() error {
			if !utils.IsAlphaStr(inputStr) {
				return errors.New("invalid characters in string")
			}
			return nil
		}

	case IntegerInput:
		defaultValidator = func() error {
			if !utils.IsIntStr(inputStr) {
				return errors.New("invalid number")
			}
			return nil
		}

	case FloatInput:
		defaultValidator = func() error {
			if !utils.IsFloatStr(inputStr) {
				return errors.New("invalid float")
			}
			return nil
		}

	case PriceInput:
		defaultValidator = func() error {
			strLen := len(inputStr)

			if utils.IsIntStr(inputStr) {
				return nil
			}

			if !utils.IsFloatStr(inputStr) {
				return errors.New("malformed price")
			}

			if inputStr[strLen-1] == '.' {
				return errors.New("trailing decimal not allowed")
			}

			if inputStr[strLen-2] == '.' {
				return errors.New("missing cents place")
			}

			if inputStr[strLen-3] != '.' {
				return errors.New("too many cents places")
			}
			return nil
		}
	}

	if defaultValidator == nil {
		return errors.New("fatal::missing form input type ")
	}

	if fi.validator != nil {
		return fi.validator(inputStr, defaultValidator)
	}

	return defaultValidator()
}

type Form struct {
	entries []FormInput
	values  []string
	header  string
	buttons struct {
		save   Button
		cancel Button
	}
	focus  formFocusState
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
			if f.focus > formFocusInput {
				return f, nil
			}
			f.entries[f.tabPos].input.Blur()
			f.tabPos = len(f.entries) - 1
			f.focus = formFocusSave
			f.buttons.save.Focus()
			return f, nil

		case "enter":
			if f.tabPos == len(f.entries)-1 {
				if f.focus > formFocusInput {
					if f.focus == formFocusSave {
						return f, f.SendSaveMsg()
					}
					if f.focus == formFocusCancel {
						return f, f.SendCancelMsg()
					}
					return f, nil
				}
				f.entries[f.tabPos].input.Blur()
				f.focus = formFocusSave
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
				f.focus = formFocusSave
			}

		case "l", "right":
			if f.buttons.save.Focused() {
				f.buttons.save.Blur()
				f.buttons.cancel.Focus()
				f.focus = formFocusCancel
			}

		case "tab":
			// Toggle buttons back and forth
			if f.focus > formFocusInput {
				if f.buttons.save.Focused() {
					f.buttons.save.Blur()
					f.buttons.cancel.Focus()
					f.focus = formFocusCancel
				} else {
					f.buttons.cancel.Blur()
					f.buttons.save.Focus()
					f.focus = formFocusSave
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
				f.focus = formFocusSave
				f.entries[f.tabPos].input.Blur()
				return f, cmd
			}

			f.entries[f.tabPos].input.Blur()
			f.tabPos++
			cmds = append(cmds, f.entries[f.tabPos].input.Focus())
			return f, tea.Batch(cmds...)

		case "backspace":
			if f.focus > formFocusInput {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				f.focus = formFocusInput
				return f, f.entries[f.tabPos].input.Focus()
			}

		case "shift+tab":
			if f.focus > formFocusInput {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				f.focus = formFocusInput
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
	cancelView := f.buttons.cancel.View() + " "
	if f.buttons.cancel.Focused() {
		cancelView = strings.TrimRight(cancelView, " ")
	}

	buttonView := f.buttons.save.View() + "   " + cancelView

	formStatusView := f.viewFormStatus("Ok", true)
	if f.err != nil {
		formStatusView = f.viewFormStatus(f.err.Error(), false)
	}

	formView := formStyles.border.Render(
		JoinVertical(lipgloss.Right,
			JoinHorizontal(lipgloss.Left, f.viewInputs(), formStatusView),
			buttonView,
		),
	)

	return Style.
		Render(JoinVertical(lipgloss.Center,
			formStyles.header.Render(f.header),
			formView,
		))
}

func (f Form) viewInputs() string {
	var sb strings.Builder
	sb.Grow((inputWidth + inputMargin) * len(f.entries))

	for i, entry := range f.entries {
		if i != 0 {
			sb.WriteByte('\n')
		}

		borderColor := formBorderColor
		if i == f.tabPos && f.focus == formFocusInput {
			borderColor = formActiveColor
		}

		b := utils.GetBorderStyle(utils.Rounded)

		sb.WriteString(formStyles.inputBorder.BorderForeground(borderColor).Render(entry.view()))
		sb.WriteByte('\n')

		// Use corner border for last input entry
		activeBorderChar := string(b.RightT)
		if i == len(f.entries)-1 {
			activeBorderChar = string(b.BottomRight)
		}

		if f.tabPos == i && f.focus == formFocusInput {
			sb.WriteString(Style.
				Width(inputWidth + inputMargin).
				Foreground(formBorderColor).
				Render(utils.GenDashedBorder(inputWidth+inputMargin-1) + string(activeBorderChar)),
			)
		} else {
			sb.WriteString(formStyles.inputBorder.Render(""))
		}
	}

	return sb.String()
}

func (f Form) viewFormStatus(status string, isGood bool) string {
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

	var statusTitle, description, statusText string

	statusChar := "\u2713 "
	statusFg := FgSuccessColor
	if !isGood || f.focus == formFocusCancel {
		statusChar = "\u2717 "
		statusFg = FgErrColor
	}

	statusTextStyle := Style.Foreground(statusFg)

	switch f.focus {
	case formFocusSave:
		statusTitle = "Save Form"
		statusText = statusTextStyle.Render(statusChar + "Form is ready to save")
		description = "Take a moment to look over the form and make sure it's correct before saving."
	case formFocusCancel:
		statusTitle = "Cancel Form"
		description = "The data you have entered will not be saved. The old data will remain intact."
		statusText = statusTextStyle.Render(statusChar + "Form will be discarded")
	default:
		statusTitle = title
		if len(f.entries[f.tabPos].description) > 0 {
			description = f.entries[f.tabPos].description
		}
		statusText = statusTextStyle.Render(statusChar + status)
	}

	return statusWrapper.Render(
		JoinVertical(
			lipgloss.Left,
			formStyles.statusTitle.Render(statusTitle),
			"",
			Style.Foreground(Blue).Render(description),
			"",
			statusText,
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

func (f Form) SendSaveMsg() tea.Cmd {
	data := make([]string, len(f.entries))
	for i, entry := range f.entries {
		data[i] = entry.input.Value()
	}

	return func() tea.Msg {
		return FormSavedMsg(data)
	}
}

func (f Form) SendCancelMsg() tea.Cmd {
	return func() tea.Msg {
		return FormCancelMsg{}
	}
}
