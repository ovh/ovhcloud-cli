# Product: Object Storage (S3-compatible)

`ovhcloud cloud storage object …` — S3-compatible storage containers and objects
in a Public Cloud project. Requires a project (`--cloud-project <id>` or
`OVH_CLOUD_PROJECT_SERVICE`).

Key sub-resources & verbs: container `list`/`get`/`create`/`edit`/`delete`,
`object` (list/get/edit/delete/copy/restore/version), `bulk-delete`,
`lifecycle`, `credentials`, `quota`, `replication-job`, `add-user`,
`generate-presigned-url`.

```bash
ovhcloud cloud storage object list -o json
ovhcloud cloud storage object get <container>
ovhcloud cloud storage object object list <container> -o json   # objects inside
ovhcloud cloud storage object create --help
ovhcloud cloud storage object generate-presigned-url <container> --help
ovhcloud cloud storage object delete <container>                # confirm first
```

> Storage is **billed**; `delete`/`bulk-delete` are destructive (data loss).
> Verify flags with `--help` (see [../references/safety.md](../references/safety.md)).
