package experience

import "testing"

func TestAssessCandidateRequiresTheCompleteSemanticMatch(t *testing.T) {
	fixtureDigest := DigestText("fixed-fixture")
	scope := DigestText("scope")
	fingerprint := DigestText("semantic")
	receipt := OutcomeReceipt{Schema: ReceiptSchema, ReceiptID: "receipt-1", CandidateID: "display-only", SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope, FixtureDigest: fixtureDigest, Outcome: StateRefuted, ObservedVersion: "v1", Immutable: true}
	receipt.ReceiptDigest = DigestReceipt(receipt)
	record := MemoryRecord{Schema: MemorySchema, RecordID: "record-1", Ordinal: 1, Kind: StateRefuted, FixtureDigest: fixtureDigest, SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope, OutcomeReceipt: receipt}
	record.RecordDigest = DigestRecord(record)
	candidate := Candidate{CandidateID: "new-display-id", Semantic: "semantic", SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope, Rank: 1}
	decision := AssessCandidate(candidate, []MemoryRecord{record}, &receipt, fixtureDigest, "v1", nil)
	if decision.State != StateRefuted || decision.Decision != "AVOID" || !decision.MatchBasis.AllThreeExact {
		t.Fatalf("exact match = %#v", decision)
	}
	drift := candidate
	drift.SemanticFingerprint = DigestText("different-semantic")
	driftDecision := AssessCandidate(drift, []MemoryRecord{record}, &receipt, fixtureDigest, "v1", nil)
	if driftDecision.State != StateUnknown || driftDecision.Unknown == nil || driftDecision.Unknown.UnknownClass != UnknownMismatch {
		t.Fatalf("fingerprint drift = %#v", driftDecision)
	}
}

func TestStaleReceiptIsUnknownAndContradictionIsRefuted(t *testing.T) {
	fixtureDigest := DigestText("fixed-fixture")
	scope := DigestText("scope")
	fingerprint := DigestText("semantic")
	receipt := OutcomeReceipt{Schema: ReceiptSchema, ReceiptID: "receipt-1", CandidateID: "candidate", SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope, FixtureDigest: fixtureDigest, Outcome: StateRefuted, ObservedVersion: "v0", Immutable: true}
	receipt.ReceiptDigest = DigestReceipt(receipt)
	candidate := Candidate{CandidateID: "candidate", Semantic: "semantic", SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope}
	stale := AssessCandidate(candidate, nil, &receipt, fixtureDigest, "v1", nil)
	if stale.State != StateUnknown || stale.Unknown == nil || stale.Unknown.UnknownClass != UnknownStale {
		t.Fatalf("stale = %#v", stale)
	}
	record := MemoryRecord{Schema: MemorySchema, RecordID: "record-1", Ordinal: 1, Kind: StateRefuted, FixtureDigest: fixtureDigest, SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope, OutcomeReceipt: receipt}
	record.RecordDigest = DigestRecord(record)
	contradiction := AssessCandidate(candidate, []MemoryRecord{record}, &receipt, fixtureDigest, "v0", &Attempt{Outcome: StateClosed, SemanticFingerprint: fingerprint, FailureClass: "KNOWN", ScopeDigest: scope})
	if contradiction.State != StateRefuted || contradiction.Reason != "CURRENT_OBSERVATION_CONTRADICTS_KNOWN_REFUTATION" {
		t.Fatalf("contradiction = %#v", contradiction)
	}
}
