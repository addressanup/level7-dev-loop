# Level 7 Dev Loop — Technology Selection

| Field | Value |
|---|---|
| Artifact ID | `L7-TEC-001` |
| Artifact type | Technology decision and compatibility baseline |
| Artifact schema | Bootstrap/pre-schema; migrate only through an approved artifact migration |
| Foundation step | 4 — Technology selection |
| Status | Proposed — conditional technology baseline; owner approval required |
| Version | 0.2.0 |
| Date | 2026-08-24 |
| Inputs | Approved `L7-REQ-001` 0.2.0, approved `L7-BKL-001` 0.1.0, approved/audited `L7-ARC-001` 0.2.0 |
| Architecture SHA-256 | `73e38782775a682f191d2bfef3ee8d239fbab2c7e10744336e2bd6065902748a` |
| Scope identity | Non-Git workspace snapshot observed on 2026-08-24; no commit identity available |
| Initial hosts | Codex CLI and Claude Code |
| Effect and risk | A1 technology-decision artifact only; no product/runtime mutation |
| Next authorization if approved | Foundation step 5 — harness construction only |

**Decision boundary:** this document selects an implementation baseline, not a support claim. It authorizes no source, prompt, skill, manifest, package, CI, dependency, installation, host-setting, experiment, deployment, publication, or release change. Every host, OS, approval, containment, receipt, and reproducibility claim remains unproved until its named actual-host test passes.

## 1. Outcome first

Select a **pure-Go, invocation-scoped local core** and be explicit that the proposed product is a **bundle of two independently installed product categories**, not one plugin with a hidden privileged mode:

1. **Level 7 host packages — A0 ceiling.** Separate generated Codex and Claude packages expose one advisory conductor from one semantic source. The Level 7 conductor does not request project reads or writes by default and contains no hook, MCP server, executable, or script in v1 A0. This is a behavioral contract, not a privacy or enforcement boundary: the surrounding stock host retains its ordinary workspace, shell, connector, memory, and provider capabilities. Installing either package alone never enables controlled mutation.
2. **Level 7 Controlled Client — A1/A2 candidate.** A user separately installs and launches the authenticated, root-owned `l7` companion. It owns the real repository, trusted confirmation, context projection, provider-auth relay, deterministic kernel, artifact writer, mutation executor, and live verifier. Each model turn is a fresh one-shot Codex/Claude child inside an outer OS containment profile. The child gets only a bounded disposable projection, provider egress through a session-bound auth-injecting relay, and no real provider credential or mutation capability. The child and its process namespace terminate before AP1 is displayed, so it cannot answer or spoof the confirmation prompt.

The selected core is **Go 1.26**, initially built with the exact patched **Go 1.26.7** toolchain, `CGO_ENABLED=0`, with Go 1.27.0 in shadow CI. The binary uses the standard library first, `os.Root` for brokered repository access, JSON Schema Draft 2020-12 for structure, UTF-8 I-JSON plus RFC 8785 JCS for semantic record digests, SHA-256 with domain separation, and ordinary repository files as canonical memory. No runtime database, graph database, vector database, Level 7 service, daemon, MCP server, native hook, subagent, external memory, or telemetry is required.

For local gate-bearing evidence, v1 selects **fresh admitted reproduction**, not a repository or same-user signing key. Durable release evidence uses **DSSE/in-toto attestations, SLSA 1.2 provenance, and Sigstore verification rooted in a selected GitHub Actions release-control plane**. The current non-Git workspace is therefore development-only; no distribution claim exists until the protected Git/GitHub prerequisites and distinct identities are actually configured and evidenced.

### 1.1 The decisive finding

Current official host controls do not establish a non-bypassable A1/A2 boundary for an ordinary stock plugin:

- a normal writable Codex workspace leaves ordinary file/shell paths available;
- Claude's Bash sandbox does not govern its built-in Edit/Write tools;
- both hosts document hooks as guardrails with bypass or failure limitations, not complete enforcement boundaries;
- neither plugin package can, by itself, produce Level 7's exact, one-use, digest-bound human approval receipt; and
- host-native reads, attachments, settings, connectors, or tools can bypass a plugin-level context gateway unless the controlling application removes or isolates them.

Therefore stock-plugin A1/A2 is rejected. `AR-001` and `AR-002` remain critical and unproved. The controlled-session design is the only selected experiment that plausibly preserves the approved architecture without making hooks, prompts, or a mandatory MCP/daemon authoritative. If it fails, stable dual-host A2 is `NO_GO` until requirements are narrowed and reapproved.

## 2. Evidence method and confidence

Three separate read-only specialist reviews covered host mechanics, runtime/platform primitives, and data/evidence/supply-chain choices. The primary decision then reconciled them against `AI-*`, `QA-*`, `AF-*`, the P0 dependency graph, and the approved prohibition on prompt-only enforcement.

Evidence labels in this document are:

| Label | Meaning |
|---|---|
| `OBSERVED` | Read directly from this workspace or a local command on 2026-08-24. |
| `DOCUMENTED` | Supported by a cited primary specification, official product document, official release, or upstream project source. |
| `INFERRED` | Engineering judgment derived from evidence; not yet tested here. |
| `UNPROVED` | Plausible mechanism whose required actual-host/platform evidence does not exist yet. |
| `REJECTED` | Fails an approved hard gate or loses to the selected alternative. |

Scores are `INFERRED`, use 1 (poor) through 5 (strong), and calculate `weight × score / 5`. A score never overrides a hard gate.

### 2.1 Local observation baseline

| Observation | State | Consequence |
|---|---|---|
| Codex CLI `0.149.1` | `OBSERVED` | C−1 seed only; not a supported or release-pinned version. |
| Claude Code `2.1.241` | `OBSERVED` | C−1 seed only; not a supported or release-pinned version. Its local help did not expose the currently documented `--max-turns`, so exact parser/behavior proof is a blocker rather than an assumed flag contract. |
| macOS `26.5.2`, Darwin `25.5.0`, arm64 | `OBSERVED` | First local smoke platform; no A1/A2 claim. |
| Node `24.19.0` | `OBSERVED` | Available locally but rejected as the trusted core runtime. |
| Go and Rust toolchains absent | `OBSERVED` | No compiler, module, dependency, or harness installation is authorized by this step. |
| No Git worktree, build, test, CI, lockfile, schema, or license file | `OBSERVED` | Bootstrap and legal/supply-chain gates remain explicit. |
| Existing manifests say `1.0.0`/MIT without release evidence or an authored license | `OBSERVED` | Those fields are prototype inputs, not accepted release truth. They remain untouched in Step 4. |

## 3. Non-negotiable technology gates

The selected stack must fail closed on all of these:

| Gate | Required outcome |
|---|---|
| `TG-01 Proposal separation` | Model, prompts, skills, repository content, retrieval, and subagents cannot mint authority, lower risk, or perform a real-repository write. |
| `TG-02 A1/A2 closure` | Every Level-7-issued artifact/source mutation reaches the same deterministic admission boundary; any alternate path caps the mode at A0. |
| `TG-03 AP1 provenance` | Mutation confirmation is a parent-controlled user event bound to exact action, target, pre-state, environment, expiry, and nonce—not text or auto-review. |
| `TG-04 Context mediation` | The model child cannot read the real repository or uncontrolled host context; it sees only a minimized projection and disclosed initial-provider ingress. |
| `TG-05 Producer authenticity` | Gate-bearing local evidence is freshly reproduced through a live protected channel or externally attested; editable self-claims cannot advance. |
| `TG-06 Rooted filesystem` | Traversal, symlink/junction escape, collision, stale target, and unexpected-effect tests pass per supported OS. |
| `TG-07 Process containment` | Approved verification commands have no undeclared filesystem, credential, network, descendant, or resource authority. Missing containment blocks command-derived evidence. |
| `TG-08 Recovery` | Multi-file interruption cannot become PASS or silently overwrite concurrent work; ambiguous recovery becomes `RECOVERY_REQUIRED`. |
| `TG-09 Reproducibility` | Two clean builders reproduce the normalized unsigned payload byte-for-byte; signing and channel metadata bind that digest. |
| `TG-10 Minimal installation` | A0 works without Level 7 SaaS, service, daemon, MCP, hook, subagent, database, vector store, telemetry, or language runtime. |
| `TG-11 Exact compatibility` | Every released host, OS, schema, core, package, and external tool version is exact and tested; ranges and “latest” are not support evidence. |
| `TG-12 A3–A5 absence` | V1 contains no executable remote, production, destructive, background, self-modifying, or autonomous-remediation interface. |

## 4. Trusted-core runtime comparison

### 4.1 Candidates and score

| Criterion | Weight | Go | Rust | TypeScript/Node |
|---|---:|---:|---:|---:|
| Memory and type safety | 14 | 4 | 5 | 3 |
| Rooted filesystem/path capability | 18 | 4 | 4 | 1 |
| Transactions, locking, process control | 16 | 3 | 4 | 2 |
| Offline, single-binary packaging | 14 | 5 | 4 | 2 |
| Reproducible build and cross-compilation | 10 | 5 | 3 | 2 |
| Dependency/supply-chain exposure | 10 | 4 | 3 | 2 |
| Tests, fuzzing, properties, static analysis | 8 | 5 | 4 | 3 |
| Maintainability and version stability | 7 | 5 | 3 | 4 |
| Startup, performance, resource use | 3 | 4 | 5 | 3 |
| **Weighted total / 100** | **100** | **84.6** | **78.0** | **44.2** |

### 4.2 Selection

**Select Go.** Its principal advantage is not raw speed; it is a small operational surface combined with a standard-library rooted filesystem API, self-contained pure-Go binaries, integrated testing/fuzzing/race tooling, module checksums/vendor mode, and the Go 1 compatibility policy. The dominant v1 hazards are authority logic, path escape, subprocess effects, stale state, and recovery rather than manual memory management.

- Use source language floor `go 1.26`.
- Use exact Go `1.26.7` for the initial reproducible baseline; verify the official distribution digest.
- Shadow-test exact Go `1.27.0`; promotion requires the full path, OS, reproducibility, and conformance suites.
- Set `GOTOOLCHAIN=local`, `CGO_ENABLED=0`, `-trimpath`, vendor mode, deterministic link metadata, and no automatic toolchain download for release builds.
- Forbid `unsafe`, cgo, generic shell execution, model-selected executables, release-time `replace` directives, and network-capable dependencies in trusted packages unless a later audited ADR explicitly admits one.
- Permit `golang.org/x/sys` only inside a narrow OS adapter if an acceptance spike proves a required primitive has no adequate standard API.

Go does not solve process-tree containment, cross-platform file locking, multi-file atomicity, host bypass, AP1, or producer authenticity. Those remain separate gates.

### 4.3 Rejected/fallback runtimes

**Rust is the fallback**, not a co-primary runtime. If `os.Root` or the pure-Go platform adapter cannot pass the matrix, compare a Rust prototype using Bytecode Alliance `cap-std` and standard file locking. Switching requires a new technology revision because it changes toolchain, dependency, audit, build, and maintenance costs.

**TypeScript/Node is rejected for the enforcement core.** Its runtime/package footprint is larger, and Node documents its permission model as a trusted-code guardrail rather than a malicious-code boundary; allowed paths can follow symlinks beyond the nominal path set. TypeScript remains suitable only for generated host-facing text or an independently isolated development tool, neither of which is selected as a v1 runtime dependency.

## 5. Host-boundary comparison

### 5.1 Deployment patterns

| Criterion | Weight | H1: stock plugin + prompts/hooks | H2: controlled companion + optional A0 packages | H3: mandatory MCP/daemon |
|---|---:|---:|---:|---:|
| A1/A2 closure plausibility | 25 | 1 | 4 | 3 |
| AP1 and context separation | 20 | 1 | 4 | 3 |
| Cross-host semantic parity | 12 | 3 | 4 | 4 |
| Native user experience | 10 | 5 | 3 | 4 |
| Offline/minimal installation fit | 10 | 5 | 5 | 1 |
| Lifecycle/distribution fit | 8 | 4 | 3 | 2 |
| C−1 learning speed | 10 | 5 | 3 | 2 |
| Maintainability | 5 | 4 | 3 | 2 |
| **Weighted total / 100** | **100** | **56.6** | **75.4** | **55.8** |
| Hard-gate result |  | **Fail `TG-02`–`04`** | **Conditional** | **Fail `TG-10`; closure still unproved** |

### 5.2 Selection: H2 as a separate controlled client

Select H2 for A1/A2 experimentation and retain thin H1 packages for A0 convenience. This is an explicit product-category decision: **Level 7 Controlled Client** is a companion application, while **Level 7 for Codex** and **Level 7 for Claude** are optional advisory host packages. They share semantics and branding but not installation, execution, authority, update, rollback, or removal. Neither host package is loaded in a controlled turn (`claude --bare` deliberately skips it), and installing a plugin alone never implies that controlled mode exists.

The supervisor is the same short-lived Go `l7` binary, not a service or daemon. It owns the real repository and trusted terminal. A controlled interaction is a sequence of one-shot turns:

1. the parent reads user intent and creates a bounded, labeled prompt projection;
2. it launches an exact host binary inside the selected model-host containment profile with pipe input/output, a fresh state root, no controlling terminal, and no real-repository mount;
3. it validates one typed proposal, closes the relay, obtains stdout/stderr EOF, reaps the namespace init, proves the dedicated cgroup is unpopulated, and sanitizes all untrusted output;
4. only after that teardown proof does the parent render AP1 on its trusted terminal; and
5. an approved action executes inside the parent/kernel and is freshly verified without restarting the model child.

The child receives no live user stdin, no capability secret, no AP1 channel, no real provider/project/cloud credentials, and no real-repository path. It receives only a non-secret, one-turn relay token/base URL and provider egress through the parent. The parent holds a dedicated spend-limited provider credential in memory for the session, strips any child-supplied authentication, injects the real credential only on the provider-side request, and closes the relay before AP1. Model visibility, endpoint compatibility, logs, rotation, revocation, billing scope, and failure behavior are release gates in Sections 13–14. Correctness never depends on plugin hooks, MCP, a background process, or host plugin-data retention.

The initial command surface is intentionally narrow:

| Command | Effect ceiling | Purpose |
|---|---:|---|
| `l7 doctor` | A0 | Exact version/capability/provider-boundary diagnosis. |
| `l7 status` | A0 | Rebuild state and show one next action. |
| `l7 codex` | A0 initially; A1/A2 only after gates | Start a controlled Codex child session. |
| `l7 claude` | A0 initially; A1/A2 only after gates | Start a controlled Claude child session. |

There is no standalone `l7 apply`, generic `l7 exec`, remote command, deployment command, signing oracle, or model-callable mutation subcommand. A process not created as the user-owned supervisor cannot mint a mutation capability.

### 5.3 Codex adapter candidate

Select the documented one-shot `codex exec` surface, not App Server, for the controlled v1 candidate. App Server is currently experimental, has a broader privileged client API, and creates unnecessary session-state/RPC risk for a turn that needs only typed proposal output. It may be researched outside the stable path, but accepting it later requires a revised dependency/security decision.

The C−1 argv candidate, with every placeholder bound into the run manifest, is:

```text
codex --ask-for-approval never exec
  --ephemeral --ignore-user-config --ignore-rules --strict-config
  --sandbox read-only --color never --json
  --output-schema <owned-schema>
  --model <qualification-approved-provider-model-id>
  --cd <owned-empty-control-root> --skip-git-repo-check
  --disable shell_tool --disable unified_exec --disable multi_agent
  --disable memories --disable skill_mcp_dependency_install
  --disable plugins --disable remote_plugin --disable workspace_dependencies
  --disable enable_request_compression --disable hooks --disable apps
  --disable auth_elicitation --disable browser_use
  --disable browser_use_external --disable browser_use_full_cdp_access
  --disable code_mode_host --disable computer_use --disable goals
  --disable fast_mode --disable in_app_chat --disable in_app_dictation
  --disable in_app_updates --disable mentions_v2 --disable personality
  --disable guardian_approval --disable image_generation
  --disable in_app_browser --disable plugin_sharing
  --disable remote_compaction_v2 --disable shell_snapshot
  --disable skill_search --disable tool_call_mcp_elicitation
  --disable tool_suggest --disable unbounded_connection_retries
  --disable view_image
  -c 'agents.enabled=false' -c 'apps._default.enabled=false'
  -c 'web_search="disabled"' -c 'tools.web_search=false'
  -c 'tools.view_image=false' -c 'feedback.enabled=false'
  -c 'check_for_update_on_startup=false' -c 'history.persistence="none"'
  -c 'file_opener="none"' -c 'shell_environment_policy.inherit="none"'
  -c 'model_instructions_file="/l7/control/codex-system.md"'
  -c 'allow_login_shell=false' -c 'analytics.enabled=false'
  -c 'otel.exporter="none"' -c 'otel.metrics_exporter="none"'
  -c 'otel.trace_exporter="none"' -c 'forced_login_method="api"'
  -c 'openai_base_url="<owned-loopback-relay>"' -
```

The owned instruction file supplies the trusted conductor. Stdin contains only the bounded user/context envelope and is closed immediately after that envelope. The supervisor does not use `--output-last-message`, resume, fork, review, images, `--add-dir`, automatic approval, or bypass flags. Invocation overrides set approval to `never`; disable the named execution, agent, memory, skill, web, browser, app, image, hook, compaction, elicitation, retry, feedback, and update surfaces; set the default app state to disabled; configure only the owned loopback provider base URL plus a non-secret one-turn relay token; and supply no MCP server, plugin, profile, hook, or remote-plugin entry. An isolated disposable `CODEX_HOME` contains no real provider credential; `--ephemeral` must leave no rollout. `codex features list` from the exact binary and the complete effective request/config inventory enter qualification; a new, unknown, or unexpectedly default-on optional surface fails the tuple rather than inheriting a default. `--strict-config` proves only that recognized configuration parsed—it does not prove ambient/managed configuration absent—so the isolated home/root, outer mounts, and semantic relay remain authoritative.

The required model-callable tool inventory is empty. `codex exec` has no documented Claude-style `--tools ""` control or pre-execution callback, so admission is based on the complete outbound request observed by the relay **before forwarding**, not successful parsing of `--disable` flags or absence of later JSONL events. Before forwarding, the relay rejects a Codex request whose tool inventory is nonempty or whose qualification-approved requested model identifier, instructions, or messages differ from the admitted digests. After receiving the provider response—but before admitting the proposal—it validates any response-reported model/service identity against the qualification contract; missing or drifting required identity fails the tuple. If the exact qualified Codex build cannot emit that zero-tool request, Controlled Codex is removed from the matrix; allowing residual “inert” tools requires a separately approved architecture revision. Outer containment—not event timing—prevents access to real or persistent state before that decision. Any tool-call, mutation, approval, hook, unknown, or malformed JSONL event also invalidates the proposal even when the sandbox denied its effect. Any shell, app, MCP, browser, web, connector, subagent, skill, hook, arbitrary file-read, or effective mutation surface kills the controlled Codex entry rather than being prompt-disabled. The parent, relay, and OS enforce time, output, request, and spend ceilings because `codex exec` exposes no sufficient complete cost/turn/output-token contract for this boundary.

Codex skills can use package-relative helpers in applicable host packaging contexts, and `codex plugin list --json` can expose an installed path. Those facts may help an A0 helper later; they are not an authenticated discovery boundary for the external parent supervisor. The controlled client uses a separately verified, root-owned absolute companion path.

### 5.4 Claude adapter candidate

Select a dependency-light, nonstreaming Claude print boundary rather than adding the TypeScript/Python Agent SDK to the trusted runtime. The C−1 argv candidate is:

```text
claude --bare --disable-slash-commands -p
  --input-format text --max-turns 1 --effort <qualified-level>
  --tools "" --disallowedTools "*" --permission-mode dontAsk
  --strict-mcp-config --mcp-config <owned-empty-config>
  --no-chrome --no-session-persistence
  --max-budget-usd <exact-per-turn-budget>
  --output-format json --json-schema <canonical-schema-argv-element>
  --system-prompt-file <owned-file>
  --model <qualification-approved-provider-model-id>
```

The parent sets the process working directory to an owned empty control root and sends the bounded user/context projection as print-mode text. It loads canonical JSON Schema bytes and passes them as one direct argv element without a shell. It passes no `--add-dir`, `--agent(s)`, plugin, skill, file, IDE, worktree, browser, resume/continue, remote-control, fallback-model, debug, hook-event, safe-mode, or dangerous-permission option. A per-run cost ceiling is mandatory. `--tools ""` removes built-ins; the wildcard deny plus strict owned-empty MCP configuration closes MCP and other unexpected tool names; `--bare` removes project memory and customization. Safe mode is not added because it explicitly retains managed policy behavior. `--max-turns 1` bounds agentic tool-use turns but does not prove one provider request; the relay independently admits exactly one request, so a transport or structured-output retry fails the turn. After exit zero, the parent parses exactly one version-qualified Claude JSON result envelope, rejects error, duplicate, trailing, or unknown result forms, extracts only `structured_output`, and independently validates its schema and semantics. The complete stdout envelope remains bounded untrusted data and terminal prose is never accepted as the proposal.

The child environment is rebuilt from an allowlist and binds `CLAUDE_CONFIG_DIR=<disposable-state>`, `ANTHROPIC_BASE_URL=<owned-relay>`, `ANTHROPIC_API_KEY=<one-turn-placeholder>`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `CLAUDE_CODE_SKIP_PROMPT_HISTORY=1`, `CLAUDE_CODE_DISABLE_CLAUDE_MDS=1`, `CLAUDE_CODE_DISABLE_BUNDLED_SKILLS=1`, and `ENABLE_CLAUDEAI_MCP_SERVERS=false`. Every variable and flag must be recognized and have the observed fail-closed effect in the exact-version C−1 manifest; an ignored/unknown control removes that tuple. No structured-output retry override is selected until its exact name and zero-value behavior are independently proved.

Claude `--bare` intentionally bypasses subscription OAuth/keychain and therefore requires `ANTHROPIC_API_KEY` or an owned `apiKeyHelper`. V1 gives the child a non-secret one-turn placeholder key and points the exact supported client to the owned loopback relay; the parent replaces the placeholder with its session-held, dedicated, spend-limited provider credential only on the outbound request. Rotation, revocation, base-URL compatibility, crash/error/log redaction, `/proc`/environment inspection, prompt-exfiltration canaries, and residue deletion are mandatory tests. If the exact Claude version bypasses the relay or requires the real key inside the child, the controlled Claude entry is unsupported.

Every flag interaction, settings-validation failure, administrative-policy precedence, unexpected tool/MCP/plugin/hook event, inherited environment/file descriptor, and version-specific behavior is `UNPROVED` until C−1. Any managed hook, managed tool, MCP, customization, or context that survives `--bare` invalidates the result and removes that enterprise configuration from the supported matrix. Nonstreaming JSON exposes no same-run initialization inventory; a separate stream-JSON startup probe may inspect initialization/tool state, but it is not equivalent admission evidence and schema-plus-stream behavior cannot silently replace the selected contract. The semantic relay must validate the actual outbound model, all messages/instructions, and empty tool definitions. An Agent SDK fallback would add a trusted Node/Python surface and requires a new technology/dependency/security review.

### 5.5 Why hooks and MCP remain optional

Hooks may improve A0 diagnostics or block obvious misuse, but a missing, timed-out, malformed, disabled, or bypassed hook cannot permit an effect. A local MCP tool has no documented privileged OS boundary merely because it uses MCP. Neither is included in the v1 A0 package or controlled correctness path; a later addition needs measured value, lifecycle evidence, and a revised attack-surface decision. Neither can satisfy `AR-001`, `AR-002`, or `AR-011` alone.

## 6. Artifact format and storage comparison

### 6.1 Serialization

| Criterion | Weight | JSON + Schema + JCS | YAML + schema | Markdown/frontmatter + custom parser |
|---|---:|---:|---:|---:|
| Deterministic cross-host semantics | 25 | 5 | 3 | 2 |
| Human readability/editability | 15 | 4 | 5 | 5 |
| Validation ecosystem | 20 | 5 | 4 | 2 |
| Cryptographic binding | 15 | 5 | 2 | 1 |
| Migration/forward compatibility | 10 | 4 | 4 | 3 |
| Implementation/dependency cost | 10 | 4 | 3 | 4 |
| Prompt/attachment friendliness | 5 | 4 | 5 | 5 |
| **Weighted total / 100** | **100** | **92.0** | **71.0** | **55.0** |

Select **UTF-8 JSON + JSON Schema Draft 2020-12 + JCS** for canonical machine records. Markdown remains available for digest-bound narrative attachments and derived status, never as the authority-bearing lifecycle state.

### 6.2 Persistence/index options

| Criterion | Weight | Canonical JSON + in-memory view | Canonical JSON + SQLite cache | Canonical JSON + graph/vector service |
|---|---:|---:|---:|---:|
| Canonical truth/human audit | 25 | 5 | 4 | 2 |
| Deterministic rebuild | 15 | 5 | 5 | 3 |
| Offline packaging simplicity | 15 | 5 | 3 | 1 |
| Query/navigation | 15 | 3 | 5 | 5 |
| Concurrency/recovery fit | 15 | 4 | 4 | 3 |
| Privacy/supply-chain surface | 10 | 5 | 3 | 2 |
| Scale runway | 5 | 3 | 5 | 5 |
| **Weighted total / 100** | **100** | **89.0** | **82.0** | **55.0** |

Select **ordinary canonical files plus an in-memory adjacency/search view**. A disposable JSON status/graph projection may be emitted, but deletion must never change a decision. SQLite is not bundled in v1 because pure-Go SQLite would materially enlarge the dependency/TCB and cgo would break the selected build. It may be reconsidered only after published benchmarks show direct reconstruction misses an approved budget. A graph or vector database is rejected for v1.

## 7. Selected stack ledger

| Concern | Selected technology | Status/constraint |
|---|---|---|
| Trusted language | Go 1.26; exact compiler 1.26.7 | `INFERRED`; toolchain absent locally; acceptance spikes required. |
| Next compiler | Go 1.27.0 in shadow CI | Not a release compiler until separately promoted. |
| Runtime packaging | Pure-Go per-OS/arch `l7` binary | `CGO_ENABLED=0`; no user Go/Node/Python runtime. |
| Local control boundary | Separately installed Level 7 Controlled Client | Invocation-scoped; no daemon/service; must be the authenticated parent of each one-shot host turn for A1/A2. |
| Host presentation | Separate generated Codex and Claude A0 packages | Optional advisory shells; one public conductor; plugin installation never enables controlled mode. |
| Controlled Codex transport | One-shot `codex exec --ephemeral` | App Server rejected from stable v1; exact flags/config/events/model must pass C−1. |
| Controlled Claude transport | One-shot `claude --bare -p` JSON | Empty tools/MCP/customization; managed residue invalidates; exact flags/settings/model must pass C−1. |
| Model-host containment | Bubblewrap `0.9.0-1build1` on exact Ubuntu 24.04 x86_64 image | Real root absent; fresh state; no TTY; supervised provider-only relay; unavailable profile means A0. |
| Provider authentication | Parent-owned protocol-aware auth-injecting relay; no real key in child | Dedicated session-held/spend-limited key; fixed endpoints, rotate/revoke/redact/delete; bypass or incompatibility blocks the host entry. |
| Semantic authoring | JSON contracts + Markdown Go templates | One authored obligation source; generated outputs are read-only. |
| Template engine | Go `text/template` with fixed function allowlist | No shell, environment expansion, dynamic includes, or model-authored templates on release path. |
| Canonical records | Deterministically pretty-printed UTF-8 I-JSON | JSON is canonical truth; Markdown is attachment/view. |
| Structural schema | JSON Schema Draft 2020-12 | Embedded/vendored only; network `$ref` forbidden. |
| Semantic record digest | RFC 8785 JCS + domain-separated SHA-256 | Distinct from raw-byte CAS and exact package digests. |
| Logical IDs | RFC 9562 UUIDv4 using `crypto/rand` | Time-independent; paths are locators, not identity. |
| Record store | Repository files under a versioned Level 7 artifact root | Current state plus policy-bounded supersession/history; retention is not indefinite and this is not event sourcing. |
| Derived graph/search | In-memory maps; optional disposable JSON view | No database/vector store in v1. |
| Rooted I/O | Go `os.Root` with relative paths | No ambient filesystem APIs in trusted repository packages. |
| Portable path collision key | `NFC(full-case-fold(NFC(segment)))` with `golang.org/x/text` | Conservative cross-platform rejection; exact original path remains identity input. |
| Concurrency | `O_CREATE` plus `O_EXCL` lease + authoritative preimage CAS | Lease is advisory; ambiguous stale lease blocks. |
| Transaction | Same-filesystem staging, sync, rename, post-delta check, recovery journal | Multi-file atomicity is not claimed; failure injection required. |
| Capability | Opaque in-memory struct + 256-bit `crypto/rand` nonce | Exact action digest, one use, expiry; never argv/env/repository/model data. |
| Local evidence | Fresh admitted reproduction over inherited anonymous handles | Persisted local claim is reverified or remains `UNVERIFIED`. |
| Protected release plane | Separate GitHub release-control/eval repositories, digest-pinned clean builders, and fresh per-case AWS VMs | Builders are credential-free; provenance, evaluator launch/result, verdict, authorization, capability grant, artifact signing, and promotion use distinct least-privilege subjects. Absent separation means prerelease only. |
| Release evidence | DSSE protocol 1.0.2 + in-toto Statement v1 + SLSA 1.2 + Level 7 verdict/authorization/grant predicates | Trust policy maps every predicate to its exact permitted issuer; hidden corpus and all privileged roles stay outside candidate authority. |
| Release signing | Predicate-restricted Sigstore identities + Cosign 3.1.3 standardized bundles | Reject <=3.1.2 and legacy JSON bundle path; a valid wrong-role signature is invalid evidence. |
| SBOM | SPDX 2.3 JSON from Syft 1.51.0 | Scan final staged package; sign SBOM digest; no format conversion. |
| Companion channel | GitHub immutable Releases + Level 7 TUF metadata; signed exact `.deb`/bundles; separate manual `l7up` | External Cosign 3.1.3 bootstrap; root-owned install; core never self-updates; TUF conformance remains a release gate. |
| Host-package channels | Official Codex/Claude managers after lifecycle proof | Update is separate per host; rollback means authenticated reinstall of an exact prior artifact, not an assumed command. |
| Public tests | Go `testing`, golden/differential/property tests, fuzzing, race detector | Provider-neutral JSON fixtures and deterministic graders. |
| Static/security checks | `gofmt`, `go vet`, govulncheck 1.7.0, Staticcheck 0.7.0, gosec 2.28.0 | Exact binaries/checksums and suppressions with owners/expiry. |
| Secret defense | Context allowlist/minimization + bounded local detectors; Gitleaks 8.30.1 in CI/release | Defense in depth; no “zero secrets” claim; redact findings. |
| Logs | Go `log/slog`, structured and sink-redacted | Local, bounded, no raw source/prompt/secret; telemetry OFF. |
| CLI output | Accessible decision-first text plus versioned JSON mode | No color-only meaning; `NO_COLOR`; stable exit/status codes. |
| Feature flag | Root-owned `/etc/level7/policy.json` contract, `controlled_local_mutation` default OFF | Explicit owner/target/expiry/metrics/failure/removal; repo/env/model cannot enable; missing/invalid/expired is OFF. |

## 8. Canonical record and digest contract

### 8.1 Physical model

The eventual migration target is conceptually:

```text
docs/artifacts/l7/
  records/<kind>/<artifact-id>.json
  history/<kind>/<artifact-id>/<semantic-digest>.json
  attachments/<raw-sha256>.<safe-extension>
  derived/status.json
  derived/graph.json
```

This document does not create that layout or migrate the bootstrap Markdown. A later approved migration must preserve every byte permitted by the applicable handling policy, plus the source digest, owner decision, and audit binding. A discovered secret, authenticated deletion obligation, or policy prohibition takes precedence over byte preservation and requires bounded redaction plus incident/policy handoff; Level 7 must never archive prohibited bytes merely to preserve history.

Current records use stable paths for usability. An update archives the prior validated record under its semantic digest only when the applicable retention policy permits history, then replaces the current record through one recoverable transaction. Evidence-bound correction uses a new record with `supersedes`; it never rewrites retained evidence silently. History is policy-bounded, not immortal, and this is current-state storage rather than a mandatory append-only event ledger.

### 8.2 Retention, privacy, and deletion

Every record and attachment carries or inherits a versioned `handling` contract: sensitivity class, data categories, access owner, `retention_policy_id`, creation/review/expiry times, legal-hold state, deletion disposition, backup/Git/remote limitations, and the minimum tombstone fields. A missing or unknown handling policy blocks persistence of personal/sensitive payloads and blocks gate advancement where retention matters.

The v1 baseline is deliberately data-minimal:

| Data class | Initial policy |
|---|---|
| Model projection, prompt staging, temporary host state | Delete after the one-shot turn; crash residue is detected on next start and becomes deletion-due no later than 24 hours. |
| Structured local operational logs | No source/prompt/secret payload; rotate at 10 MiB and delete after 7 days unless an active incident/legal hold states otherwise. |
| Protected recovery material | Keep only while `RECOVERY_REQUIRED`; review on every start, then delete within 24 hours of a durable terminal resolution. No automatic expiry may destroy the only recovery copy. |
| Superseded non-sensitive records and unreferenced attachments | Review at supersession and delete or tombstone after 365 days by default; an active evidence reference, approved project policy, or legal hold can extend it. |
| Current governance/evidence record | Retain while it governs an open project, review at least annually, then make an explicit export/delete/retain decision at project closure. |
| Minimal tombstone | Follow the governing evidence-retention policy; contain only necessary identifiers, policy/reason, time, actor class, disposition, and erasure limitations—never the deleted payload. Retain a digest only when the policy permits it; `digest_retained: false` plus a reason is valid because a digest of a low-entropy secret or personal value can itself be a confirmation oracle. |

V1 never stores a known secret in canonical artifacts and does not claim encrypted-at-rest repository storage. Content needing guaranteed cryptographic erasure is therefore ineligible for persistence; it must be redacted, represented by minimal metadata/digest, or kept in an approved external system. Attachments are reference-counted by exact digest and become deletion-eligible only when no live record or hold references them. Derived views are destroyed and rebuilt after deletion. Logs, projections, recovery, host caches, Git objects, remotes, backups, and provider retention are separate sinks and may not inherit a deletion claim from the canonical store.

Expiry creates `DELETION_DUE`; a legal hold blocks deletion and never authorizes deletion, permits secret persistence, or widens access. The kernel first classifies the exact operation. V1 may execute only policy-permitted, repository-local, recoverable or compensable A1/A2 redaction/removal after the risk matrix is satisfied: scoped AP1 is the floor, and R2/R3 identity, qualified-review, accountable domain-approval, or separation requirements block when their required AP2/AP3 evidence is unavailable. Any irreversible/destructive purge, broad sensitive-data erasure, Git-history rewrite, force-push, remote/backup/provider deletion, cryptographic erasure, or guaranteed purge is A4/external handoff only and has no v1 execution interface.

For an eligible A1/A2 operation, a trusted user starts a recoverable transaction under a dedicated `tombstones/` namespace that (1) stages a payload-free `PENDING` tombstone, (2) removes or redacts eligible current/history payloads and unreferenced attachments, (3) rebuilds derived views, (4) verifies every in-scope location, and (5) finalizes `COMPLETE`, `PARTIAL`, or `NOT_ERASED` with truthful limitations. `COMPLETE` means complete only for the explicitly enumerated, verified local scope; it never implies irreversible erasure. Git objects, remotes, backups, SSD media, provider copies, and host caches remain reported limitations until independently evidenced; Level 7 never says they were erased merely because a working-tree file disappeared or a VM was destroyed.

### 8.3 JSON rules

Before schema validation, a strict reader rejects:

- invalid UTF-8, BOM, trailing data, duplicate object keys, lone surrogate values, NaN/infinity, and values outside the selected I-JSON model;
- ambiguous or nonportable path strings where a path field is expected; and
- unknown schema major versions for gate reduction.

Records use LF, a final newline, two-space indentation, deterministic JCS key order, RFC 3339 UTC timestamp strings, and strings for large integers/decimals/durations whose exact value must survive other language implementations. JSON Unicode text is not normalized. Path collision analysis separately uses a declared portable collision key while preserving the exact path bytes/name.

Schemas have stable absolute URN `$id` values, are embedded in the authenticated binary/package, are meta-validated, and resolve `$ref` only from an allowlisted in-memory catalog. The runtime never fetches a schema. Known core fields are closed for a schema major; forward additions use a namespaced `extensions` object whose values are retained by a lossless generic-object layer. A known-major record with an unknown non-extension core field is preserved and quarantined rather than partially reduced. An unknown namespaced extension round-trips unchanged; an extension marked critical blocks semantic reduction until understood. A generated writer merges known changes into the complete parsed object and refuses the write if it cannot prove that every unknown field/value survived. Unknown future-major records are preserved byte-for-byte and quarantined; they are never rewritten or allowed to advance a gate until an approved compatible migration exists. Codex→Claude→Codex and reverse round-trip fixtures cover all three cases.

Schema validation establishes shape, not authority, freshness, transition legality, cross-record consistency, or evidence sufficiency. The deterministic reducer owns those semantics.

### 8.4 Three deliberately different digests

| Digest | Input | Purpose |
|---|---|---|
| Semantic record digest | `concat("level7:record:<schema-major>\0", JCS(record))` | Cross-render/cross-host record revision identity. |
| Raw preimage digest | `concat("level7:file-preimage:v1\0", exact_file_bytes)` | Compare-and-swap; detects formatting and out-of-band byte changes. |
| Package/candidate digest | Exact normalized archive or signed payload bytes | Build/evaluation/release lineage; never JCS-normalized. |

Attachments use exact raw-byte SHA-256 plus media type and byte length. A record's own digest is not placed inside its hashed payload. All rendered digests use `sha256:<lowercase-hex>` and declare the domain/version. Conflating these digest classes is a release-blocking defect.

### 8.5 Dependency choices for records

The initial production dependency budget is three audited modules:

1. `github.com/santhosh-tekuri/jsonschema/v6` exact `v6.0.2`, Apache-2.0, for Draft 2020-12; configure an embedded-only loader and run the official test suite.
2. `github.com/gowebpki/jcs` exact `v1.0.1`, Apache-2.0, for RFC 8785; vendor and run the RFC/reference corpus plus independent cross-language vectors.
3. `golang.org/x/text` exact `v0.41.0`, BSD-3-Clause, only for Unicode NFC and full case-fold collision keys; it does not rewrite stored paths or JSON text.

All three are provisional until Step 5 records source/checksum/license review and the harness passes. A failure does not authorize a silent alternate encoding; it triggers a revised technology decision.

## 9. Source and workspace identity

The source-identity engine produces a versioned JCS manifest of sorted entries:

- exact normalized repository-relative path;
- portable collision key (`NFC(full-case-fold(NFC(path-segment)))`, with vectors frozen by the schema) used only to detect case/Unicode aliases;
- entry type;
- executable-bit/mode policy;
- byte length;
- raw content SHA-256, or exact symlink-target bytes where symlinks are admitted; and
- declared exclusions with reason/evidence state.

Absolute paths, `.`/`..`, NUL, device/reserved names, special devices, escaping links, ambiguous mount transitions, and case/Unicode collisions block the affected scope. V1 does not mutate through symlinks or create them. The real target envelope separately binds the opened root's platform identity (device/inode or volume/file ID as available); that local identifier is not put into the portable content digest.

For Git scopes, the identity adds the applicable commit/tag and a scoped dirty-state digest. Git is invoked only by exact absolute executable under a sanitized configuration/environment and containment profile that disables network, credentials, hooks, filters, optional locks, and external helpers; otherwise Git identity is `UNVERIFIED` and the non-Git manifest is authoritative. Git is never initialized automatically.

The intake may report bounded partial inventory, but an approval or effect digest may not silently omit an in-scope path because of size, ignore files, permission, unsupported type, or timeout.

## 10. Filesystem transactions and recovery

`os.Root` is selected for brokered repository operations because it rejects relative traversal and symlink escape through its rooted APIs. It is not a sandbox: mount points, ambient `os` calls, and spawned processes remain separate threats. `os.OpenRoot` appears only in a narrow reviewed bootstrap package; trusted repository packages receive a root capability plus relative paths.

Go releases before the fixed patch lines had a trailing-slash `os.Root` symlink escape. The GO-2026-4970 reproducer becomes a permanent regression case, and the compiler/runtime floor cannot fall below a fixed release.

Every A1/A2 transaction follows this state machine:

1. create a root-scoped cooperating-writer lease with exclusive creation and a random owner token;
2. open/pin the canonical root and revalidate action, AP, feature flag, target set, source identity, path policy, and every raw preimage CAS;
3. stage complete outputs in each destination filesystem/directory;
4. sync staged files, validate their bytes/schema/semantics, and record only recovery-safe metadata;
5. perform same-filesystem replacements in deterministic order;
6. sync applicable directories where the platform contract supports it;
7. compare the actual scope delta with the admitted delta;
8. freshly verify postconditions; then commit the lifecycle marker, or enter `RECOVERY_REQUIRED`;
9. erase protected recovery material and release the lease only after the terminal state is durable.

The lease is advisory; preimage CAS is authoritative. A stale or ambiguous lease never gets broken automatically. Network filesystems, mount behavior, disk-full conditions, anti-virus/watchers, cross-filesystem targets, and platforms whose create/rename/sync semantics fail the startup fault probe are unsupported for mutation.

Recovery journals in the repository contain paths, states, digests, operation order, and safe diagnostics—not unknown raw preimages. When rollback needs sensitive/untracked bytes, controlled mode uses a supervisor-owned `0700` recovery area outside the model/host sandbox or blocks before mutation. Loss of that area cannot create PASS; it leaves explicit manual recovery evidence.

## 11. Trusted confirmation and capability binding

### 11.1 AP1 source

The only selected AP1 candidate is input read by the user-started parent `l7` supervisor after it closes the relay, obtains stdout/stderr EOF, reaps the model-host namespace init, and proves the dedicated cgroup is unpopulated. The child starts in a new session with no controlling TTY/PTY, cannot open `/dev/tty`, receives pipe input only, inherits no unrelated descriptors, and never shares the confirmation interval. On timeout, malformed output, descendant escape signal, or unexpected activity, the parent kills the complete cgroup/namespace and discards the proposal. Child bytes are never streamed directly to the trusted terminal: bounded stdout/stderr is structurally parsed as untrusted data, and ANSI/OSC/control sequences, hyperlinks, terminal titles, clipboard commands, cursor movement, and look-alike prompts are escaped or separately rendered. Only after teardown proof does the parent display AP1 on a clean trusted surface. Running `l7` from a model shell is not equivalent: without an existing parent-owned supervisor session and trusted-terminal provenance, mutation entry points do not exist.

Before confirmation the supervisor renders accessible text containing:

- effect and risk class;
- exact repository identity and targets;
- human-readable intended delta and non-goals;
- source/pre-state, plan, policy, and environment digests;
- verification and recovery plan;
- known/unknown actor authority and assurance level;
- expiry/replay behavior; and
- one approve/reject decision.

The authenticated event binds that tuple plus a 256-bit random nonce and a monotonic single-use state. A typed digest suffix may be required for higher-risk local actions; exact interaction policy is set by tested UX/risk fixtures, never by the model. AP1 proves a current local user action, not organizational role or environment ownership.

Codex auto review, Claude auto/bypass/session grants, hook output, chat phrases, files that say “approved,” model-generated input, terminal escape output, signals, background descendants, and writes to parent pipes remain AP0.

### 11.2 Capability representation

The kernel mints an unexported, non-serializable in-memory capability bound to the action-envelope digest, target root handle, pre-state, effect, executor/writer identity, deadline, and nonce. The model child never receives it. Writer/executor calls occur inside the supervisor; if an isolated child is required, the parent uses an inherited anonymous pipe/handle with an ephemeral MAC key, not argv, environment, PATH, repository IPC, or a named world-discoverable socket.

Any scope change, crash, timeout, resume, user cancel, host restart, policy change, or first use expires the capability.

### 11.3 Release capability and default-off rollout flag

`controlled_local_mutation` is not a repository boolean or environment variable. Compile-time/default behavior is A0/OFF. When enabled, the effective effect ceiling is the minimum of the binary-declared maximum, achieved assurance, the signed grant's `effect_ceiling`, and the local policy ceiling; local policy may narrow a grant but never widen it. A1/A2 requires **both**:

1. a DSSE capability grant issued only by the predicate-restricted release capability-grant identity with predicate `urn:level7:attestation:release-capability:v1`, distributed as an exact TUF target, binding the exact release authorization and `l7` digest, adapter/host/model versions, OS/image/kernel/Bubblewrap profile, policy/schema, evaluation and verdict digests, permitted `A0|A1|A2` effect ceiling, cohort, issue/expiry, revocation sequence, guardrail thresholds, and removal version; and
2. a root-owned local policy at `/etc/level7/policy.json` that references that grant and records flag type, accountable local owner, pseudonymous target selector, default OFF, enable time, expiry, targeting/privacy rationale, guardrail metrics, failure behavior, and removal work.

The separately authenticated updater verifies exact Cosign issuer/workflow identity, installs the grant at a fixed root-owned path, binds its digest into TUF target metadata and the install receipt, and persists highest-seen TUF role versions, `revocation_sequence`, and trusted-time evidence under root-owned `/var/lib/level7/`. That anti-rollback state survives ordinary package rollback/removal; an explicit privileged purge is a separately warned destructive operation. The core cannot mint, widen, refresh, or rewrite it and trusts only the verified root-owned grant/receipt.

At session start and immediately before mutation, the core validates its own/grant/policy/receipt digests; every ancestor's canonical identity, owner, and mode; absence of symlink/hardlink substitution; schema; target; time; and monotonic version/serial state. Any decrease, clock rollback, reader incompatibility, missing state, or older binary unable to interpret current trust state yields A0. A trusted administrator opts in outside any model session. Repository files, user/project host settings, prompt text, model output, argv, ordinary environment variables, and same-user plugin data cannot enable the flag. No-admin or same-user-only installation remains A0, and root compromise is outside this boundary.

V1 chooses a hard revocation/freshness bound: verified TUF timestamp refresh must be no older than **7 days**, and a capability grant may be valid for at most **30 days**. Offline operation past either bound becomes A0; manual refresh therefore is not advertised as an immediate kill switch. A newer revocation target advances the persisted sequence, and an older signed grant can never re-enable it. Missing, malformed, stale, revoked, wrong-target, wrong-version, wrong-containment, failed-guardrail, unreadable, or ambiguous state evaluates to OFF before AP1. Disabling or expiry blocks new mutations but permits only digest-bound recovery of an already admitted transaction. Every action receipt binds the exact grant and local-policy digests, and `l7 status` explains the precise OFF reason without leaking sensitive target data.

The flag is a rollout/reversal mechanism, not AP1, not a substitute for rollback/migration, and not a permanent architecture layer. Stable promotion must name its owner and removal milestone; removal or default-ON requires a new approved architecture/release decision and evidence that the safety boundary still holds.

## 12. Producer receipt and evidence selection

### 12.1 Mechanism comparison

| Criterion | Weight | Repository/same-user key | Fresh live reproduction | Platform hardware key | Protected CI identity |
|---|---:|---:|---:|---:|---:|
| Forgery resistance against proposal plane | 30 | 1 | 4 | 4 | 5 |
| Exact approval/effect binding | 15 | 3 | 5 | 5 | 5 |
| Local portability/offline use | 15 | 5 | 4 | 2 | 3 |
| Durable later verification | 15 | 2 | 2 | 5 | 5 |
| Minimal trusted surface | 10 | 4 | 4 | 2 | 3 |
| Maintainability | 10 | 4 | 3 | 2 | 3 |
| Implementation speed | 5 | 4 | 3 | 1 | 2 |
| **Weighted total / 100** | **100** | **56.0** | **74.0** | **69.0** | **83.0** |
| Selected scope |  | **Rejected** | **Local A1/A2 baseline** | Research only | **Release baseline** |

### 12.2 Local baseline: fresh admitted reproduction

After an admitted write, the supervisor starts an exact-digest verifier through a protected inherited handle. The verifier independently reopens the root, recomputes pre/post/scope/delta identities, runs the approved deterministic and contained checks, and returns a live MAC-bound response. The kernel advances only during that authenticated live session.

The repository record preserves method, inputs, environment profile, output/byproduct digests, limits, result, and reproducibility limits. On a later process, that editable record is not independently authentic; the reducer reruns the required verifier before relying on it or labels it `USER_ASSERTED`/`UNVERIFIED`. Fresh reproduction can prove current postconditions, not retroactively prove who performed an earlier mutation.

A `0600` file key, embedded key, repository key, plugin-data key, SSH agent, generic same-user signer, or self-signed receipt is rejected: the proposal plane could read it or ask the signing interface to attest arbitrary bytes. Platform hardware keys remain future research because non-exportability does not prevent signer-oracle misuse.

### 12.3 Release baseline

Select a **GitHub Enterprise Cloud release-control plane** plus a candidate-inaccessible private evaluator repository for stable v1. The current product workspace/repository supplies only an untrusted source candidate and public tests. Protected workflow, hidden corpus, labels, scorer, thresholds, release policy, signing identity, and promotion credentials live in separately administered repositories whose branch rules, CODEOWNERS, required reviews, action allowlists, and environments cannot be changed by candidate authors. If the organization/plan cannot enforce those controls and prevent self-review, output is development/prerelease only.

The exact-byte path has distinct principals and fresh substrates. “Signed” is not a generic trust bit: policy verifies each predicate type against its one permitted issuer subject, and no transport/artifact signature can substitute for a verdict, authorization, or capability grant.

1. **Builder A and Builder B** independently fetch one immutable source digest and build offline from vendored inputs in a digest-pinned Ubuntu 24.04 OCI environment on separate clean runners. They have no secrets, `id-token`, publish permission, shared mutable cache, or candidate-controlled workflow. Their normalized unsigned payload bytes must match byte-for-byte, and the raw payload SHA-256—not merely a CI artifact/archive digest—is recorded and compared.
2. **Build comparator/provenance issuer** receives only both raw payload digests plus frozen build manifests, rejects inequality, and issues only the SLSA provenance for the matching quarantined payload under its own immutable OIDC subject. It cannot execute candidate tests, decide release, grant capability, sign the artifact blob, or promote.
3. **Evaluator launcher** receives only that quarantined payload digest and a narrowly scoped AWS launch role. Each hidden case runs in a fresh Ubuntu 24.04 x86_64 AWS VM created from a frozen AMI/build manifest, with encrypted ephemeral storage, IMDS disabled, no instance profile, no inbound path, and security-group egress only to an evaluator-owned provider/auth gateway and bounded result collector. One case/projection is mounted per VM; candidate execution has no GitHub/AWS identity. The VM is destroyed after sanitized bounded output and residue evidence are collected, but destruction is not represented as guaranteed media, backup, log, or provider erasure.
4. **Evaluator result issuer** receives only sanitized bounded results after candidate termination and issues only the DSSE/in-toto Test Result for the exact payload/run manifest under its own immutable OIDC subject. It cannot launch a candidate, define thresholds, issue a release verdict, authorize, grant capability, sign the artifact blob, or promote.
5. **Structurally independent verdict issuer** has read-only candidate authority, is not the candidate author/remediator, applies the frozen thresholds without executing the candidate, and issues only a Level 7 `GO|CONDITIONAL_GO|NO_GO` release-verdict predicate under a separate immutable OIDC subject. It binds the entire frozen release packet: exact Codex/Claude/package/companion digests; provenance; public and protected Test Results; pilot/adoption; compatibility; security/privacy; SBOM/license/notices; residual risks; rollback/revocation; open conditions; reviewer identity; and release-note limitations. An unmet safety invariant, seeded blocker, evaluator boundary, package-authenticity condition, pilot threshold, declared matrix entry, or missing packet field forces `NO_GO` regardless of aggregate score. `CONDITIONAL_GO` cannot authorize or promote; all conditions require evidence and a fresh verdict over the unchanged or newly digested candidate.
6. **Release authorizer/accountable owner** reviews the immutable build/eval/security/license/matrix evidence in a protected environment with self-review disabled and issues only a release-authorization predicate for an exact `GO` verdict, payload, channel, and effect ceiling under a separate authorization identity.
7. **Capability-grant issuer** receives only a valid exact release authorization and emits the narrowly bounded `urn:level7:attestation:release-capability:v1` grant under its own predicate-restricted OIDC subject. It cannot widen the authorized ceiling or sign/promote artifacts.
8. **Artifact signer** has only blob/bundle-signing authority. It signs the exact authorized payload, SBOM, evidence index, and channel metadata with a dedicated GitHub Actions OIDC subject, but cannot mint or replace provenance, Test Result, verdict, release authorization, or capability-grant predicates; acceptance verifies their native role signatures separately.
9. **Promoter** has only channel-write authority under another immutable OIDC subject; it cannot build, execute candidates, score, authorize, grant capability, or sign. It copies only the already signed digest and verified role-issued evidence into an immutable GitHub Release and host-specific staging channel.

The AWS launcher identity, provider-gateway identity, provenance/result/verdict/authorization/grant/artifact-signing/promotion GitHub OIDC subjects, human reviewer teams, runner image/AMI/kernel, qualification-approved provider model identifier and observed service identity, workflow commit, action commit SHAs, tool digests, and artifact IDs enter the attestation and compatibility record. Every privileged stage uses its own short-lived, audience-bound identity and role; candidate-execution and builder jobs receive none. Trust policy rejects a predicate signed by the wrong subject even if its cryptography is valid. All actions are pinned to full commit SHAs, workflow permissions are explicit-minimum, caches are disabled on privileged paths, and candidate text never enters a generated shell command. Stable release remains blocked until the actual repository IDs, workflow paths and SHAs, OIDC subjects, reviewer teams, AWS account/role IDs, AMI digest, action commit SHAs, and gateway behavior are frozen and tested.

Protected evaluation/release uses DSSE protocol 1.0.2 envelopes with in-toto Statement v1 and applicable standard predicates. Build provenance uses SLSA 1.2/provenance v1, but the claimed level is capped by evidence actually achieved. Sigstore keyless signing uses exact workflow identity and issuer verification, standardized bundles, transparency verification where applicable, revocation/expiry policy, and no insecure bypass flags.

Cosign versions through 3.1.2 are rejected because the upstream high-severity advisory identifies a legacy-bundle identity-verification bypass; the initial exact baseline is 3.1.3. Detached attestations bind the exact evaluated payload digest and do not mutate it. No single CI job, workflow token, repository administrator, or human action may silently collapse build comparison/provenance, evaluator launch/result, verdict, authorization, grant issuance, artifact signing, and promotion into one identity.

## 13. Process, sandbox, and network enforcement

No language runtime supplies cross-platform child containment. The selected executor never calls a shell and never accepts a model-selected executable. An execution envelope contains an absolute allowlisted executable digest/path, typed argv, working directory, cleared/rebuilt environment, declared reads/writes, network policy, credentials policy, timeout, output cap, process-tree policy, and resource limits.

Ubuntu uses two different Bubblewrap profiles; conflating them is a blocking defect:

| Profile | Filesystem/process boundary | Network/credential boundary |
|---|---|---|
| **Model host** | Exact Codex/Claude binary and runtime libraries read-only; owned empty control root; bounded prompt/schema/config; disposable tmpfs home/state; real repository, parent state, user home, sockets, D-Bus, keychain, devices, and recovery area absent; new user/mount/PID/IPC/UTS/network namespaces; no `/proc` unless the exact host proves it indispensable, otherwise a minimized noncredential view; minimal `/dev` without `/dev/tty`; closed descriptors; dedicated namespace and cgroup. Before AP1 the relay is closed, stdout/stderr reaches EOF, namespace init is reaped, and the cgroup must be unpopulated. | Direct network is absent. A trusted in-namespace loopback HTTP bridge forwards over one inherited anonymous socket to a protocol-aware parent relay. The parent permits only the frozen provider API surface, strips child auth, injects the real key outside the namespace, enforces byte/time/request/cost budgets, and records redacted metadata. The child receives only a one-turn non-secret placeholder/relay token. |
| **Verifier/executor** | Exact allowlisted executable in a fresh namespace; declared read-only snapshot and isolated declared writable tree only; no user home, real ambient filesystem, TTY, D-Bus, sockets, or unrelated descriptors; rlimits/cgroup/process-tree teardown. | Network namespace has no relay and no egress; environment contains no provider, cloud, Git, SSH, package-manager, proxy, or signing credential. |

The provider relay is part of the trusted Go companion, not a general proxy. Before forwarding, it accepts only identity-encoded, bounded JSON for one exact qualified API operation and strict allowlisted fields. It rejects compression, redirects, `CONNECT`, WebSocket/QUIC alternatives, unexpected query parameters, alternate upstreams, unknown authority-affecting fields, non-approved requested model identifiers, nonempty tool definitions, prior-session/response identifiers, unexpected request counts, and any system/developer/user content whose canonical digest differs from the admitted request. It strips all child authorization, injects the real credential only on the provider-side request, and enforces request, input/output token, byte, time, retry, and spend ceilings. After the response arrives but before proposal admission, it verifies any response-reported model/service identity against the frozen qualification contract; a required identity that is absent or has drifted fails the tuple. Upstream TLS uses one fixed provider hostname with hostname/SNI verification and a qualification-recorded CA bundle; redirects are disabled, and every DNS resolution/connection is freshly rejected if it targets loopback, private, link-local, metadata, or any non-allowlisted address. No child-controlled proxy environment is inherited.

This boundary is deliberately **not** described as provider end-to-end TLS: protected evaluation uses two authenticated transport legs—candidate-to-evaluator gateway and gateway-to-provider—with authentication injection at the gateway; it performs no transparent TLS interception. The local in-namespace bridge is bound to one inherited socket and one-turn token, while the provider leg uses TLS. Local v1 obtains a dedicated provider key from the trusted terminal for the session and does not persist it; OS keyring, subscription OAuth, or enterprise workload-identity support requires a later credential-store review. Best-effort memory clearing is not described as forensic erasure. If either host cannot use the owned base URL, ignores the relay, opens an undeclared endpoint, requires request compression/retries, changes the admitted context, cannot emit an empty tool inventory, or receives/exposes a real provider credential, that controlled matrix entry fails.

### 13.1 Platform strategy

| Platform | Selected experiment | Current disposition |
|---|---|---|
| Ubuntu 24.04 LTS x86_64 | Ubuntu-authenticated Bubblewrap exact `0.9.0-1build1`, the two profiles above, and an exact image/kernel/prerequisite manifest | Only controlled A1/A2 candidate; every image/kernel/bwrap/user-namespace/seccomp/cgroup/relay fact must pass. |
| macOS 13+ arm64 | No independent supported Level 7 outer containment profile selected | A0/C−1 host-package target only; controlled A1/A2 is unsupported in v1. |
| WSL2 Ubuntu 24.04 x86_64 | Linux strategy plus WSL2-specific filesystem/mount/process tests | Prerelease research after native Ubuntu passes. |
| Native Windows | No selected v1 sandbox/process-tree boundary | A0 research only; A1/A2 unsupported for v1. |
| Other Linux/macOS/architectures | None | Unsupported until added by a separate matrix revision. |

Bubblewrap is a critical external prerequisite: the exact Ubuntu archive package, source version, package digest, dependencies, license, SBOM reachability, security status, install/update/removal state, and invocation profile enter the inventory. Setuid Bubblewrap is not accepted; unprivileged user namespaces, namespace nesting, seccomp/cgroup availability, kernel configuration, supervised relay behavior, and reliable teardown must pass at startup and in C−1. A rolling `ubuntu-latest`, unrecorded host kernel, or version range is never compatibility evidence. On macOS, undocumented/private sandbox profiles are not selected. Host sandboxes are defense in depth only after exact-version escape, credential, network, file, descendant, timeout, and unexpected-write tests.

Pure deterministic validation inside the trusted Go core may still verify an effect when arbitrary commands are blocked. A project build/test result cannot be synthesized from that; it remains `UNVERIFIED` where behavior evidence is required.

## 14. Context engineering and secret handling

Controlled mode never gives the child host the real repository. The supervisor:

1. performs root-bounded metadata inventory;
2. applies explicit path/type/size/sensitivity allowlists and ignore rules as hints, not authority;
3. runs bounded local secret detectors and policy checks;
4. creates a disposable, provenance-labeled prompt projection containing the smallest required snippets and safe placeholders; project filenames such as `AGENTS.md`/`CLAUDE.md` are data, never auto-discovered instructions;
5. launches the child in the model-host profile with a minimal rebuilt environment, no inherited general/project/cloud credentials or unrelated descriptors, no uncontrolled settings/memory/connectors/browser/MCP/subagents, and no real-root path;
6. admits only the declared parent-owned auth-injecting relay, binds its routes/key scope/cost budget into the run manifest, and blocks if the host can bypass it or its endpoint behavior is unknown;
7. closes the relay, obtains output EOF, reaps namespace init, proves the dedicated cgroup empty, structurally parses/sanitizes/rescans the never-directly-streamed output, and only then permits trusted-terminal review; and
8. deletes projection, prompt staging, disposable host state, auth material, relay state, and quarantined raw output under the Section 8 lifecycle.

The initial natural-language prompt and host-native attachments may already have crossed the provider boundary before a stock plugin activates. Level 7 discloses that boundary and does not claim retroactive protection. Sensitive workflows needing pre-provider filtering require controlled mode; if a safe projection cannot be produced, the path blocks.

Gitleaks 8.30.1 is selected as an offline CI/release and optional projection defense, with a vendored reviewed ruleset, full redaction, archive/decode/size/time limits, and safe findings containing rule/location metadata only. It is not an oracle, and its upstream feature-complete/security-patch-only status creates an explicit successor-review trigger. No scanner supports a “zero secrets” promise.

Provider token and cost measurement uses exact host-reported usage when available. Otherwise Level 7 records byte/character budgets and labels token/cost estimates as estimates; it does not add a tokenizer dependency or fabricate precision.

Provider credentials use separate per-host keys/accounts where the provider supports them, least billing/project scope, hard spend/rate limits, short rotation, documented revocation, and no fallback credential chain. The real key remains in the trusted parent/relay and never enters prompts, schemas, repository records, argv, parent human output, debug logs, crash dumps, or child-accessible files/environment. The relay boundary and the provider's own retention/training/region terms are disclosed before sensitive context crosses it; Level 7 does not claim that local deletion erases provider copies.

## 15. Prompt, workflow, skill, graph, loop, and memory technology

This selection makes the user's prompt/workflow/skill goals first-class implementation inputs rather than free-form Markdown conventions.

### 15.1 One semantic source

Each workflow/profile is authored as:

```text
semantic/workflows/<workflow-id>/contract.json
semantic/workflows/<workflow-id>/prompt.md.tmpl
semantic/workflows/<workflow-id>/fixtures/*.json
```

The contract declares stable obligation IDs, version, positive/negative triggers, lifecycle transitions, prerequisites, authoritative inputs, context projection, effect/risk floor, approval/gates, invariants, prohibited effects, evidence, output schema, retry/budget/stopping/escalation, success/failure, host capabilities, and fixture IDs. The Markdown template supplies concise human/model guidance. Safety policy lives in the contract/kernel, not duplicated hidden prose.

A Go compiler validates the contract and template, proves every required obligation is rendered, generates separate host-valid packages, and fails on dropped or invented safety obligations. Generated files carry source digest/version and are never reverse-edited into semantic truth.

### 15.2 Conductor and profiles

Only one public conductor is discoverable. Greenfield, existing/brownfield, legacy/refactor, live-change, release, UX, data/schema, security, performance, and later profiles are internal typed modules selected by deterministic eligibility plus evaluated model classification. Missing specialist proof returns `NOT_EVALUATED`/`BLOCKED`, not a generic pass.

The current 12 prototype skills are research inputs. Step 4 changes none of them. Their later disposition must place useful knowledge behind the conductor or exclude/deprecate the public bypass, with positive, negative, overlap, direct-call, alias, natural-language, injection, and context-budget tests.

### 15.3 Prompt contract

Every compiled prompt contains only the applicable projection of:

- goal and authoritative inputs with provenance labels;
- user/problem/outcome and acceptance criteria;
- invariants, prohibited effects, risk/effect and authority/tool bounds;
- evidence and uncertainty requirements;
- retry, cost/time/turn budgets, stopping and escalation;
- exact typed output schema; and
- one next transition or blocker.

Prompts never request or persist hidden chain-of-thought. They request decisions, sources, evidence, uncertainty, counterexamples, and reproducibility limits. Model output remains an untrusted proposal until schema and semantic admission.

### 15.4 Loop and multi-agent engineering

The deterministic reducer, not the model, owns loop state. Every loop has a transition ID, retry/turn/time/cost cap, no-progress detector, cancellation, and terminal `PASS`, `BLOCKED`, `NOT_EVALUATED`, `UNVERIFIED`, or `RECOVERY_REQUIRED` semantics. Retrying cannot lower a gate.

Subagents remain optional. A delegation manifest constrains objective, disjoint scope, minimum projection, tools, budget, output schema, verifier, and termination. Subagent output returns as untrusted proposal data; no subagent receives AP/capability/credentials or writes the real repository.

### 15.5 File/codebase/graph memory

Validated repository records are durable memory; host conversations and summaries are caches. The reducer rebuilds phase, open decisions, provenance, supersession, and evidence relationships on every invocation. An in-memory directed adjacency map provides graph queries; optional status/graph JSON is derived and disposable. Deleting every cache/index must yield the same safety decision after rebuild.

## 16. Test and evaluation technology

### 16.1 Public deterministic harness

Use Go's standard harness as the authoritative local base:

- table/unit tests for reducers, policies, serializers, paths, capabilities, transactions, adapters, and CLI contracts;
- golden tests for semantic compilation and both generated packages;
- differential Codex/Claude fixtures for all safety-critical axes;
- native fuzzing for JSON, schema boundaries, paths, JCS, migrations, capabilities, and state transitions;
- race tests and adversarial concurrency/fault injection;
- property/metamorphic tests implemented with deterministic generators before adding a property framework;
- official JSON Schema and JCS vectors;
- actual-host black-box drivers using exact binaries and isolated workspaces; and
- machine-readable run manifests with candidate, environment, seed, trial, time/cost, result, and limitations.

`go test`, `go test -race`, fuzzing, `go vet`, `govulncheck`, Staticcheck, and gosec cover complementary failure classes. No scanner result alone establishes security.

### 16.2 Agent/prompt evaluation

Public cases are provider-neutral JSON with frozen truth labels, selection rules, run counts, seeds, cost/latency capture, adjudication, and deterministic graders wherever possible. A model judge is calibrated evidence, never the sole safety/release oracle. Host runners translate the same case into exact Codex and Claude sessions and compare safety-critical outputs before prose quality.

The protected runner, corpus, labels, thresholds, scorer, credentials, result issuer, verdict issuer, authorizer, grant issuer, artifact signer, and promoter stay outside candidate read/list/write scope. It runs the exact package in a fresh sandbox per case; only after candidate termination does the distinct result issuer sign the in-toto Test Result, and only the separate non-executing verdict issuer can produce the release-verdict predicate. Inspect AI and Promptfoo are not selected as trusted v1 dependencies; either may be reconsidered as an externally isolated exploratory orchestrator, but their configuration/controller code cannot become the protected boundary without a separate review.

## 17. Build, package, supply chain, and update technology

### 17.1 Reproducible payload

The Go module uses exact `go.mod`/`go.sum`, checked-in `vendor/`, vendored schemas/vectors/licenses, `GOTOOLCHAIN=local`, offline vendor builds, `CGO_ENABLED=0`, `-trimpath`, deterministic version/link inputs, and no unreviewed initialization side effects. Two clean builders must produce identical normalized unsigned binaries.

Host packages are generated by allowlist from the frozen semantic source. Archives use sorted paths, normalized modes, normalized timestamps, fixed compression settings, and no ambient files. Platform signing/notarization and detached attestations occur after normalized payload creation; the exact signed/repackaged bytes, not a rebuild, become the candidate evaluated and promoted.

### 17.2 Package split

| Deliverable | Contents | Trust/lifecycle rule |
|---|---|---|
| Codex A0 host package | Manifest, one conductor skill, schemas/public assets needed for A0 | Self-contained advisory product; no hook/MCP/executable/script in v1; official host lifecycle only after proof. |
| Claude A0 host package | Manifest, one namespaced conductor skill, schemas/public assets needed for A0 | Self-contained advisory product; no hook/MCP/bin/script in v1; official host lifecycle only after proof. |
| `l7` controlled companion | Ubuntu x86_64 Go supervisor/core/writer/executor/verifier in an exact `.deb` | Separate root-owned authenticated application; user-started; exact host/model/platform/grant compatibility; no daemon. |
| `l7up` channel verifier | Narrow root-owned Go updater/grant verifier with embedded TUF bootstrap root | Separate executable/privilege domain; no repository/model/prompt access, arbitrary URL, shell, plugin operation, or automatic update. |
| Release capability grant | Grant-issuer-signed exact controlled-mode compatibility/eval/rollout envelope | Required with root-owned local opt-in; missing/stale/revoked/mismatch keeps mutation OFF. |
| Release evidence bundle | Exact hashes, SPDX 2.3 SBOM, provenance, Test Results, verdict, authorization, capability grant, permissions/matrix, license/notices, role-specific Sigstore bundles, rollback/revocation data | Bound to exact host package/companion candidate; wrong-role or missing evidence fails; separately promoted per host. |

The companion is not found through PATH trust alone. The supported Ubuntu package installs versioned binaries root-owned under `/opt/level7/<version>/` and an administrator-owned launcher/receipt outside same-user write authority. The receipt binds canonical path, file/package digests, signing identity, TUF target/version, platform, install transaction, grants, and compatible host/model entries. A path-shadowed, user-writable, stale, cached, substituted, revoked, or mixed version blocks before execution. A0 packages may use documented package-relative helpers in a later revision, but they cannot discover or authenticate this parent boundary.

### 17.3 SBOM and signing

Syft 1.51.0 generates SPDX 2.3 JSON directly from the final staged filesystem/archive; source-only SBOMs or lossy conversions are insufficient. Exact Cosign 3.1.3 creates/verifies standardized bundles for the artifact blob and each role-issued DSSE attestation; trust maps predicate type to exact issuer/identity, so the artifact signer cannot substitute for another issuer. Every replacement version needs separate qualification. Build actions/images/tools use immutable commit/digest references, not mutable tags.

SLSA level claims are limited to evidence actually supplied by the protected build-comparator/provenance issuer over both matching raw payload digests and frozen manifests. A candidate-controlled PR workflow cannot self-claim trusted-builder separation.

### 17.4 Channel and lifecycle

Development uses local/repository/private marketplaces. The OpenAI public directory is not selected for a mutation-capable release while it would expose untested ChatGPT/desktop/web surfaces or omit required local behavior. A public OpenAI A0 package is possible only with precise advisory/A0 labeling and independent conformance. Controlled Client distribution uses immutable GitHub Releases plus a Level 7-owned TUF repository; it is never installed by either host package.

Initial Ubuntu bootstrap is deliberately external to the code being authenticated:

1. the user obtains the exact `.deb`, SBOM, release index, and Sigstore bundles from one immutable release;
2. before `dpkg` or any package code runs, the user verifies them with exact Cosign 3.1.3 that was independently bootstrapped through Sigstore's TUF root/process, using frozen GitHub OIDC issuer/workflow identity and full bundle/claim/transparency checks;
3. a trusted administrator installs the verified package root-owned; the package contains no networked maintainer script and writes an OS/package-manager ownership receipt; and
4. `l7up` initializes only from the embedded threshold-signed TUF root and refuses arbitrary repositories, insecure mirrors, expired metadata, unknown roles, or locally supplied target digests.

The core has no self-update interface. `l7up` is manually invoked from a trusted administrator terminal, refreshes timestamp/snapshot/targets/root metadata, enforces threshold, expiry, consistent-snapshot, rollback/freeze, delegation, version, length, and hash checks, and atomically installs only a listed exact `.deb`/grant. It has no repository or model input. Emergency rollback is published as **newer signed TUF metadata** selecting a previously evaluated, nonrevoked exact package; clients never lower trusted metadata versions or accept an unsigned cached artifact. Root rotation, target revocation, freeze recovery, offline-expiry behavior, partial install compensation, and updater self-update use separate roles/tests. If the bootstrap verifier, TUF root custody, threshold keys, or privileged updater cannot be independently authenticated and lifecycle-tested, controlled distribution is `NO_GO` while A0 development may continue.

The initial TUF policy sets `consistent_snapshot=true`; root uses 2-of-3 offline hardware-held keys with 365-day expiry; the top-level `targets` role and delegated `stable-targets` role each use distinct 2-of-3 offline hardware-held keys with 90-day expiry; snapshot and timestamp use separate restricted 1-of-1 online keys with 30-day and 7-day expiries. No top-level target may bypass the stable delegation or its release-evidence policy: stable package/grant paths are terminating-delegation-only, and policy rejects a stable target authorized solely by the broader top-level role. `go-tuf/v2` 2.4.2 is selected only as a TUF **1.0.31 implementation/wire profile** and is audited against the current 1.0.36 specification. No 1.0.36 conformance claim is allowed until upstream conformance plus Level 7 root-rotation, expiry, freeze, rollback, mix-and-match, partial-update/install, revocation, delegated-target, top-level-bypass, recovery, and unknown-custom-field cases pass. TUF does not solve initial bootstrap, so the first trusted root and Cosign verifier remain independently authenticated prerequisites.

Lifecycle is host/product specific:

| Product | Install/update/rollback | Disable/remove/residue rule |
|---|---|---|
| Codex A0 package | Official plugin manager; exact installed path/digest/version observed. Rollback means remove then reinstall an authenticated exact prior package only after actual-host proof; no undocumented rollback command is assumed. | Official remove; inventory marketplace/cache/config residue. No project artifact deletion and no assumption that plugin data survives. |
| Claude A0 package | Official marketplace manager; exact scope/version/digest observed. Rollback means disable/remove and install an authenticated exact prior version only after proof. | Use explicit prepare/remove journey and `--keep-data` only when owner chose retention; disclose last-scope data deletion and orphan cache behavior rather than claiming erasure. |
| Controlled Client + `l7up` | External Cosign bootstrap, root-owned install, manual TUF update, higher-metadata rollback; host plugins are irrelevant. | `l7 prepare-remove` inventories/reconciles active transactions, then administrator package removal; root-owned policy/grants/data purge is a separate explicit choice. Canonical project artifacts remain unless separately approved. |
| Capability grant/local flag | Threshold-signed grant via TUF plus root-owned local opt-in; never automatic. | Revocation/expiry/local disable blocks new mutation. Remove grant/policy only after recovery and receipt preservation; flag-removal milestone is release-governed. |

Registration, installation, enablement, update, disablement, preparation for removal, removal, cache/data residue, rollback, revocation, and reinstall are distinct tested states. No undocumented uninstall callback or plugin-data retention is assumed. A lifecycle failure cannot be repaired by the component authenticating itself.

### 17.5 Legal gate

No repository `LICENSE` exists, so the prototype's MIT metadata is not an accepted legal source. Before any distribution, the owner must choose an open-source or proprietary license with appropriate review, add exact license text, and generate matching third-party notices. Apache-2.0 is a reasonable open-source candidate because of its explicit patent grant, but this technology artifact does not make the business/legal choice.

## 18. Proposed physical source/output layout

This is a Step 5 harness target, not a directory-creation authorization:

```text
cmd/l7/                         # only public binary entry point
cmd/l7up/                       # privileged fixed-channel verifier/updater; no project/model API
internal/
  supervisor/                  # parent host/process/session boundary
  kernel/                      # deterministic admission and reducer
  context/                     # inventory, minimization, projection, redaction
  artifact/                    # strict JSON, schema, JCS, migrations
  policy/                      # risk/effect/AP/feature/capability decisions
  transaction/                 # rooted writer, lease, CAS, journal, recovery
  executor/                    # exact admitted local effects
  receipt/                     # live reproduction and attestation verification
  platform/                    # explicit darwin/linux/unsupported adapters
  adapter/codex/               # one-shot codex exec argv/event translation
  adapter/claude/              # one-shot Claude print/JSON translation
  channel/                     # TUF verification/install logic used only by l7up
  render/                      # semantic-to-host generation
  evaluator/                   # public runner/graders only
semantic/
  taxonomy/                    # provider-neutral terms and enums
  workflows/                   # contract JSON + prompt template + fixtures
  profiles/                    # generic, feature, refactor first
schemas/                       # vendored schemas and migration catalog
fixtures/public/               # deterministic, routing, adversarial, host cases
packages/source/codex/         # authored host-only manifest inputs
packages/source/claude/        # authored host-only manifest inputs
build/generated/               # disposable generated packages; never authority
docs/artifacts/                # user-owned governance/canonical memory
```

Trusted packages use a ports-and-adapters rule: pure decision code has no filesystem, process, clock, randomness, terminal, network, or host import; those enter through narrow interfaces and explicit facts. Dependency direction is checked in the harness. `internal/executor`, `transaction`, and `receipt` cannot import render/model/host prompt packages. `cmd/l7` cannot import the privileged installer; `cmd/l7up` cannot import repository, prompt, host-adapter, executor, or transaction packages.

## 19. Compatibility strategy and initial matrix

### 19.1 Version policy

- A support entry names exact Codex/Claude executable version **and digest**, the qualification-approved provider model identifier and observed service identity, `l7`/`l7up`, host package, semantic contract, artifact schema, OS image/build, kernel/configuration, architecture, Bubblewrap/package/prerequisite, relay policy, capability grant, and external tool versions.
- C−1 starts with observed Codex `0.149.1`, Claude `2.1.241`, and macOS 26.5.2 arm64 only as development evidence.
- Initial release qualification tests exactly one host+model+platform tuple at a time; “latest,” model aliases, automatic fallback, semver ranges, mutable runner labels without recorded image identity, and documentation-only compatibility are forbidden.
- Host packages share a product version/changelog but have separate digests, matrices, evidence, rollback, and promotion.
- A host update is unsupported until the smoke, safety, differential, context, lifecycle, and degraded-mode suite reruns.
- A provider model whose full immutable build cannot be exposed is recorded with the fullest provider ID, service fingerprint/capability response, region, account tier, observation interval, and run date; any drift expires prior behavioral evidence and requires reevaluation. Level 7 does not claim immutable model weights unless the provider exposes a verifiable immutable revision.
- Product, semantic contract, artifact schema, attestation predicate, public corpus, protected corpus, host adapter, model, companion/updater pairing, and package versions remain separately identifiable.

### 19.2 Candidate matrix—not a support claim

| Host path | Platform | Qualification model/service baseline | A0 | A1 | A2 | Release disposition now |
|---|---|---|---:|---:|---:|---|
| Stock Codex package | Observed macOS arm64 | Host-controlled; record observed ID | Plausible | No | No | C−1 development only. |
| Stock Claude package | Observed macOS arm64 | Host-controlled; record observed ID | Plausible | No | No | C−1 development only. |
| Controlled Codex | Exact Ubuntu 24.04 x86_64 image/kernel/Bubblewrap tuple | `UNSET`—must freeze full ID under `L7-EVAL-009` | Plausible | `UNPROVED` | `UNPROVED` | Only stable candidate if every gate passes. |
| Controlled Claude | Exact Ubuntu 24.04 x86_64 image/kernel/Bubblewrap tuple | `UNSET`—must freeze full ID under `L7-EVAL-009` | Plausible | `UNPROVED` | `UNPROVED` | Only stable candidate if every gate passes. |
| Controlled Codex/Claude | macOS arm64 | None | A0 host package only | No | No | Controlled client unsupported in v1; no outer profile. |
| Controlled Codex/Claude | WSL2 Ubuntu 24.04 x86_64 | None | Research | No | No | Prerelease containment research after native Linux. |
| Any | Native Windows, Intel macOS, Linux arm64, other distro | None | Unknown | No | No | Out of v1 matrix unless separately selected and proven. |

Stable dual-host v1.0 does not require every OS; it requires both initial hosts to pass on every exact entry the product declares. The narrowest credible first stable controlled candidate is both hosts on one frozen Ubuntu 24.04 x86_64 tuple. The model IDs, Ubuntu image/AMI, kernel, Bubblewrap package/digest, namespace/seccomp/cgroup facts, companion/updater/package pairing, provider routes, and release grant remain deliberately `UNSET`/`UNPROVED` release blockers until Step 5/C−1 records them. macOS controlled A1/A2 would require a new selected outer boundary and revised matrix, not a documentation edit.

## 20. Dependency and upgrade policy

### 20.1 Product binary

The unprivileged `l7` core starts with at most three non-standard production modules (`jsonschema/v6`, `jcs`, `x/text`) plus a separately justified OS adapter. The privileged updater is a separate Go module/binary and adds exactly `github.com/theupdateframework/go-tuf/v2` `v2.4.2` (Apache-2.0, Go floor 1.25) plus its audited transitives; updater dependencies cannot enter the core graph. Upstream `v2.4.2` emits a TUF `1.0.31` wire profile even though the current specification tag is `1.0.36`, so compatibility with spec `1.0.36` is explicitly `UNPROVED` and stable channel use is blocked until the official conformance suite plus Level 7 root-rotation, delegation, expiry, freeze, rollback, revocation, and recovery cases pass. Every dependency requires:

- exact module version and checksum;
- source repository/release identity;
- license and notice;
- transitive inventory;
- vulnerability and maintenance/EOL review;
- reason standard library is insufficient;
- reachable API/use inventory;
- conformance/adversarial tests; and
- owner, review date, upgrade/replace trigger.

Release builds use the vendored graph with network disabled and compare it to `go.sum`/module verification. Unused/transitive dead code remains visible in the SBOM; tree shaking is not treated as a supply-chain substitute.

### 20.2 Build/evaluation tools

Initial exact research baselines are:

| Tool | Version | Role | Runtime dependency? |
|---|---:|---|---:|
| Go | 1.26.7 | compiler/test toolchain | No for users |
| govulncheck | 1.7.0 | reachable Go vulnerability analysis | No |
| Staticcheck | 0.7.0 / 2026.1 | static bug/performance/style analysis | No |
| gosec | 2.28.0 | security-oriented Go static analysis | No |
| Gitleaks | 8.30.1 | secret defense in depth | No |
| Syft | 1.51.0 | final-package SPDX 2.3 SBOM | No |
| Cosign | 3.1.3 | blob/attestation signing and verification | No |
| Bubblewrap (Ubuntu package) | 0.9.0-1build1 | model-host and verifier/executor outer profiles | **Yes, exact platform prerequisite** |
| go-tuf/v2 | 2.4.2; TUF 1.0.31 wire profile | privileged `l7up` metadata verification | **Yes, updater only** |
| TUF specification | 1.0.36 | current audit/conformance target | No; implementation compatibility `UNPROVED` |

Step 5 must resolve and record official checksums/signatures/provenance before use. It may replace a baseline only through a documented, evidence-backed harness decision; it may not install `latest`.

## 21. Degraded modes and fallback ladder

| Missing/failed capability | Safe outcome | Permitted next action |
|---|---|---|
| Stock plugin only | A0 advisory/status; no write | Start an authenticated controlled session or remain advisory. |
| Companion/updater/grant identity or TUF freshness cannot be verified | A0 diagnosis only | Use externally authenticated bootstrap/update/rollback metadata; component self-claims do not repair trust. |
| Child host cannot be isolated from real repo/context | `AR-002 BLOCKED`; no sensitive path or mutation | Use safe projection/offline planning or revise host integration. |
| Provider relay/base URL/auth injection fails or is bypassed | No controlled provider call; no real key enters child | Rotate/revoke affected key, preserve redacted incident evidence, and remove host tuple until requalified. |
| Parent AP1 channel cannot exclude child/model input | A0 only | Use a proven external confirmation client; never accept chat text. |
| A1 writer closure fails | A0 only | Fix actual-host path or narrow/reapprove requirements. |
| A2 closure or sandbox fails | A1 ceiling if independently closed | Produce plan/diff handoff; do not claim execution. |
| Fresh verifier unavailable | Effect may remain factual but evidence is `UNVERIFIED`; gate does not advance | Reproduce in an admitted verifier. |
| JSON/JCS/schema disagreement | Quarantine/block affected record | Use compatible reader/migration; no newest-file fallback. |
| Lease/CAS/atomicity probe fails | No mutation; `RECOVERY_REQUIRED` if partial | Resolve exact state under separate approval. |
| Git unsafe/unavailable | Non-Git scoped manifest with disclosed limits | Continue only if applicable requirement still met. |
| Database/index absent/corrupt | Direct canonical-file reconstruction | Rebuild optional derived view. |
| Network absent | Normal local operation | Only provider/explicit online cases report blocked network. |
| Protected evaluator/signer unavailable | Development/beta, `NOT_EVALUATED`/`NO_GO` | Wait for independent plane; self-review cannot substitute. |
| Release roles collapse or candidate gains protected corpus/identity | `NO_GO`; no signature/promotion | Restore separation, rotate credentials, rebuild and reevaluate fresh exact bytes. |
| Flag/grant missing, expired, revoked, mismatched, or guardrail failed | `controlled_local_mutation = OFF` | Diagnose or install a newly authenticated grant/policy outside the model session. |
| Retention/deletion state unknown | Quarantine sensitive persistence; no erasure claim | Resolve policy/hold/sinks, then perform separately approved lifecycle action. |
| License decision absent | No distribution | Owner/legal decision and complete notices. |

There is no prompt-only fallback for an enforcement failure.

## 22. Step 5 harness contract and future C−1 acceptance spikes

If this artifact is approved, Step 5 may construct only the minimum harness needed to falsify the selection early. It must not implement the product journey or edit current skills/prompts/manifests.

### 22.1 Harness-first order

1. Freeze toolchain/dependency checksums, public schema/fixture formats, evidence labels, and deterministic test/run manifest.
2. Establish pure package boundaries and compile-time import bans before behavior code.
3. Add JSON duplicate-key/I-JSON/schema/JCS conformance vectors and three-digest confusion tests.
4. Add `os.Root` traversal/symlink/trailing-slash-CVE/collision/mount/TOCTOU tests.
5. Add lease/CAS and crash injection at every transaction state.
6. Add parent/child protocol fixtures that prove child cannot read confirmation input, real-root files, capability data, provider credentials, or protected recovery state and cannot survive to AP1.
7. Add model-host and verifier/executor Bubblewrap profile fixtures, relay/auth injection tests, TTY/fd/process/output spoof cases, and exact prerequisite manifests.
8. Add generated Codex/Claude obligation parity, unknown-field round-trip, retention/tombstone, and byte-determinism tests using inert fixtures.
9. Add reproducible clean-build comparison and dependency/license/SBOM/channel/grant checks.
10. Record—but do not execute in Foundation Step 5—the future C−1 A0 smoke and controlled Ubuntu walking-skeleton experiments. Actual-host/provider experiments require their later approved build/evaluation wave after the deterministic harness passes.

### 22.2 Release-blocking spikes

| Spike | Required evidence | Kill result |
|---|---|---|
| `SP-01 Codex A0` | Clean local install/discovery/invocation, one conductor, no writes, exact residual lifecycle truth. | Codex path unsupported. |
| `SP-02 Claude A0` | Same as SP-01 plus namespace/cache/reload/data behavior. | Claude path unsupported. |
| `SP-03 controlled child` | Relay observes before forwarding: exact admitted requested-model identifier/instructions/messages, one request, empty tools, and no prior-session identity; after provider response but before proposal admission, any required response-reported model/service identity matches the frozen qualification contract; host has no real-root/TTY/real-provider-key/capability access or unexpected context/settings/MCP/browser/subagent/hook; only owned projection+relay; namespace residue inventory clean. | `AR-002` fails; affected controlled host is removed and overall controlled mode has A0 ceiling if dual-host remains required. |
| `SP-04 AP1` | Relay closed, output EOF, namespace init reaped, and dedicated cgroup unpopulated before prompt; `/dev/tty`, fd, signal, background-child, ANSI/OSC/clipboard/title/cursor/look-alike attempts cannot approve, replay, widen, or spoof; tuple changes stale it. | `AR-001/002` fail; A0 ceiling. |
| `SP-05 A1 closure` | Direct Edit/Write/patch/shell/Git/alias/legacy/natural-language paths cannot produce Level-7-issued artifact writes outside writer. | A0 only. |
| `SP-06 A2 closure` | Same for source/config/Git plus unexpected writes/network/credentials and mode changes. | A1 ceiling; stable C7 blocked. |
| `SP-07 rooted paths` | All declared OS cases cover absolute/traversal/dangling/internal/external symlink, GO-2026-4970, mount, case/Unicode, concurrent replacement. | OS/effect unsupported. |
| `SP-08 transactions` | Interrupt every state; disk full, permission, stale lease, concurrent writer, watcher interference; never false PASS/lost update. | A1/A2 blocked. |
| `SP-09 process containment` | Both exact Bubblewrap profiles defeat path/argv/shell/env/fd/TTY/grandchild/network/relay/output/resource/timeout escapes on the frozen image/kernel/package tuple. | Controlled host or command evidence blocked. |
| `SP-10 fresh receipt` | Repository/model/same-user imitation cannot advance; stale/wrong candidate/policy/nonce/replay fails; clean restart re-reproduces. | `AR-011` fails; C2/C7 blocked. |
| `SP-11 deterministic build` | Two protected clean builders reproduce unsigned payload; quarantine IDs, normalized archives, exact signed lineage, SBOM, and no-rebuild promotion verified. | Distribution blocked. |
| `SP-12 host lifecycle` | Per host: register/install/enable/update/disable/authenticated-prior reinstall/prepare-remove/remove/data+cache residue preserve unowned project files/artifacts. | Affected A0 host package blocked. |
| `SP-13 companion channel` | Independent Cosign bootstrap, root-owned install receipt, TUF root/threshold/expiry/freeze/rollback/revocation/delegation, partial update, higher-metadata rollback, updater self-update, and full removal pass. | Controlled distribution blocked. |
| `SP-14 data lifecycle` | Every sink obeys sensitivity/retention/hold/expiry/redaction/reference/deletion/tombstone rules; seeded secrets never archive; Git/backup/provider limits remain truthful. | Sensitive persistence and release blocked. |
| `SP-15 protected plane` | Candidate cannot read/list/write hidden corpus/workflows, gain GitHub/AWS/provider/sign/promote identities, reuse a VM/cache, alter scoring, or replace bytes between stages; self-review is impossible. | Evaluation/release blocked. |
| `SP-16 rollout flag` | Compile default, signed grant, local policy, targeting, expiry, revocation, guardrail failure, disable, recovery, and removal all fail closed; repo/env/model cannot enable. | Controlled mutation remains OFF. |

## 23. Risk register after selection

| ID | Risk | Severity/state | Owner/gate |
|---|---|---|---|
| `TR-001` | No stock plugin closes Level-7-issued A1/A2. | Critical, documented/inferred | Selected A0 ceiling; SP-05/06 controlled-session proof. |
| `TR-002` | One-shot Codex/Claude loads ambient context/tools/hooks or cannot emit the required pre-forward zero-tool request. | Critical, unproved | Semantic relay rejects context/tool drift before forwarding; affected host is removed rather than allowing residual tools; SP-03. |
| `TR-003` | Parent terminal confirmation can be influenced, spoofed, or replayed by child/agent. | Critical, unproved | Child namespace ends before sanitized trusted UI; SP-04; no standalone mutation API. |
| `TR-004` | Fresh-reproduction channel is forgeable or cannot reconstruct state after restart. | Critical, unproved | SP-10; no gate-bearing local evidence otherwise. |
| `TR-005` | Bubblewrap/profile/kernel/relay cannot contain the model host or verifier/executor. | Critical, unproved | Exact Ubuntu-only two-profile SP-09; tuple removed on failure. |
| `TR-006` | `os.Root`/path policy fails platform edge cases or mount semantics. | High, unproved | SP-07; fixed Go floor and OS-specific adapters. |
| `TR-007` | Multi-file crash recovery duplicates sensitive material or loses work. | High, unproved | SP-08; protected recovery/block. |
| `TR-008` | JCS or JSON Schema implementation diverges, is unmaintained, or introduces network behavior. | High, unproved dependency audit | Official/reference vectors, embedded-only loader, replacement ADR. |
| `TR-009` | Direct canonical-file reconstruction misses scale budgets. | Medium, empirical | Benchmark; consider disposable SQLite only after evidence. |
| `TR-010` | Companion/privileged updater/bootstrap/TUF expands lifecycle and supply-chain burden. | High, inherent | External Cosign bootstrap, root ownership, narrow `l7up`, SP-11/13. |
| `TR-011` | Public OpenAI channel exposes untested surfaces. | High, unproved | A0-only/restrictable channel or block publication. |
| `TR-012` | Product license remains undefined. | High for distribution, observed | Owner/legal decision before package promotion. |
| `TR-013` | Ubuntu-only first stable A2 matrix limits adoption. | High, product tradeoff | Pilot evidence; add macOS only after containment proof. |
| `TR-014` | Secret detector false positives/negatives cause friction or leakage. | High, inherent | Projection minimization, canaries, disclosure, safe block. |
| `TR-015` | Host bypasses auth relay or child learns a real provider key. | Critical, unproved | Protocol-aware injection, no direct network, per-turn key canaries; SP-03/09. |
| `TR-016` | Release-control roles collapse or untrusted candidate escapes hidden evaluator. | Critical, unproved | Separate GitHub/AWS identities, per-case credential-free VM, no self-review; SP-15. |
| `TR-017` | History/log/recovery/cache retention leaks payload or deletion destroys audit/recovery. | High, unproved | Explicit handling policy, no secret archive, payload tombstone, sink limits; SP-14. |
| `TR-018` | Root-owned flag/grant becomes stale, bypassable, or permanent debt. | High, unproved | Signed exact grant + local opt-in + expiry/revocation/removal; SP-16. |
| `TR-019` | GitHub Enterprise/AWS release-plane cost or availability is unacceptable. | High, product/operational choice | Development remains possible; stable release waits for owner-funded plane or revised approved technology. |

## 24. Architecture and backlog traceability

| Architecture/backlog concern | Technology owner/seam |
|---|---|
| `AR-001`, `AF-03`, `AF-04`, `AF-10`, `L7-BL-005`, `L7-BL-008`, `L7-BL-009`, `L7-BL-012`, `L7-BL-013` | Parent supervisor, child teardown before trusted AP1, in-memory capability, signed rollout grant, A1/A2 closure spikes. |
| `AR-002`, `AF-09`, `L7-BL-006`, `L7-BL-040` | Honest stock-host disclosure; one-shot actual-host walking skeleton; bounded prompt projection; outer model-host profile; auth-injecting relay. |
| `AR-003`, `AF-08`, `AF-14`, `L7-BL-004`, `L7-BL-009` | `os.Root`, path policy, two Ubuntu containment profiles, lease/CAS/journal/fault suite, explicit retention/recovery. |
| `AR-004`, `AR-005`, `AR-006`, `AF-11`, `AF-16`, `L7-BL-001`, `L7-BL-002`, `L7-BL-006`, `L7-BL-007`, `L7-BL-008`, `L7-BL-041` | Frozen scope/ownership, semantic compiler, bounded projection, decision-first CLI, prompt/skill/differential, unknown-field, accessibility and UX fixtures. |
| `AR-007`, `AF-12`, `L7-BL-003`, `L7-BL-010`, `L7-BL-015`, `L7-BL-042` | Public deterministic runner plus candidate-inaccessible GitHub Enterprise/AWS evaluator; distinct launcher, result-publisher, verdict, authorization, signer, and promoter identities; DSSE/in-toto/Sigstore. |
| `AR-008`, `AR-009`, `AF-05`, `AF-07`, `L7-BL-004` | JSON 2020-12, lossless unknown fields, JCS/raw digests, explicit migrations, Git/non-Git manifests, retention-aware reducer. |
| `AR-010`, `AR-012`, `AR-013`, `AF-06`, `AF-13`, `L7-BL-014` | Pure-Go companion/updater, generated A0 host packages, external Cosign bootstrap, TUF, SPDX/SLSA/Sigstore, per-product lifecycle. |
| `AR-011`, `AF-05`, `L7-BL-005`, `L7-BL-009`, `L7-BL-010` | Live fresh reproduction locally; protected externally rooted exact-byte release attestations and capability grant. |
| `AI-12`, `AF-15`, `L7-BL-011` | No A3–A5 interface in command/capability/registry/package; data-only handoff for deploy/expose/observe. |

Every P0 backlog ID (`L7-BL-001` through `L7-BL-015`, plus `L7-BL-040`, `L7-BL-041`, and `L7-BL-042`) appears literally above so trace extraction cannot depend on shorthand expansion.

## 25. Technology decision records

These become accepted only with owner approval of this artifact.

| ADR | Proposed decision | Consequence |
|---|---|---|
| `TDR-001` | Go 1.26.7 pure-Go core; Go 1.27.0 shadow. | Small self-contained trusted runtime; Rust is explicit fallback. |
| `TDR-002` | A0 host packages and Level 7 Controlled Client are independently installed products; A1/A2 requires the latter as parent. | Honest two-product UX/lifecycle; plugin-only installation never implies controlled mutation. |
| `TDR-003` | Each controlled turn is one-shot `codex exec`/Claude print inside the exact Ubuntu model-host profile; child dies before AP1. | Context/read/mutation/terminal separation becomes structural; App Server and long-lived sessions leave stable v1. |
| `TDR-004` | JSON 2020-12 + JCS semantic digest + raw CAS/package digests. | Human-readable validation and cross-host binding without digest conflation. |
| `TDR-005` | Canonical files plus in-memory graph; no v1 database. | Repository truth stays simple/offline; SQLite needs benchmark evidence. |
| `TDR-006` | `os.Root`, O_EXCL lease, CAS, staged recoverable transaction. | No false multi-file atomicity claim; fault/platform gates remain. |
| `TDR-007` | Parent AP1 + opaque in-memory one-use capability. | Chat/host auto-approval/environment/repository fields cannot grant mutation. |
| `TDR-008` | Fresh admitted reproduction for local gates; externally separated GitHub Enterprise/AWS DSSE/in-toto/Sigstore plane for release. | No forgeable same-user durable local receipt or candidate-owned release oracle. |
| `TDR-009` | Ubuntu 24.04 x86_64 is the first controlled A2 candidate; native Windows excluded. | Narrow credible matrix before breadth; macOS A2 remains conditional. |
| `TDR-010` | JSON semantic contracts + Go-rendered Markdown templates + one conductor. | Prompts/workflows/skills become versioned, generated, and evaluable. |
| `TDR-011` | Go public harness; protected evaluator uses one-case ephemeral credential-free AWS VMs and an external two-leg provider-auth gateway. | Candidate cannot control the oracle, hidden corpus, provider credential, or reusable evaluator state. |
| `TDR-012` | Reproducible host-specific packages, SPDX 2.3, SLSA 1.2, exact Cosign 3.1.3, root-owned `.deb`, and separate go-tuf 2.4.2/TUF-1.0.31-profile `l7up` audited against spec 1.0.36. | Exact-byte independently bootstrapped lineage and no self-updating core; channel stays blocked until conformance proves update protections. |
| `TDR-013` | `controlled_local_mutation` requires a release-signed exact capability grant plus root-owned local policy and defaults OFF. | Repo/model/env cannot enable; expiry/revocation/guardrail/removal lifecycle is explicit. |
| `TDR-014` | Product license remains an explicit owner/legal gate. | Prototype MIT metadata cannot leak into release truth. |
| `TDR-015` | Every artifact/sink uses explicit sensitivity, retention, hold, expiry, deletion, tombstone, and limitation policy. | History is not indefinite; no secret archive or unsupported Git/backup/provider erasure claim. |
| `TDR-016` | Parent-owned protocol-aware relay strips child auth and injects the real provider key outside the child namespace. | Controlled hosts require API/base-URL compatibility; bypass or real-key exposure removes the matrix entry. |

## 26. Decision quality gate

Before owner approval, this artifact must receive separate read-only reviews for:

1. architecture/requirements consistency and hidden scope changes;
2. host/API/OS realism, especially the stock-A0/controlled-session split; and
3. security/data/supply-chain correctness, especially AP1, paths, receipts, digest classes, and exact-byte lineage.

The reviews are model analysis, not qualified human security/legal review, actual-host evidence, or release independence. Any blocker/high finding must be corrected or explicitly returned to the owner; it cannot be waived silently.

## 27. Primary sources

### Hosts and integration

- [OpenAI — Build a plugin](https://developers.openai.com/plugins/build/plugins)
- [OpenAI — Build skills](https://developers.openai.com/plugins/build/skills)
- [OpenAI — Plugin submission](https://developers.openai.com/plugins/deploy/submission)
- [OpenAI — Convert a Claude Code plugin](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [OpenAI — Permissions](https://developers.openai.com/codex/permissions)
- [OpenAI — Agent approvals and security](https://developers.openai.com/codex/agent-approvals-security)
- [OpenAI — Sandboxing](https://developers.openai.com/codex/sandboxing)
- [OpenAI — Hooks](https://developers.openai.com/codex/hooks)
- [OpenAI — App Server](https://developers.openai.com/codex/app-server)
- [OpenAI — Codex CLI reference](https://developers.openai.com/codex/cli/reference)
- [OpenAI — Codex non-interactive mode](https://developers.openai.com/codex/noninteractive)
- [OpenAI — Codex configuration reference](https://developers.openai.com/codex/config-reference)
- [OpenAI — Codex features](https://developers.openai.com/codex/features)
- [Anthropic — Plugins](https://code.claude.com/docs/en/plugins)
- [Anthropic — Plugins reference](https://code.claude.com/docs/en/plugins-reference)
- [Anthropic — CLI reference](https://code.claude.com/docs/en/cli-reference)
- [Anthropic — Headless/programmatic mode](https://code.claude.com/docs/en/headless)
- [Anthropic — Environment variables](https://code.claude.com/docs/en/env-vars)
- [Anthropic — Permissions](https://code.claude.com/docs/en/permissions)
- [Anthropic — Permission modes](https://code.claude.com/docs/en/permission-modes)
- [Anthropic — Sandboxing](https://code.claude.com/docs/en/sandboxing)
- [Anthropic — Hooks](https://code.claude.com/docs/en/hooks)
- [Anthropic — Agent SDK permissions](https://code.claude.com/docs/en/agent-sdk/permissions)
- [Anthropic — Agent SDK custom tools](https://code.claude.com/docs/en/agent-sdk/custom-tools)

### Runtime, paths, and testing

- [Go — Release history](https://go.dev/doc/devel/release)
- [Go — `os` package and `Root`](https://pkg.go.dev/os)
- [Go — Traversal-resistant file APIs](https://go.dev/blog/osroot)
- [Go vulnerability GO-2026-4970](https://pkg.go.dev/vuln/GO-2026-4970)
- [Go — Modules reference](https://go.dev/ref/mod)
- [Go — Compatibility](https://go.dev/doc/go1compat)
- [Go — Fuzzing](https://go.dev/doc/security/fuzz/)
- [Go — Race detector](https://go.dev/doc/articles/race_detector)
- [Go — Security best practices](https://go.dev/doc/security/best-practices)
- [Rust releases](https://doc.rust-lang.org/releases.html)
- [Bytecode Alliance — `cap-std`](https://github.com/bytecodealliance/cap-std)
- [Node.js release schedule](https://nodejs.org/en/about/previous-releases)
- [Node.js 24 permissions](https://nodejs.org/download/release/latest-v24.x/docs/api/permissions.html)
- [Ubuntu Noble — Bubblewrap package](https://packages.ubuntu.com/noble/bubblewrap)

### Records, evidence, and supply chain

- [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12)
- [RFC 8259 — JSON](https://www.rfc-editor.org/rfc/rfc8259.html)
- [RFC 7493 — I-JSON](https://www.rfc-editor.org/rfc/rfc7493.html)
- [RFC 8785 — JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785.html)
- [RFC 9562 — UUIDs](https://www.rfc-editor.org/rfc/rfc9562.html)
- [`jsonschema/v6` upstream](https://github.com/santhosh-tekuri/jsonschema)
- [`gowebpki/jcs` upstream](https://github.com/gowebpki/jcs)
- [`golang.org/x/text` upstream](https://go.googlesource.com/text)
- [DSSE protocol](https://github.com/secure-systems-lab/dsse/blob/master/protocol.md)
- [in-toto Statement v1](https://github.com/in-toto/attestation/blob/main/spec/v1/statement.md)
- [SLSA 1.2 provenance](https://slsa.dev/spec/v1.2/provenance)
- [Sigstore threat model](https://docs.sigstore.dev/about/threat-model/)
- [Cosign GHSA-fx35-mq7g-6g98](https://github.com/sigstore/cosign/security/advisories/GHSA-fx35-mq7g-6g98)
- [SPDX 2.3 specification](https://spdx.github.io/spdx-spec/v2.3/)
- [Syft output formats](https://github.com/anchore/syft/wiki/output-formats)
- [TUF specification v1.0.36](https://github.com/theupdateframework/specification/releases/tag/v1.0.36)
- [TUF non-goals and bootstrap boundary](https://github.com/theupdateframework/specification/blob/v1.0.36/tuf-spec.md#non-goals)
- [`go-tuf/v2` v2.4.2 release](https://github.com/theupdateframework/go-tuf/releases/tag/v2.4.2)
- [`go-tuf/v2` v2.4.2 wire-profile constant](https://github.com/theupdateframework/go-tuf/blob/v2.4.2/metadata/types.go#L32)
- [Sigstore — Cosign installation and bootstrap verification](https://docs.sigstore.dev/cosign/system_config/installation/)
- [Gitleaks upstream](https://github.com/gitleaks/gitleaks)

### Protected release substrate

- [GitHub — Deployments and environments](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments)
- [GitHub — OpenID Connect claims](https://docs.github.com/en/actions/reference/security/oidc)
- [GitHub — Immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)
- [AWS — Disable EC2 instance metadata at launch](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-IMDS-new-instances.html)
- [AWS — EBS encryption by default](https://docs.aws.amazon.com/ebs/latest/userguide/encryption-by-default.html)

Sources were checked on 2026-08-24. Current-version facts must be rechecked and exact artifacts authenticated when the Step 5 harness is authorized. Documentation supplies hypotheses and constraints; only actual-host/platform evidence can promote a matrix entry.

## 28. Owner approval gate

The owner may now:

- **Approve Step 4:** accept `TDR-001`–`016` and these material product/cost constraints: (1) the Codex and Claude plugins are advisory A0 products, while A1/A2 requires a separately installed root-owned Controlled Client; (2) controlled v1 targets only an exact Ubuntu 24.04 x86_64 tuple, with macOS/Windows controlled mode unsupported; (3) local controlled use initially requires an administrator installation plus a dedicated spend-limited provider API key entered per session—subscription OAuth/keychain reuse is unsupported; (4) stable distribution requires GitHub Enterprise Cloud governance, separately administered protected repositories/identities, and per-case AWS evaluation, with their operational cost; (5) mutation remains default OFF behind both an authenticated release grant and root-owned local policy; (6) channel trust requires three role-separated 2-of-3 offline hardware-key sets—for root, top-level targets, and stable targets—plus recurring signing ceremonies, a verified TUF refresh at least every 7 days, and grant renewal within 30 days; missing a freshness/renewal bound or offline access beyond it degrades mutation to A0; and (7) TUF 1.0.31-profile conformance, exact host/model/base-URL behavior, AP1, sandbox closure, and release-plane identity details remain future build/release blockers, not support claims. Approval authorizes only Foundation Step 5: repository/harness layout, pinned dependencies, lint/type/test/CI/logging/environment/README scaffolding, and one inert proving test. Step 5 may encode future C−1 contracts but may not run actual-host/provider experiments or build product behavior.
- **Request revision:** identify the technology, operating-mode, platform, dependency, data, packaging, or product-experience concern to change; Step 5 remains unauthorized.

Approval does not assert that `AR-001`, `AR-002`, `AR-003`, or `AR-011` passes. Beyond the narrow Step 5 harness authorization above, it does not authorize product-feature implementation, prompt/skill edits, host/plugin manifest or package changes, actual-host/provider experiments, installation outside the repository harness, publication, deployment, exposure, or release.
