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
	inputWidth     = 20
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
	CurrencyInput // FloatInput keys only; validation enforces USD amounts to 2 decimal places
)

type formFocusState uint8

const (
	formFocusInput formFocusState = iota
	formFocusSave
	formFocusCancel
)

var formActiveColor = BrightYellow

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
		BorderForeground(DarkBorderColor),

	header: Style.Foreground(BrightMagenta).Align(lipgloss.Center),

	itemTitle: Style.Bold(true),

	inputBorder: Style.
		Width(inputWidth + inputMargin).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(DarkBorderColor),

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

type FormValidatorFunc func(s string, defaultValidator func() error, args ...string) error

type FormField struct {
	input       textinput.Model
	validator   FormValidatorFunc
	title       string
	description string
	descFunc    func(args ...string) string
	prompt      string
	inputType   FormInputType
	linkedInput string
	isOptional  bool
}

func NewFormField() FormField {
	fi := FormField{}
	fi.prompt = "> "
	fi.input = NewDefaultInput(inputCharLimit)
	fi.input.SetStyles(formStyles.textInput)
	return fi
}

func (fi FormField) Title(s string) FormField {
	fi.title = s
	return fi
}

func (fi FormField) Description(s string) FormField {
	fi.description = s
	return fi
}

func (fi FormField) Placeholder(s string) FormField {
	fi.input.Placeholder = s
	return fi
}

// DescriptionDyn accepts a function that should be used to create
// a dynamic description. The input of all fields prior to this
// one, will be passed as args to this func.
//
// For instance if this is the 3rd field in a form, then arg[1]
// references the 2nd field's input.
//
// 🟡 This func takes precedence over Description() when the
// field is not the first.
func (fi FormField) DescriptionDyn(d func(args ...string) string) FormField {
	fi.descFunc = d
	return fi
}

func (fi FormField) InputType(t FormInputType) FormField {
	fi.inputType = t
	if t == CurrencyInput {
		fi.prompt = "$ "
	}
	return fi
}

func (fi FormField) Value(s string) FormField {
	fi.input.SetValue(s)
	return fi
}

// Validator overwrites the default InputType validator with
// the one specified. Not all InputType's have a validator.
//
// 🔵 You can still use the default validator alongside
// your custom validator, since it is passed as an arg.
func (fi FormField) Validator(v FormValidatorFunc) FormField {
	fi.validator = v
	return fi
}

// Optional sets the field to optional. By default, all input
// fields are required.
func (fi FormField) Optional() FormField {
	fi.isOptional = true
	return fi
}

func (fi FormField) Suggestions(sugs ...string) FormField {
	fi.input.ShowSuggestions = true
	fi.input.SetSuggestions(sugs)
	return fi
}

func (fi FormField) Prompt(p string) FormField {
	fi.prompt = p
	return fi
}

// LinkTo requires the name of another field that is required
// before this field. When a field is linked, its input will
// be passed as an argument to this fields custom validator
// function.
func (fi FormField) LinkTo(name string) FormField {
	fi.linkedInput = name
	return fi
}

func (fi FormField) hasLinkedInput() bool {
	return len(fi.linkedInput) > 0
}

func (fi FormField) view() string {
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

func (fi FormField) validate(args ...string) error {
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

	case CurrencyInput:
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
		return fi.validator(inputStr, defaultValidator, args...)
	}

	return defaultValidator()
}

type Form struct {
	fields  []FormField
	values  []string
	name    string
	buttons struct {
		save   Button
		cancel Button
	}
	linkMap map[string]int
	focus   formFocusState
	tabPos  int
	isInit  bool
	err     error
}

func NewForm(name string, inputs ...FormField) Form {
	f := Form{
		name:    name,
		isInit:  true,
		fields:  inputs,
		linkMap: map[string]int{},
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

	inputMap := make(map[string]int, len(f.fields))

	for i, field := range f.fields {
		inputMap[field.title] = i
		if field.linkedInput != "" {
			inputIdx, exists := inputMap[field.linkedInput]
			if !exists {
				panic(
					fmt.Errorf(
						"fatal form error: %s field needs to be before %s",
						field.linkedInput,
						field.title,
					),
				)
			}
			f.linkMap[field.title] = inputIdx
		}
		// Remove focus from all inputs
		if i > 0 {
			f.fields[i].input.Blur()
		}
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
			f.fields[f.tabPos].input.Blur()
			f.tabPos = len(f.fields) - 1
			f.focus = formFocusSave
			f.buttons.save.Focus()
			return f, nil

		case "enter":
			if !f.validateField(f.fields[f.tabPos]) {
				return f, nil
			}

			if f.tabPos == len(f.fields)-1 {
				if f.focus > formFocusInput {
					if f.focus == formFocusSave {
						return f, f.SendSaveMsg()
					}
					if f.focus == formFocusCancel {
						return f, f.SendCancelMsg()
					}
					return f, nil
				}
				f.fields[f.tabPos].input.Blur()
				f.focus = formFocusSave
				f.buttons.save.Focus()
				return f, nil
			}

			f.fields[f.tabPos].input.Blur()
			f.tabPos++
			f.fields[f.tabPos].input.Focus()
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
			f.fields[f.tabPos].input, cmd = f.fields[f.tabPos].input.Update(msg)
			cmds = append(cmds, cmd)

			if !f.validateField(f.fields[f.tabPos]) {
				return f, nil
			}

			if f.tabPos+1 == len(f.fields) {
				f.buttons.save.Focus()
				f.focus = formFocusSave
				f.fields[f.tabPos].input.Blur()
				return f, cmd
			}

			f.fields[f.tabPos].input.Blur()
			f.tabPos++
			cmds = append(cmds, f.fields[f.tabPos].input.Focus())
			return f, tea.Batch(cmds...)

		case "backspace":
			if f.focus > formFocusInput {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				f.focus = formFocusInput
				return f, f.fields[f.tabPos].input.Focus()
			}

		case "shift+tab":
			if f.focus > formFocusInput {
				f.buttons.save.Blur()
				f.buttons.cancel.Blur()
				f.focus = formFocusInput
				return f, f.fields[f.tabPos].input.Focus()
			}

			if f.tabPos == 0 {
				return f, nil
			}

			if !f.validateField(f.fields[f.tabPos]) {
				return f, nil
			}

			f.fields[f.tabPos].input.Blur()
			f.tabPos--
			f.isInit = true
			return f, f.fields[f.tabPos].input.Focus()
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
	buttonView := Style.
		MarginRight(1).
		Render(f.buttons.save.View() + " " + f.buttons.cancel.View())

	formStatusView := f.formStatusView("Ok", true)
	if f.err != nil {
		formStatusView = f.formStatusView(f.err.Error(), false)
	}

	formView := formStyles.border.Render(JoinHorizontal(
		lipgloss.Left,
		f.inputView(),
		formStatusView,
	))

	return JoinVertical(
		lipgloss.Right,
		JoinVertical(lipgloss.Center, formStyles.header.Render(f.name+" Form"), formView),
		buttonView,
	)
}

func (f *Form) Reset() tea.Cmd {
	for i := range f.fields {
		f.fields[i].input.SetValue("")
	}
	f.focus = formFocusInput
	f.buttons.cancel.Blur()
	f.buttons.save.Blur()
	f.fields[0].input.Focus()
	f.tabPos = 0
	return textinput.Blink
}

func (f Form) SendSaveMsg() tea.Cmd {
	data := make([]string, len(f.fields))
	for i, field := range f.fields {
		data[i] = field.input.Value()
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

func (f *Form) validateField(entry FormField) bool {
	if entry.hasLinkedInput() {
		f.err = entry.validate(f.fields[f.linkMap[entry.title]].input.Value())
	} else {
		f.err = entry.validate()
	}
	return f.err == nil
}

func (f Form) inputView() string {
	var sb strings.Builder
	sb.Grow((inputWidth + inputMargin) * len(f.fields))

	for i, field := range f.fields {
		if i != 0 {
			sb.WriteByte('\n')
		}

		borderColor := DarkBorderColor
		if i == f.tabPos && f.focus == formFocusInput {
			borderColor = formActiveColor
		}

		b := utils.GetBorderStyle(utils.Rounded)

		sb.WriteString(
			formStyles.inputBorder.BorderForeground(borderColor).Render(field.view()),
		)
		sb.WriteByte('\n')

		// Use corner border for last input field
		activeBorderChar := string(b.RightT)
		if i == len(f.fields)-1 {
			activeBorderChar = string(b.BottomRight)
		}

		if f.tabPos == i && f.focus == formFocusInput {
			sb.WriteString(
				Style.
					Width(inputWidth + inputMargin).
					Foreground(DarkBorderColor).
					Render(utils.GenDashedBorder(inputWidth+inputMargin-1) + string(activeBorderChar)),
			)
		} else {
			sb.WriteString(formStyles.inputBorder.Render(""))
		}
	}

	return sb.String()
}

func (f Form) formStatusView(status string, isGood bool) string {
	const formSize = 35

	statusWrapper := Style.
		Width(formSize).
		MarginTop(0).
		Foreground(White).
		MarginLeft(1).
		Padding(0, 1, 0)

	field := f.fields[f.tabPos]
	title := field.title
	if field.isOptional {
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
		description = "No changes will be made to " + f.name + "."
		statusText = statusTextStyle.Render(statusChar + "Form will be discarded")
	default:
		statusTitle = title
		field := f.fields[f.tabPos]
		if field.descFunc != nil && f.tabPos > 0 {
			fieldInputs := make([]string, f.tabPos)
			for i := range f.tabPos {
				fieldInputs[i] = f.fields[i].input.Value()
			}
			description = field.descFunc(fieldInputs...)
		} else if len(field.description) > 0 {
			description = f.fields[f.tabPos].description
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
	for i := range f.fields {
		// Focused input
		if i == f.tabPos {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				if !f.isValidKey(f.fields[i].inputType, msg) {
					return cmd
				}
			}

			f.fields[i].input, cmd = f.fields[i].input.Update(msg)
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

	case FloatInput, CurrencyInput:
		return key == '.' || utils.IsNumber(key)
	}

	return false
}
