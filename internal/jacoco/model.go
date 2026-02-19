package jacoco

import "fmt"

// CounterType is a JaCoCo coverage counter category.
type CounterType string

const (
	CounterInstruction CounterType = "INSTRUCTION"
	CounterBranch      CounterType = "BRANCH"
	CounterLine        CounterType = "LINE"
	CounterComplexity  CounterType = "COMPLEXITY"
	CounterMethod      CounterType = "METHOD"
	CounterClass       CounterType = "CLASS"
)

var allCounterTypes = []CounterType{
	CounterInstruction,
	CounterBranch,
	CounterLine,
	CounterComplexity,
	CounterMethod,
	CounterClass,
}

// Counter holds missed/covered metrics.
type Counter struct {
	Type    CounterType
	Missed  int
	Covered int
}

func (c Counter) Total() int {
	return c.Missed + c.Covered
}

func (c Counter) CoverageRate() float64 {
	total := c.Total()
	if total == 0 {
		return 0
	}
	return float64(c.Covered) / float64(total) * 100
}

// Method corresponds to a JaCoCo method node.
type Method struct {
	Name     string
	Desc     string
	Line     int
	Counters []Counter
}

// Class corresponds to a JaCoCo class node.
type Class struct {
	Name           string
	SourceFileName string
	Methods        []Method
	Counters       []Counter
}

// Package corresponds to a JaCoCo package node.
type Package struct {
	Name     string
	Classes  []Class
	Counters []Counter
}

// Report is the root JaCoCo model.
type Report struct {
	Name     string
	Packages []Package
	Counters []Counter
}

func (m Method) Counter(t CounterType) (Counter, bool) {
	return findCounter(m.Counters, t)
}

func (c Class) Counter(t CounterType) (Counter, bool) {
	return findCounter(c.Counters, t)
}

func (p Package) Counter(t CounterType) (Counter, bool) {
	return findCounter(p.Counters, t)
}

func (r Report) Counter(t CounterType) (Counter, bool) {
	return findCounter(r.Counters, t)
}

func findCounter(counters []Counter, t CounterType) (Counter, bool) {
	for _, c := range counters {
		if c.Type == t {
			return c, true
		}
	}
	return Counter{}, false
}

func normalizeCounterType(raw string) (CounterType, error) {
	t := CounterType(raw)
	for _, supported := range allCounterTypes {
		if t == supported {
			return t, nil
		}
	}
	return "", fmt.Errorf("unsupported counter type: %s", raw)
}
