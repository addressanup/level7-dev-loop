# Level 7 v1.0 Multi-Host Orchestration — Audit Remediation

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration-remediation` |
| Risk tier | `3` — repository mutation containment, gateway credential transport, protected verification controls, and release qualification |
| Status | `proposed`; implementation requires fresh Product Owner approval bound to this exact brief addition commit |
| Base commit | `b16379e8ce609d9b1bb8fe5d3dc86ff503d66f0f` |
| Base tree | `71703b9d18f4071007e51a432e3b3d4531360fe4` |
| Proposal parent | History-preserving revert `8c933788d6cbd983f82785f84b400257809fdd1a`, whose tree exactly equals the base tree |
| Triggering audit | `NO_GO` at commit `b16379e8ce609d9b1bb8fe5d3dc86ff503d66f0f`, reviewer `apbusinessidentity-tech`, covering verified candidate `f97b2445bb966660fb3cb403521cae78f0367cbe`, tree `4aac837742726a63d8103db8da4a8fb5c6c32db8` |
| Accountable owner | Anup Pandey, Product Owner; approval of this exact proposal commit is pending |
| Implementer | `codex-root` |
| Assurance | Tier 3 exact-scope implementation, exact-head verification, and a freshly commissioned independent read-only audit before any owner GO |
| Next executable transition | Stop for explicit Product Owner approval of this exact brief addition commit |

## Problem

The independent audit recorded three blockers against the v1 orchestration
candidate:

1. `L7-ORCH-AUD-001`: patch authorization does not derive the complete path set
   for every Git-supported operation before mutation.
2. `L7-ORCH-AUD-002`: the HTTP loopback exception uses textual URL prefixes
   instead of parsed, exact authorities.
3. `L7-ORCH-AUD-003`: exact-candidate verification does not establish the
   mandatory race/fuzz and dual-host v1 package lifecycle conformance.

The prior brief, implementation, verification, and `NO_GO` audit remain
historical facts. They are not amended, superseded as evidence, or converted to
approval for this remediation.

An invalid first draft at `222458d127a2f85d387b6ab76e14a255e440132e`
omitted the machine-required Scope heading. It was never approved and was
reverted ordinarily at `8c933788d6cbd983f82785f84b400257809fdd1a`, restoring
the exact base tree. Both commits remain reachable; neither is an authority or
part of the live proposal diff.

## Outcome

Make repository patching and gateway URL admission fail closed, and make the
standard verification boundary exercise the missing adversarial and package
conformance gates. Preserve the existing v1 product behavior, default-OFF
features, package contents except for rebuilt binaries, stable v0.1.1 rollback
artifacts, and all external-effect stops.

The package conformance harness must use only disposable, test-owned roots. It
may unpack and execute the candidate's native binary there, but it must not
install into either user's Codex or Claude configuration, alter host settings,
use provider credentials, access a gateway, or make a network request.

## Scope

In scope are only the two audited security fixes, their adversarial tests, the
missing bounded race/fuzz gates, the offline dual-host v1 package-conformance
harness, and exact macOS architecture CI wiring. The harness may exercise
filesystem lifecycle and native candidate binaries only inside disposable
roots.

Out of scope are new product capabilities, dependencies, provider promotion,
real host configuration or installation, credential use, live provider or
gateway calls, Cyber active execution, publication, signing, protected merge,
release, and deployment. Existing historical records are immutable.

## Exact implementation file set

Add:

- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation-verification.md`
- `docs/artifacts/changes/l7-v1-multi-host-orchestration-remediation-audit.md`
- `internal/harness/v1candidate/main.go`
- `internal/harness/v1candidate/main_test.go`
- `scripts/harness/check-l7-fuzz.sh`
- `scripts/harness/check-v1-conformance.sh`

Modify:

- `.github/workflows/harness.yml`
- `Makefile`
- `harness/import-boundaries.tsv`
- `internal/l7/adapter/orchestrationconfig/config.go`
- `internal/l7/adapter/orchestrationconfig/config_test.go`
- `internal/l7/adapter/toolbroker/broker.go`
- `internal/l7/adapter/toolbroker/broker_test.go`

Delete: none.

## Acceptance criteria

1. The proposal parent has the exact base tree. Both base-to-proposal and
   parent-to-proposal diffs add only this brief. Fresh Product Owner approval
   binds this corrected re-add commit—not the invalid draft—before any other
   declared path changes.
2. The previous successor brief, implementation, verification, `NO_GO` audit,
   approval envelope, audit envelope, invalid draft, and ordinary revert remain
   byte-for-byte unchanged and reachable. No amend, rebase, reset, force-push,
   or history rewrite occurs.
3. Before applying a patch, the broker uses one canonical Git-aware, no-effect
   preflight to derive every affected source and destination path and operation.
   It rejects malformed, ambiguous, binary, copy, rename, mode-only, secret,
   protected, symlink, and out-of-manifest cases; no rejected patch mutates the
   worktree. The accepted post-apply delta must exactly equal the preflighted
   path set, with regression tests for every path-bearing form.
4. Gateway endpoint and catalog URLs are parsed as absolute URLs. User
   information, missing or ambiguous authority, invalid ports, and deceptive
   host spellings are rejected. HTTPS remains the general rule; HTTP is allowed
   only for an exact normalized `localhost`, `127.0.0.1`, or `::1` authority,
   with adversarial tests applying the same rule to endpoint and catalog URLs.
5. The standard technical verification target runs the relevant CGO-enabled Go
   race suite and every declared security/parser fuzz target with explicit,
   bounded durations. The fuzz driver has a fixed target inventory, fails if a
   target disappears or is duplicated, uses the pinned offline toolchain, and
   writes no tracked corpus or repository state.
6. A standalone, network-denied v1 conformance harness validates both Codex and
   Claude archives: closed inventory, modes, checksums, native launcher and JSON
   CLI contracts, MCP initialize/list/call framing, disposable installation,
   upgrade from the frozen v0.1.1 package state, exact rollback, removal, and
   cleanup. It rejects traversal, symlink, substitution, stale receipt, changed
   owned file, and unowned-file conflicts without touching user host state.
7. The macOS CI architecture matrix invokes the same v1 candidate and
   conformance gate natively on arm64 and amd64. Local or hosted results are
   credited only when bound to the exact candidate; editing the workflow does
   not itself count as hosted evidence, and publishing it requires separate
   authority.
8. Existing module, import/effect boundary, formatting, shell, vet, typecheck,
   unit, actual-provider compile-only, distribution compatibility,
   reproducibility, SBOM, and benchmark controls remain intact. New harness code
   is product-inaccessible and network-denied by the boundary table.
9. Orchestration, Sync, Cyber active mode, and Headless remain default OFF.
   Credentials remain native-host owned or referenced only by environment or
   Keychain identity. No provider support, signed provenance, installation, or
   stable-release claim is promoted.
10. Exact-head verification may add only the declared verification record and
    must record commands, platforms, outcomes, and new artifact identities. A
    later, separately commissioned reviewer may add only the declared audit
    record; the prior `NO_GO` decision cannot transfer to the remediated head.
11. The source checkout's unrelated untracked
    `docs/artifacts/foundation-rebaseline-admission-audit.md` remains untouched
    and outside this change.
12. No push, signing, release, publication, protected-branch merge, real host
    installation, deployment, or production effect occurs under this brief.

## Risks and mitigations

- **Transient out-of-scope mutation:** require complete no-effect Git preflight
  before applying, exact postcondition comparison, and negative structural-patch
  coverage.
- **URL equivalence traps:** use parsed scheme/host/port fields and adversarial
  user-info, suffix, encoding, IPv4, IPv6, and catalog fixtures.
- **Flaky or unbounded verification:** pin the toolchain, target inventory,
  duration, concurrency, temporary roots, and network-off environment.
- **False lifecycle confidence:** distinguish fixture installation from real
  host installation and require native architecture runs before crediting
  arm64 or amd64 behavior.
- **Stale assurance:** any source, test, gate, package, or workflow change after
  verification invalidates its evidence and requires a fresh exact-head run.
- **Historical record loss:** preserve the original chain, invalid draft/revert,
  and `NO_GO`; remediation starts a new change ID rather than rewriting the
  predecessor chain.

## Rollback

Before implementation, revert only the corrected brief proposal commit to
restore proposal-parent tree `71703b9d18f4071007e51a432e3b3d4531360fe4`, which
equals the exact audit base tree. After implementation, revert the audit,
verification, implementation, and corrected brief commits in reverse order,
stopping on any conflict or unexpected path and confirming the intended Git
tree after each step.

All rollback uses ordinary additive revert commits. Test-owned temporary roots
and ignored build/cache output may be discarded, but historical governance,
the frozen v0.1.1 distribution, user plugin installations, credentials, remote
refs, releases, and production state must remain untouched.

## Current transition

1. Commit only this corrected proposal on `codex/l7-v1-orchestration` as a child
   of the exact-base-tree revert.
2. Stop for explicit Product Owner approval bound to that proposal commit.
3. Only after approval, implement within the declared file set and run bounded
   local verification; stop at any scope expansion or external-effect boundary.
4. Commit a verification record only after exact-head PASS. Hosted execution,
   independent audit, owner GO, push, signing, release, publication, protected
   merge, installation, and deployment each require their applicable later
   authority.
