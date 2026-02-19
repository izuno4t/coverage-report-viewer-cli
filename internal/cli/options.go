package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	defaultThreshold = 80
	defaultSort      = "name"
)

var validSortKeys = map[string]struct{}{
	"name":     {},
	"coverage": {},
}

// Options is the normalized runtime configuration from CLI arguments.
type Options struct {
	Path        string
	Format      string
	Threshold   int
	Sort        string
	Watch       bool
	NoColor     bool
	ShowVersion bool
	ShowHelp    bool
}

func Parse(args []string) (Options, error) {
	opts := Options{
		Format:    "auto",
		Threshold: defaultThreshold,
		Sort:      defaultSort,
	}

	fs := flag.NewFlagSet("crv", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var helpShort bool
	fs.IntVar(&opts.Threshold, "threshold", defaultThreshold, "coverage threshold")
	fs.IntVar(&opts.Threshold, "t", defaultThreshold, "coverage threshold")
	fs.StringVar(&opts.Format, "format", "auto", "input format")
	fs.StringVar(&opts.Sort, "sort", defaultSort, "initial sort key")
	fs.StringVar(&opts.Sort, "s", defaultSort, "initial sort key")
	fs.BoolVar(&opts.Watch, "watch", false, "watch input report and reload")
	fs.BoolVar(&opts.NoColor, "no-color", false, "disable color output")
	fs.BoolVar(&opts.ShowVersion, "version", false, "show version")
	fs.BoolVar(&opts.ShowVersion, "v", false, "show version")
	fs.BoolVar(&opts.ShowHelp, "help", false, "show help")
	fs.BoolVar(&helpShort, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if helpShort {
		opts.ShowHelp = true
	}

	if opts.ShowHelp || opts.ShowVersion {
		return opts, nil
	}

	rest := fs.Args()
	if len(rest) > 1 {
		return Options{}, errors.New("path は1つだけ指定できます")
	}
	if len(rest) == 1 {
		opts.Path = rest[0]
	}

	if opts.Threshold < 0 || opts.Threshold > 100 {
		return Options{}, fmt.Errorf("threshold は 0 から 100 の範囲で指定してください: %d", opts.Threshold)
	}

	opts.Format = strings.ToLower(strings.TrimSpace(opts.Format))
	switch opts.Format {
	case "auto", "jacoco", "cobertura", "lcov":
	default:
		return Options{}, fmt.Errorf("format は auto / jacoco / cobertura / lcov を指定してください: %s", opts.Format)
	}

	opts.Sort = strings.ToLower(opts.Sort)
	if _, ok := validSortKeys[opts.Sort]; !ok {
		return Options{}, fmt.Errorf("sort は name または coverage を指定してください: %s", opts.Sort)
	}

	return opts, nil
}

func Usage() string {
	return strings.TrimSpace(`Usage:
  crv [options] [path]

Options:
      --format <fmt>    入力フォーマット（auto|jacoco|cobertura|lcov, default: auto）
  -t, --threshold <n>  カバレッジ閾値（0-100, default: 80）
  -s, --sort <key>     初期ソート（name|coverage, default: name）
      --watch          レポート変更を監視して自動再読み込み
      --no-color       カラー出力を無効化
  -v, --version        バージョンを表示
  -h, --help           ヘルプを表示
`)
}
