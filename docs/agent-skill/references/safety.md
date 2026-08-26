# Reference: operating safely

`ovhcloud` acts on **real infrastructure and a billed account**. Treat it with
the same care as production `kubectl`/`aws`.

## Destructive operations — confirm first
`delete`, `terminate`, `confirmTermination`, `revert`/`restore`, and anything
that removes or overwrites data are **irreversible**. Before running one:
1. Show the user exactly which resource (id + name) will be affected.
2. Get explicit confirmation.
3. Prefer a `get`/`list` to verify the target id beforehand.

## Billed operations — flag the cost
Creating resources (instances, load balancers, managed databases/registries,
Public Cloud networks in some cases, etc.) **starts billing**. Before a
`create`, tell the user it is a paid resource and confirm. Don't create
resources speculatively "to test".

## General guardrails
- **Never guess** a command or flag — verify with `--help` or `-o json`.
- **Read before write**: inspect with `list`/`get` before `create`/`edit`/`delete`.
- Scope actions to the intended **project/region/profile** (`--cloud-project`,
  `--region`, `--profile`); double-check you're not on the wrong account.
- Use `--debug` to understand a failing call instead of retrying blindly.
- Credentials are secrets: never print, log, or commit `ovh.conf` or API keys.
