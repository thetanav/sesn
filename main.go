package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("I am a tmux session manager")
	createSession("tanav")
}

func createSession(name string) {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Session create named: ", name)
}

func deleteSession(name string) {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Session killed name: ", name)
}

func renameSession(old string, new string) {
	cmd := exec.Command("tmux", "rename-session", "-t", old, new)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error", err)
	}
	fmt.Println("Session renamed to: ", new)
}

func attachSession(name string) {
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func checkTmux() {
	cmd := exec.Command("tmux", "-V")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Tmux installed")
}
