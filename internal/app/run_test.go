package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/izuno4t/coverage-report-viewer-cli/internal/jacoco"
	"github.com/izuno4t/coverage-report-viewer-cli/internal/tui"
)

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--version"}, "1.2.3", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if out.String() != "crv 1.2.3\n" {
		t.Fatalf("unexpected version output: %q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--help"}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if out.Len() == 0 {
		t.Fatal("help output should not be empty")
	}
}

func TestRunInvalidArgs(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--sort", "invalid", "a.xml"}, "dev", &out, &errOut)
	if code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
	if errOut.Len() == 0 {
		t.Fatal("error output should not be empty")
	}
}

func TestRunReturnsErrorWhenReportMissing(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{}, "dev", &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestRunAutoDetectsReportPath(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "target/site/jacoco/jacoco.xml")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte("<report name=\"x\"/>"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	origStartUI := startUI
	t.Cleanup(func() {
		startUI = origStartUI
	})
	called := false
	startUI = func(report jacoco.Report, _ tui.Config) error {
		called = true
		if report.Name != "x" {
			t.Fatalf("unexpected report name: %s", report.Name)
		}
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d (stderr=%q)", code, errOut.String())
	}
	if !called {
		t.Fatal("startUI should be called")
	}
}

func TestRunFailsWhenUIFails(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "sample.xml")
	if err := os.WriteFile(reportPath, []byte("<report name=\"x\"/>"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUI := startUI
	t.Cleanup(func() {
		startUI = origStartUI
	})
	startUI = func(_ jacoco.Report, _ tui.Config) error {
		return errors.New("boom")
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{reportPath}, "dev", &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestRunFailsOnInvalidXML(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "invalid.xml")
	if err := os.WriteFile(reportPath, []byte("<report"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUI := startUI
	t.Cleanup(func() {
		startUI = origStartUI
	})
	startUI = func(_ jacoco.Report, _ tui.Config) error {
		t.Fatal("startUI should not be called on parse error")
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{reportPath}, "dev", &out, &errOut)
	if code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
	if errOut.Len() == 0 {
		t.Fatal("error output should not be empty")
	}
}
