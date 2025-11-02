package internals

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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

type SavedSession struct {
	Name    string   `json:"name"`
	Windows []string `json:"windows"`
}

func CreateSession(name string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	_, err := cmd.CombinedOutput()
	return err
}

func DeleteSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	_, err := cmd.CombinedOutput()
	return err
}

func RenameSession(old string, new string) error {
	cmd := exec.Command("tmux", "rename-session", "-t", old, new)
	_, err := cmd.CombinedOutput()
	return err
}

func AttachSession(name string) error {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("error finding tmux: %v", err)
	}
	err = syscall.Exec(path, []string{"tmux", "attach-session", "-t", name}, os.Environ())
	if err != nil {
		return fmt.Errorf("error attaching to session: %v", err)
	}
	return nil
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

func CheckTmux() bool {
	cmd := exec.Command("tmux", "-V")
	_, err := cmd.CombinedOutput()
	return err == nil
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

func SaveSession(name string) error {
	windows, err := ListWindows(name)
	if err != nil {
		return err
	}
	var winNames []string
	for _, w := range windows {
		winNames = append(winNames, w.Name)
	}
	saved := SavedSession{Name: name, Windows: winNames}
	data, err := json.Marshal(saved)
	if err != nil {
		return err
	}
	filename := name + ".json"
	return os.WriteFile(filename, data, 0644)
}

func LoadSession(name string) error {
	filename := name + ".json"
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var saved SavedSession
	err = json.Unmarshal(data, &saved)
	if err != nil {
		return err
	}
	// Create session
	err = CreateSession(saved.Name)
	if err != nil {
		return err
	}
	// Create windows
	for i, winName := range saved.Windows {
		if i == 0 {
			// First window is already created
			cmd := exec.Command("tmux", "rename-window", "-t", fmt.Sprintf("%s:0", saved.Name), winName)
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}
		} else {
			cmd := exec.Command("tmux", "new-window", "-t", saved.Name, "-n", winName)
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CanaryFuzzy() error {
	cmd := exec.Command("bash", "-c", "tmux list-sessions | fzf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running fuzzy finder: %v", err)
	}
	line := string(out)
	parts := strings.SplitN(line, ":", 2)
	sessionName := strings.TrimSpace(parts[0])
	return AttachSession(sessionName)
}
