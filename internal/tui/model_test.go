package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/izuno4t/coverage-report-viewer-cli/internal/jacoco"
)

func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func sampleReport() jacoco.Report {
	return jacoco.Report{
		Name:     "demo",
		Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 2, Covered: 8}},
		Packages: []jacoco.Package{{
			Name:     "com/example",
			Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9}},
			Classes: []jacoco.Class{{
				Name:     "UserService",
				Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 1, Covered: 4}},
				Methods: []jacoco.Method{{
					Name:     "find",
					Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 0, Covered: 2}},
				}},
			}},
		}},
	}
}

func TestViewIncludesSections(t *testing.T) {
	m := NewModel(sampleReport(), Config{Threshold: 80, Sort: "name"})
	view := m.View()

	for _, want := range []string{"Report(demo)", "Summary (counter: instruction)", "Children", "com/example", "c: counter", "/: filter", "g/G: jump"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q", want)
		}
	}
}

func TestDrillDownAndBack(t *testing.T) {
	m := NewModel(sampleReport(), Config{Threshold: 80, Sort: "name"})

	m.applyKey("enter")
	if len(m.stack) != 2 || m.stack[1].kind != nodePackage {
		t.Fatalf("expected package level, stack=%+v", m.stack)
	}

	m.applyKey("enter")
	if len(m.stack) != 3 || m.stack[2].kind != nodeClass {
		t.Fatalf("expected class level, stack=%+v", m.stack)
	}

	m.applyKey("b")
	if len(m.stack) != 2 || m.stack[1].kind != nodePackage {
		t.Fatalf("expected back to package level, stack=%+v", m.stack)
	}

	m.applyKey("backspace")
	if len(m.stack) != 1 || m.stack[0].kind != nodeReport {
		t.Fatalf("expected back to report level, stack=%+v", m.stack)
	}
}

func TestMethodDisplayNameIncludesSignatureAndLine(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{{
			Name: "pkg",
			Classes: []jacoco.Class{{
				Name: "C",
				Methods: []jacoco.Method{{
					Name: "find",
					Desc: "(I)Ljava/lang/String;",
					Line: 42,
					Counters: []jacoco.Counter{
						{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9},
					},
				}},
			}},
		}},
	}
	m := NewModel(report, Config{Sort: "name"})
	m.applyKey("enter")
	m.applyKey("enter")

	rows := m.currentChildren()
	if len(rows) != 1 {
		t.Fatalf("method row count mismatch: %d", len(rows))
	}
	if rows[0].name != "find(I)Ljava/lang/String;:42" {
		t.Fatalf("unexpected method label: %q", rows[0].name)
	}
}

func TestBarWidth(t *testing.T) {
	got := bar(50, 10)
	if got != "█████░░░░░" {
		t.Fatalf("bar mismatch: %q", got)
	}
}

func TestSortCycle(t *testing.T) {
	m := NewModel(sampleReport(), Config{Sort: "name"})
	if m.sortID != "name-asc" {
		t.Fatalf("unexpected initial sort: %s", m.sortID)
	}
	m.applyKey("s")
	if m.sortID != "coverage-asc" {
		t.Fatalf("unexpected sort after first toggle: %s", m.sortID)
	}
	m.applyKey("s")
	if m.sortID != "coverage-desc" {
		t.Fatalf("unexpected sort after second toggle: %s", m.sortID)
	}
	m.applyKey("s")
	if m.sortID != "name-asc" {
		t.Fatalf("unexpected sort after third toggle: %s", m.sortID)
	}
}

func TestCounterTypeCycle(t *testing.T) {
	m := NewModel(sampleReport(), Config{Sort: "name"})
	if m.counterType != jacoco.CounterInstruction {
		t.Fatalf("unexpected initial counter type: %s", m.counterType)
	}
	m.applyKey("c")
	if m.counterType != jacoco.CounterBranch {
		t.Fatalf("unexpected counter type after first toggle: %s", m.counterType)
	}
	m.applyKey("c")
	if m.counterType != jacoco.CounterLine {
		t.Fatalf("unexpected counter type after second toggle: %s", m.counterType)
	}
	m.applyKey("c")
	if m.counterType != jacoco.CounterInstruction {
		t.Fatalf("unexpected counter type after third toggle: %s", m.counterType)
	}
}

func TestCoverageSortAffectsChildOrder(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{Name: "z-low", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 8, Covered: 2}}},
			{Name: "a-high", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9}}},
		},
	}
	m := NewModel(report, Config{Sort: "name"})

	rows := m.currentChildren()
	if rows[0].name != "a-high" {
		t.Fatalf("name sort mismatch: first=%s", rows[0].name)
	}

	m.applyKey("s")
	rows = m.currentChildren()
	if rows[0].name != "z-low" {
		t.Fatalf("coverage asc mismatch: first=%s", rows[0].name)
	}

	m.applyKey("s")
	rows = m.currentChildren()
	if rows[0].name != "a-high" {
		t.Fatalf("coverage desc mismatch: first=%s", rows[0].name)
	}
}

func TestCounterSwitchAffectsCoverageSort(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{
				Name: "pkg-a",
				Counters: []jacoco.Counter{
					{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9},
					{Type: jacoco.CounterBranch, Missed: 8, Covered: 2},
				},
			},
			{
				Name: "pkg-b",
				Counters: []jacoco.Counter{
					{Type: jacoco.CounterInstruction, Missed: 8, Covered: 2},
					{Type: jacoco.CounterBranch, Missed: 1, Covered: 9},
				},
			},
		},
	}
	m := NewModel(report, Config{Sort: "name"})

	m.applyKey("s")
	rows := m.currentChildren()
	if rows[0].name != "pkg-b" {
		t.Fatalf("instruction coverage asc mismatch: first=%s", rows[0].name)
	}

	m.applyKey("c")
	rows = m.currentChildren()
	if rows[0].name != "pkg-a" {
		t.Fatalf("branch coverage asc mismatch: first=%s", rows[0].name)
	}
}

func TestIncrementalFilter(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{Name: "com/example/service", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9}}},
			{Name: "com/example/repository", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 2, Covered: 8}}},
			{Name: "org/other", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 3, Covered: 7}}},
		},
	}
	m := NewModel(report, Config{Sort: "name"})

	m.applyKey("/")
	m.applyKey("e")
	m.applyKey("x")

	rows := m.currentChildren()
	if len(rows) != 2 {
		t.Fatalf("expected 2 filtered rows, got %d", len(rows))
	}
	for _, row := range rows {
		if !strings.Contains(row.name, "example") {
			t.Fatalf("unexpected filtered row: %s", row.name)
		}
	}

	m.applyKey("esc")
	rows = m.currentChildren()
	if len(rows) != 3 {
		t.Fatalf("expected filter clear to restore rows, got %d", len(rows))
	}
}

func TestFilterPromptVisibleInFilterMode(t *testing.T) {
	m := NewModel(sampleReport(), Config{Sort: "name"})
	m.applyKey("/")
	m.applyKey("a")
	view := m.View()
	if !strings.Contains(view, "filter> a (Enter: apply, Esc: clear)") {
		t.Fatalf("missing filter prompt in view: %q", view)
	}
}

func TestBandForCoverage(t *testing.T) {
	if bandForCoverage(79.9, 80) != bandLow {
		t.Fatal("expected low band")
	}
	if bandForCoverage(80, 80) != bandMid {
		t.Fatal("expected mid band")
	}
	if bandForCoverage(89.9, 80) != bandMid {
		t.Fatal("expected mid band")
	}
	if bandForCoverage(90, 80) != bandHigh {
		t.Fatal("expected high band")
	}
}

func TestCursorMoveBounds(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{{Name: "a"}, {Name: "b"}},
	}
	m := NewModel(report, Config{Sort: "name"})
	m.applyKey("up")
	if m.current().cursor != 0 {
		t.Fatalf("cursor should stay at 0, got %d", m.current().cursor)
	}
	m.applyKey("down")
	if m.current().cursor != 1 {
		t.Fatalf("cursor should move to 1, got %d", m.current().cursor)
	}
	m.applyKey("down")
	if m.current().cursor != 1 {
		t.Fatalf("cursor should stay at max, got %d", m.current().cursor)
	}
}

func TestJumpKeys(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
	}
	m := NewModel(report, Config{Sort: "name"})

	m.applyKey("G")
	if m.current().cursor != 2 {
		t.Fatalf("G should move cursor to last row, got %d", m.current().cursor)
	}

	m.applyKey("g")
	if m.current().cursor != 0 {
		t.Fatalf("g should move cursor to first row, got %d", m.current().cursor)
	}
}

func TestQuitKeys(t *testing.T) {
	m := NewModel(sampleReport(), Config{Sort: "name"})
	if !m.applyKey("q") {
		t.Fatal("q should quit")
	}
	if !m.applyKey("ctrl+c") {
		t.Fatal("ctrl+c should quit")
	}
}

func TestRenderChildrenAlignsBarByLongestName(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{Name: "a", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 2, Covered: 8}}},
			{Name: "very/long/package/name", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9}}},
			{Name: "mid/name", Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 4, Covered: 6}}},
		},
	}
	m := NewModel(report, Config{Sort: "name", NoColor: true})

	view := m.renderChildren()
	lines := strings.Split(view, "\n")
	if len(lines) < 4 {
		t.Fatalf("unexpected children rendering: %q", view)
	}

	barCol := -1
	for _, line := range lines[1:] {
		byteCol := strings.IndexAny(line, "█░")
		if byteCol < 0 {
			t.Fatalf("bar not found in line: %q", line)
		}
		col := lipgloss.Width(line[:byteCol])
		if barCol == -1 {
			barCol = col
			continue
		}
		if col != barCol {
			t.Fatalf("bar column mismatch: want=%d got=%d line=%q", barCol, col, line)
		}
	}
}

func TestRenderSummaryFitsNarrowWidth(t *testing.T) {
	report := jacoco.Report{
		Counters: []jacoco.Counter{
			{Type: jacoco.CounterInstruction, Missed: 1, Covered: 9},
			{Type: jacoco.CounterBranch, Missed: 2, Covered: 8},
			{Type: jacoco.CounterLine, Missed: 3, Covered: 7},
			{Type: jacoco.CounterMethod, Missed: 4, Covered: 6},
		},
	}
	m := NewModel(report, Config{Sort: "name", NoColor: true})
	m.width = 40

	view := m.renderSummary()
	lines := strings.Split(view, "\n")
	for _, line := range lines[1:] {
		if lipgloss.Width(line) > m.width {
			t.Fatalf("summary line exceeds width: width=%d line=%q", m.width, line)
		}
	}
}

func TestRenderChildrenFitsNarrowWidth(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{
				Name: "very/long/package/name/for/narrow/terminal/view",
				Counters: []jacoco.Counter{
					{Type: jacoco.CounterInstruction, Missed: 2, Covered: 8},
				},
			},
		},
	}
	m := NewModel(report, Config{Sort: "name", NoColor: true})
	m.width = 36

	view := m.renderChildren()
	lines := strings.Split(view, "\n")
	for _, line := range lines[1:] {
		if lipgloss.Width(line) > m.width {
			t.Fatalf("children line exceeds width: width=%d line=%q", m.width, line)
		}
	}
}

func TestRenderChildrenEllipsizesLongNames(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{
				Name:     "com/jsptags/navigation/pager/parser/very/very/very/long/path/IndexTagExport",
				Counters: []jacoco.Counter{{Type: jacoco.CounterInstruction, Missed: 0, Covered: 10}},
			},
		},
	}
	m := NewModel(report, Config{Sort: "name", NoColor: true})
	m.width = 70

	view := m.renderChildren()
	lines := strings.Split(view, "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected children rendering: %q", view)
	}
	if strings.Contains(lines[1], "...") {
		t.Fatalf("expected spring-style segment abbreviation before ellipsis: %q", lines[1])
	}
	if !strings.Contains(lines[1], "c/j/") {
		t.Fatalf("expected abbreviated directory segments in line: %q", lines[1])
	}

	byteCol := strings.IndexAny(lines[1], "█░")
	if byteCol < 0 {
		t.Fatalf("bar not found in line: %q", lines[1])
	}
	displayCol := lipgloss.Width(lines[1][:byteCol])
	if displayCol >= m.width {
		t.Fatalf("line exceeds width budget: col=%d width=%d line=%q", displayCol, m.width, lines[1])
	}
}

func TestCompactNameForDisplayFallsBackToEllipsisWhenStillTooLong(t *testing.T) {
	name := "com/jsptags/navigation/pager/parser/very/very/very/long/path/IndexTagExport"
	got := compactNameForDisplay(name, 20)
	if !strings.Contains(got, "...") {
		t.Fatalf("expected ellipsis fallback: %q", got)
	}
	if lipgloss.Width(got) > 20 {
		t.Fatalf("expected width <= 20, got=%d, value=%q", lipgloss.Width(got), got)
	}
}

func TestViewNoColorDisablesANSISequences(t *testing.T) {
	m := NewModel(sampleReport(), Config{Threshold: 80, Sort: "name", NoColor: true})
	view := m.View()
	if hasANSI(view) {
		t.Fatalf("view should not include ANSI sequences when no-color is enabled: %q", view)
	}
}

func TestWatchProbeShowsConfirmationPrompt(t *testing.T) {
	m := newModel(sampleReport(), Config{Watch: false}, func() (jacoco.Report, error) {
		return sampleReport(), nil
	}, func() (bool, error) {
		return true, nil
	})

	next, _ := m.Update(watchProbeMsg{changed: true})
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("unexpected model type: %T", next)
	}
	if !updated.watchPrompt {
		t.Fatal("watch prompt should be enabled after change detection")
	}
	if !strings.Contains(updated.View(), "reload now?") {
		t.Fatal("view should include watch confirmation prompt")
	}
}

func TestWatchConfirmAcceptTriggersReload(t *testing.T) {
	updatedReport := sampleReport()
	updatedReport.Name = "updated"
	reloadCalled := false
	m := newModel(sampleReport(), Config{Watch: false}, func() (jacoco.Report, error) {
		reloadCalled = true
		return updatedReport, nil
	}, func() (bool, error) {
		return false, nil
	})
	m.watchPrompt = true

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	intermediate := next.(Model)
	if intermediate.watchPrompt {
		t.Fatal("watch prompt should be cleared after confirmation")
	}
	if cmd == nil {
		t.Fatal("reload command should be returned on confirmation")
	}

	msg := cmd()
	reloadMsg, ok := msg.(watchReloadMsg)
	if !ok {
		t.Fatalf("unexpected command message: %T", msg)
	}
	if reloadMsg.err != nil {
		t.Fatalf("reload should succeed: %v", reloadMsg.err)
	}
	if !reloadCalled {
		t.Fatal("reload function should be called")
	}

	next, _ = intermediate.Update(reloadMsg)
	final := next.(Model)
	if final.report.Name != "updated" {
		t.Fatalf("report should be replaced after reload: %s", final.report.Name)
	}
}

func TestWatchConfirmRejectSkipsReload(t *testing.T) {
	reloadCalled := false
	m := newModel(sampleReport(), Config{Watch: false}, func() (jacoco.Report, error) {
		reloadCalled = true
		return sampleReport(), nil
	}, nil)
	m.watchPrompt = true

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := next.(Model)
	if updated.watchPrompt {
		t.Fatal("watch prompt should be cleared on reject")
	}
	if cmd != nil {
		t.Fatal("reject should not schedule reload command")
	}
	if reloadCalled {
		t.Fatal("reload should not be called on reject")
	}
}

func TestWatchProbeErrorIsShown(t *testing.T) {
	m := newModel(sampleReport(), Config{Watch: false}, nil, nil)
	next, _ := m.Update(watchProbeMsg{err: errors.New("probe failed")})
	updated := next.(Model)
	if updated.watchErr != "probe failed" {
		t.Fatalf("unexpected watch error: %s", updated.watchErr)
	}
}

func TestWatchFlagAutoReloadsWithoutConfirmation(t *testing.T) {
	reloadCalled := false
	m := newModel(sampleReport(), Config{Watch: true}, func() (jacoco.Report, error) {
		reloadCalled = true
		r := sampleReport()
		r.Name = "reloaded"
		return r, nil
	}, func() (bool, error) {
		t.Fatal("probe should not be used in auto watch mode")
		return false, nil
	})

	next, cmd := m.Update(watchTickMsg{})
	updated := next.(Model)
	if updated.watchPrompt {
		t.Fatal("watch prompt should not appear in auto watch mode")
	}
	if cmd == nil {
		t.Fatal("watch tick should schedule reload command")
	}

	msg := cmd()
	switch typed := msg.(type) {
	case tea.BatchMsg:
		if len(typed) != 2 {
			t.Fatalf("unexpected batch length: %d", len(typed))
		}
		if typed[0] == nil {
			t.Fatal("first batch command should not be nil")
		}
		reloadMsg, ok := typed[0]().(watchReloadMsg)
		if !ok {
			t.Fatalf("first batch message should be watchReloadMsg")
		}
		if reloadMsg.err != nil {
			t.Fatalf("reload should succeed: %v", reloadMsg.err)
		}
	default:
		t.Fatalf("unexpected message type: %T", msg)
	}
	if !reloadCalled {
		t.Fatal("reload function should be called in auto watch mode")
	}
}
