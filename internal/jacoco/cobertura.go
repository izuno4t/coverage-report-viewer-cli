package jacoco

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

var coberturaConditionRe = regexp.MustCompile(`\((\d+)/(\d+)\)`)

func ParseCoberturaFile(path string) (Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("open cobertura report: %w", err)
	}
	defer f.Close()
	return ParseCobertura(f)
}

func ParseCobertura(r io.Reader) (Report, error) {
	var xc xmlCoberturaCoverage
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&xc); err != nil {
		return Report{}, fmt.Errorf("decode cobertura xml: %w", err)
	}

	report := Report{Name: "cobertura"}
	for _, xp := range xc.Packages {
		pkg := Package{Name: xp.Name}
		for _, xclass := range xp.Classes {
			class := Class{Name: xclass.Name, SourceFileName: xclass.File}

			for _, xm := range xclass.Methods {
				lineCounter, branchCounter := countersFromCoberturaLines(xm.Lines)
				method := Method{
					Name:     xm.Name,
					Desc:     xm.Signature,
					Line:     firstLineNumber(xm.Lines),
					Counters: normalizeCoberturaCounters(lineCounter, branchCounter),
				}
				class.Methods = append(class.Methods, method)
			}

			if len(class.Methods) > 0 {
				class.Counters = sumMethodCounters(class.Methods)
			} else {
				lineCounter, branchCounter := countersFromCoberturaLines(xclass.Lines)
				class.Counters = normalizeCoberturaCounters(lineCounter, branchCounter)
			}

			pkg.Classes = append(pkg.Classes, class)
		}

		pkg.Counters = sumClassCounters(pkg.Classes)
		report.Packages = append(report.Packages, pkg)
	}

	report.Counters = sumPackageCounters(report.Packages)
	return report, nil
}

func firstLineNumber(lines []xmlCoberturaLineNode) int {
	if len(lines) == 0 {
		return 0
	}
	min := lines[0].Number
	for _, line := range lines[1:] {
		if line.Number < min {
			min = line.Number
		}
	}
	return min
}

func countersFromCoberturaLines(lines []xmlCoberturaLineNode) (Counter, Counter) {
	lineCounter := Counter{Type: CounterLine}
	branchCounter := Counter{Type: CounterBranch}

	for _, line := range lines {
		if line.Hits > 0 {
			lineCounter.Covered++
		} else {
			lineCounter.Missed++
		}

		if line.Branch == "true" {
			covered, missed := parseCoberturaBranchCoverage(line)
			branchCounter.Covered += covered
			branchCounter.Missed += missed
		}
	}

	return lineCounter, branchCounter
}

func parseCoberturaBranchCoverage(line xmlCoberturaLineNode) (covered int, missed int) {
	matches := coberturaConditionRe.FindStringSubmatch(line.ConditionCoverage)
	if len(matches) == 3 {
		c, err1 := strconv.Atoi(matches[1])
		t, err2 := strconv.Atoi(matches[2])
		if err1 == nil && err2 == nil && t >= c {
			return c, t - c
		}
	}

	if line.Hits > 0 {
		return 1, 0
	}
	return 0, 1
}

func normalizeCoberturaCounters(lineCounter, branchCounter Counter) []Counter {
	counters := make([]Counter, 0, 3)
	counters = append(counters, Counter{Type: CounterInstruction, Missed: lineCounter.Missed, Covered: lineCounter.Covered})
	counters = append(counters, lineCounter)
	if branchCounter.Total() > 0 {
		counters = append(counters, branchCounter)
	}
	return counters
}
