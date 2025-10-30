package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sesn/internals"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		fmt.Println("I am sesn, a tmux session manager")
		internals.CreateSession("tanav")
		return
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
