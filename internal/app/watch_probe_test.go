package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReportUpdateProbeDetectsFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jacoco.xml")
	if err := os.WriteFile(path, []byte("<report name=\"a\"/>"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	probe, err := newReportUpdateProbe([]string{path})
	if err != nil {
		t.Fatalf("newReportUpdateProbe failed: %v", err)
	}

	changed, err := probe()
	if err != nil {
		t.Fatalf("probe failed: %v", err)
	}
	if changed {
		t.Fatal("first probe should not report change")
	}

	time.Sleep(1100 * time.Millisecond)
	if err := os.WriteFile(path, []byte("<report name=\"b\"/>"), 0o644); err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}

	changed, err = probe()
	if err != nil {
		t.Fatalf("probe after update failed: %v", err)
	}
	if !changed {
		t.Fatal("probe should report change after file update")
	}
}

func TestReportUpdateProbeReturnsErrorWhenFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.xml")
	_, err := newReportUpdateProbe([]string{path})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
