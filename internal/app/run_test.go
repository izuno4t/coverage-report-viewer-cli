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

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	called := false
	startUIWatch = func(report jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
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
		t.Fatal("startUIWatch should be called")
	}
}

func TestRunAutoDetectsAndMergesMultiModuleReports(t *testing.T) {
	dir := t.TempDir()
	rootPom := `<project><modules><module>module-a</module><module>module-b</module></modules></project>`
	modulePom := `<project><build><plugins><plugin><groupId>org.jacoco</groupId><artifactId>jacoco-maven-plugin</artifactId></plugin></plugins></build></project>`
	write := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}
	write(filepath.Join(dir, "pom.xml"), rootPom)
	write(filepath.Join(dir, "module-a/pom.xml"), modulePom)
	write(filepath.Join(dir, "module-b/pom.xml"), modulePom)
	write(filepath.Join(dir, "module-a/target/site/jacoco/jacoco.xml"), `<report name="a"><package name="pkg.a"><class name="A"><method name="f" desc="()V"><counter type="INSTRUCTION" missed="1" covered="9"/></method><counter type="INSTRUCTION" missed="1" covered="9"/></class><counter type="INSTRUCTION" missed="1" covered="9"/></package><counter type="INSTRUCTION" missed="1" covered="9"/></report>`)
	write(filepath.Join(dir, "module-b/target/site/jacoco/jacoco.xml"), `<report name="b"><package name="pkg.b"><class name="B"><method name="g" desc="()V"><counter type="INSTRUCTION" missed="2" covered="8"/></method><counter type="INSTRUCTION" missed="2" covered="8"/></class><counter type="INSTRUCTION" missed="2" covered="8"/></package><counter type="INSTRUCTION" missed="2" covered="8"/></report>`)

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

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	startUIWatch = func(report jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
		if len(report.Packages) != 2 {
			t.Fatalf("expected merged package count=2, got=%d", len(report.Packages))
		}
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d (stderr=%q)", code, errOut.String())
	}
}

func TestRunFailsWhenUIFails(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "sample.xml")
	if err := os.WriteFile(reportPath, []byte("<report name=\"x\"/>"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	startUIWatch = func(_ jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
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

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	startUIWatch = func(_ jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
		t.Fatal("startUIWatch should not be called on parse error")
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

func TestRunWatchModeUsesWatchUI(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "sample.xml")
	if err := os.WriteFile(reportPath, []byte("<report name=\"x\"/>"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	called := false
	startUIWatch = func(report jacoco.Report, cfg tui.Config, reloadFn func() (jacoco.Report, error), probeFn func() (bool, error)) error {
		called = true
		if !cfg.Watch {
			t.Fatal("watch config should be true")
		}
		changed, err := probeFn()
		if err != nil {
			t.Fatalf("probe failed: %v", err)
		}
		if changed {
			t.Fatal("probe should not report change without file update")
		}
		r, err := reloadFn()
		if err != nil {
			t.Fatalf("reload failed: %v", err)
		}
		if r.Name != report.Name {
			t.Fatalf("reload report mismatch: got=%s want=%s", r.Name, report.Name)
		}
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--watch", reportPath}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d (stderr=%q)", code, errOut.String())
	}
	if !called {
		t.Fatal("watch UI should be called")
	}
}

func TestRunWithFormatCobertura(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "coverage.xml")
	content := `<coverage><packages><package name="pkg"><classes><class name="pkg.A" filename="pkg/A.py"><methods><method name="f" signature="()"><lines><line number="1" hits="1" branch="false"/></lines></method></methods></class></classes></package></packages></coverage>`
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	called := false
	startUIWatch = func(report jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
		called = true
		if len(report.Packages) != 1 {
			t.Fatalf("unexpected package count: %d", len(report.Packages))
		}
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--format", "cobertura", reportPath}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d (stderr=%q)", code, errOut.String())
	}
	if !called {
		t.Fatal("startUIWatch should be called")
	}
}

func TestRunWithFormatLCOV(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "coverage.info")
	content := "TN:\nSF:src/main.py\nDA:1,1\nend_of_record\n"
	if err := os.WriteFile(reportPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	origStartUIWatch := startUIWatch
	t.Cleanup(func() {
		startUIWatch = origStartUIWatch
	})
	called := false
	startUIWatch = func(report jacoco.Report, _ tui.Config, _ func() (jacoco.Report, error), _ func() (bool, error)) error {
		called = true
		if report.Name != "lcov" {
			t.Fatalf("unexpected report name: %s", report.Name)
		}
		return nil
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"--format", "lcov", reportPath}, "dev", &out, &errOut)
	if code != 0 {
		t.Fatalf("expected 0, got %d (stderr=%q)", code, errOut.String())
	}
	if !called {
		t.Fatal("startUIWatch should be called")
	}
}
