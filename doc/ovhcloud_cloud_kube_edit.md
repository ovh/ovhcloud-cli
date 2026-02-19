## ovhcloud cloud kube edit

Edit the given Kubernetes cluster

```
ovhcloud cloud kube edit <cluster_id> [flags]
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

* [ovhcloud cloud kube](ovhcloud_cloud_kube.md)	 - Manage Kubernetes clusters in the given cloud project

