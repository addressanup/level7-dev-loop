# Level 7 Dev Loop

Level 7 Dev Loop `0.1.1` is a dual-host instruction plugin for Codex and Claude
Code. It turns one concrete development objective into a solo-first loop:

`inspect → implement → test → repair → self-review → handoff`

## Use

In a new Codex task:

```text
$l7-next Implement and verify <your objective>.
```

In Claude Code, start a new session or run `/reload-plugins`, then use:

```text
/level7-dev-loop:l7-next Implement and verify <your objective>.
```

The plugin contains 12 Markdown instruction skills. It adds no executable,
hook, MCP server, telemetry, host setting, credential flow, or network access of
its own. The surrounding host retains its independently configured tools,
permissions, workspace boundaries, and provider behavior.

Documentation, source, releases, and issue tracking are available at
<https://github.com/addressanup/level7-dev-loop>.
