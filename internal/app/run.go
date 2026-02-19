package app

import (
	"fmt"
	"io"
	"os"

	"github.com/izuno4t/coverage-report-viewer-cli/internal/cli"
	"github.com/izuno4t/coverage-report-viewer-cli/internal/jacoco"
	"github.com/izuno4t/coverage-report-viewer-cli/internal/reportpath"
	"github.com/izuno4t/coverage-report-viewer-cli/internal/tui"
)

var startUI = tui.Start

// Run executes the CLI flow and returns the process exit code.
func Run(args []string, version string, out io.Writer, errOut io.Writer) int {
	opts, err := cli.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "error: %v\n\n", err)
		_, _ = fmt.Fprintln(errOut, cli.Usage())
		return 2
	}

	if opts.ShowHelp {
		_, _ = fmt.Fprintln(out, cli.Usage())
		return 0
	}
	if opts.ShowVersion {
		_, _ = fmt.Fprintf(out, "crv %s\n", version)
		return 0
	}

	reportPath := opts.Path
	if reportPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "error: カレントディレクトリ取得に失敗しました: %v\n", err)
			return 1
		}
		reportPath, err = reportpath.Detect(cwd)
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "error: JaCoCo XML が見つかりません: %v\n", err)
			_, _ = fmt.Fprintln(errOut, "hint: path を指定するか、target/site/jacoco/jacoco.xml の生成を確認してください")
			return 1
		}
	}

	report, err := jacoco.ParseFile(reportPath)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "error: JaCoCo XML の読み込みに失敗しました: %v\n", err)
		return 1
	}

	if err := startUI(report, tui.Config{
		Threshold: opts.Threshold,
		Sort:      opts.Sort,
		NoColor:   opts.NoColor,
	}); err != nil {
		_, _ = fmt.Fprintf(errOut, "error: TUI 起動に失敗しました: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(out, "crv finished")
	return 0
}
