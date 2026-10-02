# Product: VPS

`ovhcloud vps …` — Virtual Private Servers. No project needed (account-level).

Key sub-resources & verbs: `list`, `get <name>`, `edit <name>`, `reinstall`,
`reboot`, `start`/`stop`, `set-password`, `get-console-url`, `snapshot`,
`automated-backup`, `disk`, `image`, `ip`, `list-available-upgrades`,
`list-tasks`, `terminate`.

```bash
ovhcloud vps list -o json
ovhcloud vps get <serviceName>
ovhcloud vps reinstall <serviceName> --help      # image, --ssh-key…
ovhcloud vps reboot <serviceName>
ovhcloud vps snapshot --help
ovhcloud vps terminate <serviceName>             # confirm first
```

> `reinstall` is **destructive** (wipes the VPS); `terminate` ends billing/service.
> Verify flags with `--help`, follow ops with `list-tasks`
> (see [../references/safety.md](../references/safety.md)).
