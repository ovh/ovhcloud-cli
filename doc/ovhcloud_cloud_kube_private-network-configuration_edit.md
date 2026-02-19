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
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --output json
                                 --output yaml
                                 --output interactive
                                 --output 'id' (to extract a single field)
                                 --output 'nested.field.subfield' (to extract a nested field)
                                 --output '[id, "name"]' (to extract multiple fields as an array)
                                 --output '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --output 'name+","+type' (to extract and concatenate fields in a string)
                                 --output '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
```

### SEE ALSO

* [ovhcloud cloud kube private-network-configuration](ovhcloud_cloud_kube_private-network-configuration.md)	 - Manage private network configuration for Kubernetes clusters

