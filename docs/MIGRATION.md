# jrv から crv への移行ガイド

## 対象

- 旧コマンド `jrv` を利用していたユーザー
- Homebrew / `go install` / スクリプト実行で `jrv` を参照している環境

## 変更点

- 製品名: `JaCoCo Report Viewer CLI` から `Coverage Report Viewer` へ変更
- コマンド名: `jrv` から `crv` へ変更
- Go install パッケージ: `.../cmd/jrv` から `.../cmd/crv` へ変更
- Homebrew Tap: `homebrew-jrv` から `homebrew-crv` へ変更

## 移行手順

### Homebrew

```bash
brew untap ochyai/homebrew-jrv || true
brew tap ochyai/homebrew-crv
brew install crv
```

### go install

```bash
go install github.com/izuno4t/coverage-report-viewer-cli/cmd/crv@latest
```

### シェルエイリアス（一時互換）

既存スクリプトの段階的移行のため、必要に応じて一時的にエイリアスを設定する。

```bash
alias jrv=crv
```

## 互換方針

- 互換期間中は、旧名称 `jrv` の案内をドキュメントに残す。
- 旧コマンドの公式同梱は行わず、必要な場合はユーザー側エイリアスで吸収する。
- 新規ドキュメント・サンプルは `crv` を正とする。
- 互換案内の終了時期は、主要利用者の移行状況を見てリリースノートで告知する。
