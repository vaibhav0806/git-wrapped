package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vaibhav0806/git-wrapped/export"
	"github.com/vaibhav0806/git-wrapped/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gh-wrapped <username> [--auto]")
		os.Exit(1)
	}
	username := os.Args[1]
	auto := len(os.Args) > 2 && os.Args[2] == "--auto"

	token := os.Getenv("GITHUB_TOKEN")

	model := ui.NewModel(username, token, auto)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(ui.Model); ok && m.ExportRequested() {
		if err := export.GenerateGIF(m.Username()); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}
