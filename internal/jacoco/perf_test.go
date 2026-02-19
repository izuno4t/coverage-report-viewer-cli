package jacoco

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParsePerformance1000Classes(t *testing.T) {
	var b strings.Builder
	b.WriteString("<report name=\"perf\"><package name=\"com/example\">")
	for i := range 1000 {
		fmt.Fprintf(&b, "<class name=\"com/example/C%d\" sourcefilename=\"C%d.java\">", i, i)
		b.WriteString("<method name=\"m\" desc=\"()V\" line=\"1\">")
		b.WriteString("<counter type=\"INSTRUCTION\" missed=\"1\" covered=\"9\"/>")
		b.WriteString("</method>")
		b.WriteString("</class>")
	}
	b.WriteString("</package></report>")

	start := time.Now()
	report, err := Parse(strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	elapsed := time.Since(start)
	t.Logf("elapsed=%s", elapsed)

	if len(report.Packages) != 1 {
		t.Fatalf("unexpected package count: %d", len(report.Packages))
	}
	if got := len(report.Packages[0].Classes); got != 1000 {
		t.Fatalf("unexpected class count: %d", got)
	}
	if elapsed >= time.Second {
		t.Fatalf("parse took too long: %s", elapsed)
	}
}
