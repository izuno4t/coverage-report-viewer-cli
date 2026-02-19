# テストケース対応表（M1）

## 目的

`docs/TRACEABILITY.md` のケースIDと実装済みテストを対応付ける。

## ケース対応

| ケースID | 対応要件 | テスト名 |
| --- | --- | --- |
| TC-UNIT-001 | F-IN-01 | `internal/jacoco/parser_test.go:TestParseBasicHierarchy` |
| TC-INT-001 | F-IN-01 | `internal/app/run_test.go:TestRunFailsOnInvalidXML` |
| TC-UNIT-002 | F-IN-02 | `internal/cli/options_test.go:TestParseDefaults` |
| TC-E2E-001 | F-IN-02 | `internal/app/run_test.go:TestRunSucceedsWithPath` |
| TC-UNIT-003 | F-IN-03 | `internal/reportpath/detector_test.go:TestDetectFromPOMPluginConfig` |
| TC-INT-002 | F-IN-03 | `internal/reportpath/detector_test.go:TestDetectFromPOMExecutionReportGoal` |
| TC-UNIT-004 | F-IN-04 | `internal/reportpath/detector_test.go:TestDetectFallsBackWhenPOMMissing` |
| TC-E2E-002 | F-IN-04 | `internal/app/run_test.go:TestRunAutoDetectsReportPath` |
| TC-UNIT-005 | F-NAV-01 | `internal/tui/model_test.go:TestCursorMoveBounds` |
| TC-INT-003 | F-NAV-01 | `internal/tui/model_test.go:TestDrillDownAndBack` |
| TC-UNIT-006 | F-NAV-02 | `internal/tui/model_test.go:TestDrillDownAndBack` |
| TC-INT-004 | F-NAV-02 | `internal/tui/model_test.go:TestViewIncludesSections` |
| TC-UNIT-007 | F-NAV-03 | `internal/tui/model_test.go:TestDrillDownAndBack` |
| TC-INT-005 | F-NAV-03 | `internal/tui/model_test.go:TestDrillDownAndBack` |
| TC-UNIT-008 | F-NAV-04 | `internal/tui/model_test.go:TestQuitKeys` |
| TC-E2E-003 | F-NAV-04 | `internal/app/run_test.go:TestRunHelp` |
| TC-UNIT-009 | F-NAV-05 | `internal/tui/model_test.go:TestSortCycle` |
| TC-INT-006 | F-NAV-05 | `internal/tui/model_test.go:TestCoverageSortAffectsChildOrder` |
| TC-UNIT-010 | F-DSP-01 | `internal/tui/model_test.go:TestBandForCoverage` |
| TC-INT-007 | F-DSP-01 | `internal/tui/model_test.go:TestViewIncludesSections` |
| TC-UNIT-011 | F-DSP-02 | `internal/tui/model_test.go:TestBarWidth` |
| TC-INT-008 | F-DSP-02 | `internal/tui/model_test.go:TestViewIncludesSections` |
| TC-UNIT-012 | F-IN-05 | `internal/reportpath/detector_test.go:TestDetectAllFindsModuleReports` |
| TC-INT-009 | F-IN-05 | `internal/app/run_test.go:TestRunAutoDetectsAndMergesMultiModuleReports` |
| TC-UNIT-013 | F-IN-06 | `internal/jacoco/cobertura_test.go:TestParseCoberturaBasicHierarchy` |
| TC-UNIT-014 | F-IN-07 | `internal/jacoco/format_test.go:TestDetectFormat` |
| TC-UNIT-015 | F-IN-07 | `internal/cli/options_test.go:TestParseAcceptsLCOVFormat` |
| TC-UNIT-016 | F-IN-08 | `internal/jacoco/lcov_test.go:TestParseLCOVBasic` |
| TC-INT-010 | F-IN-08 | `internal/app/run_test.go:TestRunWithFormatLCOV` |
| TC-UNIT-017 | F-NAV-06 | `internal/tui/model_test.go:TestIncrementalFilter` |
| TC-UNIT-018 | F-NAV-07 | `internal/tui/model_test.go:TestJumpKeys` |
| TC-UNIT-019 | F-DSP-03 | `internal/tui/model_test.go:TestCounterTypeCycle` |
| TC-UNIT-020 | F-DSP-04 | `internal/tui/model_test.go:TestRenderSummaryFitsNarrowWidth` |
| TC-UNIT-021 | F-DSP-05 | `internal/tui/model_test.go:TestViewNoColorDisablesANSISequences` |
| TC-NFR-003 | NF-03 | `go.mod` と `go test ./...` 成功 |
| TC-NFR-006 | NF-06 | `internal/jacoco/parser_test.go` 各テスト |

## 未自動化（M1内で後続タスク対応）

| ケースID | 対象要件 | 状態 |
| --- | --- | --- |
| TC-NFR-001 | NF-01 | TASK-011 で実施 |
| TC-NFR-002 | NF-02 | TASK-011 で実施 |
| TC-NFR-004 | NF-04 | TASK-011 で確認 |
| TC-NFR-005 | NF-05 | TASK-012 で実施 |
| TC-NFR-007 | NF-07 | TASK-012 で実施 |
