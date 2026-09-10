# Product: Bare Metal (dedicated servers)

`ovhcloud baremetal …` — dedicated servers. No project needed (account-level).

Key sub-resources & verbs: `list`, `get <name>`, `edit <name>`, `reinstall`,
`reboot`, `reboot-rescue`, `boot`, `ipmi`, `vni`, `list-os`,
`list-compatible-os`, `list-ips`, `list-tasks`, `list-interventions`,
`list-secrets`.

```bash
ovhcloud baremetal list -o json
ovhcloud baremetal get <serverName>
ovhcloud baremetal list-compatible-os <serverName> -o name    # installable OSes
ovhcloud baremetal reinstall <serverName> --help              # --ssh-key, template…
ovhcloud baremetal reboot-rescue <serverName>
ovhcloud baremetal list-tasks <serverName> -o json            # follow operations
```

> `reinstall` is **destructive** (wipes the server). Inspect with `list`/`get`,
> verify flags with `--help`, follow long ops with `list-tasks`
> (see [../references/safety.md](../references/safety.md)).
