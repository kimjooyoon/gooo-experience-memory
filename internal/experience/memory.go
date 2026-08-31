package experience

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func LoadReceipt(path string) (OutcomeReceipt, error) {
	var receipt OutcomeReceipt
	if err := LoadJSON(path, &receipt); err != nil {
		return OutcomeReceipt{}, err
	}
	if err := ValidateReceipt(receipt); err != nil {
		return OutcomeReceipt{}, err
	}
	return receipt, nil
}

func ValidateReceipt(receipt OutcomeReceipt) error {
	if receipt.Schema != ReceiptSchema || receipt.ReceiptID == "" || receipt.CandidateID == "" || !validDigest(receipt.SemanticFingerprint) || receipt.FailureClass == "" || !validDigest(receipt.ScopeDigest) || !validDigest(receipt.FixtureDigest) || receipt.Outcome != StateRefuted || receipt.ObservedVersion == "" || !validDigest(receipt.ReceiptDigest) {
		return errors.New("outcome receipt is incomplete")
	}
	if DigestReceipt(receipt) != receipt.ReceiptDigest {
		return errors.New("outcome receipt digest mismatch")
	}
	return nil
}

func LoadMemory(path string, fixtureDigest string) ([]MemoryRecord, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	result := make([]MemoryRecord, 0)
	seen := map[string]bool{}
	previous := ""
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var record MemoryRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, "", fmt.Errorf("decode memory record: %w", err)
		}
		if record.Schema != MemorySchema || record.RecordID == "" || record.Ordinal != len(result)+1 || seen[record.RecordID] || record.FixtureDigest != fixtureDigest || record.Kind != StateRefuted {
			return nil, "", errors.New("memory is not a valid append-only ledger")
		}
		if record.PreviousRecordDigest != previous || !validDigest(record.RecordDigest) || DigestRecord(record) != record.RecordDigest {
			return nil, "", fmt.Errorf("memory digest chain failed at %s", record.RecordID)
		}
		if record.OutcomeReceipt.FixtureDigest != fixtureDigest || record.OutcomeReceipt.ReceiptDigest != DigestReceipt(record.OutcomeReceipt) {
			return nil, "", fmt.Errorf("memory receipt digest failed at %s", record.RecordID)
		}
		if record.SemanticFingerprint != record.OutcomeReceipt.SemanticFingerprint || record.FailureClass != record.OutcomeReceipt.FailureClass || record.ScopeDigest != record.OutcomeReceipt.ScopeDigest {
			return nil, "", fmt.Errorf("memory match basis failed at %s", record.RecordID)
		}
		seen[record.RecordID] = true
		result = append(result, record)
		previous = record.RecordDigest
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return result, DigestBytes(data), nil
}
