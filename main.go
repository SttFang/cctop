package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fanghanjun/cctop/internal/config"
	"github.com/fanghanjun/cctop/internal/tui"
	"github.com/fanghanjun/cctop/pkg/version"
)

func main() {
	cfg := config.Parse()

	if cfg.Version {
		fmt.Printf("cctop %s (built %s)\n", version.Version, version.BuildTime)
		os.Exit(0)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	app := tui.NewApp(cfg)

	p := tea.NewProgram(
		&app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
