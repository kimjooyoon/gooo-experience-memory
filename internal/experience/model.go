package experience

import "encoding/json"

const (
	SourceSchema     = "gooo/experience-memory/source/v1"
	IRSchema         = "gooo/experience-memory/semantic-ir/v1"
	ReportSchema     = "gooo/experience-memory/report/v1"
	FixtureSchema    = "gooo/experience-memory/fixed-fixture/v1"
	CaseSchema       = "gooo/experience-memory/case/v1"
	MemorySchema     = "gooo/experience-memory/memory-record/v1"
	ReceiptSchema    = "gooo/experience-memory/outcome-receipt/v1"
	StateClosed      = "CLOSED"
	StateUnknown     = "UNKNOWN"
	StateRefuted     = "REFUTED"
	ProofFoundation  = "FOUNDATION"
	ProofCoherence   = "COHERENCE"
	ProofRegression  = "REGRESSION"
	UnknownDirect    = "DIRECT_MISSING"
	UnknownStale     = "STALE_INPUT"
	UnknownAmbiguous = "AMBIGUOUS_EVIDENCE"
	UnknownMismatch  = "SEMANTIC_SCOPE_MISMATCH"
)

var Precedence = []string{StateRefuted, StateUnknown, StateClosed}

type Authority struct {
	RepositoryWrites          int `json:"repository_writes"`
	LocalTestExecutions       int `json:"local_test_executions"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
}

type SourceProgram struct {
	Schema       string           `json:"schema"`
	Model        string           `json:"model"`
	ModelVersion string           `json:"model_version"`
	Authority    Authority        `json:"authority"`
	Activities   []SourceActivity `json:"activities"`
}

type SourceActivity struct {
	Ordinal    int    `json:"ordinal"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Proof      string `json:"proof_choice"`
	Indicator  string `json:"indicator_class"`
	Metric     string `json:"metric_id"`
	Artifact   string `json:"artifact"`
	Evaluator  string `json:"evaluator"`
	SourceLine int    `json:"source_line"`
}

type Cell struct {
	Ordinal        int    `json:"ordinal"`
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	MetricID       string `json:"metric_id"`
	MetricPath     string `json:"metric_path"`
	Artifact       string `json:"artifact"`
	Evaluator      string `json:"evaluator"`
}

type Denominator struct {
	Schema           string         `json:"schema"`
	DenominatorID    string         `json:"denominator_id"`
	Total            int            `json:"total"`
	Fixed            bool           `json:"fixed"`
	ProofChoices     []string       `json:"proof_choices"`
	IndicatorClasses []string       `json:"indicator_classes"`
	ProofTotals      map[string]int `json:"proof_totals"`
	IndicatorTotals  map[string]int `json:"indicator_totals"`
	Precedence       []string       `json:"precedence"`
	UnknownFields    []string       `json:"unknown_fields"`
	NoScores         bool           `json:"no_scores"`
	NoPercentages    bool           `json:"no_percentages"`
	NoInference      bool           `json:"no_inference"`
	Cells            []Cell         `json:"cells"`
}

type IRNode struct {
	Ordinal        int    `json:"ordinal"`
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	SourceLine     int    `json:"source_line"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	MetricID       string `json:"metric_id"`
	MetricPath     string `json:"metric_path"`
	Artifact       string `json:"artifact"`
	Evaluator      string `json:"evaluator"`
}

type SemanticIR struct {
	Schema         string   `json:"schema"`
	Model          string   `json:"model"`
	SourcePath     string   `json:"source_path"`
	SourceDigest   string   `json:"source_digest"`
	ContractDigest string   `json:"contract_digest"`
	Nodes          []IRNode `json:"nodes"`
}

type Candidate struct {
	CandidateID         string `json:"candidate_id"`
	Semantic            string `json:"semantic"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	FailureClass        string `json:"failure_class"`
	ScopeDigest         string `json:"scope_digest"`
	Rank                int    `json:"rank"`
}

type Attempt struct {
	AttemptID           string `json:"attempt_id"`
	Phase               string `json:"phase"`
	CandidateID         string `json:"candidate_id"`
	Outcome             string `json:"outcome"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	FailureClass        string `json:"failure_class"`
	ScopeDigest         string `json:"scope_digest"`
	Observed            bool   `json:"observed"`
}

type FixedFixture struct {
	Schema          string      `json:"schema"`
	FixtureID       string      `json:"fixture_id"`
	Version         string      `json:"version"`
	ScopeDescriptor string      `json:"scope_descriptor"`
	ScopeDigest     string      `json:"scope_digest"`
	Candidates      []Candidate `json:"candidates"`
	Attempts        []Attempt   `json:"attempts"`
}

type OutcomeReceipt struct {
	Schema              string `json:"schema"`
	ReceiptID           string `json:"receipt_id"`
	CandidateID         string `json:"candidate_id"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	FailureClass        string `json:"failure_class"`
	ScopeDigest         string `json:"scope_digest"`
	FixtureDigest       string `json:"fixture_digest"`
	Outcome             string `json:"outcome"`
	ObservedVersion     string `json:"observed_version"`
	Immutable           bool   `json:"immutable"`
	ReceiptDigest       string `json:"receipt_digest"`
}

type MemoryRecord struct {
	Schema               string         `json:"schema"`
	RecordID             string         `json:"record_id"`
	Ordinal              int            `json:"ordinal"`
	Kind                 string         `json:"kind"`
	FixtureDigest        string         `json:"fixture_digest"`
	SemanticFingerprint  string         `json:"semantic_fingerprint"`
	FailureClass         string         `json:"failure_class"`
	ScopeDigest          string         `json:"scope_digest"`
	OutcomeReceipt       OutcomeReceipt `json:"outcome_receipt"`
	PreviousRecordDigest string         `json:"previous_record_digest"`
	RecordDigest         string         `json:"record_digest"`
}

type UnknownTuple struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CanonicalCase struct {
	Schema         string        `json:"schema"`
	CaseID         string        `json:"case_id"`
	Class          string        `json:"class"`
	Mode           string        `json:"mode"`
	CandidateID    string        `json:"candidate_id"`
	ExpectedState  string        `json:"expected_state"`
	ExpectedReason string        `json:"expected_reason"`
	Unknown        *UnknownTuple `json:"unknown"`
}

type CaseResult struct {
	CaseID         string        `json:"case_id"`
	Class          string        `json:"class"`
	Mode           string        `json:"mode"`
	CandidateID    string        `json:"candidate_id"`
	State          string        `json:"state"`
	Decision       string        `json:"decision"`
	Reason         string        `json:"reason"`
	ExpectedState  string        `json:"expected_state"`
	ExpectedReason string        `json:"expected_reason"`
	Unknown        *UnknownTuple `json:"unknown"`
	Activities     []string      `json:"activities"`
}

type MatchBasis struct {
	SemanticFingerprint string `json:"semantic_fingerprint"`
	FailureClass        string `json:"failure_class"`
	ScopeDigest         string `json:"scope_digest"`
	AllThreeExact       bool   `json:"all_three_exact"`
}

type CandidateDecision struct {
	CandidateID string        `json:"candidate_id"`
	State       string        `json:"state"`
	Decision    string        `json:"decision"`
	Reason      string        `json:"reason"`
	MatchBasis  MatchBasis    `json:"match_basis"`
	Unknown     *UnknownTuple `json:"unknown"`
}

type SelectionSnapshot struct {
	Phase               string              `json:"phase"`
	State               string              `json:"state"`
	SelectedCandidateID string              `json:"selected_candidate_id"`
	SelectedReason      string              `json:"selected_reason"`
	CandidateCount      int                 `json:"candidate_count"`
	Candidates          []CandidateDecision `json:"candidates"`
}

type ReplayMetrics struct {
	Comparisons int `json:"comparisons"`
	Mismatches  int `json:"mismatches"`
}

type TestMetrics struct {
	Total       int `json:"total"`
	Executed    int `json:"executed"`
	Reused      int `json:"reused"`
	Skipped     int `json:"skipped"`
	NotObserved int `json:"not_observed"`
}

type InventoryMetrics struct {
	GoPhysicalLines   int `json:"go_physical_lines"`
	GoFiles           int `json:"go_files"`
	GoooPhysicalLines int `json:"gooo_physical_lines"`
	GoooFiles         int `json:"gooo_files"`
	DescendantDirs    int `json:"descendant_dirs"`
	RegularFiles      int `json:"regular_files_root_readme_excluded"`
}

type Metrics struct {
	AttemptsObserved              int              `json:"attempts_observed"`
	MemoryRecords                 int              `json:"memory_records"`
	CandidateCount                int              `json:"candidate_count"`
	KnownRefutedRecurrencesBefore int              `json:"known_refuted_recurrences_before"`
	KnownRefutedRecurrencesAfter  int              `json:"known_refuted_recurrences_after"`
	AvoidedRefutedCandidates      int              `json:"avoided_refuted_candidates"`
	NewUnknownCandidates          int              `json:"new_unknown_candidates"`
	ReplayComparisons             int              `json:"replay_comparisons"`
	ReplayMismatches              int              `json:"replay_mismatches"`
	PeakRSSKiB                    int              `json:"peak_rss_kib"`
	WallMS                        int              `json:"wall_ms"`
	GoPhysicalLines               int              `json:"go_physical_lines"`
	GoFiles                       int              `json:"go_files"`
	GoooPhysicalLines             int              `json:"gooo_physical_lines"`
	GoooFiles                     int              `json:"gooo_files"`
	DescendantDirs                int              `json:"descendant_dirs"`
	Tests                         TestMetrics      `json:"tests"`
	Inventory                     InventoryMetrics `json:"inventory"`
	Authority                     Authority        `json:"authority"`
}

type Binding struct {
	Ordinal         int    `json:"ordinal"`
	CellID          string `json:"cell_id"`
	Activity        string `json:"activity"`
	SourcePath      string `json:"source_path"`
	SourceLine      int    `json:"source_line"`
	SourceDigest    string `json:"source_digest"`
	IRNode          string `json:"ir_node"`
	GeneratedSymbol string `json:"generated_symbol"`
	GeneratedDigest string `json:"generated_go_digest"`
	Evaluator       string `json:"evaluator"`
	MetricID        string `json:"metric_id"`
}

type Balance struct {
	ChoiceOrClass string `json:"choice_or_class"`
	Total         int    `json:"total"`
}

type Report struct {
	Schema                 string            `json:"schema"`
	Decision               string            `json:"decision"`
	State                  string            `json:"state"`
	Precedence             []string          `json:"precedence"`
	SourcePath             string            `json:"source_path"`
	SourceDigest           string            `json:"source_digest"`
	SemanticIRDigest       string            `json:"semantic_ir_digest"`
	GeneratedGoDigest      string            `json:"generated_go_digest"`
	ContractDigest         string            `json:"contract_digest"`
	FixtureDigest          string            `json:"fixture_digest"`
	MemoryDigest           string            `json:"memory_digest"`
	ReceiptDigest          string            `json:"receipt_digest"`
	FixedDenominator       int               `json:"fixed_denominator"`
	Proofs                 []Balance         `json:"proofs"`
	Indicators             []Balance         `json:"indicators"`
	Bindings               []Binding         `json:"bindings"`
	CanonicalCases         []CaseResult      `json:"canonical_cases"`
	CanonicalCounts        map[string]int    `json:"canonical_counts"`
	Baseline               SelectionSnapshot `json:"baseline"`
	After                  SelectionSnapshot `json:"after"`
	Replay                 ReplayMetrics     `json:"replay"`
	Metrics                Metrics           `json:"metrics"`
	Authority              Authority         `json:"authority"`
	AppendOnly             bool              `json:"append_only"`
	ExternalInputsReadOnly bool              `json:"external_inputs_read_only"`
	RootReadmeExcluded     bool              `json:"root_readme_excluded"`
}

func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
