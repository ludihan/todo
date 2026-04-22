package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	_ "github.com/adrg/xdg"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	NewNote         key.Binding
	DeleteNote      key.Binding
	ChangeNote      key.Binding
	ShiftNoteUp     key.Binding
	ShiftNoteDown   key.Binding
	Access          key.Binding
	GoBackHistory   key.Binding
	GoFowardHistory key.Binding
	ConfirmEdit     key.Binding
	CancelEdit      key.Binding
	Help            key.Binding
	Quit            key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "move down"),
	),
	NewNote: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new note"),
	),
	ChangeNote: key.NewBinding(
		key.WithKeys("c", "e"),
		key.WithHelp("c", "change note"),
	),
	DeleteNote: key.NewBinding(
		key.WithKeys("d", "x"),
		key.WithHelp("d", "delete note"),
	),
	ShiftNoteUp: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("shift+k", "shift note up"),
	),
	ShiftNoteDown: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("shift+j", "shift note down"),
	),
	Access: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "shift note down"),
	),
	GoBackHistory: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "go back one file"),
	),
	GoFowardHistory: key.NewBinding(
		key.WithKeys("shift+backspace"),
		key.WithHelp("shift+backspace", "go foward one file"),
	),
	ConfirmEdit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm edit"),
	),
	CancelEdit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "cancel edit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.NewNote,
		k.ChangeNote,
		k.DeleteNote,
		k.ShiftNoteUp,
		k.ShiftNoteDown,
		k.Help,
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

type model struct {
	keys     keyMap
	help     help.Model
	showHelp bool
	notes    []string
	cursor   int
	history  []string
	file     string
	editing  bool
	noteCopy string
}

func initialModel() model {
	return model{
		keys:  keys,
		help:  help.New(),
		notes: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
		file:  "default",
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		if !m.editing {
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.notes)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.NewNote):
				m.notes = append(m.notes, "")
				m.cursor = len(m.notes) - 1
				m.editing = true

			case key.Matches(msg, m.keys.ChangeNote):
				m.editing = true
				m.noteCopy = m.notes[m.cursor]

			case key.Matches(msg, m.keys.ShiftNoteUp):
				if m.cursor > 0 {
					m.notes[m.cursor-1], m.notes[m.cursor] = m.notes[m.cursor], m.notes[m.cursor-1]
					m.cursor--
				}
			case key.Matches(msg, m.keys.ShiftNoteDown):
				if m.cursor < len(m.notes)-1 {
					m.notes[m.cursor+1], m.notes[m.cursor] = m.notes[m.cursor], m.notes[m.cursor+1]
					m.cursor++
				}
			case key.Matches(msg, m.keys.ShiftNoteDown):
				m.showHelp = !m.showHelp
			case key.Matches(msg, m.keys.Help):
				m.showHelp = !m.showHelp
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
		} else {
			switch {
			case key.Matches(msg, m.keys.ConfirmEdit):
				m.editing = false
			case key.Matches(msg, m.keys.CancelEdit):
				m.editing = false
				if m.noteCopy != "" {
					m.notes[m.cursor] = m.noteCopy
				} else {
					m.notes = m.notes[:len(m.notes)-1]
					m.cursor = len(m.notes) - 1
				}
			default:
				if msg.Code == tea.KeyBackspace {
					n := &m.notes[m.cursor]
					*n = (*n)[:len(*n)-1]
				} else {
					m.notes[m.cursor] += string(msg.Text)
				}
			}
		}

	}

	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder
	fmt.Fprintf(&s, " ~  %s\n", m.file)

	for i, choice := range m.notes {

		cursor := "   "
		if m.cursor == i && !m.editing {
			cursor = "███"
		}

		if m.editing && m.cursor == i {
			fmt.Fprintf(&s, "%s %d: %s█\n", cursor, i+1, choice)
		} else {
			fmt.Fprintf(&s, "%s %d: %s\n", cursor, i+1, choice)
		}
	}

	if m.showHelp && !m.editing {
		helpView := m.help.View(m.keys)
		fmt.Fprintf(&s, "%s", helpView)
	}

	return tea.NewView(s.String())
}
