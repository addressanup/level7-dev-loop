# Level 7 Dev Loop

> Wave 2 semantic-evaluation development candidate. This repository is not a released, supported, security-qualified, or self-healing product.

Level 7 Dev Loop is being designed as a dual-host development-governance system for Codex and Claude Code. Its intended scope spans greenfield, brownfield, and legacy work while preserving evidence, explicit approval, small changes, and safe degraded behavior.

The current host manifests and skills are prototype inputs. They provide an advisory workflow only; they do not enforce filesystem isolation, trusted approval, mutation control, provider privacy, deployment safety, or release readiness. The approved architecture separates future advisory host packages from a separately installed controlled client. Neither controlled mode nor product behavior exists in this harness.

## Wave 2 candidate status

The current Wave 2 candidate freezes provider-neutral semantic/reference compiler interfaces and public local deterministic evaluation controls for development. It includes:

- versioned taxonomy, obligation, guardrail, knowledge, workflow, profile, prompt-IR, and output contracts derived from the approved 29-requirement scope;
- a pure, bounded, zero-production-dependency reference compiler with provider-neutral terminal and JSON projections;
- public semantic, boundary, degraded, interruption, and seeded broken-candidate fixtures;
- a frozen deterministic public protocol, truth labels, adjudication, grader registry, and exact 29-requirement/29-obligation coverage map; and
- fail-closed build control over the exact candidate manifest and the separately owned 21-path evaluator-control freeze.

The public protocol is local development evidence only. Its supplemental model-judge descriptor is `NOT_EVALUATED`, and its protected-holdout record is a contract boundary only: no protected corpus, labels, thresholds, credentials, operator, or result exists. No prompt, workflow, routing, model, grader, truth, adjudication, or threshold tuning was performed from comparative candidate outcomes.

This candidate has no user-facing command, host package or support result, model run, protected-corpus result, controlled mutation, security qualification, stable release, deployment, or exposure. It does not complete Wave 2: a direct evidence-only child and a fresh structurally separate independent audit are still required.

## Historical Wave 1 checkpoint

Requirements, backlog, architecture, and technology selection are recorded under `docs/artifacts/`. Foundation Step 5 established the inert developer harness. The independently audited historical Wave 1 checkpoint added a fail-closed, phase-aware build-control successor that:

- derives exactly 163 normative requirements and their `140 V1.0 / 18 V1.x / 5 Later` allocation from approved source records;
- records a non-support claim matrix and one disposition for each of the 12 prototype skills;
- binds the active phase to an immutable base commit, tree, manifest, and exact path policy;
- rejects unknown paths, deletion, protected-byte drift, non-regular files, ambiguous ownership, product paths, dependencies, and updater activation; and
- preserves the prior Foundation Step 5 checker as immutable historical source while making the Go validator the active policy gate.

The retained harness still provides:

- exact Go 1.26.7 baseline and Go 1.27.0 shadow archive locks;
- official archive digest and detached-signature verification;
- repository-scoped toolchain and Go caches—never a system install;
- formatting, built-in lint, type/compile, test, import-boundary, and same-machine repeat-build checks;
- one in-memory structured-logging proving test with no product effect;
- a digest guard over approved foundation inputs and protected prototype files; and
- a no-secret, no-authority `.env.example` that is never auto-loaded.

The root module is `github.com/addressanup/level7-dev-loop`, selected from the accountable owner's assertion that the project will live in the personal GitHub account `addressanup`. This is sufficient only for the local candidate: repository existence, account control, remote binding, publication, and compatibility remain unauthenticated and `NOT_RUN`. The module registry marks the separately required updater module `reserved`. Creating `cmd/l7up` fails the harness until a later approved wave assigns its module identity and adds module-specific dependency enforcement; the root-module boundary check cannot silently treat a nested module as covered.

There are intentionally no production dependencies, `go.sum`, vendored modules, product commands, runtime packages, generated host packages, provider calls, host experiments, release jobs, or deployment paths.

## Verify the harness

Prerequisites are a POSIX shell plus `make`, `curl`, `tar`, `awk`, `gpg`, and `gpgv`. Bootstrap downloads only a lock-matched official Go archive and verifies its size, SHA-256 digest, archive paths, Google signing-key fingerprints, and detached signature using an isolated repository-scoped GnuPG home. The signing identity is Google's shared Linux Packages Signing Authority, not a Go-exclusive release key. A fresh run extracts under ignored `.cache/toolchains/`; a later run reauthenticates the cached archive and signature but only checks the extracted toolchain's writable receipt and reported version. The extracted cache is not a trust root and must be discarded to force fresh extraction.

```sh
make bootstrap
make install
make policy-check
make import-check
make candidate-check
make lint
make typecheck
make test
make reproducible
```

Or run all blocking local gates:

```sh
make verify
```

The exact shadow toolchain can be checked separately:

```sh
make bootstrap GO_VERSION=1.27.0
make verify GO_VERSION=1.27.0
```

`make policy-check` runs the active phase-aware controller with the pinned local Go toolchain and validates the Wave 2 final-candidate closure when its manifest is present. `make candidate-check` combines the phase/scope controller and import-boundary gate. `make install` is intentionally offline and installs no production module: the current dependency count is zero. `make reproducible` compares two local builds with separate build caches; it is a smoke check, not the independent protected-build evidence required for release.

The offline settings close Go module and VCS resolution for these commands; they are not an operating-system network sandbox. The proving tests have no network import or call. Go telemetry is redirected into the ignored repository cache and pinned `off`; verification fails if either exact compiler no longer exposes that tested isolation mechanism.

The GitHub workflow is configured but remains `NOT_RUN` because no remote workflow run is part of this candidate. Its baseline job is blocking; Go 1.27.0 is an allowed-to-fail shadow observation. A mutable hosted-runner label is development evidence, not an exact production platform identity. The historical Wave 1 checkpoint retains its distinct evidence and audit lineage. This Wave 2 candidate remains incomplete until its direct evidence-only child and separately authorized independent read-only audit; it creates no product behavior or support claim.

## Prototype skill entry point

In an advisory host session, start with the Level 7 next-step skill:

```text
/l7-next
```

Codex may expose the same skill as `$l7-next`. Installation and host-lifecycle claims remain future C−1 experiments; copying or installing the prototype does not create a controlled environment.

## Working contract

- Spec before code.
- One approved Level 7 phase at a time.
- Evidence must distinguish `PASS`, `FAIL`, `BLOCKED`, and `NOT_RUN`.
- High-risk changes require an independent read-only audit.
- User-visible production behavior defaults OFF behind an approved feature flag.
- Phase completion updates `docs/artifacts/`.
- No prompt, skill, hook, or model self-claim is an authority boundary.

The approved future source/output layout and every unresolved acceptance spike remain in `docs/artifacts/technology-selection.md`; directories are created only when an approved build wave needs them.
