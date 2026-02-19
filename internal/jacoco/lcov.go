package jacoco

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type lcovMethod struct {
	name string
	line int
	hits int
}

type lcovRecord struct {
	sourcePath string
	lines      map[int]int
	branches   [][2]int // covered, missed contribution per BRDA
	methods    map[string]lcovMethod
}

func ParseLCOVFile(path string) (Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("open lcov report: %w", err)
	}
	defer f.Close()
	return ParseLCOV(f)
}

func ParseLCOV(r io.Reader) (Report, error) {
	scanner := bufio.NewScanner(r)
	records := make([]lcovRecord, 0)
	current := newLCOVRecord()
	inRecord := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "SF:"):
			if inRecord && current.sourcePath != "" {
				records = append(records, current)
			}
			current = newLCOVRecord()
			inRecord = true
			current.sourcePath = strings.TrimSpace(strings.TrimPrefix(line, "SF:"))
		case strings.HasPrefix(line, "DA:"):
			if !inRecord {
				continue
			}
			ln, hits, ok := parseLCOVDA(line)
			if ok {
				current.lines[ln] = hits
			}
		case strings.HasPrefix(line, "FN:"):
			if !inRecord {
				continue
			}
			name, lineNum, ok := parseLCOVFN(line)
			if ok {
				m := current.methods[name]
				m.name = name
				m.line = lineNum
				current.methods[name] = m
			}
		case strings.HasPrefix(line, "FNDA:"):
			if !inRecord {
				continue
			}
			name, hits, ok := parseLCOVFNDA(line)
			if ok {
				m := current.methods[name]
				m.name = name
				m.hits = hits
				current.methods[name] = m
			}
		case strings.HasPrefix(line, "BRDA:"):
			if !inRecord {
				continue
			}
			covered, missed, ok := parseLCOVBRDA(line)
			if ok {
				current.branches = append(current.branches, [2]int{covered, missed})
			}
		case line == "end_of_record":
			if inRecord && current.sourcePath != "" {
				records = append(records, current)
			}
			current = newLCOVRecord()
			inRecord = false
		}
	}
	if err := scanner.Err(); err != nil {
		return Report{}, fmt.Errorf("scan lcov: %w", err)
	}
	if inRecord && current.sourcePath != "" {
		records = append(records, current)
	}
	if len(records) == 0 {
		return Report{}, fmt.Errorf("lcov record not found")
	}

	report := Report{Name: "lcov"}
	pkgIndex := map[string]int{}

	for _, rec := range records {
		pkgName, className := normalizeLCOVNames(rec.sourcePath)
		ix, ok := pkgIndex[pkgName]
		if !ok {
			pkgIndex[pkgName] = len(report.Packages)
			report.Packages = append(report.Packages, Package{Name: pkgName})
			ix = len(report.Packages) - 1
		}

		class := lcovRecordToClass(className, rec.sourcePath, rec)
		report.Packages[ix].Classes = append(report.Packages[ix].Classes, class)
	}

	for i := range report.Packages {
		report.Packages[i].Counters = sumClassCounters(report.Packages[i].Classes)
		sort.SliceStable(report.Packages[i].Classes, func(a, b int) bool {
			return report.Packages[i].Classes[a].Name < report.Packages[i].Classes[b].Name
		})
	}
	report.Counters = sumPackageCounters(report.Packages)
	sort.SliceStable(report.Packages, func(i, j int) bool {
		return report.Packages[i].Name < report.Packages[j].Name
	})

	return report, nil
}

func newLCOVRecord() lcovRecord {
	return lcovRecord{
		lines:   map[int]int{},
		methods: map[string]lcovMethod{},
	}
}

func parseLCOVDA(line string) (lineNo int, hits int, ok bool) {
	parts := strings.Split(strings.TrimPrefix(line, "DA:"), ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	ln, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return ln, h, true
}

func parseLCOVFN(line string) (name string, lineNo int, ok bool) {
	parts := strings.SplitN(strings.TrimPrefix(line, "FN:"), ",", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	ln, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return "", 0, false
	}
	return strings.TrimSpace(parts[1]), ln, true
}

func parseLCOVFNDA(line string) (name string, hits int, ok bool) {
	parts := strings.SplitN(strings.TrimPrefix(line, "FNDA:"), ",", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return "", 0, false
	}
	return strings.TrimSpace(parts[1]), h, true
}

func parseLCOVBRDA(line string) (covered int, missed int, ok bool) {
	parts := strings.Split(strings.TrimPrefix(line, "BRDA:"), ",")
	if len(parts) != 4 {
		return 0, 0, false
	}
	taken := strings.TrimSpace(parts[3])
	if taken == "-" {
		return 0, 1, true
	}
	v, err := strconv.Atoi(taken)
	if err != nil {
		return 0, 0, false
	}
	if v > 0 {
		return 1, 0, true
	}
	return 0, 1, true
}

func normalizeLCOVNames(sourcePath string) (pkg string, class string) {
	cleaned := filepath.ToSlash(strings.TrimSpace(sourcePath))
	if cleaned == "" {
		return "default", "unknown"
	}
	dir := filepath.Dir(cleaned)
	if dir == "." || dir == "" {
		dir = "default"
	}
	base := filepath.Base(cleaned)
	if base == "" || base == "." || base == "/" {
		base = cleaned
	}
	return dir, base
}

func lcovRecordToClass(className, sourcePath string, rec lcovRecord) Class {
	lineCounter := Counter{Type: CounterLine}
	for _, hits := range rec.lines {
		if hits > 0 {
			lineCounter.Covered++
		} else {
			lineCounter.Missed++
		}
	}

	branchCounter := Counter{Type: CounterBranch}
	for _, b := range rec.branches {
		branchCounter.Covered += b[0]
		branchCounter.Missed += b[1]
	}

	methods := make([]Method, 0, len(rec.methods))
	for _, m := range rec.methods {
		counter := Counter{Type: CounterInstruction}
		if m.hits > 0 {
			counter.Covered = 1
		} else {
			counter.Missed = 1
		}
		methods = append(methods, Method{
			Name:     m.name,
			Line:     m.line,
			Counters: []Counter{counter},
		})
	}
	sort.SliceStable(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	counters := make([]Counter, 0, 3)
	counters = append(counters, Counter{Type: CounterInstruction, Missed: lineCounter.Missed, Covered: lineCounter.Covered})
	counters = append(counters, lineCounter)
	if branchCounter.Total() > 0 {
		counters = append(counters, branchCounter)
	}

	return Class{
		Name:           className,
		SourceFileName: filepath.ToSlash(sourcePath),
		Methods:        methods,
		Counters:       counters,
	}
}
