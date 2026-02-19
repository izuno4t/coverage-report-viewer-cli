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
