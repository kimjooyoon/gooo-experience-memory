package experience

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DigestText(value string) string {
	return DigestBytes([]byte(value))
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func ReceiptPayload(receipt OutcomeReceipt) string {
	return strings.Join([]string{
		receipt.Schema, receipt.ReceiptID, receipt.CandidateID,
		receipt.SemanticFingerprint, receipt.FailureClass, receipt.ScopeDigest,
		receipt.FixtureDigest, receipt.Outcome, receipt.ObservedVersion,
		strconv.FormatBool(receipt.Immutable),
	}, "\x1f")
}

func DigestReceipt(receipt OutcomeReceipt) string {
	receipt.ReceiptDigest = ""
	return DigestText(ReceiptPayload(receipt))
}

func RecordPayload(record MemoryRecord) string {
	return strings.Join([]string{
		record.Schema, record.RecordID, strconv.Itoa(record.Ordinal), record.Kind,
		record.FixtureDigest, record.SemanticFingerprint, record.FailureClass,
		record.ScopeDigest, record.OutcomeReceipt.ReceiptDigest,
		record.PreviousRecordDigest,
	}, "\x1f")
}

func DigestRecord(record MemoryRecord) string {
	record.RecordDigest = ""
	return DigestText(RecordPayload(record))
}

func FixtureDigest(fixture FixedFixture) string {
	return DigestText(strings.Join([]string{
		fixture.Schema, fixture.FixtureID, fixture.Version,
		fixture.ScopeDescriptor, fixture.ScopeDigest,
		strconv.Itoa(len(fixture.Candidates)), strconv.Itoa(len(fixture.Attempts)),
	}, "\x1f"))
}

func DigestIR(ir SemanticIR) string {
	data, err := json.Marshal(ir)
	if err != nil {
		return DigestText(fmt.Sprintf("%v", ir))
	}
	return DigestBytes(data)
}

func SnapshotDigest(snapshot SelectionSnapshot) string {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return DigestText(fmt.Sprintf("%v", snapshot))
	}
	return DigestBytes(data)
}
