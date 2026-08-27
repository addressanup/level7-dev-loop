# Standalone CLI v1 Wave 1 — Independent Audit

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1` |
| Candidate commit | `2004817cd115c2997f13aeb78ea81ce243084065` |
| Candidate tree | `eded83fa59dc0f75572d6e4a04a959b1a13df8a6` |
| Result | `NO_GO` |
| Reviewer | `independent-auditor` |

## Findings

| ID | Severity | Finding | Evidence | Required remediation |
|---|---|---|---|---|
| `CLI-AUD-001` | HIGH | The protected import/effect policy does not enforce the approved dependency contract. It is a narrow denylist, so domain, application, or presentation code can import existing non-allowed repository packages, and filesystem effects can enter through standard packages not named in the seven-prefix list while `import-check` remains green. The current Wave 1 source is pure, but the control added specifically to keep later effects outside the core can be bypassed. | The brief at `docs/artifacts/changes/standalone-cli-v1.md:574-589` requires domain to exclude filesystem/process/environment/network/clock/randomness/terminal packages and application to import only domain plus pure standard-library types. `harness/import-boundaries.tsv:75-105` forbids only selected `internal/l7/*` and seven standard prefixes; it does not prevent imports such as `internal/evaluator` or `io/ioutil`. `scripts/harness/check-import-boundaries.sh:77-99` compares only module-local direct imports against those listed prefixes and does not inspect standard-library transitive effects. In pinned Go 1.26.7, `io/ioutil` directly exposes `ReadFile`, `WriteFile`, and `ReadDir` through `os`, but an application import of `io/ioutil` does not match BND-602's `os` prefix. The verification's fail-closed probe covers only a direct `os` import, and there is no checked-in regression test for an alternate repository-package or standard-library effect bypass. | Enforce an allowlisted package closure for `internal/l7/domain`, `internal/l7/app`, and `internal/l7/presentation`, or add an equivalently fail-closed effect model that covers transitive standard-library effects. Add automated negative probes for at least a non-domain repository package and an indirect filesystem package such as `io/ioutil`, plus the intended positive closures. Keep the policy change Tier 3 and re-run verification/audit. |
| `CLI-AUD-002` | MEDIUM | The change adds and runs a macOS job, but the documented installation contract does not make that job a required merge check even though the approved brief calls for a blocking macOS baseline. A failing macOS-only test can therefore remain mergeable under the documented required-check set. | `.github/workflows/harness.yml:44-63` creates the separate `CLI macOS 15 (arm64)` check. The brief at lines 555-558 requires it to be blocking. `README.md:78-83` requires only the baseline Harness check and Trusted policy; lines 113-116 describe the macOS job but do not require its exact check or a required-workflow rule. No repository ruleset evidence is present in scope. | Add the macOS check (or the entire Harness workflow) explicitly to the required repository-rule installation contract and verify the live ruleset before calling the job blocking. |

## Passed evidence

| Area | Result |
|---|---|
| Identity and lineage | PASS — candidate commit/tree match the requested identity. Verification binds implementation commit `5e65b272e454e4b446847453ade2f042d9d631d8` and tree `24b4c277f2b9c51f247941267c2203a1063b0072`; the verification record is the only successor change. |
| Scope and authority | PASS — all 14 base-to-candidate paths are in the immutable approved brief; external owner `accountable-user` is distinct from implementer `codex-root` and bound to brief commit `28098a75c924d7360bd86dc02b32066b7c4289e4`. The Tier 3 artifact budget remains brief, verification, and this audit. |
| Technical/policy separation | PASS — `make ci` runs candidate technical checks without manufacturing authority; local `make verify` retains policy; trusted policy continues to build the evaluator from the base revision. |
| CLI behavior | PASS — help/version are deterministic and available; status truthfully returns `BLOCKED`/exit `2`; unknown flags return JSON `FAILED`/exit `1`; no repository, Git, provider, network, merge, or deployment effect exists. |
| Input/output/cancellation | PASS — argument count and byte length are bounded, untrusted fields are escaped and command context is truncated safely, short/failed writes return a stable stderr diagnostic, and pre-execution cancellation returns `CANCELLED`/exit `130`. |
| `make policy-check` | PASS — Tier 3, state `verified`, 14 changed paths, next `request an independent read-only audit`. |
| `make verify` | PASS — eight-package import check, formatting, shell syntax, vet/typecheck, all tests, harness reproducibility, and CLI reproducibility. |
| Build evidence | PASS — host and cross-built Darwin arm64 SHA-256 match `d0db0728b261ed5b7023f3e0b5bdc7ca003da8902e7068075fc7456740236c9b`; Darwin amd64 matches `87446ec4289785a3495d1f6f6178317d1c70c7c462ac8f6a8328f4a83fce3fb9`. Both binaries have the declared Mach-O architecture. |
| Workflow/manifests/diff | PASS — `actionlint`, manifest `jq empty`, and `git diff --check` passed. |
| Rollback | PASS — the brief and five scoped implementation/remediation commits are conventional and independently revertible; no dependency, runtime state, migration, provider, or repository mutation was introduced. |

## Decision

`NO_GO`. The inert CLI implementation itself is correct and truthful, but Wave 1
is a protected harness foundation. The incomplete import/effect gate is a HIGH
acceptance and safety defect, and the macOS blocking claim is not fully wired into
the documented merge contract. Remediate, produce a fresh Git-bound verification
candidate, and request another independent audit before merge.
