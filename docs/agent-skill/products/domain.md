# Product: Domains & DNS

Two account-level universes:
- `ovhcloud domain-name …` — registered domain names.
- `ovhcloud domain-zone …` — DNS zones and records.

Key verbs: domain-name `list`/`get`/`edit`; domain-zone `list`/`get`/`refresh`
and `record` (`list`/`get`/`create`/`update`/`delete`).

```bash
ovhcloud domain-name list -o json
ovhcloud domain-name get <domain>

ovhcloud domain-zone list -o json
ovhcloud domain-zone record list <zone> -o json
ovhcloud domain-zone record list <zone> --filter 'fieldType=="A"' -o json
ovhcloud domain-zone record create <zone> --help     # required flags
ovhcloud domain-zone record delete <zone> <recordId> # confirm first
ovhcloud domain-zone refresh <zone>                  # apply pending changes
```

> `record delete` and `refresh` change **live DNS**. Inspect with `list`/`get`,
> verify flags with `--help`, confirm before delete/refresh
> (see [../references/safety.md](../references/safety.md)).
