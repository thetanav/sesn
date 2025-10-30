package internals

import (
	"fmt"
	"os/exec"
)


func CreateSession(name string) {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Session create named: ", name)
}

func DeleteSession(name string) {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Session killed name: ", name)
}

func RenameSession(old string, new string) {
	cmd := exec.Command("tmux", "rename-session", "-t", old, new)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println("Session renamed to: ", new)
}

func AttachSession(name string) {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func ListSession() {
	cmd := exec.Command("tmux", "list-sessions")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(string(out))
}

func ListWindows(name string) {
	cmd := exec.Command("tmux", "list-windows", "-t", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println(string(out))
}

func CheckTmux() {
	cmd := exec.Command("tmux", "-V")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Tmux installed")
}
