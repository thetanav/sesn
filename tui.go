package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sesn/internals"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type inputMode int

const (
	inputNone inputMode = iota
	inputFuzzy
	inputCreate
	inputRename
)

type model struct {
	sessionList     list.Model
	windowList      list.Model
	sessions        []internals.Session
	windows         []internals.Window
	inputMode       inputMode
	textInput       textinput.Model
	selectedSession string
	width, height   int
}

func initialModel() model {
	sessionItems := []list.Item{}
	windowItems := []list.Item{}

	sessionList := list.New(sessionItems, list.NewDefaultDelegate(), 0, 0)
	sessionList.SetShowHelp(false)
	sessionList.SetFilteringEnabled(true)

	windowList := list.New(windowItems, list.NewDefaultDelegate(), 0, 0)
	windowList.SetShowHelp(false)

	ti := textinput.New()
	ti.Placeholder = "Name"
	ti.CharLimit = 20
	ti.Width = 20

	return model{
		sessionList: sessionList,
		windowList:  windowList,
		sessions:    []internals.Session{},
		windows:     []internals.Window{},
		inputMode:   inputNone,
		textInput:   ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadSessions, m.textInput.Focus())
}

func loadSessions() tea.Msg {
	sessions, err := internals.ListSessions()
	if err != nil {
		return errMsg{err}
	}
	return sessionsMsg{sessions}
}

func loadWindows(sessionName string) tea.Cmd {
	return func() tea.Msg {
		windows, err := internals.ListWindows(sessionName)
		if err != nil {
			return errMsg{err}
		}
		return windowsMsg{windows}
	}
}

type sessionsMsg struct{ sessions []internals.Session }
type windowsMsg struct{ windows []internals.Window }
type errMsg struct{ err error }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve one line for header
		bodyHeight := msg.Height - 1
		if bodyHeight < 1 {
			bodyHeight = 1
		}
		// Split width into two columns with a 1-column divider
		leftW := msg.Width / 2
		if leftW < 10 {
			leftW = 10
		}
		rightW := msg.Width - leftW - 1
		if rightW < 10 {
			rightW = 10
		}
		m.sessionList.SetSize(leftW, bodyHeight)
		m.windowList.SetSize(rightW, bodyHeight)
	case sessionsMsg:
		m.sessions = msg.sessions
		items := make([]list.Item, len(m.sessions))
		for i, s := range m.sessions {
			items[i] = item{title: s.Name, desc: ""}
		}
		cmd = m.sessionList.SetItems(items)
		cmds = append(cmds, cmd)
		if len(m.sessions) > 0 {
			m.selectedSession = m.sessions[0].Name
			cmds = append(cmds, loadWindows(m.selectedSession))
		}
	case windowsMsg:
		m.windows = msg.windows
		items := make([]list.Item, len(m.windows))
		for i, w := range m.windows {
			items[i] = item{title: fmt.Sprintf("%d: %s", w.Index, w.Name), desc: ""}
		}
		cmd = m.windowList.SetItems(items)
		cmds = append(cmds, cmd)
	case errMsg:
		// Handle error, maybe show in status
	case tea.KeyMsg:
		switch m.inputMode {
		case inputNone:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "c":
				m.inputMode = inputCreate
				m.textInput.Reset()
				return m, m.textInput.Focus()
			case "d":
				if m.selectedSession != "" {
					internals.DeleteSession(m.selectedSession)
					return m, loadSessions
				}
			case "r":
				if m.selectedSession != "" {
					m.inputMode = inputRename
					m.textInput.Reset()
					return m, m.textInput.Focus()
				}
			case "k":
				if m.selectedSession != "" {
					internals.DeleteSession(m.selectedSession)
					return m, loadSessions
				}
			case "/":
				m.inputMode = inputFuzzy
				m.textInput.Reset()
				return m, m.textInput.Focus()
			case "enter":
				if m.selectedSession != "" {
					internals.AttachSession(m.selectedSession)
					return m, tea.Quit
				}
			case "esc":
				// Nothing to do
			}
		case inputFuzzy:
			switch msg.String() {
			case "esc":
				m.inputMode = inputNone
				m.textInput.Reset()
				items := make([]list.Item, len(m.sessions))
				for i, s := range m.sessions {
					items[i] = item{title: s.Name, desc: ""}
				}
				m.sessionList.SetItems(items)
				return m, nil
			}
		case inputCreate, inputRename:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.textInput.Value())
				if name != "" {
					if m.inputMode == inputCreate {
						internals.CreateSession(name)
					} else {
						internals.RenameSession(m.selectedSession, name)
					}
					m.inputMode = inputNone
					return m, loadSessions
				}
			case "esc":
				m.inputMode = inputNone
				return m, nil
			}
		}
	}

	if m.inputMode == inputNone {
		var cmd1 tea.Cmd
		m.sessionList, cmd1 = m.sessionList.Update(msg)
		cmds = append(cmds, cmd1)

		// Check if session selection changed
		selectedItem := m.sessionList.SelectedItem()
		if selectedItem != nil {
			selected := selectedItem.(item).Title()
			if selected != m.selectedSession {
				m.selectedSession = selected
				cmds = append(cmds, loadWindows(selected))
			}
		}
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		if m.inputMode == inputFuzzy {
			filtered := []list.Item{}
			query := strings.ToLower(m.textInput.Value())
			for _, s := range m.sessions {
				if strings.Contains(strings.ToLower(s.Name), query) {
					filtered = append(filtered, item{title: s.Name, desc: ""})
				}
			}
			m.sessionList.SetItems(filtered)
		}
	}

	return m, tea.Batch(cmds...)
}

// truncateLines truncates each line in s to maxWidth runes.
func truncateLines(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		r := []rune(line)
		if len(r) > maxWidth {
			out = append(out, string(r[:maxWidth]))
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func (m model) View() string {
	ascii := `  ___  ___  ___ _ __  
 / __|/ _ \/ __| '_ \ 
 \__ \  __/\__ \ | | |
 |___/\___||___/_| |_|
`		
	var header string
	if m.inputMode == inputNone {
		header = "c: create  d: delete  r: rename\nk: kill  enter: attach  /: fuzzy find\n"
	} else {
		var prompt string
		switch m.inputMode {
		case inputFuzzy:
			prompt = "Find: "
		case inputCreate:
			prompt = "Create: "
		case inputRename:
			prompt = "Rename: "
		}
		header = fmt.Sprintf("%s %s\n", prompt, m.textInput.View())
	}


		// Determine column widths (fall back if not set yet)
		leftW := m.width/2
		if leftW <= 0 {
			leftW = 20
		}
		rightW := m.width - leftW - 1
		if rightW <= 0 {
			rightW = 20
		}

		// Compose titled columns
		sessionTitle := lipgloss.NewStyle().Bold(true).Render("sessions")
		windowTitle := lipgloss.NewStyle().Bold(true).Render("windows")

		// Render compact session list (one line per session, minimal padding)
		leftLines := []string{sessionTitle}
		for i, s := range m.sessions {
			prefix := "  "
			if m.sessionList.Index() == i {
				prefix = "> "
			}
			leftLines = append(leftLines, prefix+s.Name)
		}
		left := strings.Join(leftLines, "\n")

		// Render compact window list
		rightLines := []string{windowTitle}
		for _, w := range m.windows {
			// Show index and name compactly without selection indicator
			rightLines = append(rightLines, fmt.Sprintf("%d: %s", w.Index, w.Name))
		}
		right := strings.Join(rightLines, "\n")

		// Ensure each column renders within its width
	// Truncate each column to its width to keep layout compact
	leftStyled := lipgloss.NewStyle().Width(leftW).Align(lipgloss.Left).Render(truncateLines(left, leftW))
	rightStyled := lipgloss.NewStyle().Width(rightW).Align(lipgloss.Left).Render(truncateLines(right, rightW))

		body := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
		return ascii + "\n" + header + "\n" + body
}
