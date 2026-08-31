# RFC: append-only experience memory for Gooo

## Scope

This repository evaluates one narrow self-improvement loop: a known refuted
candidate must not recur after an immutable outcome receipt is added. The
fixture is deliberately fixed so the result can be stated as exact integers,
not as a score or percentage.

## Semantic identity

`candidate_id` is a label. It is never sufficient for avoidance. The semantic
identity is the tuple:

```text
(semantic_fingerprint, failure_class, scope_digest)
```

The fixed fixture digest and receipt digest bind that tuple to one observed
context and one immutable outcome. A candidate may be avoided only if every
coordinate and binding matches. Related evidence with only a partial match is
not silently generalized.

## Resolution

`REFUTED > UNKNOWN > CLOSED` is the report precedence. `UNKNOWN` is a
structured, actionable result with six required fields:
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
`blocked_by`. `DIRECT_MISSING`, `STALE_INPUT`, `AMBIGUOUS_EVIDENCE`, and
`SEMANTIC_SCOPE_MISMATCH` remain distinguishable.

## Metaprogramming boundary

The `.gooo` program declares all twelve meta-activities, proof axes, indicator
classes, metric IDs, artifacts, and evaluator names. Compilation lowers those
declarations to semantic IR. Generation projects the IR to Go markers and
bindings. Evaluation verifies every source activity has exactly one IR node,
one generated binding marker, and one evaluator binding. Handwritten evaluator
logic may interpret evidence, but it cannot introduce an unbound activity.

## Exact evidence

The report includes attempts observed, memory records, candidate count, the
before/after recurrence pair, avoided candidates, new unknown candidates,
replay comparisons and mismatches, peak RSS, wall time, Go/Gooo inventory,
descendant directories, and test status counts. Input repositories and other
tools are consumed through read-only digests. Generated output is caller-owned.

