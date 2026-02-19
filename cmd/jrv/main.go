package main

import (
	"os"

	"github.com/izuno4t/jacoco-report-viewer-cli/internal/app"
)

var version = "dev"

func main() {
	os.Exit(app.Run(os.Args[1:], version, os.Stdout, os.Stderr))
}
