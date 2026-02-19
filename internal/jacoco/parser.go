package jacoco

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

func ParseFile(path string) (Report, error) {
	f, err := os.Open(path)
	if err != nil {
		return Report{}, fmt.Errorf("open report: %w", err)
	}
	defer f.Close()

	return Parse(f)
}

func Parse(r io.Reader) (Report, error) {
	var xr xmlReport
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&xr); err != nil {
		return Report{}, fmt.Errorf("decode xml: %w", err)
	}

	report := Report{Name: xr.Name}

	counters, err := decodeCounters(xr.Counters)
	if err != nil {
		return Report{}, err
	}
	report.Counters = counters

	for _, xp := range xr.Packages {
		pkg := Package{Name: xp.Name}
		pkg.Counters, err = decodeCounters(xp.Counters)
		if err != nil {
			return Report{}, err
		}

		for _, xc := range xp.Classes {
			class := Class{Name: xc.Name, SourceFileName: xc.SourceFileName}
			class.Counters, err = decodeCounters(xc.Counters)
			if err != nil {
				return Report{}, err
			}

			for _, xm := range xc.Methods {
				method := Method{Name: xm.Name, Desc: xm.Desc, Line: xm.Line}
				method.Counters, err = decodeCounters(xm.Counters)
				if err != nil {
					return Report{}, err
				}
				class.Methods = append(class.Methods, method)
			}

			if len(class.Counters) == 0 {
				class.Counters = sumMethodCounters(class.Methods)
			}

			pkg.Classes = append(pkg.Classes, class)
		}

		if len(pkg.Counters) == 0 {
			pkg.Counters = sumClassCounters(pkg.Classes)
		}

		report.Packages = append(report.Packages, pkg)
	}

	if len(report.Counters) == 0 {
		report.Counters = sumPackageCounters(report.Packages)
	}

	return report, nil
}

func decodeCounters(raw []xmlCounter) ([]Counter, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]Counter, 0, len(raw))
	for _, rc := range raw {
		t, err := normalizeCounterType(rc.Type)
		if err != nil {
			return nil, err
		}
		out = append(out, Counter{Type: t, Missed: rc.Missed, Covered: rc.Covered})
	}
	return out, nil
}

func sumMethodCounters(methods []Method) []Counter {
	agg := map[CounterType]Counter{}
	for _, m := range methods {
		mergeCounters(agg, m.Counters)
	}
	return mapToCounters(agg)
}

func sumClassCounters(classes []Class) []Counter {
	agg := map[CounterType]Counter{}
	for _, c := range classes {
		mergeCounters(agg, c.Counters)
	}
	return mapToCounters(agg)
}

func sumPackageCounters(pkgs []Package) []Counter {
	agg := map[CounterType]Counter{}
	for _, p := range pkgs {
		mergeCounters(agg, p.Counters)
	}
	return mapToCounters(agg)
}

func mergeCounters(agg map[CounterType]Counter, counters []Counter) {
	for _, c := range counters {
		base := agg[c.Type]
		base.Type = c.Type
		base.Missed += c.Missed
		base.Covered += c.Covered
		agg[c.Type] = base
	}
}

func mapToCounters(agg map[CounterType]Counter) []Counter {
	out := make([]Counter, 0, len(allCounterTypes))
	for _, t := range allCounterTypes {
		if c, ok := agg[t]; ok {
			out = append(out, c)
		}
	}
	return out
}
