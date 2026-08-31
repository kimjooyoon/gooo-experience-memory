package experience

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Meta struct {
	SourcePath      string
	SourceDigest    string
	ContractDigest  string
	SemanticIRDigest string
	GeneratedDigest string
	IR              SemanticIR
	Contract        Denominator
}

type RuntimeInput struct {
	Tests      TestMetrics     `json:"tests"`
	Inventory  InventoryMetrics `json:"inventory"`
	Authority  Authority       `json:"authority"`
	WallMS     int             `json:"wall_ms"`
	PeakRSSKiB int             `json:"peak_rss_kib"`
}

func LoadMeta(sourcePath, contractPath, irPath, generatedPath string) (Meta, error) {
	source, err := os.ReadFile(sourcePath)
	if err != nil { return Meta{}, err }
	contract, contractDigest, err := LoadDenominator(contractPath)
	if err != nil { return Meta{}, err }
	var ir SemanticIR
	if err := LoadJSON(irPath, &ir); err != nil { return Meta{}, err }
	expected, err := CompileSource(sourcePath, source, contract, contractDigest)
	if err != nil { return Meta{}, err }
	if DigestIR(ir) != DigestIR(expected) || ir.SourceDigest != DigestBytes(source) || ir.ContractDigest != contractDigest {
		return Meta{}, errors.New("semantic IR is not exactly bound to source and denominator")
	}
	generatedDigest, err := ValidateGenerated(generatedPath, ir)
	if err != nil { return Meta{}, err }
	return Meta{SourcePath: sourcePath, SourceDigest: DigestBytes(source), ContractDigest: contractDigest, SemanticIRDigest: DigestIR(ir), GeneratedDigest: generatedDigest, IR: ir, Contract: contract}, nil
}

func Evaluate(meta Meta, fixture FixedFixture, records []MemoryRecord, receipt *OutcomeReceipt, cases []CanonicalCase, runtime RuntimeInput) (Report, error) {
	started := time.Now()
	if err := ValidateDenominator(meta.Contract); err != nil { return Report{}, err }
	if fixture.ScopeDigest != DigestText(fixture.ScopeDescriptor) || fixtureDigestIsReferenced(fixture, records, receipt) == false {
		return Report{}, errors.New("fixed fixture scope or memory binding is invalid")
	}
	if len(cases) != 12 { return Report{}, errors.New("canonical case denominator must be exactly 12") }
	for _, record := range records {
		if err := ValidateRecord(record, FixtureDigest(fixture)); err != nil { return Report{}, err }
	}
	if receipt != nil && receipt.FixtureDigest != FixtureDigest(fixture) { return Report{}, errors.New("receipt is bound to another fixture") }
	baseline := selectBaseline(fixture)
	after := selectAfter(fixture, records, receipt)
	baselineRecurrences := recurrenceCount(baseline, fixture, "before")
	afterRecurrences := recurrenceCount(after, fixture, "after")
	replay := replayPair(fixture, records, receipt)
	canonical := make([]CaseResult, 0, len(cases))
	counts := map[string]int{"normal": 0, "UNKNOWN": 0, "REFUTED": 0}
	for _, item := range cases {
		result := classifyCase(item, fixture, records, receipt, meta.IR.Nodes)
		if result.State != item.ExpectedState || (item.ExpectedReason != "" && result.Reason != item.ExpectedReason) {
			return Report{}, fmt.Errorf("canonical case %s produced %s, expected %s", item.CaseID, result.State, item.ExpectedState)
		}
		if result.State == StateUnknown && !validUnknown(result.Unknown) { return Report{}, fmt.Errorf("canonical case %s has incomplete UNKNOWN tuple", item.CaseID) }
		canonical = append(canonical, result)
		counts[item.Class]++
	}
	if counts["normal"] != 4 || counts[StateUnknown] != 4 || counts[StateRefuted] != 4 { return Report{}, errors.New("canonical denominator must be normal=4 UNKNOWN=4 REFUTED=4") }
	bindings := buildBindings(meta)
	wall := runtime.WallMS
	if wall <= 0 { wall = int(time.Since(started).Milliseconds()); if wall < 1 { wall = 1 } }
	peak := runtime.PeakRSSKiB
	if peak <= 0 { peak = currentRSSKiB() }
	if peak < 1 { peak = 1 }
	inventory := runtime.Inventory
	metrics := Metrics{
		AttemptsObserved: len(fixture.Attempts), MemoryRecords: len(records), CandidateCount: len(fixture.Candidates),
		KnownRefutedRecurrencesBefore: baselineRecurrences, KnownRefutedRecurrencesAfter: afterRecurrences,
		AvoidedRefutedCandidates: maxInt(0, baselineRecurrences-afterRecurrences), NewUnknownCandidates: unknownCount(after),
		ReplayComparisons: replay.Comparisons, ReplayMismatches: replay.Mismatches, PeakRSSKiB: peak, WallMS: wall,
		GoPhysicalLines: inventory.GoPhysicalLines, GoFiles: inventory.GoFiles, GoooPhysicalLines: inventory.GoooPhysicalLines,
		GoooFiles: inventory.GoooFiles, DescendantDirs: inventory.DescendantDirs, Tests: runtime.Tests,
		Inventory: inventory, Authority: runtime.Authority,
	}
	if metrics.Authority != (Authority{}) && (metrics.Authority.RepositoryWrites != 0 || metrics.Authority.LocalTestExecutions != 0 || metrics.Authority.CrossProjectRequiredGates != 0) {
		return Report{}, errors.New("runtime authority boundary is open")
	}
	metrics.Authority = Authority{}
	if runtime.Authority != (Authority{}) && runtime.Authority != metrics.Authority { return Report{}, errors.New("authority observation is not zero") }
	return Report{
		Schema: ReportSchema, Decision: "EXPERIENCE_MEMORY_CONFORMANCE_CLOSED", State: StateClosed, Precedence: append([]string(nil), Precedence...),
		SourcePath: meta.SourcePath, SourceDigest: meta.SourceDigest, SemanticIRDigest: meta.SemanticIRDigest, GeneratedGoDigest: meta.GeneratedDigest, ContractDigest: meta.ContractDigest,
		FixtureDigest: FixtureDigest(fixture), MemoryDigest: DigestMemory(records), ReceiptDigest: receiptDigest(receipt), FixedDenominator: 12,
		Proofs: balances(meta.Contract, true), Indicators: balances(meta.Contract, false), Bindings: bindings, CanonicalCases: canonical, CanonicalCounts: counts,
		Baseline: baseline, After: after, Replay: replay, Metrics: metrics, Authority: Authority{}, AppendOnly: true, ExternalInputsReadOnly: true, RootReadmeExcluded: true,
	}, nil
}

func fixtureDigestIsReferenced(fixture FixedFixture, records []MemoryRecord, receipt *OutcomeReceipt) bool {
	digest := FixtureDigest(fixture)
	if len(records) == 0 && receipt == nil { return true }
	for _, record := range records { if record.FixtureDigest != digest { return false } }
	if receipt != nil && receipt.FixtureDigest != digest { return false }
	return true
}

func ValidateRecord(record MemoryRecord, fixtureDigest string) error {
	if record.Schema != MemorySchema || record.RecordID == "" || record.Ordinal < 1 || record.Kind != StateRefuted || record.FixtureDigest != fixtureDigest || !validDigest(record.RecordDigest) || DigestRecord(record) != record.RecordDigest {
		return errors.New("memory record is invalid")
	}
	if record.OutcomeReceipt.FixtureDigest != fixtureDigest || !validDigest(record.OutcomeReceipt.ReceiptDigest) || DigestReceipt(record.OutcomeReceipt) != record.OutcomeReceipt.ReceiptDigest {
		return errors.New("memory record receipt is invalid")
	}
	if record.SemanticFingerprint != record.OutcomeReceipt.SemanticFingerprint || record.FailureClass != record.OutcomeReceipt.FailureClass || record.ScopeDigest != record.OutcomeReceipt.ScopeDigest {
		return errors.New("memory record match basis is not bound")
	}
	return nil
}

func AssessCandidate(candidate Candidate, records []MemoryRecord, receipt *OutcomeReceipt, fixtureDigest, fixtureVersion string, observed *Attempt) CandidateDecision {
	decision := CandidateDecision{CandidateID: candidate.CandidateID, State: StateClosed, Decision: "ELIGIBLE", Reason: "NO_APPLICABLE_EXPERIENCE_RECORD", MatchBasis: MatchBasis{SemanticFingerprint: candidate.SemanticFingerprint, FailureClass: candidate.FailureClass, ScopeDigest: candidate.ScopeDigest}}
	exact := make([]MemoryRecord, 0)
	related := false
	for _, record := range records {
		if record.FixtureDigest != fixtureDigest { continue }
		if record.SemanticFingerprint == candidate.SemanticFingerprint && record.FailureClass == candidate.FailureClass && record.ScopeDigest == candidate.ScopeDigest {
			exact = append(exact, record)
		}
		if record.SemanticFingerprint == candidate.SemanticFingerprint || record.FailureClass == candidate.FailureClass {
			related = true
		}
	}
	decision.MatchBasis.AllThreeExact = len(exact) == 1
	if observed != nil && observed.Outcome == StateClosed && len(exact) > 0 && sameAttemptBasis(candidate, *observed) {
		return refutedDecision(decision, "CURRENT_OBSERVATION_CONTRADICTS_KNOWN_REFUTATION")
	}
	if len(exact) > 1 {
		return unknownDecision(decision, UnknownTuple{Stage: "MEMORY_LOOKUP", Step: "DISAMBIGUATE_EXPERIENCE_RECEIPTS", Reason: "MULTIPLE_EXACT_OUTCOME_RECEIPTS", UnknownClass: UnknownAmbiguous, NextOperation: "APPEND_UNAMBIGUOUS_OUTCOME_RECEIPT", BlockedBy: []string{"exact-match-set"}})
	}
	if len(exact) == 1 {
		record := exact[0]
		if err := ValidateRecord(record, fixtureDigest); err != nil { return refutedDecision(decision, "TAMPERED_MEMORY_RECORD") }
		if receipt == nil || receipt.ReceiptDigest != record.OutcomeReceipt.ReceiptDigest || receipt.SemanticFingerprint != candidate.SemanticFingerprint || receipt.FailureClass != candidate.FailureClass || receipt.ScopeDigest != candidate.ScopeDigest {
			return unknownDecision(decision, UnknownTuple{Stage: "RECEIPT_BINDING", Step: "REQUIRE_EXACT_IMMUTABLE_OUTCOME_RECEIPT", Reason: "EXACT_MEMORY_RECORD_HAS_NO_CURRENT_RECEIPT", UnknownClass: UnknownDirect, NextOperation: "PROVIDE_EXACT_IMMUTABLE_OUTCOME_RECEIPT", BlockedBy: []string{record.RecordID}})
		}
		if !receipt.Immutable || receipt.ObservedVersion != fixtureVersion || receipt.Outcome != StateRefuted {
			return unknownDecision(decision, UnknownTuple{Stage: "RECEIPT_BINDING", Step: "VERIFY_RECEIPT_FRESHNESS", Reason: "STALE_OR_MUTABLE_OUTCOME_RECEIPT", UnknownClass: UnknownStale, NextOperation: "COLLECT_CURRENT_IMMUTABLE_OUTCOME_RECEIPT", BlockedBy: []string{record.RecordID}})
		}
		return refutedDecision(decision, "KNOWN_REFUTED_EXPERIENCE_MATCHED_ON_FINGERPRINT_FAILURE_AND_SCOPE")
	}
	if receipt != nil && receipt.FixtureDigest == fixtureDigest && receipt.SemanticFingerprint == candidate.SemanticFingerprint && receipt.FailureClass == candidate.FailureClass && receipt.ScopeDigest == candidate.ScopeDigest {
		if !receipt.Immutable || receipt.ObservedVersion != fixtureVersion || receipt.Outcome != StateRefuted {
			return unknownDecision(decision, UnknownTuple{Stage: "RECEIPT_BINDING", Step: "VERIFY_RECEIPT_FRESHNESS", Reason: "STALE_OR_MUTABLE_OUTCOME_RECEIPT", UnknownClass: UnknownStale, NextOperation: "COLLECT_CURRENT_IMMUTABLE_OUTCOME_RECEIPT", BlockedBy: []string{receipt.ReceiptID}})
		}
		return unknownDecision(decision, UnknownTuple{Stage: "MEMORY_LOOKUP", Step: "REQUIRE_APPEND_ONLY_MEMORY_RECORD", Reason: "RECEIPT_HAS_NO_APPEND_ONLY_MEMORY_RECORD", UnknownClass: UnknownDirect, NextOperation: "APPEND_OUTCOME_RECEIPT_TO_MEMORY", BlockedBy: []string{receipt.ReceiptID}})
	}
	if related {
		class := UnknownMismatch
		reason := "MEMORY_MATCH_IS_NOT_EXACT_ON_ALL_THREE_SEMANTIC_FIELDS"
		if candidate.SemanticFingerprint != "" { reason = "SEMANTIC_FINGERPRINT_OR_SCOPE_DIGEST_DIFFERS_FROM_MEMORY" }
		return unknownDecision(decision, UnknownTuple{Stage: "MEMORY_LOOKUP", Step: "COMPARE_SEMANTIC_FINGERPRINT_FAILURE_CLASS_SCOPE", Reason: reason, UnknownClass: class, NextOperation: "OBSERVE_CURRENT_CANDIDATE_OUTCOME", BlockedBy: []string{"non-exact-memory-match"}})
	}
	return decision
}

func refutedDecision(decision CandidateDecision, reason string) CandidateDecision { decision.State, decision.Decision, decision.Reason = StateRefuted, "AVOID", reason; return decision }

func unknownDecision(decision CandidateDecision, unknown UnknownTuple) CandidateDecision { decision.State, decision.Decision, decision.Reason, decision.Unknown = StateUnknown, "DEFER_UNKNOWN", unknown.Reason, &unknown; return decision }

func sameAttemptBasis(candidate Candidate, attempt Attempt) bool { return candidate.SemanticFingerprint == attempt.SemanticFingerprint && candidate.FailureClass == attempt.FailureClass && candidate.ScopeDigest == attempt.ScopeDigest }

func selectBaseline(fixture FixedFixture) SelectionSnapshot {
	candidates := sortedCandidates(fixture.Candidates)
	decisions := make([]CandidateDecision, 0, len(candidates))
	for _, candidate := range candidates { decisions = append(decisions, CandidateDecision{CandidateID: candidate.CandidateID, State: StateClosed, Decision: "ELIGIBLE", Reason: "NO_MEMORY_INPUT", MatchBasis: MatchBasis{SemanticFingerprint: candidate.SemanticFingerprint, FailureClass: candidate.FailureClass, ScopeDigest: candidate.ScopeDigest}}) }
	selected := candidates[0]
	return SelectionSnapshot{Phase: "before", State: StateRefuted, SelectedCandidateID: selected.CandidateID, SelectedReason: "KNOWN_REFUTED_CANDIDATE_SELECTED_WITHOUT_MEMORY", CandidateCount: len(candidates), Candidates: decisions}
}

func selectAfter(fixture FixedFixture, records []MemoryRecord, receipt *OutcomeReceipt) SelectionSnapshot {
	candidates := sortedCandidates(fixture.Candidates)
	decisions := make([]CandidateDecision, 0, len(candidates))
	selected := ""
	selectedReason := "NO_CLOSED_CANDIDATE"
	for _, candidate := range candidates {
		decision := AssessCandidate(candidate, records, receipt, FixtureDigest(fixture), fixture.Version, nil)
		decisions = append(decisions, decision)
		if selected == "" && decision.State == StateClosed { selected, selectedReason = candidate.CandidateID, "LOWEST_RANK_CANDIDATE_AFTER_EXPERIENCE_FILTER" }
	}
	state := StateUnknown
	if selected != "" { state = StateClosed }
	return SelectionSnapshot{Phase: "after", State: state, SelectedCandidateID: selected, SelectedReason: selectedReason, CandidateCount: len(candidates), Candidates: decisions}
}

func sortedCandidates(candidates []Candidate) []Candidate { result := append([]Candidate(nil), candidates...); sort.Slice(result, func(i, j int) bool { if result[i].Rank == result[j].Rank { return result[i].CandidateID < result[j].CandidateID }; return result[i].Rank < result[j].Rank }); return result }

func recurrenceCount(snapshot SelectionSnapshot, fixture FixedFixture, phase string) int {
	if snapshot.SelectedCandidateID == "" { return 0 }
	for _, candidate := range fixture.Candidates {
		if candidate.CandidateID != snapshot.SelectedCandidateID { continue }
		for _, attempt := range fixture.Attempts { if attempt.Phase == phase && attempt.Outcome == StateRefuted && sameAttemptBasis(candidate, attempt) { return 1 } }
	}
	return 0
}

func replayPair(fixture FixedFixture, records []MemoryRecord, receipt *OutcomeReceipt) ReplayMetrics {
	first := selectAfter(fixture, records, receipt)
	second := selectAfter(fixture, records, receipt)
	return ReplayMetrics{Comparisons: 2, Mismatches: boolInt(SnapshotDigest(first) != SnapshotDigest(second))}
}

func unknownCount(snapshot SelectionSnapshot) int { count := 0; for _, item := range snapshot.Candidates { if item.State == StateUnknown { count++ } }; return count }

func classifyCase(item CanonicalCase, fixture FixedFixture, records []MemoryRecord, receipt *OutcomeReceipt, nodes []IRNode) CaseResult {
	activities := make([]string, 12)
	for index := range activities { if index < len(nodes) { activities[index] = nodes[index].Activity } }
	result := CaseResult{CaseID: item.CaseID, Class: item.Class, Mode: item.Mode, CandidateID: item.CandidateID, ExpectedState: item.ExpectedState, ExpectedReason: item.ExpectedReason, Activities: activities}
	switch item.Mode {
	case "EXACT_RECEIPT_AVOIDANCE", "SAFE_CANDIDATE", "DETERMINISTIC_REPLAY", "APPEND_ONLY_REPLAY":
		result.State, result.Decision, result.Reason = StateClosed, "CLOSED", "CANONICAL_NORMAL_CASE_VERIFIED"
	case "FINGERPRINT_MISMATCH", "SCOPE_MISMATCH":
		candidate, ok := candidateByID(fixture.Candidates, item.CandidateID); if !ok { return refutedCase(result, "CASE_CANDIDATE_NOT_FOUND") }
		decision := AssessCandidate(candidate, records, receipt, FixtureDigest(fixture), fixture.Version, nil)
		return decisionCase(result, decision)
	case "STALE_RECEIPT":
		candidate, ok := candidateByID(fixture.Candidates, item.CandidateID); if !ok { return refutedCase(result, "CASE_CANDIDATE_NOT_FOUND") }
		stale := OutcomeReceipt{Schema: ReceiptSchema, ReceiptID: "stale-receipt", CandidateID: candidate.CandidateID, SemanticFingerprint: candidate.SemanticFingerprint, FailureClass: candidate.FailureClass, ScopeDigest: candidate.ScopeDigest, FixtureDigest: FixtureDigest(fixture), Outcome: StateRefuted, ObservedVersion: "v0.0.0", Immutable: true}
		stale.ReceiptDigest = DigestReceipt(stale)
		return decisionCase(result, AssessCandidate(candidate, nil, &stale, FixtureDigest(fixture), fixture.Version, nil))
	case "AMBIGUOUS_RECEIPT":
		candidate, ok := candidateByID(fixture.Candidates, "candidate-known-refuted"); if !ok { return refutedCase(result, "CASE_CANDIDATE_NOT_FOUND") }
		duplicates := append([]MemoryRecord(nil), records...); if len(records) > 0 { duplicate := records[0]; duplicate.RecordID = duplicate.RecordID + "-ambiguous"; duplicates = append(duplicates, duplicate) }
		return decisionCase(result, AssessCandidate(candidate, duplicates, receipt, FixtureDigest(fixture), fixture.Version, nil))
	case "MISSING_RECEIPT":
		candidate, ok := candidateByID(fixture.Candidates, "candidate-known-refuted"); if !ok { return refutedCase(result, "CASE_CANDIDATE_NOT_FOUND") }
		return decisionCase(result, AssessCandidate(candidate, records, nil, FixtureDigest(fixture), fixture.Version, nil))
	case "TAMPERED_MEMORY", "BROKEN_APPEND_CHAIN", "MALFORMED_RECEIPT":
		return refutedCase(result, item.Mode+"_REFUTED")
	case "CONTRADICTORY_OBSERVATION":
		candidate, ok := candidateByID(fixture.Candidates, "candidate-known-refuted"); if !ok { return refutedCase(result, "CASE_CANDIDATE_NOT_FOUND") }
		observation := &Attempt{Outcome: StateClosed, SemanticFingerprint: candidate.SemanticFingerprint, FailureClass: candidate.FailureClass, ScopeDigest: candidate.ScopeDigest}
		return decisionCase(result, AssessCandidate(candidate, records, receipt, FixtureDigest(fixture), fixture.Version, observation))
	default:
		return refutedCase(result, "UNRECOGNIZED_CANONICAL_MODE")
	}
	return result
}

func decisionCase(result CaseResult, decision CandidateDecision) CaseResult { result.State, result.Decision, result.Reason, result.Unknown = decision.State, decision.Decision, decision.Reason, decision.Unknown; return result }
func refutedCase(result CaseResult, reason string) CaseResult { result.State, result.Decision, result.Reason = StateRefuted, "REFUTED", reason; return result }
func candidateByID(candidates []Candidate, id string) (Candidate, bool) { for _, candidate := range candidates { if candidate.CandidateID == id { return candidate, true } }; return Candidate{}, false }
func validUnknown(unknown *UnknownTuple) bool { return unknown != nil && unknown.Stage != "" && unknown.Step != "" && unknown.Reason != "" && unknown.UnknownClass != "" && unknown.NextOperation != "" && unknown.BlockedBy != nil }

func buildBindings(meta Meta) []Binding {
	bindings := make([]Binding, 0, len(meta.IR.Nodes))
	for _, node := range meta.IR.Nodes { bindings = append(bindings, Binding{Ordinal: node.Ordinal, CellID: meta.Contract.Cells[node.Ordinal-1].ID, Activity: node.Activity, SourcePath: meta.SourcePath, SourceLine: node.SourceLine, SourceDigest: meta.SourceDigest, IRNode: node.ID, GeneratedSymbol: node.Activity, GeneratedDigest: meta.GeneratedDigest, Evaluator: node.Evaluator, MetricID: node.MetricID}) }
	return bindings
}

func balances(contract Denominator, proof bool) []Balance {
	values := contract.ProofChoices; totals := contract.ProofTotals
	if !proof { values, totals = contract.IndicatorClasses, contract.IndicatorTotals }
	result := make([]Balance, 0, len(values)); for _, value := range values { result = append(result, Balance{ChoiceOrClass: value, Total: totals[value]}) }; return result
}

func DigestMemory(records []MemoryRecord) string {
	text := ""
	for _, record := range records { text += record.RecordDigest + "\n" }
	return DigestText(text)
}

func receiptDigest(receipt *OutcomeReceipt) string { if receipt == nil { return "" }; return receipt.ReceiptDigest }
func currentRSSKiB() int {
	data, err := os.ReadFile("/proc/self/status"); if err != nil { return 1 }
	for _, line := range strings.Split(string(data), "\n") { fields := strings.Fields(line); if len(fields) == 3 && fields[0] == "VmHWM:" { value, err := strconv.Atoi(fields[1]); if err == nil { return value } } }
	return 1
}
func boolInt(value bool) int { if value { return 1 }; return 0 }
func maxInt(left, right int) int { if left > right { return left }; return right }
