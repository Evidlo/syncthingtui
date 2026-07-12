package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Browse-then-edit form: ↑↓/jk move between fields, enter starts editing the
// focused field (or toggles a checkbox), enter commits / esc cancels the edit.
// Readonly fields render dim and cannot be edited.

type fieldKind int

const (
	ftText fieldKind = iota
	ftPass
	ftRO
	ftSelect
	ftBool
	ftArea
)

type FormField struct {
	kind        fieldKind
	label, desc string
	key         string // config key for Save (empty = not saved)
	value       string
	numeric     bool
	boolVal     bool
	options     []string // display strings
	vals        []string // config values parallel to options (Sel only)
	optIdx      int
	input       textinput.Model
	area        textarea.Model
}

// K sets the config key this field saves to.
func (f FormField) K(key string) FormField { f.key = key; return f }

// Num marks a text field as numeric (saved as int).
func (f FormField) Num() FormField { f.numeric = true; return f }

// Vals sets the config values corresponding to a select's display options and
// re-resolves the initial selection against them (value holds the raw init).
func (f FormField) Vals(vals ...string) FormField {
	f.vals = vals
	for i, v := range vals {
		if v == f.value {
			f.optIdx = i
		}
	}
	return f
}

func Text(label, desc, value string) FormField {
	return FormField{kind: ftText, label: label, desc: desc, value: value}
}
func Pass(label, desc, value string) FormField {
	return FormField{kind: ftPass, label: label, desc: desc, value: value}
}
func RO(label, desc, value string) FormField {
	return FormField{kind: ftRO, label: label, desc: desc, value: value}
}

// Sel builds a select; value may match an option (display string) or, after
// .Vals(...), a config value.
func Sel(label, desc string, options []string, value string) FormField {
	idx := 0
	for i, o := range options {
		if o == value {
			idx = i
		}
	}
	return FormField{kind: ftSelect, label: label, desc: desc, options: options, optIdx: idx, value: value}
}
func Bool(label, desc string, value bool) FormField {
	return FormField{kind: ftBool, label: label, desc: desc, boolVal: value}
}
func Area(label, desc, value string) FormField {
	return FormField{kind: ftArea, label: label, desc: desc, value: value}
}

func (f FormField) display() string {
	switch f.kind {
	case ftPass:
		return strings.Repeat("•", max(4, len(f.value)))
	case ftSelect:
		return f.options[f.optIdx] + " ▾"
	case ftBool:
		if f.boolVal {
			return "[x]"
		}
		return "[ ]"
	case ftArea:
		lines := strings.Split(f.value, "\n")
		return fmt.Sprintf("%s … (%d lines)", lines[0], len(lines))
	}
	return f.value
}

// Form is one section's field list plus a trailing button row.
type Form struct {
	fields  []FormField
	buttons []string
	cursor  int // 0..len(fields); len(fields) = button row
	btn     int
	editing bool
	width   int
	flash   string
}

func NewForm(buttons []string, fields ...FormField) Form {
	return Form{fields: fields, buttons: buttons}
}

func (f Form) Editing() bool { return f.editing }

// OnButtons reports whether the cursor sits on the button row (where ←→
// select buttons rather than switching sections).
func (f Form) OnButtons() bool { return f.cursor == len(f.fields) }

// buttonMsg reports an activated form button ("Save", "Close", ...).
type buttonMsg struct{ label string }

// values collects the savable fields into config-key → value.
func (f Form) values() map[string]any {
	m := map[string]any{}
	for _, fld := range f.fields {
		if fld.key == "" || fld.kind == ftRO {
			continue
		}
		switch fld.kind {
		case ftBool:
			m[fld.key] = fld.boolVal
		case ftSelect:
			if len(fld.vals) == len(fld.options) {
				m[fld.key] = fld.vals[fld.optIdx]
			} else {
				m[fld.key] = fld.options[fld.optIdx]
			}
		default: // text, pass, area
			if fld.numeric {
				n, _ := strconv.Atoi(strings.TrimSpace(fld.value))
				m[fld.key] = n
			} else {
				m[fld.key] = fld.value
			}
		}
	}
	return m
}

func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = ws.Width
		return f, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		if f.editing { // cursor blink etc.
			return f.updateEdit(msg)
		}
		return f, nil
	}
	if f.editing {
		return f.updateEditKey(key)
	}
	switch key.String() {
	case "up", "k":
		f.cursor = clamp(f.cursor-1, 0, len(f.fields))
	case "down", "j":
		f.cursor = clamp(f.cursor+1, 0, len(f.fields))
	case "left", "h":
		if f.cursor == len(f.fields) {
			f.btn = clamp(f.btn-1, 0, len(f.buttons)-1)
		}
	case "right", "l":
		if f.cursor == len(f.fields) {
			f.btn = clamp(f.btn+1, 0, len(f.buttons)-1)
		}
	case "enter":
		if f.cursor == len(f.fields) {
			label := f.buttons[f.btn]
			return f, func() tea.Msg { return buttonMsg{label} }
		}
		fld := &f.fields[f.cursor]
		switch fld.kind {
		case ftRO:
			// not editable
		case ftBool:
			fld.boolVal = !fld.boolVal
		case ftArea:
			fld.area = textarea.New()
			fld.area.SetValue(fld.value)
			fld.area.SetWidth(min(60, f.width-20))
			fld.area.SetHeight(6)
			fld.area.Focus()
			f.editing = true
		default:
			fld.input = textinput.New()
			fld.input.SetValue(fld.value)
			fld.input.Prompt = ""
			if fld.kind == ftPass {
				fld.input.EchoMode = textinput.EchoPassword
			}
			fld.input.Width = min(48, f.width-24)
			fld.input.Focus()
			f.editing = true
		}
	}
	return f, nil
}

func (f Form) updateEditKey(key tea.KeyMsg) (Form, tea.Cmd) {
	fld := &f.fields[f.cursor]
	switch key.String() {
	case "esc":
		f.editing = false // discard
		return f, nil
	case "enter":
		if fld.kind != ftArea { // textarea: enter inserts newline, esc commits
			switch fld.kind {
			case ftSelect:
			default:
				fld.value = fld.input.Value()
			}
			f.editing = false
			return f, nil
		}
	case "ctrl+s":
		if fld.kind == ftArea {
			fld.value = fld.area.Value()
			f.editing = false
			return f, nil
		}
	}
	if fld.kind == ftSelect {
		switch key.String() {
		case "up", "k":
			fld.optIdx = clamp(fld.optIdx-1, 0, len(fld.options)-1)
		case "down", "j":
			fld.optIdx = clamp(fld.optIdx+1, 0, len(fld.options)-1)
		}
		return f, nil
	}
	return f.updateEdit(key)
}

func (f Form) updateEdit(msg tea.Msg) (Form, tea.Cmd) {
	fld := &f.fields[f.cursor]
	var cmd tea.Cmd
	if fld.kind == ftArea {
		fld.area, cmd = fld.area.Update(msg)
	} else {
		fld.input, cmd = fld.input.Update(msg)
	}
	return f, cmd
}

const labelW = 32 // fits the longest label ("Sync Protocol Listen Addresses")

func (f Form) View() string {
	var b strings.Builder
	for i, fld := range f.fields {
		focused := i == f.cursor
		cursor := "  "
		if focused && !f.editing {
			cursor = keyStyle.Render("▸ ")
		}
		label := fmt.Sprintf("%-*s", labelW, fld.label)
		switch {
		case fld.kind == ftRO:
			b.WriteString(cursor + dimStyle.Render(label+fld.value+"  (readonly)") + "\n")
		case focused && f.editing && fld.kind == ftSelect:
			b.WriteString("  " + keyStyle.Render(label) + "\n")
			for j, o := range fld.options {
				marker, style := "    ", lipgloss.NewStyle()
				if j == fld.optIdx {
					marker, style = "  "+keyStyle.Render("▸ "), keyStyle
				}
				b.WriteString(marker + style.Render(o) + "\n")
			}
		case focused && f.editing && fld.kind == ftArea:
			b.WriteString("  " + keyStyle.Render(fld.label) + "\n" +
				lipgloss.NewStyle().PaddingLeft(2).Render(fld.area.View()) + "\n")
		case focused && f.editing:
			b.WriteString("  " + keyStyle.Render(label) + fld.input.View() + "\n")
		default:
			labelSt, valueSt := dimStyle, lipgloss.NewStyle()
			if focused {
				labelSt, valueSt = keyStyle, boldSt
			}
			b.WriteString(cursor + labelSt.Render(label) + valueSt.Render(fld.display()) + "\n")
		}
		if focused && fld.desc != "" && !(f.editing && fld.kind == ftArea) {
			b.WriteString("  " + strings.Repeat(" ", labelW) + dimStyle.Render(fld.desc) + "\n")
		}
		b.WriteString("\n")
	}
	sel := -1
	if f.cursor == len(f.fields) {
		sel = f.btn
	}
	cursor := "  "
	if f.cursor == len(f.fields) {
		cursor = keyStyle.Render("▸ ")
	}
	b.WriteString(cursor + buttonRow(f.buttons, sel) + "\n")
	return b.String()
}
