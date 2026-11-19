## ovhcloud cloud managed-kubernetes restart

Restart control plane apiserver to invalidate cache without downtime

```
ovhcloud cloud managed-kubernetes restart <cluster_id> [flags]
```

### Options

```
      --force   Force restart the Kubernetes cluster (will create a slight downtime)
  -h, --help    help for restart
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

