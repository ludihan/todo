package main

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
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
	Confirm         key.Binding
	Cancel          key.Binding
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
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm edit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "cancel edit"),
	),
	FileView: key.NewBinding(
		key.WithKeys("tab", "-"),
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
	files      map[string]struct{}
	fileView   bool
	noteCopy   string
	root       *os.Root
	err        error
	shouldQuit bool
}

const defaultFile = "general"

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

func configureTextInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	styles := ti.Styles()
	styles.Cursor.Blink = false
	styles.Cursor.Color = lipgloss.BrightWhite
	ti.SetVirtualCursor(true)
	ti.SetStyles(styles)
	return ti
}

func initialModel() model {
	root, err := createRootAndDefaultFile()
	if err != nil {
		return model{
			err: err,
		}
	}

	ti := configureTextInput()
	notes, err := readNotes(root, defaultFile)
	files, err := readFiles(root)

	return model{
		textInput: ti,
		keys:      keys,
		help:      help.New(),
		notes:     notes,
		file:      defaultFile,
		files:     files,
		root:      root,
		err:       err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) saveFile() {
	err := m.root.WriteFile(m.file, []byte(strings.Join(m.notes, "\n")+"\n"), 0777)
	if err != nil {
		m.err = err
	}
}

func readNotes(root *os.Root, file string) ([]string, error) {
	data, err := root.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if !utf8.Valid(data) {
		return nil, errors.New("invalid utf8")
	}

	noteData := string(data)
	notes := []string{}
	for line := range strings.Lines(noteData) {
		notes = append(notes, strings.TrimSpace(line))
	}
	return notes, nil
}

func readFiles(root *os.Root) (map[string]struct{}, error) {
	fsRoot := root.FS()
	dirEntries, err := fs.ReadDir(fsRoot, ".")
	if err != nil {
		return nil, err
	}

	dirs := map[string]struct{}{}
	for _, dir := range dirEntries {
		dirs[dir.Name()] = struct{}{}
	}

	return dirs, nil
}

func mapFileViewToFs(root *os.Root, files []string) error {
	filesMap := map[string]struct{}{}
	for _, f := range files {
		if f != "" {
			filesMap[f] = struct{}{}
		}
	}
	existingFiles, err := readFiles(root)
	if err != nil {
		return err
	}

	for v := range filesMap {
		_, ok := existingFiles[v]
		if !ok {
			f, err := root.Create(v)
			f.Close()
			if err != nil {
				return err
			}
		}
	}

	for k := range existingFiles {
		_, ok := filesMap[k]
		if !ok {
			err := root.Remove(k)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		m.root.Close()
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

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
				m.cursor = min(max(0, m.cursor+1), len(m.notes))
				m.notes = slices.Insert(m.notes, m.cursor, "")
				m.textInput.Focus()

			case key.Matches(msg, m.keys.NewNoteAbove):
				m.noteCopy = ""
				m.notes = slices.Insert(m.notes, m.cursor, "")
				m.textInput.Focus()

			case key.Matches(msg, m.keys.ChangeNote):
				m.textInput.Focus()
				m.noteCopy = m.notes[m.cursor]
				m.textInput.SetValue(m.noteCopy)

			case key.Matches(msg, m.keys.DeleteNote):
				if len(m.notes) > 0 {
					if !(m.fileView && m.cursor == 0) {
						m.notes = slices.Delete(m.notes, m.cursor, m.cursor+1)
						m.cursor = max(min(m.cursor, len(m.notes)-1), 0)
					}
					if !m.fileView {
						m.saveFile()
					}
				}

			case key.Matches(msg, m.keys.ShiftNoteUp):
				if m.cursor > 0 && !m.fileView {
					m.notes[m.cursor-1], m.notes[m.cursor] = m.notes[m.cursor], m.notes[m.cursor-1]
					m.cursor--
				}

			case key.Matches(msg, m.keys.ShiftNoteDown):
				if m.cursor < len(m.notes)-1 && !m.fileView {
					m.notes[m.cursor+1], m.notes[m.cursor] = m.notes[m.cursor], m.notes[m.cursor+1]
					m.cursor++
				}

			case key.Matches(msg, m.keys.FileView):
				m.fileView = !m.fileView
				if m.fileView {
					files, err := readFiles(m.root)
					m.files = files
					m.err = err
					notes := []string{}
					for k := range m.files {
						if k != defaultFile {
							notes = append(notes, k)
						}
					}
					slices.Sort(notes)
					notes = append([]string{defaultFile}, notes...)
					m.notes = notes
					for i, v := range notes {
						if v == m.file {
							m.cursor = i
						}
					}
				} else {
					m.cursor = 0
					notes, err := readNotes(m.root, m.file)
					m.err = err
					m.notes = notes
				}

			case key.Matches(msg, m.keys.Help):
				m.showHelp = !m.showHelp

			case key.Matches(msg, m.keys.Quit):
				m.root.Close()
				return m, tea.Quit

			case key.Matches(msg, m.keys.Confirm):
				if m.fileView {
					file := m.notes[m.cursor]
					m.file = file
					notes, err := readNotes(m.root, m.notes[m.cursor])
					m.fileView = !m.fileView
					m.err = err
					m.notes = notes
					m.cursor = 0
				}
			}
		} else {
			switch {
			case key.Matches(msg, m.keys.Confirm):
				m.notes[m.cursor] = strings.TrimSpace(m.textInput.Value())
				m.textInput.Blur()
				m.textInput.Reset()
				if !m.fileView {
					m.saveFile()
				}

			case key.Matches(msg, m.keys.Cancel):
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

	if m.fileView {
		mapFileViewToFs(m.root, m.notes)
	}
	return m, cmd
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView(m.err.Error() + "\n")
	}

	var v tea.View
	var s strings.Builder

	if !m.fileView {
		fmt.Fprintf(&s, "~~~ %s ~~~\n", path.Base(m.file))
	} else {
		fmt.Fprintf(&s, "@@@@@@ \n")
	}

	for i, note := range m.notes {
		prefix := math.Log10(float64(len(m.notes)))
		noteSelection := "   "
		if m.cursor == i && !m.textInput.Focused() {
			noteSelection = "███"
		}

		if m.textInput.Focused() && m.cursor == i {
			fmt.Fprintf(&s, "%s %*d: %s\n", noteSelection, int(prefix)+1, i+1, m.textInput.View())
		} else {
			fmt.Fprintf(&s, "%s %*d: %s\n", noteSelection, int(prefix)+1, i+1, note)
		}
	}

	if m.showHelp && !m.textInput.Focused() {
		helpView := m.help.View(m.keys)
		fmt.Fprintf(&s, "%s", helpView)
	}

	v.SetContent(s.String())

	return v
}
