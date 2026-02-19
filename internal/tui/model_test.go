package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/izuno4t/jacoco-report-viewer-cli/internal/jacoco"
)

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

	for _, want := range []string{"Report(demo)", "Summary", "Children", "com/example"} {
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

func TestRenderChildrenEllipsizesLongNames(t *testing.T) {
	report := jacoco.Report{
		Packages: []jacoco.Package{
			{
				Name: "com/jsptags/navigation/pager/parser/very/very/very/long/path/IndexTagExport",
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
