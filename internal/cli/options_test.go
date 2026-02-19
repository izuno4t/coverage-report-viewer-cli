package cli

import "testing"

func TestParseDefaults(t *testing.T) {
	opts, err := Parse([]string{"report.xml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Path != "report.xml" {
		t.Fatalf("path mismatch: %s", opts.Path)
	}
	if opts.Format != "auto" {
		t.Fatalf("format mismatch: %s", opts.Format)
	}
	if opts.Threshold != 80 {
		t.Fatalf("threshold mismatch: %d", opts.Threshold)
	}
	if opts.Sort != "name" {
		t.Fatalf("sort mismatch: %s", opts.Sort)
	}
}

func TestParseRejectsInvalidFormat(t *testing.T) {
	_, err := Parse([]string{"--format", "unknown", "report.xml"})
	if err == nil {
		t.Fatal("expected format error")
	}
}

func TestParseAcceptsLCOVFormat(t *testing.T) {
	opts, err := Parse([]string{"--format", "lcov", "report.info"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Format != "lcov" {
		t.Fatalf("format mismatch: %s", opts.Format)
	}
}

func TestParseWatchFlag(t *testing.T) {
	opts, err := Parse([]string{"--watch", "report.xml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Watch {
		t.Fatal("watch flag should be true")
	}
}

func TestParseRejectsInvalidThreshold(t *testing.T) {
	_, err := Parse([]string{"--threshold", "101", "report.xml"})
	if err == nil {
		t.Fatal("expected threshold error")
	}
}

func TestParseRejectsInvalidSort(t *testing.T) {
	_, err := Parse([]string{"--sort", "unknown", "report.xml"})
	if err == nil {
		t.Fatal("expected sort error")
	}
}

func TestParseVersionSkipsValidation(t *testing.T) {
	opts, err := Parse([]string{"--version", "--sort", "unknown"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.ShowVersion {
		t.Fatal("version flag should be true")
	}
}
