# Level 7 Dev Loop — Wave 1 Digest-Binding Design Amendment

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-W01-DES-001` |
| Artifact type | Finding-specific successor amendment to the Wave 1 design |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.3.0 |
| Date | 2026-08-26 |
| Status | **APPROVED FOR THE AUTHORIZED WAVE 1 AUDIT REMEDIATION** |
| Predecessor design | `L7-W01-DES-001` 0.1.0, SHA-256 `07953b2319635846505a018c3e4cc66705e0c263ab01b0a5c79e75cdaf1fb8e8` |
| Finding | `W01-DES-DIGEST-001`: the candidate manifest cannot hash evidence that embeds that manifest's final SHA-256 |
| Successor findings | `AUD-W01-016`: capped repository enumeration used an unordered pre-sort subset; `AUD-W01-017`: accepted fixture/effect description was stale; `AUD-W01-020`: verifier effect roots were lexically but not physically contained |
| Accountable owner | Anup Pandey |
| Authority event | Accountable-owner approval of `level7-dev-loop:l7-release` Mode C remediation for `AUD-W01-020` and `AUD-W01-021` on 2026-08-26 |
| Effect | Preserve the acyclic binding; retain stable aggregate-bound behavior; enforce physical repository containment before verifier writes; record actual local fixture/process/clock effects |

## 1. Finding

`L7-W01-DES-001` §15.2 required `wave-01-candidate.sha256` to list every changed file other than itself and the later audit while also requiring `wave-01-evidence.md` to embed that manifest's final SHA-256. If the manifest hashes evidence, inserting the manifest hash changes the evidence hash and therefore changes the manifest. The contract has no truthful finite construction.

The same principle applies to a commit identity embedded in a file contained by that commit. A candidate file cannot truthfully contain the final commit/tree identity that depends on the candidate file's own bytes.

Implementation stopped before the phase/scope gate when the defect was found. The accountable owner explicitly authorized this narrow correction and continuation.

## 2. Corrected non-circular binding

The original design remains byte-preserved. These rules supersede only its circular manifest/evidence/commit wording:

1. `docs/artifacts/wave-01-design-amendment.md` is added to `harness/wave-01-paths.tsv` and is included in the candidate manifest.
2. `docs/artifacts/wave-01-candidate.sha256` lists every changed regular candidate file relative to the approved base except:
   - `docs/artifacts/wave-01-candidate.sha256` itself;
   - `docs/artifacts/wave-01-evidence.md`; and
   - the absent later `docs/artifacts/wave-01-audit.md`.
3. The implementation candidate is committed first. That commit contains the manifest and every manifest-covered file, but not the evidence or audit record.
4. `docs/artifacts/wave-01-evidence.md` is then created in one evidence-only direct-child commit. It binds the implementation-candidate commit, tree, parent, candidate-manifest SHA-256, commands, environment, results, effects, and limits.
5. The evidence record does not embed its own SHA-256 or its containing evidence-commit identity. Its final SHA-256 and evidence-commit/tree/parent are reported in the completion handoff and become mandatory inputs to the later independent audit.
6. The independent audit binds both exact identities:
   - implementation-candidate commit/tree plus candidate-manifest SHA-256; and
   - evidence-only commit/tree plus evidence SHA-256.
7. Any candidate change requires a new implementation-candidate commit, manifest, full verification, evidence-only binding, and audit. Any evidence change requires a new evidence digest/commit binding and invalidates an earlier audit.

This creates an acyclic chain:

```text
approved base
  -> implementation candidate + manifest
  -> evidence-only binding of candidate/manifest
  -> independent audit binding candidate + evidence
```

## 3. Scope and non-authorization

The digest-binding correction does not change product scope, acceptance criteria, module identity, branch, writer ownership, grant semantics, audit independence, merge authority, or the Wave 1 checkpoint. Section 4 corrects the test-effect description, stable-rule contract, and verifier-effect boundary only; it creates no product behavior or broader authority. The fifth successor adds the harness-only `scripts/harness/prepare-cache.sh` path to enforce that boundary before a verifier write. This amendment authorizes no independent audit verdict, merge, remote, publication, release, deployment, exposure, Wave 2 work, network access, credential access, provider/host action, or ambient cleanup.

All other clauses of the approved change contract, specification, design, approval record, and module decision remain in force.

## 4. Successor bounded-enumeration and test-effect correction

### 4.1 Confirmed design drift

The original design at `L7-W01-DES-001` §4.2 described every test fixture as immutable `testing/fstest.MapFS` data needing no temporary files, external process, or clock. Finding-specific regressions now correctly use rooted real-filesystem symlink/FIFO/replacement and enumeration fixtures, the current local test binary in child processes, pinned local Go subprocesses for import/process proof, and clocks for bounded-completion or injected deterministic deadline checks. The earlier sentence is historical and does not describe the successor suite.

### 4.2 Approved successor effect model

For the authorized Wave 1 remediation and verification envelope:

1. `testing/fstest.MapFS` remains preferred for format and pure-policy fixtures. Real filesystem fixtures are permitted only where link, node, identity, replacement, enumeration, process, or OS-root behavior is the fact under test.
2. `Makefile` exports `TMPDIR=$(GOTMPDIR)` and `GOTMPDIR=$(PROJECT_ROOT)/.cache/go/tmp`, then delegates preparation to `scripts/harness/prepare-cache.sh`. Before any verifier write, that script must reject a noncanonical project root plus every existing symlink or non-directory component in the declared Go, temporary, telemetry, reproducibility, and selected toolchain roots. It creates missing effect directories one component at a time, verifies every resulting physical path equals its repository path, and replaces telemetry mode through a newly created repository-local file. `t.TempDir`, local replacement modules, test binaries, and their automatically cleaned temporary roots must remain physically below that ignored repository-scoped root.
3. A process fixture may execute only the current local test binary or a pinned repository-local Go binary with offline module settings. No shell-selected executable, user program, provider, host plugin, credential, network, remote, or external sink is admitted.
4. Wall-clock use is limited to fail-closed completion ceilings. Deterministic policy tests inject fixed clocks. A timeout never produces a pass or suppresses a stricter finding.
5. Go testing owns cleanup of `t.TempDir` roots. Repository `.cache` remains ignored retained verifier state; no ambient host cleanup is authorized.

### 4.3 Stable aggregate-bound rule

`SCOPE-338` means: a sentinel-filled repository directory batch proves that the remaining combined Wave 1 directory/file budget is exceeded before entry-category inspection. Its subject is the stable repository-relative directory, its result is blocking, and its recovery is to remove unapproved paths or approve a bounded successor. This rule is evaluated before sorting or inspecting the filesystem-selected subset, so directory order cannot choose a file/directory-specific rule or subject.

Permanent regressions must prove:

- ascending and descending creation of an oversized real-file directory render the same complete `SCOPE-338` result;
- mixed capped subsets representing an omitted directory and an omitted file render the same complete `SCOPE-338` result without inspecting an entry; and
- separate child processes exit 1 with byte-identical rule, subject, message, and recovery text.

### 4.4 Physical verifier-effect containment

Physical containment is an admission rule, not a lexical naming claim. Preparation must fail before its first write when `.cache`, a Go cache/temp/telemetry component, the telemetry-mode path, the toolchain root, or the selected toolchain is a symlink or has an unexpected file type. A safe fixture must create only the declared repository-local directories and exact telemetry mode. Redirect fixtures must leave their external test-owned targets byte-for-byte unchanged and must not create a later cache sibling before rejection.
