# Coverage Report Viewer (`crv`)

JaCoCo XML レポートをターミナル上でインタラクティブに閲覧する CLI ツールです。  
ブラウザに切り替えず、階層をドリルダウンしてカバレッジを確認できます。

## 主な機能

- JaCoCo XML レポートの読み込み（引数指定または自動検出）
- `Report -> Package -> Class -> Method` の階層ナビゲーション
- カバレッジ率とバー表示
- 閾値ベースの色分け表示
- ソート切り替え（名前 / カバレッジ）

## インストール

### Homebrew

```bash
brew tap ochyai/homebrew-crv
brew install crv
```

### go install

```bash
go install github.com/izuno4t/coverage-report-viewer-cli/cmd/crv@latest
```

## 使い方

```bash
crv [options] [path]
```

- `path`: JaCoCo XML レポートのパス（省略時は自動検出）

### オプション

- `-t, --threshold <n>`: カバレッジ閾値（デフォルト: `80`）
- `-s, --sort <key>`: 初期ソート（`name` / `coverage`、デフォルト: `name`）
- `--no-color`: カラー表示を無効化
- `-v, --version`: バージョン表示
- `-h, --help`: ヘルプ表示

## キーバインド

- `↑` / `↓` または `k` / `j`: カーソル移動
- `Enter`: 子ノードへ移動
- `b` または `Backspace`: 親ノードへ戻る
- `s`: ソート切り替え
- `q` または `Ctrl+C`: 終了

## 開発コマンド（Make）

```bash
make help
make build
make test
make test-perf
make lint-md
```

- `make build`: `./bin/crv` をビルド
- `make test`: 全テスト実行（`go test ./...`）
- `make test-perf`: パーサ性能テスト実行
- `make lint-md`: Markdown lint 実行

## レポート自動検出

`path` 未指定時は以下の順で探索します。

1. カレントディレクトリの `pom.xml` を検出
2. `jacoco-maven-plugin` 設定から出力先を解決（`<pluginManagement>` も対象）
3. 解決先で `jacoco.xml` を探索
4. 見つからない場合は次のデフォルトパスへフォールバック
   - `target/site/jacoco/jacoco.xml`（Maven）
   - `build/reports/jacoco/test/jacocoTestReport.xml`（Gradle）

## 色分けルール

- 閾値未満: 赤
- 閾値以上 90% 未満: 黄
- 90% 以上: 緑

## 動作要件

- OS: macOS（arm64 / x86_64）, Linux（x86_64）
- Go: 1.22 以上（ビルド時）
- 対応 JaCoCo XML: 0.8.x 系

## 将来拡張（予定）

- ソースコード行カバレッジ表示
- マルチモジュール XML マージ
- diff モード
- インクリメンタル検索
- テキスト / JSON エクスポート
- Watch モード
