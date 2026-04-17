package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/app"
	"github.com/menegas/lumina/cli"
	"github.com/menegas/lumina/components/layout"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

// version is injected at build time via -ldflags "-X main.version=...".
// Defaults to "dev" when built without the flag (e.g. local `go build`).
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println(version)
			return
		case "--help", "-h":
			fmt.Print(cli.UsageText())
			return
		}
	}

	overrides, err := cli.ParseArgs(os.Args[1:], os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lumina: config error: %v\n", err)
		os.Exit(1)
	}

	layoutOpts := buildLayoutOpts(overrides)

	model, err := app.New(cfg, layoutOpts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lumina: init error: %v\n", err)
		os.Exit(1)
	}

	var initialCmd tea.Cmd
	if overrides.FilePath != "" {
		path := overrides.FilePath
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

// buildLayoutOpts maps the CLI overrides into layout.Option values.
func buildLayoutOpts(o cli.StartupOverrides) []layout.Option {
	var opts []layout.Option
	opts = append(opts, layout.WithMaxPanes(o.EffectiveMaxPanes()))
	if o.StartCommand != "" {
		opts = append(opts, layout.WithStartCommand(o.StartCommand))
	}
	if o.StartPanes > 1 {
		opts = append(opts, layout.WithInitialLayout(orientToSplitDir(o.StartOrient), o.StartPanes))
	}
	return opts
}

func orientToSplitDir(o cli.Orient) msgs.SplitDir {
	if o == cli.OrientVertical {
		return msgs.SplitVertical
	}
	return msgs.SplitHorizontal
}
