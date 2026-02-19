package jacoco

import (
	"sort"
	"strconv"
)

// MergeReports merges multiple JaCoCo reports by package/class/method identity.
func MergeReports(reports ...Report) Report {
	if len(reports) == 0 {
		return Report{}
	}

	merged := Report{Name: reports[0].Name}
	pkgIndex := map[string]int{}

	for _, report := range reports {
		merged.Counters = mergeCounterSlices(merged.Counters, report.Counters)
		for _, pkg := range report.Packages {
			ix, ok := pkgIndex[pkg.Name]
			if !ok {
				pkgIndex[pkg.Name] = len(merged.Packages)
				merged.Packages = append(merged.Packages, Package{Name: pkg.Name})
				ix = len(merged.Packages) - 1
			}
			mergePackage(&merged.Packages[ix], pkg)
		}
	}

	sort.SliceStable(merged.Packages, func(i, j int) bool {
		return merged.Packages[i].Name < merged.Packages[j].Name
	})
	return merged
}

func mergePackage(dst *Package, src Package) {
	dst.Counters = mergeCounterSlices(dst.Counters, src.Counters)

	classIndex := map[string]int{}
	for i, class := range dst.Classes {
		classIndex[class.Name] = i
	}

	for _, class := range src.Classes {
		ix, ok := classIndex[class.Name]
		if !ok {
			dst.Classes = append(dst.Classes, Class{Name: class.Name, SourceFileName: class.SourceFileName})
			ix = len(dst.Classes) - 1
			classIndex[class.Name] = ix
		}
		mergeClass(&dst.Classes[ix], class)
	}

	sort.SliceStable(dst.Classes, func(i, j int) bool {
		return dst.Classes[i].Name < dst.Classes[j].Name
	})
}

func mergeClass(dst *Class, src Class) {
	if dst.SourceFileName == "" {
		dst.SourceFileName = src.SourceFileName
	}
	dst.Counters = mergeCounterSlices(dst.Counters, src.Counters)

	methodIndex := map[string]int{}
	for i, method := range dst.Methods {
		methodIndex[methodMergeKey(method)] = i
	}

	for _, method := range src.Methods {
		key := methodMergeKey(method)
		ix, ok := methodIndex[key]
		if !ok {
			dst.Methods = append(dst.Methods, Method{Name: method.Name, Desc: method.Desc, Line: method.Line})
			ix = len(dst.Methods) - 1
			methodIndex[key] = ix
		}
		dst.Methods[ix].Counters = mergeCounterSlices(dst.Methods[ix].Counters, method.Counters)
	}

	sort.SliceStable(dst.Methods, func(i, j int) bool {
		return methodMergeKey(dst.Methods[i]) < methodMergeKey(dst.Methods[j])
	})
}

func mergeCounterSlices(dst []Counter, src []Counter) []Counter {
	if len(src) == 0 {
		return dst
	}
	index := map[CounterType]int{}
	for i, c := range dst {
		index[c.Type] = i
	}
	for _, c := range src {
		if ix, ok := index[c.Type]; ok {
			dst[ix].Missed += c.Missed
			dst[ix].Covered += c.Covered
			continue
		}
		index[c.Type] = len(dst)
		dst = append(dst, Counter{Type: c.Type, Missed: c.Missed, Covered: c.Covered})
	}
	sort.SliceStable(dst, func(i, j int) bool {
		return dst[i].Type < dst[j].Type
	})
	return dst
}

func methodMergeKey(m Method) string {
	return m.Name + "\x00" + m.Desc + "\x00" + strconv.Itoa(m.Line)
}
