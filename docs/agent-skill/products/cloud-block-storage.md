# Product: Block Storage (volumes)

`ovhcloud cloud storage block …` — block storage volumes in a Public Cloud
project. Requires a project (`--cloud-project <id>` or
`OVH_CLOUD_PROJECT_SERVICE`).

Key sub-resources & verbs: `list`, `get <id>`, `create <region>`, `edit <id>`
(name/size/type — retype), `delete <id>`, `attach`/`detach`, `snapshot`,
`backup`.

```bash
ovhcloud cloud storage block list -o json
ovhcloud cloud storage block get <id>
ovhcloud cloud storage block create GRA11 --name data --size 50 --type HIGH_SPEED_GEN2 --wait
ovhcloud cloud storage block attach <volumeId> <instanceId>
ovhcloud cloud storage block edit <id> --type HIGH_SPEED_GEN2   # retype
ovhcloud cloud storage block delete <id>                        # confirm first
```

> Volumes are **billed**; `delete` is destructive (data loss). Size can only be
> increased. Verify flags with `--help`
> (see [../references/safety.md](../references/safety.md)).
