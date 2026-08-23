## ovhcloud cloud managed-kubernetes edit

Edit the given Kubernetes cluster

```
ovhcloud cloud managed-kubernetes edit <cluster_id> [flags]
```

### Options

```
      --editor                 Use a text editor to define edit parameters
  -h, --help                   help for edit
      --name string            Name of the Kubernetes cluster
      --update-policy string   Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)
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

