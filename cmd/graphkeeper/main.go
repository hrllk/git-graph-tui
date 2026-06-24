package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"hrllk/graphkeeper/internal/app"
	"hrllk/graphkeeper/internal/git"
)

func main() {
	repo, err := git.Open(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	model, err := app.New(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	prog := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
