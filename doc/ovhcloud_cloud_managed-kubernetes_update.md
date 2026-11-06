## ovhcloud cloud managed-kubernetes update

Update the given Kubernetes cluster

```
ovhcloud cloud managed-kubernetes update <cluster_id> [flags]
```

### Options

```
      --force             Force redeploying the control plane / reinstalling the nodes regardless of their current version
  -h, --help              help for update
      --strategy string   Update strategy to apply on your service (LATEST_PATCH, NEXT_MINOR)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-kubernetes](ovhcloud_cloud_managed-kubernetes.md)	 - Manage Kubernetes clusters in the given cloud project

