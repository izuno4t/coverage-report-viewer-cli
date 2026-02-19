package jacoco

import (
	"strings"
	"testing"
)

func TestParseBasicHierarchy(t *testing.T) {
	xmlText := `
<report name="demo">
  <package name="com/example">
    <class name="com/example/UserService" sourcefilename="UserService.java">
      <method name="find" desc="()V" line="10">
        <counter type="INSTRUCTION" missed="2" covered="8"/>
        <counter type="LINE" missed="1" covered="4"/>
      </method>
      <counter type="INSTRUCTION" missed="2" covered="8"/>
      <counter type="LINE" missed="1" covered="4"/>
    </class>
    <counter type="INSTRUCTION" missed="2" covered="8"/>
    <counter type="LINE" missed="1" covered="4"/>
  </package>
  <counter type="INSTRUCTION" missed="2" covered="8"/>
  <counter type="LINE" missed="1" covered="4"/>
</report>`

	report, err := Parse(strings.NewReader(xmlText))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if report.Name != "demo" {
		t.Fatalf("report name mismatch: %s", report.Name)
	}
	if len(report.Packages) != 1 {
		t.Fatalf("package count mismatch: %d", len(report.Packages))
	}
	pkg := report.Packages[0]
	if pkg.Name != "com/example" {
		t.Fatalf("package name mismatch: %s", pkg.Name)
	}
	if len(pkg.Classes) != 1 {
		t.Fatalf("class count mismatch: %d", len(pkg.Classes))
	}
	class := pkg.Classes[0]
	if len(class.Methods) != 1 {
		t.Fatalf("method count mismatch: %d", len(class.Methods))
	}

	counter, ok := class.Counter(CounterInstruction)
	if !ok {
		t.Fatal("instruction counter missing")
	}
	if counter.CoverageRate() != 80 {
		t.Fatalf("coverage rate mismatch: %v", counter.CoverageRate())
	}
}

func TestParseAggregatesWhenMissingUpperCounters(t *testing.T) {
	xmlText := `
<report name="demo">
  <package name="com/example">
    <class name="com/example/UserService" sourcefilename="UserService.java">
      <method name="a" desc="()V" line="10">
        <counter type="INSTRUCTION" missed="1" covered="3"/>
      </method>
      <method name="b" desc="()V" line="20">
        <counter type="INSTRUCTION" missed="2" covered="4"/>
      </method>
    </class>
  </package>
</report>`

	report, err := Parse(strings.NewReader(xmlText))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	rc, ok := report.Counter(CounterInstruction)
	if !ok {
		t.Fatal("report instruction counter missing")
	}
	if rc.Missed != 3 || rc.Covered != 7 {
		t.Fatalf("report aggregate mismatch: missed=%d covered=%d", rc.Missed, rc.Covered)
	}

	pc, ok := report.Packages[0].Counter(CounterInstruction)
	if !ok {
		t.Fatal("package instruction counter missing")
	}
	if pc.Missed != 3 || pc.Covered != 7 {
		t.Fatalf("package aggregate mismatch: missed=%d covered=%d", pc.Missed, pc.Covered)
	}
}

func TestParseRejectsUnknownCounterType(t *testing.T) {
	xmlText := `
<report name="demo">
  <counter type="UNKNOWN" missed="1" covered="1"/>
</report>`

	_, err := Parse(strings.NewReader(xmlText))
	if err == nil {
		t.Fatal("expected error for unknown counter type")
	}
}
