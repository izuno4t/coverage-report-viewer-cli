# 検証記録（M1）

## 実施日

- 2026-02-19

## 実行環境

- OS: macOS (arm64)
- Go: `go version go1.26.0 darwin/arm64`

## NF-01 起動性能（1,000クラス規模）

- コマンド: `go test ./internal/jacoco -run TestParsePerformance1000Classes -count=1 -v`
- 結果: `elapsed=13.303166ms`
- 判定: 1秒以内を満たす

## NF-02 対応OSビルド確認

- コマンド: `GOOS=darwin GOARCH=arm64 go build -o /tmp/jrv-darwin-arm64 ./cmd/jrv`
- 結果: 成功

- コマンド: `GOOS=darwin GOARCH=amd64 go build -o /tmp/jrv-darwin-amd64 ./cmd/jrv`
- 結果: 成功

- コマンド: `GOOS=linux GOARCH=amd64 go build -o /tmp/jrv-linux-amd64 ./cmd/jrv`
- 結果: 成功

## 補足

- Linux実機での実行検証は未実施（クロスビルド成功まで確認）。
