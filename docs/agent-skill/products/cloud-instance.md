# Product: Public Cloud instances

`ovhcloud cloud instance …` — virtual machines in a Public Cloud project.
Requires a project (`--cloud-project <id>` or `OVH_CLOUD_PROJECT_SERVICE`).

Key sub-resources & verbs: `list`, `get <id>`, `create`, `edit <id>`,
`delete <id>`, `start`/`stop`/`reboot`, `reinstall`, `set-flavor`, `set-name`,
`shelve`/`unshelve`, `backup`, `autobackup`, `group`, `interface`, `image`,
`flavor`.

```bash
ovhcloud cloud instance list -o json
ovhcloud cloud instance get <id>
ovhcloud cloud instance flavor            # available flavors in the project
ovhcloud cloud instance image             # available images
ovhcloud cloud instance create --help     # region, flavor, image, ssh-key…
ovhcloud cloud instance reinstall <id> --help
ovhcloud cloud instance delete <id>       # confirm first
```

> Creating an instance is **billed**; `delete`/`reinstall` are destructive.
> Inspect with `list`/`get`, verify flags with `--help`
> (see [../references/safety.md](../references/safety.md)).
