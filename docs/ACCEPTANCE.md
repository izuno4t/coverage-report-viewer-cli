# 受け入れ確認記録（M1）

## 実施日

- 2026-02-19

## 対象

- 要件定義: `docs/REQUIREMENTS.md`
- トレーサビリティ: `docs/TRACEABILITY.md`
- テスト対応: `docs/TESTCASES.md`
- 検証記録: `docs/VERIFICATION.md`

## 受け入れ結果（必須機能）

| 要件ID | 結果 | 根拠 |
| --- | --- | --- |
| F-IN-01 | ✅ | `internal/jacoco` パーサ実装とテスト |
| F-IN-02 | ✅ | `internal/cli` 引数解析実装とテスト |
| F-IN-03 | ✅ | `internal/reportpath` POM解析実装とテスト |
| F-IN-04 | ✅ | デフォルトパスフォールバック実装とテスト |
| F-NAV-01 | ✅ | `internal/tui` カーソル移動実装とテスト |
| F-NAV-02 | ✅ | Enterによる階層遷移実装とテスト |
| F-NAV-03 | ✅ | b/Backspace戻り実装とテスト |
| F-NAV-04 | ✅ | q/Ctrl+C終了実装とテスト |
| F-NAV-05 | ✅ | sキーのソート切替実装とテスト |
| F-DSP-01 | ✅ | 閾値色分け実装とテスト |
| F-DSP-02 | ✅ | カバレッジバー表示実装とテスト |

## 受け入れ結果（非機能）

| 要件ID | 結果 | 根拠 |
| --- | --- | --- |
| NF-01 | ✅ | 1000クラス性能テストで 13.303166ms |
| NF-02 | ✅ | darwin/linux クロスビルド成功 |
| NF-03 | ✅ | Go 1.22+（現行 1.26.0）でビルド/テスト成功 |
| NF-04 | 🧪 | 依存最小化は運用レビュー継続 |
| NF-05 | ✅ | `.goreleaser.yaml` と `docs/RELEASE.md` 追加 |
| NF-06 | ✅ | JaCoCo 0.8系XMLを想定したパーサ実装 |
| NF-07 | ✅ | `CGO_ENABLED=0` で単一バイナリ設定 |

## 未対応（M1対象外）

- 任意要件: F-NAV-07, F-DSP-03, F-DSP-04, F-DSP-05
- 将来要件: F-IN-05, F-NAV-06, FUT-01〜FUT-06
