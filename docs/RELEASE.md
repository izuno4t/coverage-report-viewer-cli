# リリース手順（M1）

## 配布経路

- Homebrew Tap: `izuno4t/homebrew-tap`
- Go install: `go install github.com/izuno4t/coverage-report-viewer-cli/cmd/crv@latest`

## GoReleaser

- 設定ファイル: `.goreleaser.yaml`
- 対象OS/Arch:
  - darwin/arm64
  - darwin/amd64
  - linux/amd64
- 生成バイナリ名: `crv`
- バージョン注入: `-X main.version={{.Version}}`

## リリース前チェック（ローカル）

```bash
go test ./...
go build ./cmd/crv
```

ローカルに `goreleaser` がある場合は設定検証も実行:

```bash
goreleaser check
```

## 公開リリース手順

1. `main` の最新を取り込む
2. リリースタグを作成する（例: `v1.0.0`）
3. タグを push する

```bash
git tag v1.0.0
git push origin v1.0.0
```

タグ push をトリガーに GitHub Actions の Release workflow が `goreleaser release --clean` を実行し、GitHub Releases と Homebrew Tap を更新します。

## 手動ドライラン（公開なし）

GitHub Actions の `Release` workflow を `workflow_dispatch` で手動実行すると、
`goreleaser build --snapshot --clean` が実行され、公開せずにビルド成立のみを確認できます。

## GitHub Actions

- `.github/workflows/ci.yml`: push/PR (`main`) で `gofmt` / `go test ./...` / Markdown lint を実行
- `.github/workflows/release.yml`:
  - `v*` タグ push で公開リリース（`goreleaser release --clean`）
  - `workflow_dispatch` で公開なしドライラン（`goreleaser build --snapshot --clean`）
- Release workflow で Homebrew Tap 連携するため、リポジトリ Secrets に `HOMEBREW_TAP_GITHUB_TOKEN` を設定する

## 補足

- Homebrew Formula は `Formula/crv.rb` として生成される。
- `go install` は `cmd/crv` パッケージを指定する。
- `--format` オプションで `jacoco` / `cobertura` / `lcov` を明示指定できる。
- 旧コマンド移行方針は `docs/MIGRATION.md` を参照する。
- Homebrew Tap へ push するため、`HOMEBREW_TAP_GITHUB_TOKEN`（tapリポジトリへの書き込み権限付き）を設定する。
