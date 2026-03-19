package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"github.com/vaibhav/gh-wrapped/export"
	"github.com/vaibhav/gh-wrapped/github"
	"github.com/vaibhav/gh-wrapped/personality"
	"github.com/vaibhav/gh-wrapped/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gh-wrapped <username> [--auto]")
		os.Exit(1)
	}
	username := os.Args[1]
	auto := len(os.Args) > 2 && os.Args[2] == "--auto"

	token := os.Getenv("GITHUB_TOKEN")
	client := github.NewClient("https://api.github.com", token)

	var (
		user     github.User
		events   []github.Event
		repos    []github.Repo
		calendar []github.ContributionDay
	)

	g := new(errgroup.Group)

	g.Go(func() error {
		var err error
		user, err = client.FetchUser(username)
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		events, err = client.FetchEvents(username)
		if err != nil {
			events = nil
		}
		return nil
	})

	g.Go(func() error {
		var err error
		repos, err = client.FetchRepos(username)
		if err != nil {
			repos = nil
		}
		return nil
	})

	g.Go(func() error {
		var err error
		calendar, err = client.FetchContributions(username)
		if err != nil {
			calendar = nil
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(events) == 0 && len(calendar) == 0 && len(repos) == 0 {
		fmt.Printf("Looks like @%s had a quiet year. Nothing to wrap!\n", username)
		os.Exit(0)
	}

	stats := github.ComputeStats(user, events, repos, calendar)

	p := personality.Compute(stats)
	stats.Archetype = p.Archetype
	stats.Traits = p.Traits

	model := ui.NewModel(stats, auto)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(ui.Model); ok && m.ExportRequested() {
		if err := export.GenerateGIF(username); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}
