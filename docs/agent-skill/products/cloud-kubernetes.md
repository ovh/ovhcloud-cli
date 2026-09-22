# Product: Managed Kubernetes (MKS)

`ovhcloud cloud managed-kubernetes …` — managed Kubernetes clusters in a Public
Cloud project. Requires a project (`--cloud-project <id>` or
`OVH_CLOUD_PROJECT_SERVICE`).

Key sub-resources & verbs: `list`, `get <id>`, `create`, `edit <id>`,
`delete <id>`, `kubeconfig`, `nodepool`, `node`, `oidc`, `ip-restrictions`,
`private-network-configuration`, `reset`, `restart`.

```bash
ovhcloud cloud managed-kubernetes list -o json
ovhcloud cloud managed-kubernetes get <id>
ovhcloud cloud managed-kubernetes kubeconfig --help   # fetch kubeconfig
ovhcloud cloud managed-kubernetes nodepool list <clusterId> -o json
ovhcloud cloud managed-kubernetes create --help       # region, version, nodepool…
ovhcloud cloud managed-kubernetes delete <id>         # confirm first
```

> Clusters and node pools are **billed**; `delete`/`reset` are destructive.
> Verify flags with `--help` (see [../references/safety.md](../references/safety.md)).
