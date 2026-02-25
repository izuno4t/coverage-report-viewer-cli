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

var startUIWatch = tui.StartWatch

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
	reportPaths := make([]string, 0, 1)
	if reportPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "error: カレントディレクトリ取得に失敗しました: %v\n", err)
			return 1
		}
		reportPaths, err = reportpath.DetectAll(cwd)
		if err != nil {
			_, _ = fmt.Fprintf(errOut, "error: JaCoCo XML が見つかりません: %v\n", err)
			_, _ = fmt.Fprintln(errOut, "hint: path を指定するか、target/site/jacoco/jacoco.xml の生成を確認してください")
			return 1
		}
	} else {
		reportPaths = append(reportPaths, reportPath)
	}

	loadReport := func() (jacoco.Report, error) {
		reports := make([]jacoco.Report, 0, len(reportPaths))
		for _, path := range reportPaths {
			report, err := jacoco.ParseWithFormatFile(path, jacoco.InputFormat(opts.Format))
			if err != nil {
				return jacoco.Report{}, err
			}
			reports = append(reports, report)
		}
		return jacoco.MergeReports(reports...), nil
	}
	report, err := loadReport()
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "error: カバレッジレポートの読み込みに失敗しました: %v\n", err)
		return 1
	}

	uiConfig := tui.Config{
		Threshold: opts.Threshold,
		Sort:      opts.Sort,
		NoColor:   opts.NoColor,
		Watch:     opts.Watch,
	}

	probe, err := newReportUpdateProbe(reportPaths)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "error: 監視対象レポートの状態取得に失敗しました: %v\n", err)
		return 1
	}
	if err := startUIWatch(report, uiConfig, loadReport, probe); err != nil {
		_, _ = fmt.Fprintf(errOut, "error: TUI 起動に失敗しました: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(out, "crv finished")
	return 0
}
