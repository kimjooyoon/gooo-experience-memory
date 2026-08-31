# Experience memory protocol v1

The loop has two selection observations over one fixed fixture.

1. The baseline has no memory input. Candidates are ordered by their declared
   rank; the first candidate is selected even though the fixed fixture records
   its observed result as `REFUTED`.
2. The after observation loads an append-only memory chain and an immutable
   outcome receipt. A candidate is avoided only when its semantic fingerprint,
   failure class, and scope digest all match the memory record, and the receipt
   is bound to the same fixture and current version.
3. Candidates with related but non-exact evidence are `UNKNOWN`; they are not
   treated as safe and are not selected. A stale or ambiguous receipt follows
   the same conservative path.
4. A current `CLOSED` observation that contradicts an exact known refutation is
   `REFUTED`, regardless of missing or ambiguous evidence.

Memory records are append-only. Every record carries an ordinal, previous
record digest, its own record digest, and the receipt digest. The receipt
itself is immutable evidence of the `REFUTED` outcome. Neither the evaluator
nor the conformance script edits the source repository.

The report's `state=CLOSED` means the evaluator closed the conformance proof;
it does not turn the four canonical `UNKNOWN` or four canonical `REFUTED`
examples into successful candidate outcomes. Those examples prove that the
resolution lattice and precedence rules are exercised.

