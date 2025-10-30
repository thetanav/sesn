package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tmuxly/internals"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type mode int

const (
	modeNormal mode = iota
	modeCreate
	modeRename
	modeDelete
)

type model struct {
	sessionList     list.Model
	windowList      list.Model
	sessions        []internals.Session
	windows         []internals.Window
	mode            mode
	textInput       textinput.Model
	selectedSession string
	width, height   int
}

func initialModel() model {
	sessionItems := []list.Item{}
	windowItems := []list.Item{}

	sessionList := list.New(sessionItems, list.NewDefaultDelegate(), 0, 0)
	sessionList.SetShowHelp(false)

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
		mode:        modeNormal,
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
		switch m.mode {
		case modeNormal:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "c":
				m.mode = modeCreate
				m.textInput.Reset()
				return m, m.textInput.Focus()
			case "d":
				if m.selectedSession != "" {
					internals.DeleteSession(m.selectedSession)
					return m, loadSessions
				}
			case "r":
				if m.selectedSession != "" {
					m.mode = modeRename
					m.textInput.Reset()
					return m, m.textInput.Focus()
				}
			case "k":
				if m.selectedSession != "" {
					internals.DeleteSession(m.selectedSession)
					return m, loadSessions
				}
			case "enter":
				if m.selectedSession != "" {
					internals.AttachSession(m.selectedSession)
					return m, tea.Quit
				}
			}
		case modeCreate, modeRename:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.textInput.Value())
				if name != "" {
					if m.mode == modeCreate {
						internals.CreateSession(name)
					} else {
						internals.RenameSession(m.selectedSession, name)
					}
					m.mode = modeNormal
					return m, loadSessions
				}
			case "esc":
				m.mode = modeNormal
				return m, nil
			}
		}
	}

	if m.mode == modeNormal {
		var cmd1, cmd2 tea.Cmd
		m.sessionList, cmd1 = m.sessionList.Update(msg)
		m.windowList, cmd2 = m.windowList.Update(msg)
		cmds = append(cmds, cmd1, cmd2)

		// Check if session selection changed
		if m.sessionList.Index() >= 0 && m.sessionList.Index() < len(m.sessions) {
			selected := m.sessions[m.sessionList.Index()].Name
			if selected != m.selectedSession {
				m.selectedSession = selected
				cmds = append(cmds, loadWindows(selected))
			}
		}
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
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
	if m.mode == modeNormal {
		header := "c: create  d: delete  r: rename  k: kill  enter: attach"

		// Determine column widths (fall back if not set yet)
		leftW := m.width/2
		if leftW <= 0 {
			leftW = 30
		}
		rightW := m.width - leftW - 1
		if rightW <= 0 {
			rightW = 30
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
		for i, w := range m.windows {
			prefix := "  "
			if m.windowList.Index() == i {
				prefix = "> "
			}
			// Show index and name compactly
			rightLines = append(rightLines, prefix+fmt.Sprintf("%d: %s", w.Index, w.Name))
		}
		right := strings.Join(rightLines, "\n")

		// Ensure each column renders within its width
	// Truncate each column to its width to keep layout compact
	leftStyled := lipgloss.NewStyle().Width(leftW).Align(lipgloss.Left).Render(truncateLines(left, leftW))
	rightStyled := lipgloss.NewStyle().Width(rightW).Align(lipgloss.Left).Render(truncateLines(right, rightW))

		body := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, rightStyled)
		return header + "\n" + body
	} else {
		var prompt string
		if m.mode == modeCreate {
			prompt = "Create:"
		} else {
			prompt = "Rename:"
		}
		return fmt.Sprintf("%s %s", prompt, m.textInput.View())
	}
}
