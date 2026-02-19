package jacoco

import (
	"strings"
	"testing"
)

func TestParseCoberturaBasicHierarchy(t *testing.T) {
	xmlText := `
<coverage>
  <packages>
    <package name="pkg.alpha">
      <classes>
        <class name="pkg.alpha.A" filename="pkg/alpha/A.py">
          <methods>
            <method name="foo" signature="()">
              <lines>
                <line number="10" hits="1" branch="false"/>
                <line number="11" hits="0" branch="false"/>
                <line number="12" hits="1" branch="true" condition-coverage="50% (1/2)"/>
              </lines>
            </method>
          </methods>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	report, err := ParseCobertura(strings.NewReader(xmlText))
	if err != nil {
		t.Fatalf("parse cobertura failed: %v", err)
	}
	if report.Name != "cobertura" {
		t.Fatalf("report name mismatch: %s", report.Name)
	}
	if len(report.Packages) != 1 {
		t.Fatalf("package count mismatch: %d", len(report.Packages))
	}

	pkg := report.Packages[0]
	if len(pkg.Classes) != 1 {
		t.Fatalf("class count mismatch: %d", len(pkg.Classes))
	}
	class := pkg.Classes[0]
	if len(class.Methods) != 1 {
		t.Fatalf("method count mismatch: %d", len(class.Methods))
	}

	mc, ok := class.Methods[0].Counter(CounterLine)
	if !ok || mc.Covered != 2 || mc.Missed != 1 {
		t.Fatalf("method line counter mismatch: %#v", mc)
	}
	bc, ok := class.Methods[0].Counter(CounterBranch)
	if !ok || bc.Covered != 1 || bc.Missed != 1 {
		t.Fatalf("method branch counter mismatch: %#v", bc)
	}
	ic, ok := class.Methods[0].Counter(CounterInstruction)
	if !ok || ic.Covered != 2 || ic.Missed != 1 {
		t.Fatalf("method instruction counter mismatch: %#v", ic)
	}
}

func TestParseCoberturaClassLinesFallback(t *testing.T) {
	xmlText := `
<coverage>
  <packages>
    <package name="pkg.beta">
      <classes>
        <class name="pkg.beta.B" filename="pkg/beta/B.py">
          <lines>
            <line number="1" hits="1" branch="false"/>
            <line number="2" hits="0" branch="false"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	report, err := ParseCobertura(strings.NewReader(xmlText))
	if err != nil {
		t.Fatalf("parse cobertura failed: %v", err)
	}
	if len(report.Packages) != 1 || len(report.Packages[0].Classes) != 1 {
		t.Fatalf("unexpected hierarchy: %#v", report)
	}

	class := report.Packages[0].Classes[0]
	c, ok := class.Counter(CounterLine)
	if !ok || c.Covered != 1 || c.Missed != 1 {
		t.Fatalf("class line counter mismatch: %#v", c)
	}
}

func TestParseCoberturaRejectsInvalidXML(t *testing.T) {
	_, err := ParseCobertura(strings.NewReader("<coverage"))
	if err == nil {
		t.Fatal("expected error for invalid cobertura xml")
	}
}
