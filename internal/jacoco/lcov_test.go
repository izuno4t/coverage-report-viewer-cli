package jacoco

import (
	"strings"
	"testing"
)

func TestParseLCOVBasic(t *testing.T) {
	text := `
TN:
SF:src/pkg/foo.py
FN:10,foo
FNDA:3,foo
DA:10,3
DA:11,0
BRDA:12,0,0,1
BRDA:12,0,1,-
end_of_record
`

	report, err := ParseLCOV(strings.NewReader(text))
	if err != nil {
		t.Fatalf("parse lcov failed: %v", err)
	}
	if report.Name != "lcov" {
		t.Fatalf("report name mismatch: %s", report.Name)
	}
	if len(report.Packages) != 1 {
		t.Fatalf("package count mismatch: %d", len(report.Packages))
	}
	pkg := report.Packages[0]
	if pkg.Name != "src/pkg" {
		t.Fatalf("package name mismatch: %s", pkg.Name)
	}
	if len(pkg.Classes) != 1 {
		t.Fatalf("class count mismatch: %d", len(pkg.Classes))
	}
	class := pkg.Classes[0]
	if class.Name != "foo.py" {
		t.Fatalf("class name mismatch: %s", class.Name)
	}
	lineCounter, ok := class.Counter(CounterLine)
	if !ok || lineCounter.Covered != 1 || lineCounter.Missed != 1 {
		t.Fatalf("line counter mismatch: %#v", lineCounter)
	}
	branchCounter, ok := class.Counter(CounterBranch)
	if !ok || branchCounter.Covered != 1 || branchCounter.Missed != 1 {
		t.Fatalf("branch counter mismatch: %#v", branchCounter)
	}
	if len(class.Methods) != 1 || class.Methods[0].Name != "foo" {
		t.Fatalf("method mismatch: %#v", class.Methods)
	}
}

func TestParseLCOVRejectsEmptyInput(t *testing.T) {
	_, err := ParseLCOV(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error")
	}
}
