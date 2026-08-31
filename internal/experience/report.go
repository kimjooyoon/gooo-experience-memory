package experience

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteReport(outputDir string, report Report) error {
	if !filepath.IsAbs(outputDir) {
		return errors.New("report output must be an absolute caller-owned directory")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	data, err := report.JSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "evaluation.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "summary.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "human-report.md"), []byte(RenderMarkdown(report)), 0o644)
}

func RenderMarkdown(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Gooo experience memory\n\n- decision: `%s`\n- state: `%s`\n- pipeline: `GOOO_SOURCE -> SEMANTIC_IR -> GENERATED_GO -> EVALUATOR`\n- fixed denominator: `%d` cells\n- precedence: `%s`\n\n", report.Decision, report.State, report.FixedDenominator, strings.Join(report.Precedence, " > "))
	fmt.Fprintf(&b, "source digest: `%s`\nsemantic IR digest: `%s`\ngenerated Go digest: `%s`\ncontract digest: `%s`\nfixture digest: `%s`\nmemory digest: `%s`\n\n", report.SourceDigest, report.SemanticIRDigest, report.GeneratedGoDigest, report.ContractDigest, report.FixtureDigest, report.MemoryDigest)
	b.WriteString("## Exact closure pair\n\n")
	fmt.Fprintf(&b, "| metric | before | after |\n|---|---:|---:|\n| candidate count | %d | %d |\n| known REFUTED recurrences | %d | %d |\n| selected candidate | `%s` | `%s` |\n| replay mismatches | %d | %d |\n\n", report.Baseline.CandidateCount, report.After.CandidateCount, report.Metrics.KnownRefutedRecurrencesBefore, report.Metrics.KnownRefutedRecurrencesAfter, report.Baseline.SelectedCandidateID, report.After.SelectedCandidateID, report.Replay.Mismatches, report.Replay.Mismatches)
	fmt.Fprintf(&b, "avoided REFUTED candidates: `%d`\nnew UNKNOWN candidates: `%d`\nreplay comparisons: `%d`\nreplay mismatches: `%d`\n\n", report.Metrics.AvoidedRefutedCandidates, report.Metrics.NewUnknownCandidates, report.Metrics.ReplayComparisons, report.Metrics.ReplayMismatches)
	b.WriteString("## Canonical cases\n\n| case | class | state | reason |\n|---|---|---|---|\n")
	for _, item := range report.CanonicalCases {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", item.CaseID, item.Class, item.State, item.Reason)
	}
	b.WriteString("\n## UNKNOWN coordinates\n\n")
	for _, item := range report.CanonicalCases {
		if item.Unknown != nil {
			u := item.Unknown
			fmt.Fprintf(&b, "- `%s`: stage=`%s`; step=`%s`; reason=`%s`; unknown_class=`%s`; next_operation=`%s`; blocked_by=`%s`\n", item.CaseID, u.Stage, u.Step, u.Reason, u.UnknownClass, u.NextOperation, strings.Join(u.BlockedBy, ","))
		}
	}
	b.WriteString("\n## 12-cell bindings\n\n")
	for _, binding := range report.Bindings {
		fmt.Fprintf(&b, "- `%02d %s` -> IR `%s` -> generated `%s` -> evaluator `%s`\n", binding.Ordinal, binding.Activity, binding.IRNode, binding.GeneratedSymbol, binding.Evaluator)
	}
	fmt.Fprintf(&b, "\nExact integer metrics: attempts_observed=`%d`, memory_records=`%d`, candidate_count=`%d`, peak_rss_kib=`%d`, wall_ms=`%d`.\n", report.Metrics.AttemptsObserved, report.Metrics.MemoryRecords, report.Metrics.CandidateCount, report.Metrics.PeakRSSKiB, report.Metrics.WallMS)
	fmt.Fprintf(&b, "Inventory: Go files=`%d`, Go physical lines=`%d`, Gooo files=`%d`, Gooo physical lines=`%d`, descendant dirs=`%d`; tests total=`%d`, executed=`%d`, reused=`%d`, skipped=`%d`, not_observed=`%d`. Root README is excluded.\n", report.Metrics.GoFiles, report.Metrics.GoPhysicalLines, report.Metrics.GoooFiles, report.Metrics.GoooPhysicalLines, report.Metrics.DescendantDirs, report.Metrics.Tests.Total, report.Metrics.Tests.Executed, report.Metrics.Tests.Reused, report.Metrics.Tests.Skipped, report.Metrics.Tests.NotObserved)
	b.WriteString("\nAuthority: repository_writes=`0`, local_test_executions=`0`, cross_project_required_gates=`0`. External inputs are read-only immutable digests.\n")
	return b.String()
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
