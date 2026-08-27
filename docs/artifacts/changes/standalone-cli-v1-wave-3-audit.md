# Standalone CLI v1 Wave 3 — Independent Audit

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-3` |
| Candidate commit | `f7dbc2e2206f7c35faa8bb32a596c1d6f4d8bc86` |
| Candidate tree | `e69526be849ddf8ec725612b5dc21d465050ce7f` |
| Result | `GO` |
| Reviewer | `claude-code-opus-5` |
| Implementation candidate | `e2c0cdba4f0f5d02377695e179c5596148b6c423` |
| Implementation tree | `6c4269e06f4329e05e906421a58d729aad5225a0` |

## Audit boundary

The accountable owner authorized one real Claude Code session as the distinct
independent read-only auditor. Claude Code `2.1.241` (executable SHA-256
`1495eb7c42d3b4451f5f1cd38b6d498d22a4a38c802bc2be5c1cf1795e64820d`),
using `claude-opus-5`, reviewed a fresh detached, remote-free, recursively
read-only clone on macOS `26.5.2` arm64. Edit, write, browser, web, subagent,
Codex, Level 7, merge, deploy, and provider-trial capabilities were absent or
forbidden. No session transcript is retained in this record.

Executable verification was `NOT_RUN` in the frozen audit clone. The auditor
instead inspected the implementation and tests, ran bounded read-only Git
inspection, and cross-checked the committed verification record. This session
was an independent source audit, not a Level 7 adapter actual-host trial or a
provider-order validation.

## Findings

No blocking finding was reported. Claude classified the following six findings
as non-blocking for this default-OFF wave:

| ID | Severity | Finding and evidence | Recommended disposition |
|---|---|---|---|
| `W3-AUD-001` | MEDIUM | A descendant that deliberately calls `setsid` can escape the supervised process group while retaining inherited output pipes. `Cmd.Wait` then may not return, and `stopProcessGroup` waits without a second bound after `SIGKILL`, so the command can outlive its timeout and retain the mutation lock. See `internal/l7/adapter/process/process.go:134-161` and `internal/l7/adapter/process/process_unix.go:25-35`. The brief and README disclose session escape, but not this hang-and-lock consequence. | Add a bounded `Cmd.WaitDelay` path and a regression test before broadening containment claims; explicitly document the lock-retention consequence. The auditor treated this as the consequence of the already accepted macOS session-escape residual, so it did not block this wave. |
| `W3-AUD-002` | MEDIUM | The Claude review role retains `Bash` under `--permission-mode plan` (`internal/l7/adapter/claude/adapter.go:64-70`), whereas the Codex reviewer requests a hard read-only sandbox. Real Claude review semantics were not exercised, so mutation prevention depends on provisional provider behavior; post-review Git checks detect but do not undo mutation. | Remove `Bash` from the Claude reviewer tools, or make the isolation asymmetry an explicit prerequisite of a separately authorized actual-host envelope before claiming Claude reviewer support. |
| `W3-AUD-003` | LOW | `VerificationEvidence` binds commit and tree but stores no digest of the resolved verification configuration (`internal/l7/domain/execution.go:142-149`). A repository that ignores `.l7/config.json` could change verification argv without Git-derived invalidation. | Bind verification evidence to a digest of the resolved verification argv and relevant limits rather than depending on configuration being Git-visible. |
| `W3-AUD-004` | LOW | The brief's “normal repository hooks” wording overstates the environment: hooks run, but `internal/l7/adapter/git/commit.go:109-114` strips system/global Git configuration and uses the minimal process environment. Environment-dependent hooks or global-only Git identity can therefore differ from a plain `git commit`. | Document that hooks run under the bounded environment and that repository-local identity may be required. |
| `W3-AUD-005` | LOW | The `NO_GO` return-to-`building` path exists in `internal/l7/app/execution.go`, but no application- or CLI-level test directly drives it; the only Wave 3 `DecisionNoGO` test renders stored evidence. | Add a focused fake-reviewer test asserting `L7-REVIEW-010`, `building`, the remediation transition, and reconstructed status. |
| `W3-AUD-006` | LOW | `internal/l7/domain/lifecycle_test.go` was approved in the Modify set but is unchanged. | No action required; this is unused authorization, not scope expansion. |

## Evidence

| Area | Result |
|---|---|
| Identity and lineage | PASS — `HEAD` resolved to verified successor `f7dbc2e2206f7c35faa8bb32a596c1d6f4d8bc86` and tree `e69526be849ddf8ec725612b5dc21d465050ce7f`; its parent is implementation candidate `e2c0cdba4f0f5d02377695e179c5596148b6c423`, whose tree is `6c4269e06f4329e05e906421a58d729aad5225a0`. |
| Verification successor | PASS — the candidate-to-successor delta contains only `docs/artifacts/changes/standalone-cli-v1-wave-3-verification.md`, with 55 insertions and no deletion. |
| Scope and artifacts | PASS — all 42 base-to-successor paths are in the approved file set. `.l7/config.json`, `Makefile`, CI, plugin/skill, and historical-governance paths are unchanged. The future audit record and excluded user-owned foundation audit were correctly absent from the audit clone. |
| Default-OFF effects | PASS — `run`, `verify`, and `review` all enter through the disabled lifecycle gate before provider, verification, commit, or lock effects; tests assert zero such effects while disabled. |
| Verification and process controls | PASS with `W3-AUD-001` — verification dispatches exact argv without a shell and bounds environment, output, timeout, cancellation, and inherited process groups. Output overflow and terminal parser failures cannot report success. Deliberate session escape remains an explicitly limited macOS residual. |
| Controlled Git effects | PASS — expected HEAD/tree, clean index, exact path set, normal hook invocation, post-commit parent/path checks, and recovery-preserving failure behavior are present. No merge, amend, reset, rebase, push, or history rewrite capability was found. |
| Provider contracts | PASS with `W3-AUD-002` — a neutral contract isolates Codex and Claude flags/parsers, capability probing distinguishes unavailable, degraded, and available, required-version mismatch blocks, and event parsing fails closed. Compatibility remains fixture-bound and provisional. |
| Orchestration and authority | PASS — fake-provider coverage exercises both provider orders, restart reconstruction, exact identities, stale evidence, self-review rejection, scope/review mutation, and the three-artifact ceiling. Status stops at `reviewed`; Wave 3 exposes no ready or merge command. |
| Verification-record claims | PASS — named integration, adversarial, race, fuzz, cross-build, benchmark, and build-tagged probe artifacts exist and match the committed claims. The record truthfully marks real provider trials and daemon/session-escape experiments `NOT_RUN`. |
| Audit execution | `NOT_RUN` — this audit did not run `make`, Go tests, hooks, Level 7 commands, Codex, provider adapters, or actual-host/provider-order trials. |

## Residual risks

- Repository-defined verification is authorized arbitrary user code, not an OS
  sandbox. It retains the user's `HOME` and `PATH` within a bounded environment.
- A deliberately session-escaped descendant can survive group termination and,
  as `W3-AUD-001` records, may also retain output pipes and the mutation lock.
- A narrow digest-check-to-exec TOCTOU window remains without an `fexecve`-style
  primitive.
- Git postconditions reject scope expansion but intentionally do not erase
  provider-created work; operator recovery remains required.
- The mutation lock coordinates Level 7 processes, not arbitrary external Git
  actors; late index/HEAD races are detected rather than prevented.
- Adapter compatibility is bound only to fake fixtures for Codex CLI `0.149.1`
  and Claude Code `2.1.241`. No provider-support claim follows from this audit.
- Real providers retain custody of their own credentials, network access, and
  billing. Level 7 passes no API keys but preserves `HOME` for provider custody.
- This GO relies on committed executable-verification evidence plus independent
  static review; this audit session itself executed no product verification.

## Decision

`GO`. The independent auditor found the verified successor exactly in scope,
all ten acceptance criteria materially satisfied, no blocking issue, no merge
capability, and the current provider/process limitations truthfully constrained.
The six findings above remain hardening work and do not authorize remediation or
scope expansion in Wave 3.

This decision authorizes only the audit-only successor and its external binding.
It does not establish real-provider compatibility, authorize actual-host trials,
make the candidate ready, or authorize merge, deployment, or release.
