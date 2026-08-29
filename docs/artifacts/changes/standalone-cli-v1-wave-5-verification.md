# Standalone CLI v1 Wave 5 — Verification

| Field | Value |
|---|---|
| Change ID | `standalone-cli-v1-wave-5` |
| Candidate commit | `43badc1d9681f392b8e4cc3e86346d8c20a784f4` |
| Candidate tree | `a5db3b063b8898a4a131cf83a8bf4a9f353cb37a` |
| Result | `PASS` |
| Reviewer | `addressanup` |
| Verified at | `2026-08-29T16:59:02Z` |
| Local host | `darwin/arm64` |
| Toolchain | Repository-pinned Go `1.26.7` |

## Checks

| Check | Result |
|---|---|
| Approval and controller binding | PASS locally — external active-user approval binds owner `apbusinessidentity-tech`, implementer `addressanup`, and exact brief addition `20d956075d85ff0c3439b4613e25488214a34120`; the controller selected Tier 3, base `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, candidate/tree above, 22 changed paths, and truthful pre-verification state `building` |
| Lineage and scope | PASS — the candidate is six commits above the exact base; every changed path is in the approved Wave 5 set; production CLI/runtime, provider adapters and pins, `.l7/config.json`, module metadata, skills, and trusted-policy/controller code are byte-for-byte unchanged |
| Module, lint, and boundary gates | PASS — offline module download reported no dependencies, module verification and tidy diff passed, all harness shell files parsed, formatting and `git diff --check` passed, `go vet` passed, `actionlint` passed, the 21-package import/effect boundary set passed, and the dedicated three-package Level 7 closure plus two negative probes passed |
| Distribution drift and reproducibility | PASS — one strict descriptor reproduced both manifests and catalogs, two clean builds were byte-identical, every archive entry was reinspected against its allowlist, and both development package digests remained stable |
| Lifecycle and containment fixtures | PASS — clean install, identical reinstall, upgrade, both interruption points, recovery, rollback, preview, and removal passed in disposable roots; missing/malformed receipts, changed or unowned files, package substitution, same-digest version mismatch, symlinked parents, and receipt reclassification failed closed while preserving bytes outside the owned set |
| Full test suite | PASS in an isolated exact-head run across all packages. A preceding diagnostic run executed the non-race and race suites simultaneously; its non-race fake-provider test hit one version-probe timeout under shared-host contention while the race suite passed. No candidate byte changed, and the single isolated non-race rerun passed without retrying any real provider |
| Race suite | PASS — CGO-enabled race instrumentation passed `./...` on Darwin arm64 with the exact harness identity link flags |
| Tagged actual-host coverage | PASS as compile-only evidence — `l7_actual_provider` files compiled across `./...` with `-run '^$'`; every package reported no selected tests and no provider executable ran |
| macOS architecture builds | PASS — the pinned offline toolchain produced Mach-O arm64 and x86_64 CLI binaries |
| Binary reproducibility | PASS — two independent clean-cache builds produced byte-identical harness test binaries and byte-identical Darwin arm64 CLI binaries |
| Hosted push Harness | PASS — run `33264206679`, event `push`, exact head above; baseline, shadow, arm64, and amd64 jobs succeeded, while the PR-only paired benchmark job was correctly skipped |
| Hosted PR Harness | PASS — run `33264221566`, event `pull_request`, exact head above; baseline, shadow, arm64, amd64, and paired same-host benchmark jobs all succeeded |
| Trusted policy pre-record boundary | EXPECTED FAIL-CLOSED — labeled run `33264222322` evaluated the exact head and stopped at `AUTH-001` because PR #6 has no exact-head GitHub owner review. Conversation-local authority is deliberately unavailable to hosted CI. The candidate remains blocked and this result is not represented as ready |
| Git and workspace hygiene | PASS — the implementation worktree and index were clean before this record; remote branch and PR #6 remained at the exact implementation head, remote `main` remained at the exact base, no review or merge occurred, and the unrelated user-owned foundation audit remained untouched, uninspected, and unstaged |

## Deterministic identities

| Output | SHA-256 |
|---|---|
| Codex development package | `9e54fff83a4ef3812bcfeb8737ec095305c828c7fd33e35926ae54588df39fd0` |
| Claude development package | `718ea9366ac6d286a954e655275f994de9d6e9fd2679123efda903c8f6881acb` |
| Harness test binary | `e46823dcaebf66cb798f7da0d65aba345cabfe55bb375d072508341018ba26da` |
| CLI Darwin arm64 | `5cf178c9fcea14e78f3c6885db3cdef938fe50aa6a87289d5e7cb8f4309713cb` |
| CLI Darwin amd64 | `ea82462fb51e1a55b84adac0b89c8a57f34e7268095914e4b60073c94f75c7f9` |

## Hosted evidence

| Evidence | State |
|---|---|
| PR | `https://github.com/addressanup/level7-dev-loop/pull/6`; open, Tier 3 label present, base `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, head `43badc1d9681f392b8e4cc3e86346d8c20a784f4`, review required, merge blocked |
| Push Harness | `https://github.com/addressanup/level7-dev-loop/actions/runs/33264206679`; `success`; completed `2026-08-29T16:59:02Z` |
| PR Harness | `https://github.com/addressanup/level7-dev-loop/actions/runs/33264221566`; `success`; completed `2026-08-29T16:58:31Z` |
| Trusted policy | `https://github.com/addressanup/level7-dev-loop/actions/runs/33264222322`; expected pre-record `failure` at `AUTH-001`; completed `2026-08-29T16:54:48Z` |

The successful Harness results are exact-head technical evidence. They do not
replace owner approval, independent audit, final exact-head reviews, or a
successful ready-state trusted-policy evaluation. The verification successor is
not pushed by this transition, so none of the implementation-head checks or
future reviews are silently transferred to it.

## Provider, package, and release boundary

No real Codex or Claude executable, version/help probe, prompt, stdin, model
session, retry, fallback, provider network, host package manager, installation,
update, removal, signing, notarization, publication, deployment, pilot, or
release participated. Provider execution and real host lifecycle remain
`NOT_RUN`; support remains `WITHHELD`. The compile-only tagged check and
filesystem fixtures cannot promote either host or establish a dual-host,
Intel-runtime, security, authenticity, supported-installation, or stable-v1.0
claim.

## Rollback proof

The verified implementation chain is:

1. base `f92c560cbe89e8d318e5521d9fc620f6153e9e14`, tree `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0`;
2. brief `20d956075d85ff0c3439b4613e25488214a34120`, tree `4e787f4e54294e58235abfe3c7d9415df69182a8`;
3. boundary tests `fe82eccb6f09c980798c94412cc70e1c2c25b453`;
4. deterministic packages `3037adb88ae1bbf505fb33a35958bd2d59de88e4`;
5. lifecycle fixtures `a2e9c4c0e33d755510f879a0ccdade948f9e1fad`;
6. CI and documentation wiring `62d05f9fef01dcedbfbf43b3c7b0a47a905e3772`; and
7. containment remediation candidate `43badc1d9681f392b8e4cc3e86346d8c20a784f4`, tree `a5db3b063b8898a4a131cf83a8bf4a9f353cb37a`.

Before audit, revert this verification-record commit and then the six Wave 5
commits in reverse order. After audit, revert its record first. Use ordinary
reviewed reverts, fail closed on conflict or unexpected paths, and confirm the
restored tree equals exact base tree `3b4f7fe9dd09fbb53102e82473d392dcb2745ba0`;
never rewrite preserved history.

## Verification boundary

This is implementer-run technical verification, not an independent audit,
owner review, release decision, merge authority, deployment approval,
installation authority, publication, or support claim. Any implementation
change invalidates this record and requires a new verification successor. The
only next Level 7 transition is a separately authorized independent read-only
`l7-release` audit of this verification successor.
