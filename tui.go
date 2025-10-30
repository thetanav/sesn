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
		listHeight := 10
		if msg.Height < 10 {
			listHeight = msg.Height
		}
		halfHeight := listHeight / 2
		m.sessionList.SetSize(msg.Width, halfHeight)
		m.windowList.SetSize(msg.Width, halfHeight)
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
			case "q", "ctrl+c":
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

func (m model) View() string {
	if m.mode == modeNormal {
		header := "c: create  d: delete  r: rename  k: kill  enter: attach  q: quit"
		sessionView := m.sessionList.View()
		windowView := m.windowList.View()
		body := lipgloss.JoinVertical(lipgloss.Left, sessionView, windowView)
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
