# gooo-experience-memory

`gooo-experience-memory` is a metaprogrammed, append-only experience memory
for the Gooo self-improvement loop. A previous known-`REFUTED` outcome is
carried forward as an immutable receipt and can prevent the same semantic
failure from being selected again.

The authority chain is explicit and one-to-one:

```text
.gooo source -> semantic IR -> generated Go -> evaluator
```

The evaluator never treats a candidate ID as proof. Avoidance requires an
exact match on all three semantic coordinates: `semantic_fingerprint`,
`failure_class`, and `scope_digest`, plus the fixed fixture digest and a
fresh immutable outcome receipt. A fingerprint or scope mismatch, stale
receipt, or ambiguous receipt becomes `UNKNOWN`. A current observation that
contradicts a known refutation is `REFUTED` and takes precedence over
`UNKNOWN`.

## Closed fixture

The fixed fixture contains five candidates. With no memory, the rank-one
candidate is selected and its known-`REFUTED` recurrence is `1`. After the
append-only memory record and immutable outcome receipt are supplied, that
candidate is avoided and the safe candidate is selected. The exact integer
pair is:

```text
known_refuted_recurrences_before/after = 1/0
avoided_refuted_candidates = 1
new_unknown_candidates = 2
replay_comparisons/mismatches = 2/0
```

The canonical case corpus has 12 cases: 4 normal, 4 `UNKNOWN`, and 4
`REFUTED`. The denominator has exactly 12 cells with proof balances
`FOUNDATION/COHERENCE/REGRESSION = 4/4/4` and indicator balances
`DRIVER/OUTCOME/GUARDRAIL = 4/4/4`.

All `UNKNOWN` results preserve `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`. No score, percentage,
or inference-based improvement claim is emitted.

## Authority and verification

The source repository and all supplied fixtures are read-only inputs. Compile,
generation, reports, and evidence assets are written only to caller-owned
absolute output paths. The conformance evidence records
`repository_writes=0`, `local_test_executions=0`, and
`cross_project_required_gates=0`. The root README is excluded from physical
inventory counts.

The only verification boundary is GitHub Actions with Go 1.27. It runs format,
vet, tests, race tests, build, and the fixed conformance script. Local Go
tests are intentionally not part of the workflow contract.

The release workflow requires an annotated tag and publishes an archive,
`SHA256SUMS`, and an immutable release manifest without overwriting an
existing tag or release.

## Commands

The CLI has three stages. All output paths are required to be absolute
caller-owned paths.

```text
go run ./cmd/gooo-experience-memory compile --source examples/experience-memory/main.gooo --contract contracts/experience-memory-denominator-v1.json --output /tmp/semantic-ir.json
go run ./cmd/gooo-experience-memory generate --ir /tmp/semantic-ir.json --output /tmp/semantic.gooo.go
go run ./cmd/gooo-experience-memory evaluate --source examples/experience-memory/main.gooo --contract contracts/experience-memory-denominator-v1.json --ir /tmp/semantic-ir.json --generated /tmp/semantic.gooo.go --fixture fixtures/fixed-fixture.json --memory fixtures/memory.ndjson --receipt fixtures/outcome-receipt.json --cases fixtures/cases --output-dir /tmp/gooo-experience-memory-evidence
```

See [docs/protocol-v1.md](docs/protocol-v1.md) and
[docs/rfc-v1.md](docs/rfc-v1.md) for the data contract. The checked-in
workflow is the verification authority.

