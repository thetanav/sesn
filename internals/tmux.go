package internals

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Session struct {
	Name     string
	Windows  int
	Created  string
	Attached bool
}

type Window struct {
	Index   int
	Name    string
	Panes   int
	Size    string
	Created string
	Active  bool
}

func CreateSession(name string) {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func DeleteSession(name string) {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func RenameSession(old string, new string) {
	cmd := exec.Command("tmux", "rename-session", "-t", old, new)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error", err)
	}
}

func AttachSession(name string) {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func ListSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ParseSessions(string(out)), nil
}

func ListWindows(name string) ([]Window, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ParseWindows(string(out)), nil
}

func CheckTmux() {
	cmd := exec.Command("tmux", "-V")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Tmux installed")
}

func ParseSessions(output string) []Session {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var sessions []Session
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		details := parts[1]
		// Parse details: "1 windows (created Thu Oct 30 14:23:45 2025) [80x24] (attached)"
		// Remove (attached) if present
		attached := strings.Contains(details, "(attached)")
		details = strings.Replace(details, " (attached)", "", 1)
		// Split by spaces
		detailParts := strings.Fields(details)
		if len(detailParts) < 4 {
			continue
		}
		windows, _ := strconv.Atoi(detailParts[0])
		created := strings.Join(detailParts[2:len(detailParts)-1], " ")
		sessions = append(sessions, Session{
			Name:     name,
			Windows:  windows,
			Created:  created,
			Attached: attached,
		})
	}
	return sessions
}

func ParseWindows(output string) []Window {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var windows []Window
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		indexStr := parts[0]
		details := parts[1]
		index, _ := strconv.Atoi(indexStr)
		// Parse details: "bash* (1 panes) [80x24] created Thu Oct 30 14:23:45 2025"
		active := strings.Contains(details, "*")
		details = strings.Replace(details, "*", "", 1)
		detailParts := strings.Fields(details)
		if len(detailParts) < 5 {
			continue
		}
		name := detailParts[0]
		panes, _ := strconv.Atoi(detailParts[1])
		size := detailParts[3]
		created := strings.Join(detailParts[5:], " ")
		windows = append(windows, Window{
			Index:   index,
			Name:    name,
			Panes:   panes,
			Size:    size,
			Created: created,
			Active:  active,
		})
	}
	return windows
}
