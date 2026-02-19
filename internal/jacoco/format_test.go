package jacoco

import (
	"strings"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	cases := []struct {
		name string
		xml  string
		want InputFormat
	}{
		{name: "jacoco", xml: `<report name="x"></report>`, want: FormatJaCoCo},
		{name: "cobertura", xml: `<coverage></coverage>`, want: FormatCobertura},
		{name: "lcov", xml: "TN:\nSF:src/main.py\nDA:1,1\nend_of_record\n", want: FormatLCOV},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DetectFormat(strings.NewReader(tc.xml))
			if err != nil {
				t.Fatalf("detect failed: %v", err)
			}
			if got != tc.want {
				t.Fatalf("format mismatch: got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestDetectFormatRejectsUnknownRoot(t *testing.T) {
	_, err := DetectFormat(strings.NewReader(`<unknown/>`))
	if err == nil {
		t.Fatal("expected error for unknown root")
	}
}
