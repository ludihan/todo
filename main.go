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
	Up            key.Binding
	Down          key.Binding
	NewNote       key.Binding
	DeleteNote    key.Binding
	ChangeNote    key.Binding
	ShiftNoteUp   key.Binding
	ShiftNoteDown key.Binding
	Help          key.Binding
	Quit          key.Binding
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
	choices  []string
	cursor   int
}

func initialModel() model {
	return model{
		keys:    keys,
		help:    help.New(),
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.ShiftNoteUp):
			if m.cursor > 0 {
				m.choices[m.cursor-1], m.choices[m.cursor] = m.choices[m.cursor], m.choices[m.cursor-1]
				m.cursor--
			}
		case key.Matches(msg, m.keys.ShiftNoteDown):
			if m.cursor < len(m.choices)-1 {
				m.choices[m.cursor+1], m.choices[m.cursor] = m.choices[m.cursor], m.choices[m.cursor+1]
				m.cursor++
			}
		case key.Matches(msg, m.keys.ShiftNoteDown):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

	}

	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder

	for i, choice := range m.choices {

		cursor := "   "
		if m.cursor == i {
			cursor = "███"
		}

		fmt.Fprintf(&s, "%s %d: %s\n", cursor, i+1, choice)
	}

	if m.showHelp {
		helpView := m.help.View(m.keys)
		fmt.Fprintf(&s, "%s", helpView)
	}

	return tea.NewView(s.String())
}
