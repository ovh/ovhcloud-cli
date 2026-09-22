---
name: ovhcloud-cli
description: >-
  Use the `ovhcloud` CLI to manage OVHcloud services (Public Cloud, domains, web
  hosting, bare metal, network/vRack, email, telephony, account/API keys…). Load
  this when a task involves running `ovhcloud`, scripting against OVHcloud, or
  deciding which command/flag to use.
---

# OVHcloud CLI skill

The full, tool-agnostic guide lives in **[`AGENTS.md`](../../../AGENTS.md)** at the
repository root, with load-on-demand material under
**[`docs/agent-skill/`](../../../docs/agent-skill/)**. This file is the Claude
Code entry point; it deliberately mirrors that single source of truth.

## Essentials

Invocation pattern:
```
ovhcloud <universe> [sub-resource ...] <verb> [positional args] [flags]
```
Consistent verbs: `list`/`ls`, `get <id>`, `create`, `edit <id>`, `delete <id>`
(async actions take `--wait`). Confirm exact flags with `--help`; read output
with `-o json`.

Golden rules: never invent a flag (verify with `--help`); inspect with
`list`/`get` before mutating; confirm before `delete`/`terminate` or creating
**billed** resources (see
[`docs/agent-skill/references/safety.md`](../../../docs/agent-skill/references/safety.md));
Public Cloud needs a project (`--cloud-project` / `OVH_CLOUD_PROJECT_SERVICE`).

For install/auth, the universe map, global flags, and recipes, read
[`AGENTS.md`](../../../AGENTS.md) and the files under
[`docs/agent-skill/`](../../../docs/agent-skill/).
