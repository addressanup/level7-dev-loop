# Level 7 v1.0 Multi-Host Orchestration — Audit Remediation Verification

| Field | Value |
|---|---|
| Change ID | `l7-v1-multi-host-orchestration-remediation` |
| Candidate commit | `b5ebe9ea96ed8f41782fcc0fb3a0292f74314cd7` |
| Candidate tree | `32e850aef741188108fb4e5c2097f7af684ceb96` |
| Result | `PASS` |
| Reviewer | `codex-root` |
| Verified at | `2026-08-31T17:55:10Z` |
| Brief commit | `7947514a6a817a9e8490e7702288873ae8de249a` |
| Triggering audit | `b16379e8ce609d9b1bb8fe5d3dc86ff503d66f0f` (`NO_GO`) |
| Implementation commits | `0dc40eeae6d86685ca214792bf44c8bca01c9d58`, `fb50d947f153c63f62fae4eae8afec2c807ead16`, `b5ebe9ea96ed8f41782fcc0fb3a0292f74314cd7` |

## Checks

| Check | Result |
|---|---|
| Approval and policy binding | PASS — the external active-user approval envelope names Anup Pandey (Product Owner), implementer `codex-root`, remediation change ID, and exact brief commit `7947514a6a817a9e8490e7702288873ae8de249a`. Team policy selected Tier 3 and accepted exact candidate `b5ebe9ea96ed8f41782fcc0fb3a0292f74314cd7`, tree `32e850aef741188108fb4e5c2097f7af684ceb96`, in `building` state. |
| Lineage and exact scope | PASS — relative to audit base `b16379e8ce609d9b1bb8fe5d3dc86ff503d66f0f`, the candidate changes exactly 12 declared paths: the approved remediation brief plus 11 implementation, test, harness, and workflow paths. Historical briefs, verification records, and the triggering `NO_GO` audit remain unchanged. The declared remediation audit record is absent. |
| `L7-ORCH-AUD-001` | PASS — the patch broker obtains the affected file set through bounded, no-effect `git apply --check --numstat -z`, rejects ambiguous source/destination, rename, copy, binary, mode-only, symlink, submodule, secret, protected, duplicate, and out-of-manifest operations before mutation, and verifies the same reverse path set after apply with a checked rollback on mismatch. Adversarial real-Git, symlink, rollback, parser, and fuzz coverage passed. |
| `L7-ORCH-AUD-002` | PASS — endpoint and catalog admission use the same strict absolute-URL parser. It rejects user information, encoded or ambiguous authorities, noncanonical and invalid ports, deceptive loopback spellings, malformed IPv4/IPv6, queries, fragments, and non-HTTPS remote endpoints; HTTP admits only exact normalized `localhost`, `127.0.0.1`, or `::1`. Adversarial tests passed. |
| `L7-ORCH-AUD-003` | PASS — the standard technical gate now requires serialized CGO-enabled race coverage and all eight fixed parser/security fuzz targets. The v1 package gate validates both host archives and both Mach-O architectures, with native host execution under macOS `sandbox-exec` using `(deny network*)`; a test confirmed a loopback socket attempt fails with `EPERM`. |
| `L7_ASSURANCE_MODE=team make verify GO_VERSION=1.26.7` | PASS — exact-head policy; offline module integrity; import/effect boundaries; formatting and shell syntax; vet; type and actual-provider tagged compile-only checks; complete unit tests; CGO-enabled serialized race tests; all eight fixed fuzz targets for five seconds each in an archived disposable checkout; reproducible harness/CLI builds; and frozen v0.1.1 distribution compatibility passed. |
| `L7_ASSURANCE_MODE=team make v1-candidate-check GO_VERSION=1.26.7` | PASS — arm64/amd64 Go and Swift binaries and both host archives reproduced. On native Darwin arm64, the OS-network-denied harness validated closed inventory, modes, checksums, SBOM/provenance, JSON CLI, MCP initialize/list/invalid-call framing, disposable stable installation, upgrade, exact rollback, conflict-safe removal, and cleanup for Codex and Claude. |
| Negative lifecycle coverage | PASS — unit tests rejected traversal, absolute/backslash paths, symlinks, duplicate and colliding archive entries, inventory substitution, duplicate/trailing JSON, stale receipts, changed owned files, unowned files, and symlink substitution. All roots were test-owned and disposable. |
| User-owned state | PASS — the source checkout's unrelated untracked `docs/artifacts/foundation-rebaseline-admission-audit.md` remained untouched and unstaged, with SHA-256 `9f2a89ce869dd69bcd326272608be972ea6c974563c639b766b8542ef03a9884`. |

The first exact-head attempt against implementation commit
`0dc40eeae6d86685ca214792bf44c8bca01c9d58` exposed an existing five-second
handshake deadline under concurrent package-level race loading. The affected
package then passed three consecutive race runs. Additive commit
`fb50d947f153c63f62fae4eae8afec2c807ead16` serialized package-level race
execution with `-p=1`; the complete gates passed, but a pre-record evidence
review found that environment-only network controls did not establish the
brief's OS-level network-denial requirement. No verification record was
committed for either earlier candidate. Additive commit
`b5ebe9ea96ed8f41782fcc0fb3a0292f74314cd7` added and tested OS-level denial,
and both required gates were rerun from that exact clean head. Only those final
runs are credited here.

## Reproducible identities

| Output | SHA-256 |
|---|---|
| Harness test binary | `72fa589b7eca46364eb6e71803051456a12603b0405aa89b754c3ac8bf4f0b70` |
| Darwin arm64 `l7` / reproducible native CLI | `70137a3872929b5ead065f30c9686db5cb94752cfff8da28db667f1588f87590` |
| Darwin amd64 `l7` | `412a1d09f657066aab55fdab33c4b111e1a64f3cdbf731ed11b00cc45c5babb4` |
| Darwin arm64 `l7-embed` | `898db0787c397cb15508e0f91b8081f458522cb3734abd8a1bfdc953b920bc3b` |
| Darwin amd64 `l7-embed` | `c7c33ff35e05ad2b5a037374276c01d8c610401e7300136db8c08e118884ef7d` |
| `level7-dev-loop-1.0.0-dev-codex.zip` | `ce6ead70d4cbec718c2737bddcd0abe7e5cc984b81549f8188af5008c7b4f1fd` |
| `level7-dev-loop-1.0.0-dev-claude.zip` | `d7ae491f869ee0346a5437500272d26fd90a472c494d141ea129d6addab4df3d` |
| Frozen `level7-dev-loop-codex-0.1.1.zip` | `58ec422efd1b672f3c5d2aa6d1e7672077fb7741c68abcc548179c188f329dba` |
| Frozen `level7-dev-loop-claude-0.1.1.zip` | `0a589d5566ffb6498f0501f76cd198ac0100edc3570a07f094fe1de595241c49` |

## Platform and evidence boundary

| Evidence | Status |
|---|---|
| macOS 26.5.2 arm64, Go 1.26.7, Swift 6.3.3 native execution | PASS |
| Darwin amd64 cross-build, Mach-O identity, archive validation, and reproducibility | PASS |
| Darwin amd64 native execution | `NOT_RUN` — requires a separately authorized hosted or physical amd64 run |
| Hosted CI execution | `NOT_RUN` — workflow wiring is present but hosted execution was not authorized |
| Live Codex, Claude, or gateway provider calls | `NOT_RUN` |
| Real Codex/Claude host installation or settings mutation | `NOT_RUN` |

This is implementer-run technical verification, not an independent audit or
owner GO. No independent reviewer was commissioned, and no audit record, push,
signing, release, publication, protected-branch merge, installation, or
deployment was performed.

The next executable transition is to stop for separate Product Owner authority
to commission one independent read-only remediation audit by a reviewer
distinct from `codex-root`, bound to the verification-record commit and tree.
Any implementation or gate change invalidates this record and requires fresh
exact-head verification.
