# Benchmark Paired Gate — Change Brief

| Field | Value |
|---|---|
| Change ID | `benchmark-paired-gate` |
| Risk tier | `3` — changes the decision rule of a required CI gate (`scripts/harness/`) |
| Status | `approved` by the active user on 2026-10-09 after reviewing the evidence and the proposed rule |
| Base commit | `5073586f08cfd4ba606d773ca5ebe4304b821417` |
| Accountable owner | Active user; authority is recorded outside candidate-controlled repository text |
| Implementer | `devin` |
| Assurance | `solo` — self-review only; no independent audit is claimed |

## Problem

The required `CLI paired benchmark gate` compares the median of five base
samples with the median of five candidate samples and blocks above +10%.
Hosted runners slow down and speed up during a run, so the two medians drift
apart even when the measured code is identical. Across PRs #21–#23 the gate
failed six times. Each time, the benchmarked package
`internal/l7/adapter/git` was byte-identical between base and candidate, and
a re-run passed. On a local machine the same two builds measured 73–104%
slower in one round and 5–6% faster in the next.

## Decision

This deliberately changes the decision rule, not just the sample count. The
threshold stays at 10%.

- **Pairs, not groups.** Sample *i* of the base and sample *i* of the
  candidate run back to back (the alternation is unchanged), so each pair
  sees the same machine state. The gate divides each candidate sample by its
  partner and takes the median of those ratios.
- **Consistency.** A pair counts as slower when its ratio exceeds 1. The gate
  blocks only when the median paired slowdown exceeds the threshold **and**
  at least 8 of 9 pairs are slower. A real slowdown of 10% or more shifts
  nearly every adjacent pair; one that shows in fewer pairs cannot be told
  apart from noise at this sample size.
- **More, longer samples.** Nine alternating pairs instead of five, and
  `BenchmarkSnapshot10000Paths` runs `30x` per sample instead of `10x`
  (`BenchmarkParseStatus10000Paths` keeps `250x`). The job takes about four to
  five minutes instead of about one and a half.

Re-scoring the failed runs that were captured from CI:

| Run | Current rule (group medians) | Paired median | Slower pairs | Paired rule |
|---|---|---|---|---|
| #22 first run, Snapshot | +20.8%, blocked | +6.1% | 5/5 | pass |
| #22 re-run, ParseStatus | +11.3%, blocked | +8.0% | 4/5 | pass |
| #22 re-run, Snapshot | +14.6%, blocked | +3.9% | 3/5 | pass |
| #23 first run, ParseStatus | +16.2%, blocked | +6.9% | 4/5 | pass |
| #23 re-run, Snapshot | +11.9%, blocked | +11.2% | 5/5 | blocked |

The last row would still block. It was slower in every pair, which five
samples cannot distinguish from a small real slowdown; nine samples with the
8-of-9 rule decide such cases with more evidence.

Unchanged: the 10% threshold, the benchmark names and package, the job name,
offline disposable build roots, direct failure propagation, and the trusted
policy's exact-head owner acceptance marker.

## Exact implementation file set

Add:

- `docs/artifacts/changes/benchmark-paired-gate.md`

Modify:

- `internal/harness/benchgate/main.go`
- `internal/harness/benchgate/main_test.go`
- `scripts/harness/check-cli-benchmarks.sh`
- `CHANGELOG.md`

## Acceptance criteria

1. `benchgate` reports, per benchmark, both group medians, the median paired
   change, and the slower-pair count. It blocks only when the paired change
   exceeds the threshold and the slower pairs reach `--minimum-slower-pairs`.
   A paired change of exactly the threshold passes.
2. It still fails closed on unpaired, undersampled, or malformed data, and
   rejects a `--minimum-slower-pairs` outside 1 to `--minimum-samples`.
3. Fixtures built from the captured CI runs above score as in the table, a
   consistent +15% slowdown blocks, and one slow outlier pair does not.
4. The script runs nine alternating pairs with Snapshot at `30x` and calls
   `benchgate --threshold-percent 10 --minimum-samples 9
   --minimum-slower-pairs 8`.
5. The benchgate tests, `make technical-lint`, `make policy-check`, a local run
   of the script, and `git diff --check` pass.

## Risks and mitigations

- **Missing a real regression:** a slowdown that shows in seven or fewer of
  nine pairs passes. Such a result is indistinguishable from runner noise at
  this sample size; the threshold itself is unchanged, and a consistent
  slowdown still blocks.
- **Self-judging change:** the workflow runs the candidate's script, so this
  PR is judged by its own new rule. The owner approved the rule before
  implementation, and the brief records the evidence.
- **Longer CI:** about three extra minutes per PR, within the job's
  20-minute limit.

## Rollback

Revert the implementation commit and this brief; the previous five-sample,
group-median rule returns unchanged.
