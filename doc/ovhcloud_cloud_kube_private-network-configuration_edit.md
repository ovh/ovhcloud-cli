## ovhcloud cloud kube private-network-configuration edit

Edit the private network configuration for the given Kubernetes cluster

```
ovhcloud cloud kube private-network-configuration edit <cluster_id> [flags]
```

### Options

```
      --default-vrack-gateway string   If defined, all egress traffic will be routed towards this IP address, which should belong to the private network
      --editor                         Use a text editor to define parameters
  -h, --help                           help for edit
      --routing-as-default             Set private network routing as default
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud kube private-network-configuration](ovhcloud_cloud_kube_private-network-configuration.md)	 - Manage private network configuration for Kubernetes clusters

