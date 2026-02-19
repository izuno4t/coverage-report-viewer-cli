package reportpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFromPOMPluginConfig(t *testing.T) {
	dir := t.TempDir()

	pom := `<project><build><plugins><plugin><groupId>org.jacoco</groupId><artifactId>jacoco-maven-plugin</artifactId><configuration><outputDirectory>target/custom-jacoco</outputDirectory></configuration></plugin></plugins></build></project>`
	writeFile(t, filepath.Join(dir, "pom.xml"), pom)
	writeFile(t, filepath.Join(dir, "target/custom-jacoco/jacoco.xml"), "<report name=\"x\"/>")

	path, err := Detect(dir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	want := filepath.Join(dir, "target/custom-jacoco/jacoco.xml")
	if path != want {
		t.Fatalf("path mismatch: got=%s want=%s", path, want)
	}
}

func TestDetectFromPOMExecutionReportGoal(t *testing.T) {
	dir := t.TempDir()

	pom := `<project><build><plugins><plugin><groupId>org.jacoco</groupId><artifactId>jacoco-maven-plugin</artifactId><executions><execution><goals><goal>report</goal></goals><configuration><outputDirectory>target/site/exec-jacoco</outputDirectory></configuration></execution></executions></plugin></plugins></build></project>`
	writeFile(t, filepath.Join(dir, "pom.xml"), pom)
	writeFile(t, filepath.Join(dir, "target/site/exec-jacoco/jacoco.xml"), "<report name=\"x\"/>")

	path, err := Detect(dir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	want := filepath.Join(dir, "target/site/exec-jacoco/jacoco.xml")
	if path != want {
		t.Fatalf("path mismatch: got=%s want=%s", path, want)
	}
}

func TestDetectFallsBackWhenPOMMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "target/site/jacoco/jacoco.xml"), "<report name=\"x\"/>")

	path, err := Detect(dir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	want := filepath.Join(dir, "target/site/jacoco/jacoco.xml")
	if path != want {
		t.Fatalf("path mismatch: got=%s want=%s", path, want)
	}
}

func TestDetectReturnsErrorWhenNoCandidate(t *testing.T) {
	dir := t.TempDir()
	_, err := Detect(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
