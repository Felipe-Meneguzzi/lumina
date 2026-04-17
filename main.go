package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/app"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lumina: config error: %v\n", err)
		os.Exit(1)
	}

	model, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lumina: init error: %v\n", err)
		os.Exit(1)
	}

	// Open a file if provided as argument.
	var initialCmd tea.Cmd
	if len(os.Args) > 1 {
		path := os.Args[1]
		initialCmd = func() tea.Msg {
			return msgs.OpenFileMsg{Path: path}
		}
	}

	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	}
	if initialCmd != nil {
		opts = append(opts, tea.WithoutSignalHandler())
	}

	p := tea.NewProgram(model, opts...)

	if initialCmd != nil {
		go func() {
			p.Send(initialCmd())
		}()
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lumina: %v\n", err)
		os.Exit(1)
	}
}
