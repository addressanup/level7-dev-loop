# Standalone CLI v1 Wave 1 — Independent Audit

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1` |
| Candidate commit | `7bb3088aace63cb38eed97dcd81b4ed1c0c9f775` |
| Candidate tree | `a72bc1d7fbb968bd1b65557903735f8e18c54eb0` |
| Result | `GO` |
| Reviewer | `independent-auditor` |

## Prior finding closure

| ID | Status | Evidence |
|---|---|---|
| `CLI-AUD-001` | CLOSED | `l7-import-closure-check` compares direct imports of domain, application, and presentation against exact package-specific allowlists on the host, Darwin arm64, and Darwin amd64. TSV `allowclosure` rules independently restrict all non-standard dependencies to the domain package. The current closures contain no unexpected package. Checked-in negative probes reject both `internal/evaluator` and `io/ioutil`; the same validator controls real `go list` output. `make import-check` reports three valid closures and both negative probes. |
| `CLI-AUD-002` | CLOSED | README now names `Go 1.26.7 (baseline)`, `CLI macOS 15 (arm64)`, and `Trusted policy` as required installation checks, requires live-ruleset verification during installation/upgrade, and truthfully states that the local repository contains no evidence that those remote rules are currently active. It no longer calls the macOS job blocking without that evidence. |

## Findings

No BLOCKER, CRITICAL, HIGH, MEDIUM, or LOW findings remain.

## Evidence

| Area | Result |
|---|---|
| Identity and lineage | PASS — verification commit `7bb3088aace63cb38eed97dcd81b4ed1c0c9f775` resolves to tree `a72bc1d7fbb968bd1b65557903735f8e18c54eb0`. Its record binds implementation/remediation commit `34ba6284121374fdf5e8efc13a4872f29fa6ab0c` and tree `19c711edffba61929995b4894e70e9807698dec4`; the verification record is the only successor change. Prior NO_GO and superseded verification remain in Git history. |
| Scope and authority | PASS — all 14 base-to-candidate paths are within the immutable approved brief. External owner `accountable-user` is distinct from implementer `codex-root` and is bound to brief commit `28098a75c924d7360bd86dc02b32066b7c4289e4`. The Tier 3 artifact budget remains brief, verification, and this audit. |
| Import and effect boundaries | PASS — direct imports exactly match the three allowlists; local closures contain only the applicable product package and domain; non-domain repository and indirect-filesystem probes fail closed; existing BND-601 through BND-606 checks pass. |
| Technical/policy separation | PASS — `make ci` remains technical-only, local `make verify` retains policy, and trusted policy continues to evaluate protected changes from the base controller rather than candidate code. |
| CLI behavior and bounds | PASS — help/version are deterministic; status truthfully returns `BLOCKED`/exit `2`; invalid flags return JSON `FAILED`/exit `1`; argument count/bytes and diagnostic context are bounded; escaping, failed writes, and cancellation behave as specified. No repository, Git, provider, network, review, merge, release, or deployment effect exists. |
| `make policy-check` | PASS — Tier 3, state `verified`, 14 paths, next `request an independent read-only audit`. |
| `make verify` | PASS — offline modules, exact import closures, legacy boundaries, formatting, shell syntax, vet/typecheck, all tests, harness reproducibility, and CLI reproducibility. |
| Reproducible builds | PASS — host and Darwin arm64 SHA-256 are `d0db0728b261ed5b7023f3e0b5bdc7ca003da8902e7068075fc7456740236c9b`; Darwin amd64 is `87446ec4289785a3495d1f6f6178317d1c70c7c462ac8f6a8328f4a83fce3fb9`. File inspection confirms the declared Mach-O architectures. |
| `make ready-check` before audit | Expected BLOCKED — state `verified`, next `request an independent read-only audit`. |
| Workflow/manifests/diff | PASS — `actionlint`, manifest `jq empty`, and `git diff --check` passed. |
| Rollback and claims | PASS — remediation is scoped and conventionally committed; no dependency, state, migration, or product effect was added. README and verification limit macOS, provider, lifecycle, release, and live-ruleset claims to observed evidence. |

## Decision

`GO`. The remediated Wave 1 candidate satisfies the approved inert-CLI and
protected-harness acceptance criteria. The prior audit remains preserved in Git,
the current decision is bound to the exact verified commit/tree, and merge still
requires the external exact-head review and repository rules described by the
installation contract.
