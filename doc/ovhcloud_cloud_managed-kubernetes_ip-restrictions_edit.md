## ovhcloud cloud managed-kubernetes ip-restrictions edit

Edit IP restrictions for the given Kubernetes cluster

```
ovhcloud cloud managed-kubernetes ip-restrictions edit <cluster_id> [flags]
```

### Options

```
      --editor        Use a text editor to define parameters
  -h, --help          help for edit
      --ips strings   List of IPs to restrict access to the Kubernetes cluster
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

* [ovhcloud cloud managed-kubernetes ip-restrictions](ovhcloud_cloud_managed-kubernetes_ip-restrictions.md)	 - Manage IP restrictions for Kubernetes clusters

