# AGENTS.md — driving the `ovhcloud` CLI

This file is the canonical guide for **AI coding agents** (Codex, Cursor, Gemini
CLI, Aider, Zed, Claude Code, Copilot…) that need to use the `ovhcloud` CLI to
manage OVHcloud services. Other agent files in this repo (`CLAUDE.md`,
`GEMINI.md`, `.claude/skills/…`, `.github/copilot-instructions.md`) point back
here so every tool shares the same instructions.

Detailed, load-on-demand material lives in **[`docs/agent-skill/`](docs/agent-skill/)**.

## The one pattern to remember

```
ovhcloud <universe> [sub-resource ...] <verb> [positional args] [flags]
```
- Verbs are consistent: `list` (alias `ls`), `get <id>`, `create`, `edit <id>`,
  `delete <id>`. Async actions accept `--wait`.
- Confirm the exact subcommand/flags with `--help` at any level; read results
  with `-o json`.

```bash
ovhcloud login                  # authenticate
ovhcloud --help                 # list universes
ovhcloud cloud instance --help  # verbs + flags of a resource
ovhcloud cloud instance list -o json
```

## Golden rules (read before acting)

1. **Never invent a flag or subcommand** — verify with `ovhcloud <path> --help`,
   or inspect fields with `-o json` on a `list`/`get`.
2. **Inspect before you mutate**: run `list`/`get` first, then `create`/`edit`/`delete`.
3. **Destructive or billed operations**: confirm with the user before `delete`,
   `terminate`, or creating paid resources. See
   [`docs/agent-skill/references/safety.md`](docs/agent-skill/references/safety.md).
4. Public Cloud commands need a project — set `--cloud-project <id>` or
   `OVH_CLOUD_PROJECT_SERVICE` (find it with `ovhcloud cloud project list`).

## Global flags (work on every command)

- `-o, --output json|yaml|interactive` — machine-readable output (use `json` to script).
- `--output '<expr>'` — extract/transform fields (gval), e.g. `--output 'id'`.
- `--filter '<expr>'` — filter list results, e.g. `--filter 'status=="running"'`.
- `--wait` — for async create/actions, block until ready.
- `-d, --debug` — log the full HTTP request/response.
- `--profile <name>` — use a specific profile (multi-account).

## Learn more (load on demand)

- **[references/commands.md](docs/agent-skill/references/commands.md)** — install,
  auth, the universe map, verbs, discovery.
- **[references/flags.md](docs/agent-skill/references/flags.md)** — global flags in detail.
- **[references/safety.md](docs/agent-skill/references/safety.md)** — destructive/billed ops.
- **[recipes/public-cloud.md](docs/agent-skill/recipes/public-cloud.md)** — Public Cloud workflows.
- **[recipes/account-and-domains.md](docs/agent-skill/recipes/account-and-domains.md)** — auth, profiles, domains/DNS.

> This guide stays small on purpose. It teaches the pattern and points to the
> CLI's own `--help` (the always-up-to-date source of truth) instead of copying
> the ~900 generated command pages.
