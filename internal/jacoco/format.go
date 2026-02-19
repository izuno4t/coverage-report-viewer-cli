package jacoco

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type InputFormat string

const (
	FormatAuto      InputFormat = "auto"
	FormatJaCoCo    InputFormat = "jacoco"
	FormatCobertura InputFormat = "cobertura"
	FormatLCOV      InputFormat = "lcov"
)

func ParseWithFormatFile(path string, format InputFormat) (Report, error) {
	switch format {
	case FormatJaCoCo:
		return ParseFile(path)
	case FormatCobertura:
		return ParseCoberturaFile(path)
	case FormatLCOV:
		return ParseLCOVFile(path)
	case FormatAuto:
		detected, err := DetectFormatFile(path)
		if err != nil {
			return Report{}, err
		}
		return ParseWithFormatFile(path, detected)
	default:
		return Report{}, fmt.Errorf("unsupported input format: %s", format)
	}
}

func DetectFormatFile(path string) (InputFormat, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open report for format detection: %w", err)
	}
	defer f.Close()
	return DetectFormat(f)
}

func DetectFormat(r io.Reader) (InputFormat, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read report for format detection: %w", err)
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("unsupported or empty report format")
	}
	if strings.HasPrefix(trimmed, "<") {
		return detectXMLFormat(bytes.NewReader(data))
	}
	return detectTextFormat(trimmed)
}

func detectXMLFormat(r io.Reader) (InputFormat, error) {
	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("detect format: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "report":
			return FormatJaCoCo, nil
		case "coverage":
			return FormatCobertura, nil
		default:
			return "", fmt.Errorf("unsupported xml root element: %s", start.Name.Local)
		}
	}
	return "", fmt.Errorf("unsupported xml report format")
}

func detectTextFormat(text string) (InputFormat, error) {
	for _, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "TN:") ||
			strings.HasPrefix(trim, "SF:") ||
			strings.HasPrefix(trim, "DA:") ||
			strings.HasPrefix(trim, "FN:") ||
			strings.HasPrefix(trim, "BRDA:") {
			return FormatLCOV, nil
		}
		break
	}
	return "", fmt.Errorf("unsupported text report format")
}
