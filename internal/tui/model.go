package tui

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/izuno4t/jacoco-report-viewer-cli/internal/jacoco"
)

type Config struct {
	Threshold int
	Sort      string
	NoColor   bool
}

type nodeKind int

const (
	nodeReport nodeKind = iota
	nodePackage
	nodeClass
)

const (
	draculaComment = "#6272A4"
	draculaCyan    = "#8BE9FD"
	draculaGreen   = "#50FA7B"
	draculaPink    = "#FF79C6"
	draculaPurple  = "#BD93F9"
	draculaRed     = "#FF5555"
	draculaYellow  = "#F1FA8C"
)

type navNode struct {
	kind      nodeKind
	packageIx int
	classIx   int
	cursor    int
	offset    int
}

type Model struct {
	report jacoco.Report
	config Config
	stack  []navNode
	sortID string
	width  int
	height int

	titleStyle  lipgloss.Style
	headerStyle lipgloss.Style
	itemStyle   lipgloss.Style
	cursorStyle lipgloss.Style
	helpStyle   lipgloss.Style
}

func NewModel(report jacoco.Report, cfg Config) Model {
	m := Model{
		report: report,
		config: cfg,
		stack:  []navNode{{kind: nodeReport, cursor: 0, offset: 0}},
		sortID: normalizeInitialSort(cfg.Sort),
		width:  100,
		height: 30,
		titleStyle: lipgloss.NewStyle().
			Bold(true),
		headerStyle: lipgloss.NewStyle().
			Bold(true),
		itemStyle: lipgloss.NewStyle(),
		cursorStyle: lipgloss.NewStyle().
			Bold(true),
		helpStyle: lipgloss.NewStyle().
			Faint(true),
	}
	if !cfg.NoColor {
		m.titleStyle = m.titleStyle.Foreground(lipgloss.Color(draculaPurple))
		m.headerStyle = m.headerStyle.Foreground(lipgloss.Color(draculaCyan))
		m.cursorStyle = m.cursorStyle.Foreground(lipgloss.Color(draculaPink))
		m.helpStyle = m.helpStyle.Foreground(lipgloss.Color(draculaComment))
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureCursorVisible(m.childCount(*m.current()))
		return m, nil
	case tea.KeyMsg:
		if quit := m.applyKey(msg.String()); quit {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) applyKey(key string) (quit bool) {
	switch key {
	case "q", "ctrl+c":
		return true
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "enter":
		m.enterChild()
	case "b", "backspace":
		m.goBack()
	case "s":
		m.toggleSort()
	}
	return false
}

func normalizeInitialSort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "coverage":
		return "coverage-asc"
	default:
		return "name-asc"
	}
}

func (m *Model) toggleSort() {
	switch m.sortID {
	case "name-asc":
		m.sortID = "coverage-asc"
	case "coverage-asc":
		m.sortID = "coverage-desc"
	default:
		m.sortID = "name-asc"
	}
	m.current().cursor = 0
	m.current().offset = 0
}

func (m *Model) moveCursor(delta int) {
	current := m.current()
	childCount := m.childCount(*current)
	if childCount == 0 {
		current.cursor = 0
		current.offset = 0
		return
	}
	current.cursor += delta
	if current.cursor < 0 {
		current.cursor = 0
	}
	if current.cursor > childCount-1 {
		current.cursor = childCount - 1
	}
	m.ensureCursorVisible(childCount)
}

func (m *Model) enterChild() {
	current := m.current()
	children := m.currentChildren()
	if len(children) == 0 {
		return
	}
	if current.cursor < 0 || current.cursor >= len(children) {
		return
	}
	selected := children[current.cursor]

	switch current.kind {
	case nodeReport:
		m.stack = append(m.stack, navNode{
			kind:      nodePackage,
			packageIx: selected.index,
			cursor:    0,
			offset:    0,
		})
	case nodePackage:
		m.stack = append(m.stack, navNode{
			kind:      nodeClass,
			packageIx: current.packageIx,
			classIx:   selected.index,
			cursor:    0,
			offset:    0,
		})
	case nodeClass:
		// Method level is the leaf in current milestone.
		return
	}
}

func (m *Model) goBack() {
	if len(m.stack) <= 1 {
		return
	}
	m.stack = m.stack[:len(m.stack)-1]
}

func (m *Model) current() *navNode {
	return &m.stack[len(m.stack)-1]
}

func (m Model) View() string {
	summary := m.renderSummary()
	parts := []string{
		m.renderBreadcrumb(),
		"",
		summary,
		"",
		m.renderChildren(),
		"",
		m.helpStyle.Render(fmt.Sprintf("sort: %s | ↑/↓ or j/k: move  Enter: open  b: back  s: sort  q: quit", m.sortLabel())),
	}
	return strings.Join(parts, "\n")
}

func (m Model) renderBreadcrumb() string {
	labels := []string{"Report"}
	for i := 1; i < len(m.stack); i++ {
		n := m.stack[i]
		if n.packageIx >= 0 && n.packageIx < len(m.report.Packages) {
			pkg := m.report.Packages[n.packageIx]
			if n.kind == nodePackage {
				labels = append(labels, pkg.Name)
				continue
			}
			if n.classIx >= 0 && n.classIx < len(pkg.Classes) {
				labels = append(labels, pkg.Name, pkg.Classes[n.classIx].Name)
			}
		}
	}
	if m.report.Name != "" {
		labels[0] = fmt.Sprintf("Report(%s)", m.report.Name)
	}
	return m.titleStyle.Render(strings.Join(uniqueOrdered(labels), " > "))
}

func uniqueOrdered(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (m Model) renderSummary() string {
	lines := []string{m.headerStyle.Render("Summary")}
	counters := m.currentCounters()
	for _, t := range []jacoco.CounterType{jacoco.CounterInstruction, jacoco.CounterBranch, jacoco.CounterLine, jacoco.CounterMethod} {
		if c, ok := findCounter(counters, t); ok {
			rate := c.CoverageRate()
			line := fmt.Sprintf("%-12s %6.1f%%  %s", t, rate, bar(rate, 20))
			lines = append(lines, m.styleForCoverage(rate).Render(line))
		}
	}
	if len(lines) == 1 {
		lines = append(lines, "(no counters)")
	}
	return strings.Join(lines, "\n")
}

func findCounter(counters []jacoco.Counter, t jacoco.CounterType) (jacoco.Counter, bool) {
	for _, c := range counters {
		if c.Type == t {
			return c, true
		}
	}
	return jacoco.Counter{}, false
}

func (m Model) currentCounters() []jacoco.Counter {
	current := m.stack[len(m.stack)-1]
	switch current.kind {
	case nodeReport:
		return m.report.Counters
	case nodePackage:
		return m.report.Packages[current.packageIx].Counters
	case nodeClass:
		return m.report.Packages[current.packageIx].Classes[current.classIx].Counters
	default:
		return nil
	}
}

func (m Model) renderChildren() string {
	lines := []string{m.headerStyle.Render(fmt.Sprintf("Children (%s)", m.sortLabel()))}
	children := m.currentChildren()
	if len(children) == 0 {
		return lines[0] + "\n(no children)"
	}
	nameWidth := maxChildNameWidth(children)
	maxNameWidth := m.width - 20
	if maxNameWidth < 12 {
		maxNameWidth = 12
	}
	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}
	current := m.stack[len(m.stack)-1]
	maxRows := m.maxVisibleChildren()
	if maxRows > len(children) {
		maxRows = len(children)
	}
	offset := current.offset
	maxOffset := len(children) - maxRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + maxRows
	for i, c := range children[offset:end] {
		rowIx := offset + i
		marker := " "
		style := m.itemStyle
		if rowIx == current.cursor {
			marker = "❯"
			style = style.Inherit(m.cursorStyle)
		}
		name := compactNameForDisplay(c.name, nameWidth)
		line := fmt.Sprintf("%s %s %6.1f%% %s", marker, padRightDisplay(name, nameWidth), c.coverage, bar(c.coverage, 10))
		style = style.Inherit(m.styleForCoverage(c.coverage))
		lines = append(lines, style.Render(line))
	}
	return strings.Join(lines, "\n")
}

func (m Model) summaryLineCount() int {
	count := 1
	counters := m.currentCounters()
	for _, t := range []jacoco.CounterType{jacoco.CounterInstruction, jacoco.CounterBranch, jacoco.CounterLine, jacoco.CounterMethod} {
		if _, ok := findCounter(counters, t); ok {
			count++
		}
	}
	if count == 1 {
		return 2
	}
	return count
}

func (m Model) maxVisibleChildren() int {
	available := m.height - (m.summaryLineCount() + 6)
	if available < 1 {
		return 1
	}
	return available
}

func (m *Model) ensureCursorVisible(childCount int) {
	current := m.current()
	if childCount <= 0 {
		current.cursor = 0
		current.offset = 0
		return
	}
	maxRows := m.maxVisibleChildren()
	maxOffset := childCount - maxRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if current.cursor < current.offset {
		current.offset = current.cursor
	}
	if current.cursor >= current.offset+maxRows {
		current.offset = current.cursor - maxRows + 1
	}
	if current.offset < 0 {
		current.offset = 0
	}
	if current.offset > maxOffset {
		current.offset = maxOffset
	}
}

func maxChildNameWidth(children []childRow) int {
	maxWidth := 0
	for _, c := range children {
		w := lipgloss.Width(c.name)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

func padRightDisplay(s string, width int) string {
	padding := width - lipgloss.Width(s)
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
}

func compactNameForDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}

	abbreviated := abbreviateSpringLikePath(s, width)
	if lipgloss.Width(abbreviated) <= width {
		return abbreviated
	}
	return ellipsizeMiddleDisplay(abbreviated, width)
}

func abbreviateSpringLikePath(s string, width int) string {
	sep := detectPathSeparator(s)
	if sep == "" {
		return s
	}
	parts := strings.Split(s, sep)
	if len(parts) <= 1 {
		return s
	}
	out := append([]string(nil), parts...)
	for i := 0; i < len(out)-1; i++ {
		if lipgloss.Width(strings.Join(out, sep)) <= width {
			break
		}
		out[i] = abbreviateSegment(out[i])
	}
	return strings.Join(out, sep)
}

func detectPathSeparator(s string) string {
	switch {
	case strings.Contains(s, "/"):
		return "/"
	case strings.Contains(s, "."):
		return "."
	default:
		return ""
	}
}

func abbreviateSegment(seg string) string {
	if seg == "" {
		return seg
	}
	r, _ := utf8.DecodeRuneInString(seg)
	return string(r)
}

func ellipsizeMiddleDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	ellipsis := "..."
	target := width - lipgloss.Width(ellipsis)
	leftTarget := target/2 + target%2
	rightTarget := target / 2

	left := ""
	leftWidth := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if leftWidth+rw > leftTarget {
			break
		}
		left += string(r)
		leftWidth += rw
	}

	right := ""
	rightWidth := 0
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		rw := lipgloss.Width(string(runes[i]))
		if rightWidth+rw > rightTarget {
			break
		}
		right = string(runes[i]) + right
		rightWidth += rw
	}

	out := left + ellipsis + right
	for lipgloss.Width(out) > width && len(right) > 0 {
		_, size := utf8.DecodeRuneInString(right)
		right = right[size:]
		out = left + ellipsis + right
	}
	return out
}

type childRow struct {
	index    int
	name     string
	coverage float64
}

func (m Model) currentChildren() []childRow {
	current := m.stack[len(m.stack)-1]
	rows := make([]childRow, 0)
	switch current.kind {
	case nodeReport:
		rows = make([]childRow, 0, len(m.report.Packages))
		for i, p := range m.report.Packages {
			rows = append(rows, childRow{index: i, name: p.Name, coverage: instructionCoverage(p.Counters)})
		}
	case nodePackage:
		pkg := m.report.Packages[current.packageIx]
		rows = make([]childRow, 0, len(pkg.Classes))
		for i, c := range pkg.Classes {
			rows = append(rows, childRow{index: i, name: c.Name, coverage: instructionCoverage(c.Counters)})
		}
	case nodeClass:
		class := m.report.Packages[current.packageIx].Classes[current.classIx]
		rows = make([]childRow, 0, len(class.Methods))
		for i, m := range class.Methods {
			rows = append(rows, childRow{index: i, name: m.Name, coverage: instructionCoverage(m.Counters)})
		}
	default:
		return nil
	}
	m.sortRows(rows)
	return rows
}

func (m Model) sortRows(rows []childRow) {
	switch m.sortID {
	case "coverage-asc":
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].coverage == rows[j].coverage {
				return rows[i].name < rows[j].name
			}
			return rows[i].coverage < rows[j].coverage
		})
	case "coverage-desc":
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].coverage == rows[j].coverage {
				return rows[i].name < rows[j].name
			}
			return rows[i].coverage > rows[j].coverage
		})
	default:
		sort.SliceStable(rows, func(i, j int) bool {
			return rows[i].name < rows[j].name
		})
	}
}

func (m Model) sortLabel() string {
	switch m.sortID {
	case "coverage-asc":
		return "coverage asc"
	case "coverage-desc":
		return "coverage desc"
	default:
		return "name asc"
	}
}

func (m Model) childCount(node navNode) int {
	switch node.kind {
	case nodeReport:
		return len(m.report.Packages)
	case nodePackage:
		return len(m.report.Packages[node.packageIx].Classes)
	case nodeClass:
		return len(m.report.Packages[node.packageIx].Classes[node.classIx].Methods)
	default:
		return 0
	}
}

func instructionCoverage(counters []jacoco.Counter) float64 {
	if c, ok := findCounter(counters, jacoco.CounterInstruction); ok {
		return c.CoverageRate()
	}
	return 0
}

func bar(percentage float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int((percentage / 100) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

type coverageBand int

const (
	bandLow coverageBand = iota
	bandMid
	bandHigh
)

func bandForCoverage(rate float64, threshold int) coverageBand {
	if rate >= 90 {
		return bandHigh
	}
	if rate >= float64(threshold) {
		return bandMid
	}
	return bandLow
}

func (m Model) styleForCoverage(rate float64) lipgloss.Style {
	if m.config.NoColor {
		return lipgloss.NewStyle()
	}
	switch bandForCoverage(rate, m.config.Threshold) {
	case bandHigh:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(draculaGreen))
	case bandMid:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(draculaYellow))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(draculaRed))
	}
}
