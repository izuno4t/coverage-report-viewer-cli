package jacoco

import "testing"

func TestMergeReportsAggregatesCountersByHierarchy(t *testing.T) {
	r1 := Report{
		Name:     "a",
		Counters: []Counter{{Type: CounterInstruction, Missed: 1, Covered: 9}},
		Packages: []Package{{
			Name:     "com/example",
			Counters: []Counter{{Type: CounterInstruction, Missed: 2, Covered: 8}},
			Classes: []Class{{
				Name:     "UserService",
				Counters: []Counter{{Type: CounterInstruction, Missed: 3, Covered: 7}},
				Methods: []Method{{
					Name:     "find",
					Desc:     "()V",
					Counters: []Counter{{Type: CounterInstruction, Missed: 4, Covered: 6}},
				}},
			}},
		}},
	}
	r2 := Report{
		Name:     "b",
		Counters: []Counter{{Type: CounterInstruction, Missed: 10, Covered: 90}},
		Packages: []Package{{
			Name:     "com/example",
			Counters: []Counter{{Type: CounterInstruction, Missed: 20, Covered: 80}},
			Classes: []Class{{
				Name:     "UserService",
				Counters: []Counter{{Type: CounterInstruction, Missed: 30, Covered: 70}},
				Methods: []Method{{
					Name:     "find",
					Desc:     "()V",
					Counters: []Counter{{Type: CounterInstruction, Missed: 40, Covered: 60}},
				}},
			}},
		}},
	}

	merged := MergeReports(r1, r2)

	root, ok := merged.Counter(CounterInstruction)
	if !ok || root.Missed != 11 || root.Covered != 99 {
		t.Fatalf("root counter mismatch: %#v", root)
	}
	if len(merged.Packages) != 1 {
		t.Fatalf("package count mismatch: %d", len(merged.Packages))
	}

	pkg := merged.Packages[0]
	pc, ok := pkg.Counter(CounterInstruction)
	if !ok || pc.Missed != 22 || pc.Covered != 88 {
		t.Fatalf("package counter mismatch: %#v", pc)
	}
	if len(pkg.Classes) != 1 {
		t.Fatalf("class count mismatch: %d", len(pkg.Classes))
	}

	class := pkg.Classes[0]
	cc, ok := class.Counter(CounterInstruction)
	if !ok || cc.Missed != 33 || cc.Covered != 77 {
		t.Fatalf("class counter mismatch: %#v", cc)
	}
	if len(class.Methods) != 1 {
		t.Fatalf("method count mismatch: %d", len(class.Methods))
	}
	mc, ok := class.Methods[0].Counter(CounterInstruction)
	if !ok || mc.Missed != 44 || mc.Covered != 66 {
		t.Fatalf("method counter mismatch: %#v", mc)
	}
}

func TestMergeReportsKeepsDistinctNodes(t *testing.T) {
	r1 := Report{Packages: []Package{{Name: "a"}}}
	r2 := Report{Packages: []Package{{Name: "b"}}}

	merged := MergeReports(r2, r1)
	if len(merged.Packages) != 2 {
		t.Fatalf("package count mismatch: %d", len(merged.Packages))
	}
	if merged.Packages[0].Name != "a" || merged.Packages[1].Name != "b" {
		t.Fatalf("unexpected package ordering: %#v", merged.Packages)
	}
}
