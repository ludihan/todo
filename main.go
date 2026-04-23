package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/adrg/xdg"
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
	UpFirst         key.Binding
	DownLast        key.Binding
	NextFile        key.Binding
	PreviousFile    key.Binding
	NewNoteBelow    key.Binding
	NewNoteAbove    key.Binding
	DeleteNote      key.Binding
	ChangeNote      key.Binding
	ShiftNoteUp     key.Binding
	ShiftNoteDown   key.Binding
	Access          key.Binding
	GoBackHistory   key.Binding
	GoFowardHistory key.Binding
	ConfirmEdit     key.Binding
	CancelEdit      key.Binding
	FileView        key.Binding
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
	UpFirst: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to first note"),
	),
	DownLast: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to last note"),
	),
	NextFile: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "go to next file"),
	),
	PreviousFile: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "go to previous file"),
	),
	NewNoteBelow: key.NewBinding(
		key.WithKeys("n", "o"),
		key.WithHelp("n", "new note"),
	),
	NewNoteAbove: key.NewBinding(
		key.WithKeys("N", "O"),
		key.WithHelp("N", "new note"),
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
	FileView: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "show file view"),
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
		k.NewNoteBelow,
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
	keys       keyMap
	textInput  textinput.Model
	help       help.Model
	showHelp   bool
	notes      []string
	cursor     int
	history    []string
	file       string
	noteCopy   string
	root       *os.Root
	err        error
	shouldQuit bool
}

var defaultFile = "@"

func createRootAndDefaultFile() (*os.Root, error) {
	rootPath := path.Join(xdg.DataHome, "todo")
	_, err := os.Stat(rootPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(rootPath, 0777)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	_, err = root.Stat(defaultFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := root.Create(defaultFile)
			if err != nil {
				panic(err)
			}
			defer f.Close()
		}
	}
	return root, nil
}

func initialModel() model {
	root, err := createRootAndDefaultFile()
	if err != nil {
		return model{
			err: err,
		}
	}
	ti := textinput.New()
	ti.Prompt = ""
	styles := ti.Styles()
	styles.Cursor.Blink = false
	styles.Cursor.Color = lipgloss.BrightWhite
	ti.SetVirtualCursor(true)
	ti.SetStyles(styles)

	data, err := root.ReadFile(defaultFile)
	if !utf8.Valid(data) {
		err = errors.New("invalid utf8")
	}
	noteData := string(data)
	notes := []string{}
	for line := range strings.Lines(noteData) {
		notes = append(notes, strings.TrimSpace(line))
	}

	return model{
		textInput: ti,
		keys:      keys,
		help:      help.New(),
		notes:     notes,
		file:      defaultFile,
		root:      root,
		err:       err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) saveFile() {
	err := m.root.WriteFile(m.file, []byte(strings.Join(m.notes, "\n")), 0777)
	if err != nil {
		m.err = err
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if !m.textInput.Focused() {
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}

			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.notes)-1 {
					m.cursor++
				}

			case key.Matches(msg, m.keys.UpFirst):
				m.cursor = 0

			case key.Matches(msg, m.keys.DownLast):
				m.cursor = len(m.notes) - 1

			case key.Matches(msg, m.keys.NewNoteBelow):
				m.noteCopy = ""
				m.notes = append(m.notes, "")
				if len(m.notes) > 1 {
					m.cursor++
					m.notes[len(m.notes)-1], m.notes[m.cursor] = m.notes[m.cursor], m.notes[len(m.notes)-1]
				}
				m.textInput.Focus()
				return m, nil

			case key.Matches(msg, m.keys.NewNoteAbove):
				m.noteCopy = ""
				m.notes = slices.Insert(m.notes, m.cursor, "")
				m.textInput.Focus()
				return m, nil

			case key.Matches(msg, m.keys.ChangeNote):
				m.textInput.Focus()
				m.noteCopy = m.notes[m.cursor]
				m.textInput.SetValue(m.noteCopy)
				return m, nil

			case key.Matches(msg, m.keys.DeleteNote):
				if len(m.notes) > 0 {
					m.notes = slices.Delete(m.notes, m.cursor, m.cursor+1)
					m.cursor = max(min(m.cursor, len(m.notes)-1), 0)
					m.saveFile()
					return m, nil
				}

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
				m.notes[m.cursor] = strings.TrimSpace(m.textInput.Value())
				m.textInput.Blur()
				m.textInput.Reset()
				m.saveFile()
				return m, nil

			case key.Matches(msg, m.keys.CancelEdit):
				m.textInput.Blur()
				m.textInput.Reset()
				if m.noteCopy != "" {
					m.notes[m.cursor] = m.noteCopy
				} else {
					m.notes = append(m.notes[:m.cursor], m.notes[m.cursor+1:]...)
					m.cursor = max(m.cursor-1, 0)
				}

			}
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView(m.err.Error() + "\n")
	}

	var v tea.View
	var s strings.Builder

	fmt.Fprintf(&s, " ~  %s\n", path.Base(m.file))

	for i, note := range m.notes {
		noteSelection := "   "
		if m.cursor == i && !m.textInput.Focused() {
			noteSelection = "███"
		}

		if m.textInput.Focused() && m.cursor == i {
			fmt.Fprintf(&s, "%s %d: %s\n", noteSelection, i+1, m.textInput.View())
		} else {
			fmt.Fprintf(&s, "%s %d: %s\n", noteSelection, i+1, note)
		}
	}

	if m.showHelp && !m.textInput.Focused() {
		helpView := m.help.View(m.keys)
		fmt.Fprintf(&s, "%s", helpView)
	}

	v.SetContent(s.String())

	return v
}
