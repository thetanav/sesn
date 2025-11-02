package main

import (
	"flag"
	"fmt"
	"os"

	"sesn/internals"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if !internals.CheckTmux() {
		fmt.Println("Error: tmux not found. Please install tmux.")
		os.Exit(1)
	}

	fuzzy := flag.Bool("f", false, "run fuzzy finder")
	flag.Parse()

	if *fuzzy {
		internals.CanaryFuzzy()
		return
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
