# Level 7 Dev Loop — Wave 1 Digest-Binding Design Amendment

| Field | Value |
|---|---|
| Artifact ID | `L7-AMD-W01-DES-001` |
| Artifact type | Finding-specific successor amendment to the Wave 1 design |
| Artifact schema | Bootstrap/pre-schema; migrate only through a later approved transition |
| Version | 0.1.0 |
| Date | 2026-08-25 |
| Status | **APPROVED FOR THE AUTHORIZED WAVE 1 IMPLEMENTATION** |
| Predecessor design | `L7-W01-DES-001` 0.1.0, SHA-256 `07953b2319635846505a018c3e4cc66705e0c263ab01b0a5c79e75cdaf1fb8e8` |
| Finding | `W01-DES-DIGEST-001`: the candidate manifest cannot hash evidence that embeds that manifest's final SHA-256 |
| Accountable owner | Anup Pandey |
| Authority event | Current-session explicit authorization of this amendment and continuation of approved Slices 2–4 |
| Effect | Add this amendment, extend the exact path policy, and replace only the circular candidate/evidence binding |

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

This amendment does not change product scope, acceptance criteria, module identity, test effects, branch, writer ownership, grant semantics, audit independence, merge authority, or the Wave 1 checkpoint. It authorizes no independent audit, audit-record write, merge, remote, publication, release, deployment, exposure, Wave 2 work, or external effect.

All other clauses of the approved change contract, specification, design, approval record, and module decision remain in force.
