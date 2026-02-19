# リリース手順（M1）

## 配布経路

- Homebrew Tap: `ochyai/homebrew-jrv`
- Go install: `go install github.com/izuno4t/jacoco-report-viewer-cli/cmd/jrv@latest`

## GoReleaser

- 設定ファイル: `.goreleaser.yaml`
- 対象OS/Arch:
  - darwin/arm64
  - darwin/amd64
  - linux/amd64
- 生成バイナリ名: `jrv`
- バージョン注入: `-X main.version={{.Version}}`

## 実行例

```bash
goreleaser release --clean
```

## 補足

- Homebrew Formula は `Formula/jrv.rb` として生成される。
- `go install` は `cmd/jrv` パッケージを指定する。
