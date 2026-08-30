# Claude Package Validation Qualification

| Field | Value |
|---|---|
| Change ID | `claude-package-validation-qualification` |
| Risk tier | `3` — actual-host process execution and qualification evidence |
| Status | Proposed; awaiting fresh external accountable-owner approval of this exact brief commit |
| Base commit | `1ffe87310320d10103aaa625d44d10d217652974` |
| Base tree | `53be6fe314c9ec57d12a82c2db3cd8a273d49d66` |
| Assurance | `team` required for this Tier 3 wave; trusted team mode is not yet established for this branch |
| Accountable owner | Configured GitHub user `apbusinessidentity-tech`; configuration is not approval |
| Independent reviewer | Unresolved; must be a trusted forge identity distinct from owner and implementer before audit |
| Actual-host gate | Default OFF: `l7_actual_provider` build tag, exact run-binding envelope, and active-TTY confirmation; authority is evaluated externally |
| Runtime feature flag | Not applicable — test-only host validation; production behavior is unchanged |

## Problem

The current distribution report proves deterministic offline package bytes and
synthetic filesystem lifecycle behavior. It does not prove that the target host
accepts those bytes. Existing actual-host provider gates exercise bounded
diagnostic interfaces but do not consume a generated plugin package.

Claude Code documents a noninteractive `plugin validate` command for a plugin
directory containing a manifest. Current Codex documentation directs local
plugin installation and testing through the ChatGPT desktop app, so Codex host
qualification is deferred rather than represented by undocumented automation.

## Scope

After fresh accountable-owner approval of this brief commit, and only under
trusted team assurance, add a Claude-only, test-only qualification harness. A
separately authorized future run makes exactly two Claude process calls, in
this order, with empty stdin:

```text
claude --version
claude plugin validate <physical-generated-Claude-plugin-root>
```

The version preflight is limited to 10 seconds and 64 KiB combined output. Its
trimmed output must equal the repository-accepted literal `2.1.247` or
`2.1.247 (Claude Code)`; the observation binds the raw value and the exact
compatibility-matrix target `2.1.247 (Claude Code)`. The validation call is
limited to 30 seconds and 1 MiB combined output and occurs exactly once.

Execution uses a clean, detached, no-remote checkout of the final audit commit,
not the earlier implementation commit. Before package inspection, the harness
requires a team-mode controller `ready` decision for that exact head and binds
the complete lineage: brief commit and owner-approval envelope; implementation
commit/tree; verification commit/tree and record blob; audit commit/tree,
record blob, `GO` decision, reviewer identity, and audit-authority envelope.
Any missing, changed, non-ancestor, self-issued, or non-independent link fails
before a Claude process starts.

Before either call, the harness will consume a strict canonical schema-1
offline qualification report generated from the exact execution audit head by the
existing distribution `--check --json` path. It will require the report digest
bound by the run-binding envelope, select exactly its Claude entry, and
reject reordered, extra, cross-host, promoted, or noncanonical input.

The harness will independently inspect the generated output rather than attach
unrelated digests to a validated tree. It will:

- require the exact archive
  `build/distributions/level7-dev-loop-claude-<version>.zip` and the exact
  catalog `build/distributions/claude-marketplace/.claude-plugin/marketplace.json`;
- require the physical plugin directory
  `build/distributions/claude-marketplace/plugins/level7-dev-loop`;
- reopen the archive and prove exact path, mode, and content equality between
  every safe regular archive entry and every file in that physical tree;
- strictly decode the catalog, require one Claude plugin with the bound name and
  version, and prove its sole `./plugins/level7-dev-loop` source resolves to the
  same physical tree; and
- cross-check `DISTRIBUTION.json`, the report, archive, catalog, and tree before
  and after every process outcome.

The complete identity binding includes that approved and audited lineage and
exact execution commit/tree; package version and development channel; source
digest; offline-report, archive, catalog, and materialized-tree SHA-256 values
and physical paths; `GOOS` and `GOARCH`; and the physical Claude executable
path, SHA-256 digest, raw observed version, and normalized target.

The child-process environment contains exactly these keys and no inherited
entries:

```text
HOME=<owned-run-root>/home
CLAUDE_CONFIG_DIR=<owned-run-root>/claude
TMPDIR=<owned-run-root>/tmp
LANG=C
LC_ALL=C
TERM=dumb
NO_COLOR=1
GIT_TERMINAL_PROMPT=0
DISABLE_UPDATES=1
DISABLE_TELEMETRY=1
DISABLE_ERROR_REPORTING=1
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
ENABLE_CLAUDEAI_MCP_SERVERS=false
```

The harness uses an empty physical working directory under the owned run root.
It inherits no `PATH`, proxy, provider selector, provider credential,
`FORCE_AUTOUPDATE_PLUGINS`, or other environment variable. `HOME`,
`CLAUDE_CONFIG_DIR`, its plugin cache, `TMPDIR`, and the working directory are
empty physical non-symlink directories created for the run.

`CLAUDE_CONFIG_DIR` does not isolate credentials stored in the macOS system
Keychain. The environment controls also do not prove network denial. The
external immediate-effect authorization must acknowledge both residuals; the observation marks
`provider_execution`, `network_effects`, and `credential_access` as
`NOT_EVALUATED`. No prompt, model request, provider subcommand, installation,
activation, or marketplace mutation is requested, but absence of those latent
effects is not claimed.

The future run accepts one strict, nonsecret, canonical run-binding envelope
outside the candidate tree. It binds the change ID; 128-bit lowercase-hex run
ID; issue time no more than five minutes before validation; expiry strictly
after issue and no more than ten minutes later; the complete
brief/approval/implementation/verification/audit lineage; exact audit
head/tree; platform; executable tuple; normalized version target;
offline-report digest; package identities and paths; full archive digest; exact
argv; and explicit network/Keychain residual acknowledgment. The active
operator must type the full archive digest and run ID on a TTY.

The envelope and TTY exchange bind intent to bytes; they do not authenticate an
owner, prevent replay across processes, or constitute authority. The harness
therefore records `execution_authority=NOT_EVALUATED`. The orchestrator must
obtain fresh explicit immediate-effect authorization for the exact audited head,
command, executable, package digest, platform, and acknowledged residuals before
launching the tagged test. A file, token, passing test, prior approval, or TTY
possession alone never supplies that authority.

After both calls, postconditions, and owned-run-root cleanup succeed, the
harness emits exactly one canonical schema-1 JSON observation to verbose test
stdout with a stable marker and writes no observation file. It records the run
ID and envelope digest, complete audited lineage, confirmation binding, all
package and executable identities, exact argv and environment digest, both exit
codes, separate stdout/stderr digests for each call, cleanup result, macOS
residual, and `execution_authority=NOT_EVALUATED`. Raw child output is not
persisted. A failed, interrupted, drifted, or incompletely cleaned run cannot
emit `PASS`. External capture, retention, and review of the observation are out
of this change's scope.

Only `host_package_validation` may be `PASS`. Marketplace lifecycle,
installation, activation, behavioral activation, signing, and publication
remain `NOT_RUN`; provider execution, network effects, credential access,
execution authority, and release authority remain `NOT_EVALUATED`; support
remains `WITHHELD`; and `release_ready` remains false. The observation is
separate from and cannot promote or replace the offline report.

The actual Claude calls do not run during implementation, ordinary CI,
verification, or independent audit.

## Exact implementation file set

Add now:

- `docs/artifacts/changes/claude-package-validation-qualification.md`

Add only after valid owner approval and trusted team assurance:

- `internal/l7/adapter/claude/package_validation_test.go`
- `internal/l7/adapter/claude/package_validation_actual_host_test.go`

Add after implementation verification:

- `docs/artifacts/changes/claude-package-validation-qualification-verification.md`

Add only through the independent read-only audit:

- `docs/artifacts/changes/claude-package-validation-qualification-audit.md`

Modify no existing file. In particular, do not change production adapters,
distribution code or its import boundaries, constants, versions, manifests,
catalogs, compatibility claims, the offline qualification report, `Makefile`,
workflows, policy controls, or dependencies. Do not add Codex automation,
marketplace lifecycle, install/activation behavior, provider/model execution,
signing, publication, deployment, release, or support claims. The runtime
observation is not a repository governance artifact.

## Acceptance criteria

1. Strict parsing selects the sole Claude entry from the exact canonical offline
   report and binds it to the expected archive, catalog, and physical plugin
   tree. Archive-to-tree path/mode/content equality and exact catalog source
   resolution are proven before process execution.
2. A pure request/report evaluator and fake runner prove exactly two calls in
   order: one bounded `--version` preflight and one
   `plugin validate <physical-plugin-root>` call. Both use empty stdin, the same
   exact executable, empty physical working directory, and exact environment.
3. The tagged test binds the complete brief-through-audit lineage and execution
   head/tree, package version/channel/source/report/archive/catalog/tree
   identities, `GOOS`/`GOARCH`, and executable path/digest plus the exact
   observed and normalized Claude version.
4. The run-binding envelope is canonical, nonsecret, within the exact
   five-minute issue and ten-minute expiry bounds, and specific to the complete
   audited run tuple. Active-TTY confirmation rejects tuple mismatch. The
   harness never promotes those controls into an authority or anti-replay claim;
   explicit immediate-effect authorization is an external launch prerequisite.
5. Source, package, report, archive, catalog, materialized tree, executable,
   envelope, and output identities are checked immediately before and after
   every outcome. Any drift fails closed. Cleanup removes only the owned run
   root and no repository or external authority path.
6. Fake-runner and filesystem tests reject wrong call count/order/argv/env,
   cross-host identities, malformed or noncanonical reports/envelopes/catalogs,
   archive/tree differences, unsafe paths/modes/symlinks, digest/version/root
   substitutions, invalid or stale lineage, nonzero or invalid output,
   oversized output, timeout, cancellation, pre/post drift, missing
   postconditions, and cleanup failure.
7. The canonical observation records separate output digests and exact outcomes;
   only `host_package_validation=PASS` may pass. Every broader state remains
   exactly `NOT_RUN`, `NOT_EVALUATED`, `WITHHELD`, or false as scoped above.
8. Ordinary, full, and race tests skip the actual host calls. The existing
   tagged compile target compiles them without selecting the test. No Claude,
   provider, plugin, credential, or network operation runs during implementation.
9. Focused and race tests, `make verify`, declared cross-builds,
   `make distribution-check`, team-mode policy checks, and diff hygiene pass on
   the exact implementation candidate. The sole verification record binds those
   results and makes no actual-host claim.
10. A distinct trusted reviewer performs one `l7-release` read-only audit of the
    exact verification head and adds only the sole audit record. A later run
    binding must name that exact `GO` lineage. Neither the audit,
    passing CI, nor a locally supplied envelope authorizes the Claude calls;
    fresh explicit immediate-effect authorization remains mandatory.

## Risks and mitigations

- **Credential-capable host:** ambient credentials are stripped and writable
  roots isolated, but macOS Keychain access remains possible. Authorization
  explicitly acknowledges it and the result stays `NOT_EVALUATED`.
- **Network or telemetry effects:** documented controls disable updates,
  telemetry, error reporting, connector fetches, and nonessential traffic, but
  are not network containment. Authorization acknowledges the residual and the
  result stays `NOT_EVALUATED`.
- **Package identity substitution:** canonical offline input plus exact
  archive/tree/catalog equality and pre/post identity checks prevent unrelated
  digests from being attached to the validated directory.
- **Invalid authority or replay:** the harness makes no authentication or
  one-use claim. The orchestrator stops before launch unless fresh explicit
  authority covers the exact audited tuple; the observation truthfully records
  authority as `NOT_EVALUATED`.
- **Semantic overreach:** host validation is independent from install,
  marketplace, activation, behavior, provider, signing, publication, support,
  and release states.
- **Cleanup damage:** only a newly created physical child of the exact authorized
  temporary parent may be recursively removed; repository, envelope, and
  external capture paths are never cleanup roots.
- **Host/output drift:** exact version targeting, bounded outputs, distinct
  digests, and before/after checks fail closed.
- **One-host inference:** Claude evidence cannot qualify Codex or the dual-host
  distribution.

## Rollback

Before integration, revert the brief, implementation, verification, and audit
commits in reverse order. The implementation and its verification execute no
host and create no installed, published, or deployed state. A later authorized
trial owns only its temporary run root. It emits an observation to stdout and
creates no retained evidence file; any external capture or cleanup is a
separately authorized concern.

The only accepted transitions are:

1. commit this brief only and stop;
2. establish trusted team assurance, resolve a distinct reviewer, and obtain
   fresh accountable-owner approval bound to the brief commit;
3. add only the two test files without executing Claude;
4. verify the exact candidate and add the sole verification record;
5. perform one independent `l7-release` audit and add the sole audit record; and
6. request separate immediate-effect authorization before any actual Claude
   invocation.

## Primary documentation

- [Anthropic Claude Code plugins reference](https://code.claude.com/docs/en/plugins-reference), accessed 2026-08-30.
- [Anthropic Claude Code environment variables](https://code.claude.com/docs/en/env-vars), accessed 2026-08-30.
- [OpenAI plugin packaging contract](https://developers.openai.com/plugins/build/plugins), accessed 2026-08-30.
