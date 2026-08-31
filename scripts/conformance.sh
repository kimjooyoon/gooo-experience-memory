#!/usr/bin/env bash
set -Eeuo pipefail

if test "$#" -ne 6; then
  echo "usage: conformance.sh BINARY SUBJECT_SHA GO_VERSION GO_TEST_JSON TEST_MS BUILD_MS" >&2
  exit 2
fi

binary=$1
subject_sha=$2
go_version=$3
go_test_json=$4
test_ms=$5
build_ms=$6
artifact_root=${RUNNER_TEMP:-/tmp}/gooo-experience-memory-artifact
work_root=${RUNNER_TEMP:-/tmp}/gooo-experience-memory-work
rm -rf "$artifact_root" "$work_root"
mkdir -p "$artifact_root/generated" "$work_root"

before=$(git status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')

"$binary" compile --source examples/experience-memory/main.gooo --contract contracts/experience-memory-denominator-v1.json --output "$work_root/semantic-ir.json" > "$work_root/compile.json"
"$binary" generate --ir "$work_root/semantic-ir.json" --output "$work_root/semantic.gooo.go" > "$work_root/generate.json"

go_files=0
go_lines=0
gooo_files=0
gooo_lines=0
regular_files=0
while IFS= read -r -d '' file; do
  case "$file" in
    ./README.md) continue ;;
  esac
  regular_files=$((regular_files + 1))
  case "$file" in
    *.go) go_files=$((go_files + 1)); go_lines=$((go_lines + $(awk 'END {print NR+0}' "$file"))) ;;
    *.gooo) gooo_files=$((gooo_files + 1)); gooo_lines=$((gooo_lines + $(awk 'END {print NR+0}' "$file"))) ;;
  esac
done < <(find . -type f ! -path './.git/*' ! -path './.git' -print0 | sort -z)
descendant_dirs=$(find . -mindepth 1 -type d ! -path './.git' ! -path './.git/*' | wc -l | tr -d ' ')
tests_total=$(jq -s '[.[] | select(.Test != null and (.Action == "run" or .Action == "skip"))] | length' "$go_test_json")
tests_executed=$(jq -s '[.[] | select(.Test != null and (.Action == "pass" or .Action == "fail"))] | length' "$go_test_json")
tests_skipped=$(jq -s '[.[] | select(.Test != null and .Action == "skip")] | length' "$go_test_json")

jq -S -n --arg subject_sha "$subject_sha" --arg go_version "$go_version" --argjson tests_total "$tests_total" --argjson tests_executed "$tests_executed" --argjson tests_skipped "$tests_skipped" --argjson go_files "$go_files" --argjson go_lines "$go_lines" --argjson gooo_files "$gooo_files" --argjson gooo_lines "$gooo_lines" --argjson descendant_dirs "$descendant_dirs" --argjson regular_files "$regular_files" --argjson test_ms "$test_ms" --argjson build_ms "$build_ms" \
  '{subject_sha:$subject_sha,go_version:$go_version,tests:{total:$tests_total,executed:$tests_executed,reused:0,skipped:$tests_skipped,not_observed:0},inventory:{go_files:$go_files,go_physical_lines:$go_lines,gooo_files:$gooo_files,gooo_physical_lines:$gooo_lines,descendant_dirs:$descendant_dirs,regular_files_root_readme_excluded:$regular_files},build_wall_ms:$build_ms,test_wall_ms:$test_ms,authority:{repository_writes:0,local_test_executions:0,cross_project_required_gates:0}}' > "$work_root/runtime.json"

"$binary" evaluate --source examples/experience-memory/main.gooo --contract contracts/experience-memory-denominator-v1.json --ir "$work_root/semantic-ir.json" --generated "$work_root/semantic.gooo.go" --fixture fixtures/fixed-fixture.json --memory fixtures/memory.ndjson --receipt fixtures/outcome-receipt.json --cases fixtures/cases --runtime "$work_root/runtime.json" --output-dir "$work_root/evaluation" > "$work_root/evaluate.json"

cp "$work_root/semantic-ir.json" "$artifact_root/semantic-ir.json"
cp "$work_root/semantic.gooo.go" "$artifact_root/generated/semantic.gooo.go"
cp "$work_root/evaluation/evaluation.json" "$artifact_root/evaluation.json"
cp "$work_root/evaluation/summary.json" "$artifact_root/summary.json"
cp "$work_root/evaluation/human-report.md" "$artifact_root/human-report.md"
cp fixtures/memory.ndjson "$artifact_root/memory.ndjson"
cp fixtures/outcome-receipt.json "$artifact_root/outcome-receipt.json"
cp fixtures/fixed-fixture.json "$artifact_root/fixed-fixture.json"
cp "$work_root/runtime.json" "$artifact_root/runtime.json"

jq -e '
  .schema == "gooo/experience-memory/report/v1" and
  .decision == "EXPERIENCE_MEMORY_CONFORMANCE_CLOSED" and
  .state == "CLOSED" and
  .precedence == ["REFUTED", "UNKNOWN", "CLOSED"] and
  .fixed_denominator == 12 and
  (.proofs | map(.total)) == [4,4,4] and
  (.indicators | map(.total)) == [4,4,4] and
  (.bindings | length) == 12 and
  ([.bindings[].activity] | unique | length) == 12 and
  .canonical_counts.normal == 4 and
  .canonical_counts.UNKNOWN == 4 and
  .canonical_counts.REFUTED == 4 and
  .metrics.attempts_observed == 2 and
  .metrics.memory_records == 1 and
  .metrics.candidate_count == 5 and
  .metrics.known_refuted_recurrences_before == 1 and
  .metrics.known_refuted_recurrences_after == 0 and
  .metrics.avoided_refuted_candidates == 1 and
  .metrics.new_unknown_candidates == 2 and
  .metrics.replay_comparisons == 2 and
  .metrics.replay_mismatches == 0 and
  .metrics.peak_rss_kib > 0 and
  .metrics.wall_ms > 0 and
  .metrics.tests.total == (.metrics.tests.executed + .metrics.tests.skipped) and
  .metrics.tests.reused == 0 and
  .metrics.tests.not_observed == 0 and
  .metrics.authority == {repository_writes:0,local_test_executions:0,cross_project_required_gates:0} and
  .baseline.selected_candidate_id == "candidate-known-refuted" and
  .baseline.state == "REFUTED" and
  .after.selected_candidate_id == "candidate-safe" and
  .after.state == "CLOSED" and
  (any(.after.candidates[]; .candidate_id == "candidate-known-refuted" and .state == "REFUTED" and .match_basis.all_three_exact == true)) and
  (any(.after.candidates[]; .candidate_id == "candidate-fingerprint-drift" and .state == "UNKNOWN")) and
  (any(.after.candidates[]; .candidate_id == "candidate-scope-drift" and .state == "UNKNOWN")) and
  (all(.canonical_cases[] | select(.state == "UNKNOWN"); .unknown.stage != null and .unknown.step != null and .unknown.reason != null and .unknown.unknown_class != null and .unknown.next_operation != null and .unknown.blocked_by != null)) and
  (any(.canonical_cases[]; .case_id == "refuted-contradictory-observation" and .state == "REFUTED")) and
  .append_only == true and .external_inputs_read_only == true and .root_readme_excluded == true
' "$artifact_root/evaluation.json" >/dev/null

manifest_lines="$work_root/manifest.ndjson"
: > "$manifest_lines"
while IFS= read -r -d '' file; do
  name=${file#"$artifact_root/"}
  digest="sha256:$(sha256sum "$file" | awk '{print $1}')"
  bytes=$(wc -c < "$file" | tr -d ' ')
  jq -cn --arg name "$name" --arg digest "$digest" --argjson bytes "$bytes" '{name:$name,digest:$digest,bytes:$bytes}' >> "$manifest_lines"
done < <(find "$artifact_root" -type f ! -name manifest.json -print0 | sort -z)
jq -S -n --arg subject_sha "$subject_sha" --slurpfile files "$manifest_lines" '{schema:"gooo/experience-memory/evidence-manifest/v1",subject_sha:$subject_sha,files:$files}' > "$artifact_root/manifest.json"

after=$(git status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
test "$before" = "$after"
jq -r '"decision: " + .decision + "\nfixed denominator: " + (.fixed_denominator|tostring) + "\nknown REFUTED recurrence pair: " + (.metrics.known_refuted_recurrences_before|tostring) + " -> " + (.metrics.known_refuted_recurrences_after|tostring) + "\navoided REFUTED candidates: " + (.metrics.avoided_refuted_candidates|tostring) + "\nnew UNKNOWN candidates: " + (.metrics.new_unknown_candidates|tostring) + "\nreplay comparisons/mismatches: " + (.metrics.replay_comparisons|tostring) + "/" + (.metrics.replay_mismatches|tostring)' "$artifact_root/evaluation.json" > "$artifact_root/summary.md"
