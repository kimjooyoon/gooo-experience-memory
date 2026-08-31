package experience

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func ParseProgram(source []byte) (SourceProgram, error) {
	program := SourceProgram{Schema: SourceSchema}
	seen := map[int]bool{}
	scanner := bufio.NewScanner(strings.NewReader(string(source)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "program":
			if len(fields) != 3 || fields[1] != "gooo-experience-memory" || fields[2] != "v1" {
				return SourceProgram{}, fmt.Errorf("line %d: invalid program declaration", lineNumber)
			}
		case "model":
			if len(fields) != 3 {
				return SourceProgram{}, fmt.Errorf("line %d: model requires name and version", lineNumber)
			}
			program.Model, program.ModelVersion = fields[1], fields[2]
		case "authority":
			values, err := keyValues(fields[1:])
			if err != nil {
				return SourceProgram{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			program.Authority, err = parseAuthority(values)
			if err != nil {
				return SourceProgram{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "activity":
			values, err := keyValues(fields[1:])
			if err != nil {
				return SourceProgram{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			activity, err := parseActivity(values, lineNumber)
			if err != nil {
				return SourceProgram{}, err
			}
			if seen[activity.Ordinal] {
				return SourceProgram{}, fmt.Errorf("line %d: duplicate activity ordinal %d", lineNumber, activity.Ordinal)
			}
			seen[activity.Ordinal] = true
			program.Activities = append(program.Activities, activity)
		default:
			return SourceProgram{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return SourceProgram{}, err
	}
	if program.Model != "ExperienceMemory" || program.ModelVersion != "v1" {
		return SourceProgram{}, errors.New("source must declare model ExperienceMemory v1")
	}
	if len(program.Activities) != 12 {
		return SourceProgram{}, fmt.Errorf("source must declare exactly 12 activities, got %d", len(program.Activities))
	}
	sort.Slice(program.Activities, func(i, j int) bool { return program.Activities[i].Ordinal < program.Activities[j].Ordinal })
	for index, activity := range program.Activities {
		if activity.Ordinal != index+1 {
			return SourceProgram{}, fmt.Errorf("activity ordinals are not contiguous at %d", index+1)
		}
	}
	return program, nil
}

func keyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || values[parts[0]] != "" {
			return nil, fmt.Errorf("invalid or duplicate key-value %q", field)
		}
		values[parts[0]] = parts[1]
	}
	return values, nil
}

func parseAuthority(values map[string]string) (Authority, error) {
	read := func(name string) (int, error) {
		value, ok := values[name]
		if !ok {
			return 0, fmt.Errorf("authority missing %s", name)
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return 0, fmt.Errorf("authority %s must be a non-negative integer", name)
		}
		return parsed, nil
	}
	writes, err := read("repository_writes")
	if err != nil {
		return Authority{}, err
	}
	local, err := read("local_test_executions")
	if err != nil {
		return Authority{}, err
	}
	gates, err := read("cross_project_required_gates")
	if err != nil {
		return Authority{}, err
	}
	return Authority{RepositoryWrites: writes, LocalTestExecutions: local, CrossProjectRequiredGates: gates}, nil
}

func parseActivity(values map[string]string, line int) (SourceActivity, error) {
	read := func(name string) (string, error) {
		value := values[name]
		if value == "" {
			return "", fmt.Errorf("line %d: activity missing %s", line, name)
		}
		return value, nil
	}
	ordinalText, err := read("ordinal")
	if err != nil {
		return SourceActivity{}, err
	}
	ordinal, err := strconv.Atoi(ordinalText)
	if err != nil || ordinal < 1 {
		return SourceActivity{}, fmt.Errorf("line %d: activity ordinal is invalid", line)
	}
	id, err := read("id")
	if err != nil {
		return SourceActivity{}, err
	}
	name, err := read("name")
	if err != nil {
		return SourceActivity{}, err
	}
	proof, err := read("proof")
	if err != nil {
		return SourceActivity{}, err
	}
	indicator, err := read("indicator")
	if err != nil {
		return SourceActivity{}, err
	}
	metric, err := read("metric")
	if err != nil {
		return SourceActivity{}, err
	}
	artifact, err := read("artifact")
	if err != nil {
		return SourceActivity{}, err
	}
	evaluator, err := read("evaluator")
	if err != nil {
		return SourceActivity{}, err
	}
	return SourceActivity{Ordinal: ordinal, ID: id, Name: name, Proof: proof, Indicator: indicator, Metric: metric, Artifact: artifact, Evaluator: evaluator, SourceLine: line}, nil
}

func LoadDenominator(path string) (Denominator, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Denominator{}, "", err
	}
	var contract Denominator
	if err := json.Unmarshal(data, &contract); err != nil {
		return Denominator{}, "", err
	}
	if err := ValidateDenominator(contract); err != nil {
		return Denominator{}, "", err
	}
	return contract, DigestBytes(data), nil
}

func ValidateDenominator(contract Denominator) error {
	if contract.Schema != "gooo/experience-memory/denominator/v1" || contract.DenominatorID == "" || contract.Total != 12 || !contract.Fixed || len(contract.Cells) != 12 {
		return errors.New("denominator must be fixed at exactly 12 cells")
	}
	if len(contract.ProofChoices) != 3 || len(contract.IndicatorClasses) != 3 || len(contract.Precedence) != 3 || contract.Precedence[0] != StateRefuted || contract.Precedence[1] != StateUnknown || contract.Precedence[2] != StateClosed {
		return errors.New("denominator precedence or axes are invalid")
	}
	for _, field := range []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"} {
		found := false
		for _, actual := range contract.UnknownFields {
			if actual == field {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("unknown field %q is not preserved", field)
		}
	}
	if !contract.NoScores || !contract.NoPercentages || !contract.NoInference {
		return errors.New("score, percentage, and inference claims must be disabled")
	}
	proofs := map[string]int{}
	indicators := map[string]int{}
	seenIDs := map[string]bool{}
	for index, cell := range contract.Cells {
		if cell.Ordinal != index+1 || cell.ID == "" || cell.Activity == "" || cell.MetricID == "" || cell.Artifact == "" || cell.Evaluator == "" || seenIDs[cell.ID] {
			return fmt.Errorf("invalid denominator cell at ordinal %d", index+1)
		}
		seenIDs[cell.ID] = true
		proofs[cell.ProofChoice]++
		indicators[cell.IndicatorClass]++
	}
	for _, choice := range contract.ProofChoices {
		if proofs[choice] != 4 || contract.ProofTotals[choice] != 4 {
			return fmt.Errorf("proof balance for %s is not 4/4", choice)
		}
	}
	for _, class := range contract.IndicatorClasses {
		if indicators[class] != 4 || contract.IndicatorTotals[class] != 4 {
			return fmt.Errorf("indicator balance for %s is not 4/4", class)
		}
	}
	return nil
}

func CompileSource(sourcePath string, source []byte, contract Denominator, contractDigest string) (SemanticIR, error) {
	if err := ValidateDenominator(contract); err != nil {
		return SemanticIR{}, err
	}
	program, err := ParseProgram(source)
	if err != nil {
		return SemanticIR{}, err
	}
	if program.Authority != (Authority{}) {
		return SemanticIR{}, errors.New("source authority must be zero")
	}
	if contractDigest == "" {
		return SemanticIR{}, errors.New("contract digest is required")
	}
	ir := SemanticIR{Schema: IRScheme, Model: program.Model, SourcePath: sourcePath, SourceDigest: DigestBytes(source), ContractDigest: contractDigest, Nodes: make([]IRNode, 0, 12)}
	for index, cell := range contract.Cells {
		activity := program.Activities[index]
		if activity.Name != cell.Activity || activity.ID != cell.ID || activity.Proof != cell.ProofChoice || activity.Indicator != cell.IndicatorClass || activity.Metric != cell.MetricID || activity.Artifact != cell.Artifact || activity.Evaluator != cell.Evaluator {
			return SemanticIR{}, fmt.Errorf("source to contract binding failed at cell %d", cell.Ordinal)
		}
		ir.Nodes = append(ir.Nodes, IRNode{Ordinal: cell.Ordinal, ID: fmt.Sprintf("ir/%02d/%s", cell.Ordinal, cell.ID), Activity: activity.Name, SourceLine: activity.SourceLine, ProofChoice: cell.ProofChoice, IndicatorClass: cell.IndicatorClass, MetricID: cell.MetricID, MetricPath: cell.MetricPath, Artifact: cell.Artifact, Evaluator: cell.Evaluator})
	}
	return ir, nil
}

func LoadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func LoadFixture(path string) (FixedFixture, error) {
	var fixture FixedFixture
	if err := LoadJSON(path, &fixture); err != nil {
		return FixedFixture{}, err
	}
	if fixture.Schema != FixtureSchema || fixture.FixtureID == "" || fixture.Version == "" || fixture.ScopeDescriptor == "" || !validDigest(fixture.ScopeDigest) || fixture.ScopeDigest != DigestText(fixture.ScopeDescriptor) || len(fixture.Candidates) != 5 || len(fixture.Attempts) != 2 {
		return FixedFixture{}, errors.New("fixed fixture identity, scope, or cardinality is invalid")
	}
	seen := map[string]bool{}
	for _, candidate := range fixture.Candidates {
		if candidate.CandidateID == "" || candidate.Semantic == "" || !validDigest(candidate.SemanticFingerprint) || candidate.SemanticFingerprint != DigestText(candidate.Semantic) || candidate.FailureClass == "" || !validDigest(candidate.ScopeDigest) || candidate.Rank < 1 || seen[candidate.CandidateID] {
			return FixedFixture{}, fmt.Errorf("invalid candidate %q", candidate.CandidateID)
		}
		seen[candidate.CandidateID] = true
	}
	for _, attempt := range fixture.Attempts {
		if !attempt.Observed || attempt.Phase == "" || attempt.Outcome == "" || !validDigest(attempt.SemanticFingerprint) || attempt.ScopeDigest == "" {
			return FixedFixture{}, errors.New("invalid observed attempt")
		}
	}
	return fixture, nil
}

func LoadCases(path string) ([]CanonicalCase, error) {
	paths, err := filepath.Glob(filepath.Join(path, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	if len(paths) != 12 {
		return nil, fmt.Errorf("canonical case denominator requires exactly 12 files, found %d", len(paths))
	}
	result := make([]CanonicalCase, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		var item CanonicalCase
		if err := LoadJSON(path, &item); err != nil {
			return nil, err
		}
		if item.Schema != CaseSchema || item.CaseID == "" || seen[item.CaseID] || item.Class == "" || item.Mode == "" || item.ExpectedState == "" {
			return nil, fmt.Errorf("invalid canonical case %s", path)
		}
		seen[item.CaseID] = true
		result = append(result, item)
	}
	return result, nil
}
