# v0.1.0 release checklist

The release is valid only after the single implementation pull request is
merged and the Go 1.27 conformance workflow succeeds on the merge commit.

- repository: public `kimjooyoon/gooo-experience-memory`
- tag: annotated immutable `v0.1.0`
- assets: source archive, `SHA256SUMS`, and `release-manifest.json`
- manifest: binds tag, peeled commit SHA, archive name, archive digest, and
  `immutable=true`
- evidence: the CI artifact contains the semantic IR, generated Go, report,
  fixture, memory chain, receipt, runtime metrics, and per-file digests

If a repository or tag already exists, the operator must inspect it first and
use the next patch version. Existing releases and tags must never be replaced.
The release workflow itself uses `--verify-tag` and does not delete or update
release assets.

