package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/izuno4t/coverage-report-viewer-cli/internal/jacoco"
)

func Start(report jacoco.Report, cfg Config) error {
	m := NewModel(report, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func StartWatch(report jacoco.Report, cfg Config, reloadFn func() (jacoco.Report, error), probeFn func() (bool, error)) error {
	m := newModel(report, cfg, reloadFn, probeFn)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
